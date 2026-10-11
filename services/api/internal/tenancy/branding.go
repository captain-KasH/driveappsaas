package tenancy

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

// Brand is the public, runtime-configurable identity of a tenant (prompt §18). Consumed by the
// booking page and the mobile apps; never contains private contact details of staff.
type Brand struct {
	Name         string  `json:"name"`
	Color        string  `json:"color"`
	OnColor      string  `json:"onColor"` // black or white, whichever reads better on Color
	LogoURL      *string `json:"logoUrl"`
	Tagline      *string `json:"tagline"`
	SupportPhone *string `json:"supportPhone"`
	SupportEmail *string `json:"supportEmail"`
	WebsiteURL   *string `json:"websiteUrl"`
	TermsURL     *string `json:"termsUrl"`
	PrivacyURL   *string `json:"privacyUrl"`
	Language     string  `json:"language"`
	Currency     string  `json:"currency"`
}

const brandCols = `name, brand_color, logo_file_id, tagline, support_phone, support_email, website_url, terms_url, privacy_url, language, currency`

func scanBrand(row pgx.Row, slug string) (Brand, error) {
	var b Brand
	var logo *string
	err := row.Scan(&b.Name, &b.Color, &logo, &b.Tagline, &b.SupportPhone, &b.SupportEmail, &b.WebsiteURL, &b.TermsURL, &b.PrivacyURL, &b.Language, &b.Currency)
	if err != nil {
		return b, err
	}
	b.OnColor = onColor(b.Color)
	if logo != nil {
		u := "/v1/public/tenants/" + slug + "/logo"
		b.LogoURL = &u
	}
	return b, nil
}

// onColor picks black or white text for a hex background using WCAG relative luminance.
func onColor(hex string) string {
	lin := func(i int) float64 {
		v, _ := strconv.ParseUint(hex[i:i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	l := 0.2126*lin(1) + 0.7152*lin(3) + 0.0722*lin(5)
	// Contrast with white = 1.05/(l+0.05); with black = (l+0.05)/0.05. White wins below l≈0.179.
	if (1.05 / (l + 0.05)) >= ((l + 0.05) / 0.05) {
		return "#ffffff"
	}
	return "#000000"
}

// PublicBrand loads the brand of an active tenant by slug (platform scope for the single lookup).
func PublicBrand(ctx context.Context, pool *pgxpool.Pool, slug string) (id string, b Brand, err error) {
	slug = strings.ToLower(slug)
	err = db.InTx(ctx, pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		var status string
		var logo *string
		err := tx.QueryRow(ctx, `SELECT id, status, `+brandCols+` FROM tenants WHERE slug = $1`, slug).Scan(
			&id, &status, &b.Name, &b.Color, &logo, &b.Tagline, &b.SupportPhone, &b.SupportEmail, &b.WebsiteURL, &b.TermsURL, &b.PrivacyURL, &b.Language, &b.Currency)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && status != "active") {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		b.OnColor = onColor(b.Color)
		if logo != nil {
			u := "/v1/public/tenants/" + slug + "/logo"
			b.LogoURL = &u
		}
		return nil
	})
	return id, b, err
}

var hexRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *Service) getBranding(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var slug string
	if err := tx.QueryRow(r.Context(), `SELECT slug FROM tenants WHERE id=$1`, m.TenantID).Scan(&slug); err != nil {
		return Result{}, err
	}
	b, err := scanBrand(tx.QueryRow(r.Context(), `SELECT `+brandCols+` FROM tenants WHERE id=$1`, m.TenantID), slug)
	var logoID *string
	tx.QueryRow(r.Context(), `SELECT logo_file_id FROM tenants WHERE id=$1`, m.TenantID).Scan(&logoID)
	return Result{http.StatusOK, map[string]any{"brand": b, "logoFileId": logoID}}, err
}

// updateBranding: owner/admin set colour, logo (a file uploaded to this tenant), tagline and public links.
func (s *Service) updateBranding(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Color        *string `json:"color"`
		LogoFileID   *string `json:"logoFileId"` // "" removes the logo
		Tagline      *string `json:"tagline"`
		SupportPhone *string `json:"supportPhone"`
		SupportEmail *string `json:"supportEmail"`
		WebsiteURL   *string `json:"websiteUrl"`
		TermsURL     *string `json:"termsUrl"`
		PrivacyURL   *string `json:"privacyUrl"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	bad := func(field, msg string) (Result, error) {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_"+field, "%s", msg)
	}
	sets, args, changed := []string{"updated_at=now()"}, []any{m.TenantID}, map[string]any{}
	add := func(col, field string, v any) {
		args = append(args, v)
		sets = append(sets, col+"=$"+strconv.Itoa(len(args)))
		changed[field] = v
	}
	if in.Color != nil {
		c := strings.ToLower(strings.TrimSpace(*in.Color))
		if !hexRe.MatchString(c) {
			return bad("color", "colour must be a hex value like #0f766e")
		}
		add("brand_color", "color", c)
	}
	if in.LogoFileID != nil {
		if *in.LogoFileID == "" {
			add("logo_file_id", "logoFileId", nil)
		} else {
			var ok bool
			if !uuidRe.MatchString(*in.LogoFileID) {
				return bad("logo", "upload the logo first, then attach its file id")
			}
			// RLS already limits the lookup to this tenant's files; also require an image.
			if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM files WHERE id=$1 AND content_type LIKE 'image/%')`, *in.LogoFileID).Scan(&ok); err != nil {
				return Result{}, err
			}
			if !ok {
				return bad("logo", "the logo must be a PNG, JPEG or WebP uploaded to this business")
			}
			add("logo_file_id", "logoFileId", *in.LogoFileID)
		}
	}
	opt := func(col, field string, v *string, max int, check func(string) bool) bool {
		if v == nil {
			return true
		}
		t := strings.TrimSpace(*v)
		if t != "" && (len(t) > max || (check != nil && !check(t))) {
			return false
		}
		if t == "" {
			add(col, field, nil)
		} else {
			add(col, field, t)
		}
		return true
	}
	isURL := func(s string) bool { return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") }
	isEmail := func(s string) bool { a, err := mail.ParseAddress(s); return err == nil && a.Address == s }
	if !opt("tagline", "tagline", in.Tagline, 160, nil) {
		return bad("tagline", "tagline must be at most 160 characters")
	}
	if !opt("support_phone", "supportPhone", in.SupportPhone, 20, phoneRe.MatchString) {
		return bad("support_phone", "support phone must be 7-20 digits")
	}
	if !opt("support_email", "supportEmail", in.SupportEmail, 254, isEmail) {
		return bad("support_email", "support email must be a valid address")
	}
	if !opt("website_url", "websiteUrl", in.WebsiteURL, 300, isURL) {
		return bad("website_url", "website must start with https://")
	}
	if !opt("terms_url", "termsUrl", in.TermsURL, 300, isURL) {
		return bad("terms_url", "terms link must start with https://")
	}
	if !opt("privacy_url", "privacyUrl", in.PrivacyURL, 300, isURL) {
		return bad("privacy_url", "privacy link must start with https://")
	}
	if len(changed) == 0 {
		return bad("input", "nothing to update")
	}
	if _, err := tx.Exec(r.Context(), `UPDATE tenants SET `+strings.Join(sets, ", ")+` WHERE id=$1`, args...); err != nil {
		if pgCode(err) == "23514" {
			return bad("input", "a field is out of range")
		}
		return Result{}, err
	}
	if err := audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID, Action: "branding.updated",
		TargetType: "tenant", TargetID: m.TenantID, Metadata: changed}); err != nil {
		return Result{}, err
	}
	return s.getBranding(r, tx, m)
}
