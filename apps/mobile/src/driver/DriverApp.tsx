import { useCallback, useEffect, useRef, useState } from "react";
import { Alert, Linking, Platform, View } from "react-native";
import * as Location from "expo-location";
import * as ImagePicker from "expo-image-picker";
import { ago, api, config, km, money, onSignedOut, session, signIn, signOut, upload, type Me, type Membership, type PublicTenant } from "../api";
import { Badge, Button, Card, Chip, ErrorText, FareLines, Field, H1, H2, Header, Loading, Note, P, Row, Screen, ThemeProvider, useAction, usePoll } from "../ui";

const ISSUE_TEXT: Record<string, string> = {
  driver_not_approved: "your operator hasn't approved you yet",
  no_vehicle: "no vehicle assigned",
  vehicle_not_approved: "your vehicle isn't approved",
  documents_missing_or_expired: "documents missing or expired",
};
const NEXT: Record<string, { action: string; label: string }[]> = {
  driver_assigned: [{ action: "en-route", label: "I'm on my way" }, { action: "arrived", label: "I've arrived" }],
  en_route: [{ action: "arrived", label: "I've arrived" }],
  arrived: [{ action: "start", label: "Passenger on board — start trip" }],
  in_progress: [{ action: "complete", label: "Complete trip" }],
};
const DOC_TYPES: [string, string][] = [["driving_licence", "Driving licence"], ["national_id", "NIC"], ["vehicle_insurance", "Insurance"], ["revenue_licence", "Revenue licence"], ["police_report", "Police report"]];
const ABORT_REASONS = ["Passenger didn't show up", "Passenger cancelled with me", "Vehicle problem", "Safety concern"];

type DriverMe = {
  driver: { id: string; fullName: string; status: string; availability: string; complianceIssues: string[]; locationAt: string | null; locationFresh: boolean };
  documents: { id: string; docType: string; number: string; expiresOn: string | null; status: string; reviewNote: string | null; fileId: string | null }[];
  vehicle: { plate: string; make: string; model: string; colour: string } | null;
};
type Offer = { id: string; expiresAt: string; distanceToPickupM: number | null; pickupAddress: string; dropoffAddress: string; estimatedTotalMinor: number; currency: string; customerFirstName: string; scheduledFor: string | null };

/** Driver app: sign in (invited by the operator) → choose business → go online → offers → trip steps → cash. */
export default function DriverApp() {
  const [state, setState] = useState<"loading" | "signedOut" | "ready">("loading");
  const [memberships, setMemberships] = useState<Membership[]>([]);
  const [tenant, setTenant] = useState<Membership | null>(null);

  const loadMe = useCallback(async () => {
    if (!(await session.get())) return setState("signedOut");
    try {
      const me = await api<Me>("GET", "/v1/me");
      const mine = me.memberships.filter((m) => m.role === "driver" && (!config.tenant || m.tenantSlug === config.tenant));
      setMemberships(mine);
      setTenant(mine.length === 1 ? mine[0] : null);
      setState("ready");
    } catch { setState("signedOut"); }
  }, []);
  useEffect(() => { loadMe(); }, [loadMe]);
  useEffect(() => onSignedOut(() => setState("signedOut")), []);

  const out = async () => { await signOut(); setTenant(null); setState("signedOut"); };
  if (state === "loading") return <Screen><Loading /></Screen>;
  if (state === "signedOut") return <SignIn onSignedIn={loadMe} />;
  if (memberships.length === 0) return (
    <Screen header={<Header right={<Button kind="link" title="Sign out" onPress={out} />} />}>
      <H1>No driver account yet</H1>
      <P>This login isn't a driver for {config.tenant ? "this operator" : "any operator"} on Taxios. Ask your operator to add you as a driver — they'll send you an invitation link to set your password.</P>
      <Button kind="plain" title="Check again" onPress={loadMe} />
    </Screen>
  );
  if (!tenant) return (
    <Screen header={<Header right={<Button kind="link" title="Sign out" onPress={out} />} />}>
      <H1>Drive for</H1>
      {memberships.map((m) => <Button key={m.tenantId} kind="plain" title={m.tenantName} onPress={() => setTenant(m)} />)}
    </Screen>
  );
  return <Branded m={tenant}><Driving m={tenant} onSignOut={out} onSwitch={memberships.length > 1 ? () => setTenant(null) : undefined} /></Branded>;
}

function Branded({ m, children }: { m: Membership; children: React.ReactNode }) {
  const [op, setOp] = useState<PublicTenant | null>(null);
  useEffect(() => { api<PublicTenant>("GET", `/v1/public/tenants/${m.tenantSlug}`).then(setOp).catch(() => {}); }, [m.tenantSlug]);
  const theme = { color: op?.brand.color ?? "#0f766e", onColor: op?.brand.onColor ?? "#ffffff", name: m.tenantName, logoUrl: op?.brand.logoUrl ?? null };
  return <ThemeProvider value={theme}>{children}</ThemeProvider>;
}

function SignIn({ onSignedIn }: { onSignedIn: () => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const act = useAction();
  return (
    <Screen header={<Header />}>
      <H1>Driver sign in</H1>
      <P muted>Use the email and password you set from your operator's invitation link.</P>
      <Field label="Email" value={email} onChangeText={setEmail} keyboardType="email-address" autoCapitalize="none" autoComplete="email" textContentType="username" />
      <Field label="Password" value={password} onChangeText={setPassword} secureTextEntry textContentType="password" />
      <ErrorText error={act.error} />
      <Button title="Sign in" busy={act.busy} onPress={() => act.run(async () => { await signIn(email, password); onSignedIn(); })} />
      <Button kind="link" title="Forgot password?" onPress={() => Linking.openURL(`${config.apiUrl}/forgot/`)} />
    </Screen>
  );
}

function Driving({ m, onSignOut, onSwitch }: { m: Membership; onSignOut: () => void; onSwitch?: () => void }) {
  const base = `/v1/tenants/${m.tenantId}`;
  const [me, setMe] = useState<DriverMe | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [trip, setTrip] = useState<any>(null);
  const [finished, setFinished] = useState<any>(null);
  const [gps, setGps] = useState<"off" | "waiting" | "ok" | "denied">("off");
  const act = useAction();

  const refresh = useCallback(async () => {
    try {
      const [d, w] = await Promise.all([api<DriverMe>("GET", `${base}/driver/me`), api<{ offer: Offer | null; activeTrip: any }>("GET", `${base}/driver/work`)]);
      setMe(d); setOffer(w.offer); setTrip(w.activeTrip);
    } catch (e) { act.setError(e); }
  }, [base, act.setError]);
  usePoll(refresh, 4000);

  // Foreground location while online or on a trip, at most one update per 10 s.
  // ponytail: background tracking (expo-task-manager + "always" permission) needs a dev build — T-152b.
  const sharing = !!me && (me.driver.availability !== "offline" || !!trip);
  const last = useRef(0);
  useEffect(() => {
    if (!sharing) { setGps("off"); return; }
    let sub: Location.LocationSubscription | null = null;
    let cancelled = false;
    (async () => {
      setGps("waiting");
      const perm = await Location.requestForegroundPermissionsAsync();
      if (perm.status !== "granted") { setGps("denied"); return; }
      sub = await Location.watchPositionAsync({ accuracy: Location.Accuracy.High, timeInterval: 5000, distanceInterval: 10 }, (p) => {
        setGps("ok");
        if (Date.now() - last.current < 10_000) return;
        last.current = Date.now();
        api("POST", `${base}/driver/location`, { lat: p.coords.latitude, lng: p.coords.longitude, recordedAt: new Date(p.timestamp).toISOString() }).catch(() => {});
      });
      if (cancelled) sub.remove();
    })();
    return () => { cancelled = true; sub?.remove(); };
  }, [sharing, base]);

  const header = <Header right={<Row>{onSwitch && <Button kind="link" title="Switch" onPress={onSwitch} />}<Button kind="link" title="Sign out" onPress={onSignOut} /></Row>} />;
  if (!me) return <Screen header={header}>{act.error ? <ErrorText error={act.error} /> : <Loading />}</Screen>;
  const d = me.driver;
  const online = d.availability !== "offline";
  const go = (fn: () => Promise<unknown>) => act.run(async () => { await fn(); await refresh(); });

  return (
    <Screen header={header}>
      <Card>
        <Row style={{ justifyContent: "space-between" }}>
          <H1>{d.fullName.split(" ")[0]}</H1>
          <Row><Badge value={d.availability} />{me.vehicle && <P small muted>{me.vehicle.plate}</P>}</Row>
        </Row>
        {d.complianceIssues.length > 0 ? (
          <Note>You can't go online yet: {d.complianceIssues.map((i) => ISSUE_TEXT[i] ?? i).join(", ")}. Add your documents below; your operator reviews them.</Note>
        ) : !trip && (
          <Button kind={online ? "plain" : "primary"} title={online ? "Go offline" : "Go online"} busy={act.busy}
            onPress={() => go(() => api("POST", `${base}/driver/availability`, { available: !online }))} />
        )}
        {sharing && (
          <P small muted>
            {gps === "ok" ? "Sharing location" : gps === "waiting" ? "Waiting for GPS…" : gps === "denied" ? "" : ""} · last update {ago(d.locationAt)}
          </P>
        )}
        {gps === "denied" && <Note tone="bad">Location is off for this app, so you won't get nearby offers. Turn it on in Settings.</Note>}
        {gps === "denied" && <Button kind="plain" title="Open Settings" onPress={() => Linking.openSettings()} />}
        <ErrorText error={act.error} />
      </Card>

      {offer && !trip && <OfferCard offer={offer} busy={act.busy}
        onAccept={() => go(() => api("POST", `${base}/driver/offers/${offer.id}/accept`))}
        onReject={() => go(() => api("POST", `${base}/driver/offers/${offer.id}/reject`))} />}

      {trip && (
        <Card>
          <Row style={{ justifyContent: "space-between" }}><H2>Current trip</H2><Badge value={trip.trip?.status} /></Row>
          <P><P style={{ fontWeight: "700" }}>Passenger </P>{trip.customer?.name}</P>
          <P><P style={{ fontWeight: "700" }}>Pickup </P>{trip.pickup.address}</P>
          <Button kind="link" title="Navigate to pickup" onPress={() => navigate(trip.pickup.lat, trip.pickup.lng)} />
          <P><P style={{ fontWeight: "700" }}>Drop-off </P>{trip.dropoff.address}</P>
          <Button kind="link" title="Navigate to drop-off" onPress={() => navigate(trip.dropoff.lat, trip.dropoff.lng)} />
          {trip.note ? <Note>{trip.note}</Note> : null}
          <P small muted>Estimate {money(trip.estimate?.totalMinor, trip.estimate?.currency)} · cash</P>
          {(NEXT[trip.trip?.status] ?? []).map((n) => (
            <Button key={n.action} title={n.label} busy={act.busy} onPress={() => act.run(async () => {
              const r = await api<{ booking: any }>("POST", `${base}/driver/trips/${trip.id}/${n.action}`);
              if (n.action === "complete") setFinished(r.booking);
              await refresh();
            })} />
          ))}
          <Button kind="danger" title="End trip early" onPress={() => Alert.alert("Why are you ending this trip?", undefined, [
            ...ABORT_REASONS.map((reason) => ({ text: reason, onPress: () => go(() => api("POST", `${base}/driver/trips/${trip.id}/abort`, { reason })) })),
            { text: "Keep driving", style: "cancel" as const },
          ])} />
        </Card>
      )}

      {finished && <CashCard base={base} booking={finished} onDone={() => setFinished(null)} />}
      {!trip && !offer && online && d.complianceIssues.length === 0 && <Card><P muted>Waiting for ride offers. Keep the app open.</P></Card>}
      <Documents base={base} me={me} onChange={refresh} />
    </Screen>
  );
}

function navigate(lat: number, lng: number) {
  const google = `https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`;
  Linking.openURL(Platform.OS === "ios" ? `comgooglemaps://?daddr=${lat},${lng}` : `google.navigation:q=${lat},${lng}`).catch(() => Linking.openURL(google));
}

function OfferCard({ offer, busy, onAccept, onReject }: { offer: Offer; busy: boolean; onAccept: () => void; onReject: () => void }) {
  const [left, setLeft] = useState(0);
  useEffect(() => {
    const tick = () => setLeft(Math.max(0, Math.round((new Date(offer.expiresAt).getTime() - Date.now()) / 1000)));
    tick();
    const id = setInterval(tick, 500);
    return () => clearInterval(id);
  }, [offer.expiresAt]);
  return (
    <Card style={{ borderWidth: 2 }}>
      <Row style={{ justifyContent: "space-between" }}><H2>New ride offer</H2><Badge value={`${left}s`} /></Row>
      <P><P style={{ fontWeight: "700" }}>Pickup </P>{offer.pickupAddress} · {km(offer.distanceToPickupM)} away</P>
      <P><P style={{ fontWeight: "700" }}>Drop-off </P>{offer.dropoffAddress}</P>
      <P><P style={{ fontWeight: "700" }}>Passenger </P>{offer.customerFirstName}</P>
      <P><P style={{ fontWeight: "700" }}>Estimate </P>{money(offer.estimatedTotalMinor, offer.currency)} · cash</P>
      <Button title="Accept" busy={busy} disabled={left === 0} onPress={onAccept} />
      <Button kind="plain" title="Decline" disabled={busy || left === 0} onPress={onReject} />
    </Card>
  );
}

function CashCard({ base, booking, onDone }: { base: string; booking: any; onDone: () => void }) {
  const total: number = booking.trip?.finalTotalMinor ?? 0;
  const [amount, setAmount] = useState((total / 100).toFixed(2));
  const [note, setNote] = useState("");
  const act = useAction();
  const [r, c = ""] = amount.trim().split(".");
  const minor = Number(r || 0) * 100 + Number((c + "00").slice(0, 2));
  return (
    <Card>
      <H2>Collect {money(total, booking.trip?.currency)} in cash</H2>
      <FareLines b={booking.trip?.finalBreakdown} title="Final fare" />
      <Field label={`Amount received (${booking.trip?.currency ?? "LKR"})`} value={amount} onChangeText={setAmount} keyboardType="decimal-pad" />
      {minor !== total && <Field label="Why is it different?" value={note} onChangeText={setNote} />}
      <ErrorText error={act.error} />
      <Button title="Confirm cash received" busy={act.busy} onPress={() => act.run(async () => {
        await api("POST", `${base}/driver/trips/${booking.id}/cash`, { amountMinor: minor, note });
        onDone();
      })} />
    </Card>
  );
}

function Documents({ base, me, onChange }: { base: string; me: DriverMe; onChange: () => void }) {
  const [open, setOpen] = useState(me.driver.complianceIssues.length > 0);
  const [docType, setDocType] = useState("driving_licence");
  const [number, setNumber] = useState("");
  const [expires, setExpires] = useState("");
  const [file, setFile] = useState<{ id: string; fileName: string } | null>(null);
  const act = useAction();

  const pick = (camera: boolean) => act.run(async () => {
    const perm = camera ? await ImagePicker.requestCameraPermissionsAsync() : await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!perm.granted) throw new Error(camera ? "Camera access is off for this app." : "Photo access is off for this app.");
    const res = camera
      ? await ImagePicker.launchCameraAsync({ quality: 0.6 })
      : await ImagePicker.launchImageLibraryAsync({ mediaTypes: ["images"], quality: 0.6 });
    if (res.canceled || !res.assets[0]) return;
    const a = res.assets[0];
    setFile(await upload(base, { uri: a.uri, name: a.fileName ?? "document.jpg", type: a.mimeType ?? "image/jpeg" }));
  });

  return (
    <Card>
      <Row style={{ justifyContent: "space-between" }}><H2>My documents</H2><Button kind="link" title={open ? "Hide" : "Show"} onPress={() => setOpen(!open)} /></Row>
      {open && <>
        {me.documents.length === 0 && <P muted small>Nothing submitted yet. Your operator needs your licence, NIC and insurance before you can go online.</P>}
        {me.documents.map((x) => (
          <View key={x.id} style={{ gap: 2 }}>
            <Row><P style={{ fontWeight: "700" }}>{DOC_TYPES.find(([k]) => k === x.docType)?.[1] ?? x.docType}</P><Badge value={x.status} />{x.fileId && <P small muted>photo attached</P>}</Row>
            <P small muted>{x.number}{x.expiresOn ? ` · expires ${x.expiresOn}` : ""}{x.reviewNote ? ` · “${x.reviewNote}”` : ""}</P>
          </View>
        ))}
        <P muted small>Add a document</P>
        <Row>{DOC_TYPES.map(([k, label]) => <Chip key={k} label={label} on={docType === k} onPress={() => setDocType(k)} />)}</Row>
        <Field label="Document number" value={number} onChangeText={setNumber} autoCapitalize="characters" />
        <Field label="Expiry date (YYYY-MM-DD)" value={expires} onChangeText={setExpires} placeholder="2027-12-31" keyboardType="numbers-and-punctuation" />
        <Row>
          <Button kind="plain" title="Take photo" onPress={() => pick(true)} />
          <Button kind="plain" title="Choose photo" onPress={() => pick(false)} />
        </Row>
        {file && <Note tone="ok">Photo attached: {file.fileName}</Note>}
        <ErrorText error={act.error} />
        <Button title="Submit for review" busy={act.busy} disabled={!number.trim()} onPress={() => act.run(async () => {
          if (expires && !/^\d{4}-\d{2}-\d{2}$/.test(expires)) throw new Error("Write the expiry date as YYYY-MM-DD, e.g. 2027-12-31.");
          await api("POST", `${base}/driver/documents`, { docType, number, expiresOn: expires || null, fileId: file?.id ?? null });
          setNumber(""); setExpires(""); setFile(null);
          onChange();
        })} />
      </>}
    </Card>
  );
}
