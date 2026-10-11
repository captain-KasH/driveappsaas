import { useCallback, useEffect, useState } from "react";
import { Alert, Linking, Pressable, View } from "react-native";
import { ApiError, api, config, km, mins, money, newKey, onSignedOut, session, signIn, signOut, store, when, type Me, type PublicTenant, type Breakdown } from "../api";
import { Badge, Button, Card, Chip, ErrorText, FareLines, Field, H1, H2, Header, Loading, Note, P, Row, Screen, ThemeProvider, useAction, usePoll } from "../ui";
import { currentPlace, PlaceSearch, type Place } from "../places";

const OPERATOR_KEY = "taxios_operator";

const STATUS_TEXT: Record<string, string> = {
  scheduled: "Scheduled — we'll start finding a driver 15 minutes before pickup.",
  searching: "Finding you a driver…",
  driver_assigned: "Driver assigned and getting ready.",
  en_route: "Your driver is on the way.",
  arrived: "Your driver has arrived.",
  in_progress: "On the way to your destination.",
  completed: "Trip complete. Thank you for riding!",
  cancelled: "This booking was cancelled.",
  expired: "Sorry — no driver accepted in time. Please try again.",
};

/** Passenger app: operator (pinned at build time or entered once) → join/sign in → book → track → receipt. */
export default function PassengerApp() {
  const [slug, setSlug] = useState<string | null>(config.tenant || null);
  const [operator, setOperator] = useState<PublicTenant | null>(null);
  const [opError, setOpError] = useState<unknown>(null);
  const [member, setMember] = useState<"unknown" | "yes" | "no">("unknown");

  useEffect(() => { if (!config.tenant) store.get(OPERATOR_KEY).then((v) => setSlug(v ?? "")); }, []);
  useEffect(() => onSignedOut(() => setMember("no")), []);
  useEffect(() => {
    if (!slug) return;
    setOperator(null); setOpError(null);
    api<PublicTenant>("GET", `/v1/public/tenants/${encodeURIComponent(slug)}`).then(setOperator).catch(setOpError);
  }, [slug]);

  const checkMember = useCallback(async (op: PublicTenant) => {
    if (!(await session.get())) return setMember("no");
    try {
      const me = await api<Me>("GET", "/v1/me");
      setMember(me.memberships.some((m) => m.tenantId === op.tenant.id && m.role === "passenger") ? "yes" : "no");
    } catch (e) {
      // A server/network error must not look like "signed out"; keep the session and let Home retry.
      setMember(e instanceof ApiError && e.status === 401 ? "no" : "yes");
    }
  }, []);
  useEffect(() => { if (operator) checkMember(operator); }, [operator, checkMember]);

  const forgetOperator = async () => { await store.del(OPERATOR_KEY); setOperator(null); setSlug(""); setMember("unknown"); };

  if (slug === null) return <Screen><Loading /></Screen>;
  if (slug === "" ) return <ChooseOperator onChosen={async (s) => { await store.set(OPERATOR_KEY, s); setSlug(s); }} />;
  if (opError) {
    const gone = opError instanceof ApiError && opError.status === 404;
    return (
      <Screen>
        <H1>Taxios</H1>
        <Note tone="bad">{gone ? "This operator isn't taking bookings on Taxios right now." : String((opError as Error).message)}</Note>
        {!config.tenant && <Button kind="plain" title="Use a different operator code" onPress={forgetOperator} />}
        {!gone && <Button title="Try again" onPress={() => setSlug(slug + "")} />}
      </Screen>
    );
  }
  if (!operator || member === "unknown") return <Screen><Loading /></Screen>;

  const theme = { color: operator.brand.color, onColor: operator.brand.onColor, name: operator.tenant.name, logoUrl: operator.brand.logoUrl };
  return (
    <ThemeProvider value={theme}>
      {member === "no"
        ? <Join slug={slug} op={operator} onJoined={() => setMember("yes")} onSwitch={config.tenant ? undefined : forgetOperator} />
        : <Home op={operator} onSignOut={async () => { await signOut(); setMember("no"); }} onSwitch={config.tenant ? undefined : forgetOperator} />}
    </ThemeProvider>
  );
}

function ChooseOperator({ onChosen }: { onChosen: (slug: string) => void }) {
  const [code, setCode] = useState("");
  const act = useAction();
  return (
    <Screen>
      <H1>Book a ride</H1>
      <P muted>Enter the code your taxi operator gave you. It's the last part of their booking link, e.g. taxios.packline.lk/book/?b=<P style={{ fontWeight: "700" }}>colombo-cabs</P>.</P>
      <Field label="Operator code" value={code} onChangeText={(t) => setCode(t.toLowerCase().replace(/[^a-z0-9-]/g, ""))} autoCapitalize="none" autoCorrect={false} placeholder="colombo-cabs" />
      <ErrorText error={act.error} />
      <Button title="Continue" busy={act.busy} disabled={code.length < 3} onPress={() => act.run(async () => {
        await api("GET", `/v1/public/tenants/${code}`).catch((e) => { throw e instanceof ApiError && e.status === 404 ? new Error("No operator with that code is taking bookings.") : e; });
        onChosen(code);
      })} />
    </Screen>
  );
}

function Join({ slug, op, onJoined, onSwitch }: { slug: string; op: PublicTenant; onJoined: () => void; onSwitch?: () => void }) {
  const [mode, setMode] = useState<"register" | "signin">("register");
  const [f, setF] = useState({ fullName: "", phone: "", email: "", password: "" });
  const act = useAction();
  const submit = () => act.run(async () => {
    if (mode === "signin") {
      await signIn(f.email, f.password);
      const me = await api<Me>("GET", "/v1/me");
      if (me.memberships.some((m) => m.tenantId === op.tenant.id && m.role === "passenger")) return onJoined();
      setMode("register");
      throw new Error(`You're signed in. Add your name and phone once to ride with ${op.tenant.name}.`);
    }
    const r = await api<{ token: string }>("POST", `/v1/public/tenants/${encodeURIComponent(slug)}/passengers`, { ...f, client: "mobile" });
    await session.set(r.token);
    onJoined();
  });
  return (
    <Screen header={<Header />}>
      <H1>{mode === "register" ? `Ride with ${op.tenant.name}` : "Welcome back"}</H1>
      {op.brand.tagline && <P muted>{op.brand.tagline}</P>}
      <Row>
        <Chip label="New here" on={mode === "register"} onPress={() => setMode("register")} />
        <Chip label="I have an account" on={mode === "signin"} onPress={() => setMode("signin")} />
      </Row>
      {mode === "register" && <>
        <Field label="Full name" value={f.fullName} onChangeText={(v) => setF({ ...f, fullName: v })} autoComplete="name" textContentType="name" />
        <Field label="Mobile number" value={f.phone} onChangeText={(v) => setF({ ...f, phone: v })} keyboardType="phone-pad" placeholder="077 123 4567" autoComplete="tel" textContentType="telephoneNumber" />
      </>}
      <Field label="Email" value={f.email} onChangeText={(v) => setF({ ...f, email: v })} keyboardType="email-address" autoCapitalize="none" autoComplete="email" textContentType="emailAddress" />
      <Field label={mode === "register" ? "Password (10+ characters)" : "Password"} value={f.password} onChangeText={(v) => setF({ ...f, password: v })} secureTextEntry
        textContentType={mode === "register" ? "newPassword" : "password"} />
      <ErrorText error={act.error} />
      <Button title={mode === "register" ? "Create account" : "Sign in"} busy={act.busy} onPress={submit} />
      {mode === "register" && <P muted small>Already ride with another Taxios operator? Use the same email and password — we'll link it.</P>}
      <Legal op={op} />
      {onSwitch && <Button kind="link" title="Not your operator? Change code" onPress={onSwitch} />}
    </Screen>
  );
}

function Legal({ op }: { op: PublicTenant }) {
  const b = op.brand;
  const links: [string, string][] = [];
  if (b.supportPhone) links.push([`Call ${b.supportPhone}`, `tel:${b.supportPhone}`]);
  if (b.supportEmail) links.push(["Email support", `mailto:${b.supportEmail}`]);
  if (b.termsUrl) links.push(["Terms", b.termsUrl]);
  if (b.privacyUrl) links.push(["Privacy", b.privacyUrl]);
  if (!links.length) return null;
  return <Row>{links.map(([label, url]) => <Button key={url} kind="link" title={label} onPress={() => Linking.openURL(url)} />)}</Row>;
}

type Booking = { id: string; reference: string; status: string; createdAt: string; dropoffAddress: string };

function Home({ op, onSignOut, onSwitch }: { op: PublicTenant; onSignOut: () => void; onSwitch?: () => void }) {
  const base = `/v1/tenants/${op.tenant.id}`;
  const [active, setActive] = useState<string | null>(null);
  const [history, setHistory] = useState<Booking[] | null>(null);
  const load = useCallback(() => api<{ bookings: Booking[] }>("GET", `${base}/my/bookings`).then((r) => {
    setHistory(r.bookings);
    const live = r.bookings.find((b) => ["scheduled", "searching", "assigned"].includes(b.status));
    if (live) setActive((cur) => cur ?? live.id);
  }).catch(() => {}), [base]);
  useEffect(() => { load(); }, [load]);

  const header = <Header right={<Button kind="link" title="Sign out" onPress={() => Alert.alert("Sign out?", undefined, [{ text: "Cancel", style: "cancel" }, { text: "Sign out", style: "destructive", onPress: onSignOut }])} />} />;
  if (active) return <Screen header={header}><Active base={base} id={active} onClose={() => { setActive(null); load(); }} /></Screen>;
  return (
    <Screen header={header}>
      <Book op={op} base={base} onBooked={(id) => { setActive(id); load(); }} />
      {history && history.length > 0 && (
        <Card>
          <H2>Your rides</H2>
          {history.slice(0, 10).map((b) => (
            <Pressable key={b.id} accessibilityRole="button" onPress={() => setActive(b.id)} style={{ paddingVertical: 8 }}>
              <Row style={{ justifyContent: "space-between" }}><P small>{when(b.createdAt)} · {b.dropoffAddress.split(",")[0]}</P><Badge value={b.status} /></Row>
            </Pressable>
          ))}
        </Card>
      )}
      <Legal op={op} />
      {onSwitch && <Button kind="link" title="Switch operator" onPress={onSwitch} />}
    </Screen>
  );
}

const WHEN: [string, number][] = [["Now", 0], ["In 30 min", 30], ["In 1 hour", 60], ["In 2 hours", 120], ["Tomorrow", 24 * 60]];

function Book({ op, base, onBooked }: { op: PublicTenant; base: string; onBooked: (id: string) => void }) {
  const [pickup, setPickup] = useState<Place | null>(null);
  const [dropoff, setDropoff] = useState<Place | null>(null);
  const [categoryId, setCategoryId] = useState(op.categories[0]?.id ?? "");
  const [later, setLater] = useState(0);
  const [note, setNote] = useState("");
  const [quote, setQuote] = useState<any>(null);
  const [key, setKey] = useState(newKey);
  const act = useAction();
  const reset = () => setQuote(null);

  return (
    <Card>
      <H1>Where to?</H1>
      <PlaceSearch label="Pickup" value={pickup} onChange={(p) => { setPickup(p); reset(); }} placeholder="Search a place or address" />
      <Button kind="link" title="Use my current location" onPress={() => act.run(async () => { setPickup(await currentPlace()); reset(); })} />
      <PlaceSearch label="Drop-off" value={dropoff} onChange={(p) => { setDropoff(p); reset(); }} placeholder="Where are you going?" />
      <P muted small>Vehicle</P>
      <Row>{op.categories.map((c) => <Chip key={c.id} label={`${c.name} · ${c.seats}`} on={categoryId === c.id} onPress={() => { setCategoryId(c.id); reset(); }} />)}</Row>
      <P muted small>When</P>
      <Row>{WHEN.map(([label, m]) => <Chip key={label} label={label} on={later === m} onPress={() => { setLater(m); reset(); }} />)}</Row>
      {!quote ? (
        <Button title="See fare" busy={act.busy} disabled={!pickup || !dropoff || !categoryId} onPress={() => act.run(async () => {
          const r = await api<{ quote: any }>("POST", `${base}/quotes`, { pickup, dropoff, categoryId, ...(later && { scheduledFor: new Date(Date.now() + later * 60_000).toISOString() }) });
          setQuote(r.quote); setKey(newKey());
        })} />
      ) : (
        <View style={{ gap: 10 }}>
          <FareLines b={quote.estimate as Breakdown} title={`Estimated fare · about ${km(quote.estimatedDistanceM)}, ${mins(quote.estimatedDurationS)}`} />
          <P muted small>This is an estimate. Your final fare is calculated from the actual trip using the same rates. Pay the driver in cash.</P>
          <Field label="Note for the driver (optional)" value={note} onChangeText={setNote} maxLength={500} />
          <Button title={`Book · ${money(quote.estimate.totalMinor, quote.estimate.currency)} est.`} busy={act.busy} onPress={() => act.run(async () => {
            const r = await api<{ booking: { id: string } }>("POST", `${base}/my/bookings`, { quoteId: quote.id, paymentMethod: "cash", note }, { "Idempotency-Key": key });
            setQuote(null); setPickup(null); setDropoff(null); setNote(""); setLater(0);
            onBooked(r.booking.id);
          })} />
        </View>
      )}
      <ErrorText error={act.error} />
    </Card>
  );
}

function Active({ base, id, onClose }: { base: string; id: string; onClose: () => void }) {
  const [b, setB] = useState<any>(null);
  const [receipt, setReceipt] = useState<any>(null);
  const act = useAction();
  const load = useCallback(() => api<{ booking: any }>("GET", `${base}/my/bookings/${id}`).then((r) => setB(r.booking)).catch(act.setError), [base, id, act.setError]);
  // Poll the booking until it's done, then the receipt until the driver has confirmed the cash.
  const loadReceipt = useCallback(() => api<{ receipt: any }>("GET", `${base}/my/bookings/${id}/receipt`).then((r) => setReceipt(r.receipt)).catch(() => {}), [base, id]);
  usePoll(() => {
    if (b?.status !== "completed") load();
    else if (receipt?.paymentStatus !== "paid_cash") loadReceipt();
  }, 4000);
  if (!b) return <Loading />;
  const state = b.status === "assigned" ? b.trip?.status : b.status;
  const cancellable = ["scheduled", "searching"].includes(b.status) || (b.status === "assigned" && b.trip?.status !== "in_progress");
  return (
    <>
      <Card>
        <Row style={{ justifyContent: "space-between" }}><H2>{b.reference}</H2><Badge value={state} /></Row>
        <P>{STATUS_TEXT[state] ?? state}</P>
        {b.driver && b.status === "assigned" && (
          <Card style={{ backgroundColor: "#eaf1f0" }}>
            <P style={{ fontWeight: "700" }}>{b.driver.firstName} · {b.driver.vehiclePlate}</P>
            <P small>{b.driver.vehicleDescription}</P>
            <P muted small>{b.driver.locationFresh ? "Driver location is live" : "Driver location not available right now"}</P>
          </Card>
        )}
        <P small><P small style={{ fontWeight: "700" }}>Pickup </P>{b.pickup.address}</P>
        <P small><P small style={{ fontWeight: "700" }}>Drop-off </P>{b.dropoff.address}</P>
        {b.scheduledFor && <P small><P small style={{ fontWeight: "700" }}>Pickup time </P>{when(b.scheduledFor)}</P>}
        {receipt ? <Receipt r={receipt} base={base} id={id} /> : <FareLines b={b.estimate} title="Estimated fare" />}
        <ErrorText error={act.error} />
        {cancellable && <Button kind="danger" title="Cancel booking" busy={act.busy} onPress={() => Alert.alert("Cancel this booking?", undefined, [
          { text: "Keep it", style: "cancel" },
          { text: "Cancel booking", style: "destructive", onPress: () => act.run(async () => { await api("POST", `${base}/my/bookings/${id}/cancel`, { reason: "cancelled by passenger" }); await load(); }) },
        ])} />}
        <Button kind="plain" title={cancellable ? "Back" : "Book another ride"} onPress={onClose} />
      </Card>
    </>
  );
}

function Receipt({ r, base, id }: { r: any; base: string; id: string }) {
  const [stars, setStars] = useState(0);
  const [comment, setComment] = useState("");
  const [rated, setRated] = useState(false);
  const act = useAction();
  return (
    <View style={{ gap: 8 }}>
      <H2>Receipt · {r.business}</H2>
      <P muted small>{km(r.distanceM)} · {mins(r.durationS)} · {r.distanceSource === "gps_track" ? "distance measured by GPS" : "estimated distance"}</P>
      <FareLines b={r.fare} />
      <P small>Payment: {r.paymentStatus === "paid_cash" ? `paid in cash (${money(r.paidMinor)})` : "cash, not yet confirmed by the driver"}</P>
      {rated ? <Note tone="ok">Thanks for your feedback.</Note> : (
        <>
          <Row>{[1, 2, 3, 4, 5].map((n) => <Chip key={n} label={`★ ${n}`} on={stars >= n} onPress={() => setStars(n)} />)}</Row>
          {stars > 0 && <>
            <Field label="Comment (optional)" value={comment} onChangeText={setComment} maxLength={500} />
            <Button title="Rate trip" busy={act.busy} onPress={() => act.run(async () => { await api("POST", `${base}/my/bookings/${id}/rating`, { stars, comment }); setRated(true); })} />
          </>}
          <ErrorText error={act.error} />
        </>
      )}
    </View>
  );
}
