package identity

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/comms"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

// PublicURL is where emailed links point (the web app origin).
func PublicURL() string {
	if u := os.Getenv("TAXIOS_PUBLIC_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://taxios.packline.lk"
}

// RegisterAccount adds self-service account routes: sign-up, email verification, password reset.
func RegisterAccount(mux *http.ServeMux, pool *pgxpool.Pool, mailer comms.Mailer) {
	a := &accounts{pool: pool, mailer: mailer}
	mux.Handle("POST /v1/auth/register", httpx.Handler(a.register))
	mux.Handle("POST /v1/auth/verify-email", httpx.Handler(a.verifyEmail))
	mux.Handle("POST /v1/auth/resend-verification", RequireAuth(pool, a.resendVerification))
	mux.Handle("POST /v1/auth/forgot-password", httpx.Handler(a.forgotPassword))
	mux.Handle("POST /v1/auth/reset-password", httpx.Handler(a.resetPassword))
}

type accounts struct {
	pool   *pgxpool.Pool
	mailer comms.Mailer
}

func validEmail(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) > 254 {
		return "", false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", false
	}
	return strings.ToLower(s), true
}

func (a *accounts) register(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		FullName string `json:"fullName"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	email, ok := validEmail(in.Email)
	if !ok {
		return httpx.Errorf(http.StatusUnprocessableEntity, "invalid_email", "a valid email is required")
	}
	name := strings.TrimSpace(in.FullName)
	if name == "" || len(name) > 200 {
		return httpx.Errorf(http.StatusUnprocessableEntity, "name_required", "full name is required")
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return err
	}
	out := map[string]any{"emailVerified": false}
	err = db.InTx(r.Context(), a.pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		var userID string
		err := tx.QueryRow(r.Context(), `INSERT INTO users(email, full_name, password_hash) VALUES ($1,$2,$3)
			ON CONFLICT (email) DO NOTHING RETURNING id`, email, name, hash).Scan(&userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Errorf(http.StatusConflict, "email_taken", "an account with this email already exists; sign in or reset your password")
		}
		if err != nil {
			return err
		}
		if err := a.sendVerification(r, tx, userID, email, out); err != nil {
			return err
		}
		if err := audit.Record(r.Context(), tx, audit.Event{ActorUserID: userID, Action: "user.registered", TargetType: "user", TargetID: userID}); err != nil {
			return err
		}
		token, err := CreateSession(r.Context(), tx, userID, r.UserAgent(), RemoteIP(r))
		out["token"], out["userId"] = token, userID
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusCreated, out)
}

func (a *accounts) sendVerification(r *http.Request, tx pgx.Tx, userID, email string, out map[string]any) error {
	token, err := IssueToken(r.Context(), tx, userID, PurposeVerifyEmail)
	if err != nil {
		return err
	}
	if comms.DevTokens() {
		out["verifyToken"] = token
	}
	return a.mailer.Send(r.Context(), comms.Email{To: email, Subject: "Verify your Taxios email",
		Text: fmt.Sprintf("Confirm your email to continue setting up your transport business:\n\n%s/verify/?token=%s\n\nThe link works once and expires in 24 hours.", PublicURL(), token)})
}

func (a *accounts) verifyEmail(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Token string `json:"token"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	err := db.InTx(r.Context(), a.pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		userID, err := ConsumeToken(r.Context(), tx, in.Token, PurposeVerifyEmail)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE users SET email_verified_at = coalesce(email_verified_at, now()), updated_at = now() WHERE id = $1`, userID); err != nil {
			return err
		}
		return audit.Record(r.Context(), tx, audit.Event{ActorUserID: userID, Action: "user.email_verified", TargetType: "user", TargetID: userID})
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"emailVerified": true})
}

func (a *accounts) resendVerification(w http.ResponseWriter, r *http.Request) error {
	p, _ := FromContext(r.Context())
	out := map[string]any{"sent": true}
	err := db.InTx(r.Context(), a.pool, db.Scope{UserID: p.UserID}, func(tx pgx.Tx) error {
		var verified bool
		if err := tx.QueryRow(r.Context(), `SELECT email_verified_at IS NOT NULL FROM users WHERE id = $1`, p.UserID).Scan(&verified); err != nil {
			return err
		}
		if verified {
			return httpx.Errorf(http.StatusConflict, "already_verified", "this email is already verified")
		}
		return a.sendVerification(r, tx, p.UserID, p.Email, out)
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, out)
}

// forgotPassword always answers 200 so it cannot be used to discover accounts.
func (a *accounts) forgotPassword(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email string `json:"email"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	out := map[string]any{"sent": true}
	email, ok := validEmail(in.Email)
	if ok {
		err := db.InTx(r.Context(), a.pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
			var userID string
			err := tx.QueryRow(r.Context(), `SELECT id FROM users WHERE email = $1 AND disabled_at IS NULL`, email).Scan(&userID)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			token, err := IssueToken(r.Context(), tx, userID, PurposeResetPassword)
			if err != nil {
				return err
			}
			if comms.DevTokens() {
				out["resetToken"] = token
			}
			return a.mailer.Send(r.Context(), comms.Email{To: email, Subject: "Reset your Taxios password",
				Text: fmt.Sprintf("Choose a new password here:\n\n%s/reset/?token=%s\n\nThe link works once and expires in 1 hour. If you didn't ask for this, ignore it.", PublicURL(), token)})
		})
		if err != nil {
			return err
		}
	}
	return httpx.WriteJSON(w, http.StatusOK, out)
}

// resetPassword sets a new password, counts as proof of email ownership, and signs out every
// existing session (the old password may have been stolen). A fresh session is returned.
func (a *accounts) resetPassword(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return err
	}
	var out map[string]any
	err = db.InTx(r.Context(), a.pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		userID, err := ConsumeToken(r.Context(), tx, in.Token, PurposeResetPassword)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE users SET password_hash = $2, email_verified_at = coalesce(email_verified_at, now()), updated_at = now() WHERE id = $1`, userID, hash); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
			return err
		}
		if err := audit.Record(r.Context(), tx, audit.Event{ActorUserID: userID, Action: "user.password_reset", TargetType: "user", TargetID: userID}); err != nil {
			return err
		}
		token, err := CreateSession(r.Context(), tx, userID, r.UserAgent(), RemoteIP(r))
		out = map[string]any{"token": token, "userId": userID}
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, out)
}
