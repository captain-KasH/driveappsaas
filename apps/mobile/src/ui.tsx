import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import { ActivityIndicator, AppState, Image, KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View, type TextInputProps, type ViewStyle } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { config, type Breakdown, money } from "./api";

// Theme = the operator's runtime brand (T-007). Defaults to Taxios teal until the brand loads.
export type Theme = { color: string; onColor: string; name: string; logoUrl: string | null };
const ThemeCtx = createContext<Theme>({ color: "#0f766e", onColor: "#ffffff", name: "Taxios", logoUrl: null });
export const ThemeProvider = ThemeCtx.Provider;
export const useTheme = () => useContext(ThemeCtx);

export const c = { bg: "#f4f7f7", surface: "#ffffff", text: "#0f1a1c", muted: "#52676a", line: "#d9e3e2", bad: "#b3261e", badBg: "#fbe9e7", ok: "#1e7a3c", okBg: "#e5f4ea", warn: "#8a5a00", warnBg: "#fdf3d8" };

export function Screen({ children, scroll = true, header }: { children: ReactNode; scroll?: boolean; header?: ReactNode }) {
  const body = scroll ? <ScrollView contentContainerStyle={s.content} keyboardShouldPersistTaps="handled">{children}</ScrollView> : <View style={[s.content, { flex: 1 }]}>{children}</View>;
  return (
    <SafeAreaView style={s.screen} edges={["top", "left", "right"]}>
      {header}
      <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === "ios" ? "padding" : undefined}>{body}</KeyboardAvoidingView>
    </SafeAreaView>
  );
}

export function Header({ right }: { right?: ReactNode }) {
  const t = useTheme();
  return (
    <View style={[s.header, { backgroundColor: t.color }]}>
      {t.logoUrl ? <Image source={{ uri: config.apiUrl + t.logoUrl }} style={s.logo} /> : <View style={[s.logo, s.logoMark]}><Text style={{ color: t.color, fontWeight: "800" }}>{t.name.slice(0, 1)}</Text></View>}
      <Text style={[s.headerTitle, { color: t.onColor }]} numberOfLines={1}>{t.name}</Text>
      <View style={{ flex: 1 }} />
      {right}
    </View>
  );
}

export function H1({ children }: { children: ReactNode }) { return <Text style={s.h1}>{children}</Text>; }
export function H2({ children }: { children: ReactNode }) { return <Text style={s.h2}>{children}</Text>; }
export function P({ children, muted, small, style }: { children: ReactNode; muted?: boolean; small?: boolean; style?: object }) {
  return <Text style={[s.p, muted && { color: c.muted }, small && { fontSize: 13 }, style]}>{children}</Text>;
}
export function Card({ children, style }: { children: ReactNode; style?: ViewStyle }) { return <View style={[s.card, style]}>{children}</View>; }
export function Row({ children, style }: { children: ReactNode; style?: ViewStyle }) { return <View style={[{ flexDirection: "row", alignItems: "center", gap: 8, flexWrap: "wrap" }, style]}>{children}</View>; }

export function Button({ title, onPress, kind = "primary", disabled, busy }: { title: string; onPress: () => void; kind?: "primary" | "plain" | "danger" | "link"; disabled?: boolean; busy?: boolean }) {
  const t = useTheme();
  const bg = kind === "primary" ? t.color : kind === "danger" ? c.badBg : kind === "link" ? "transparent" : c.surface;
  const fg = kind === "primary" ? t.onColor : kind === "danger" ? c.bad : kind === "link" ? t.color : c.text;
  return (
    <Pressable accessibilityRole="button" accessibilityState={{ disabled: !!(disabled || busy) }} disabled={disabled || busy} onPress={onPress}
      style={({ pressed }) => [s.btn, kind === "link" ? s.linkBtn : null, { backgroundColor: bg, borderColor: kind === "plain" ? c.line : bg, opacity: disabled ? 0.5 : pressed ? 0.8 : 1 }]}>
      {busy ? <ActivityIndicator color={fg} /> : <Text style={[s.btnText, { color: fg }]}>{title}</Text>}
    </Pressable>
  );
}

export function Field({ label, ...props }: TextInputProps & { label: string }) {
  return (
    <View style={{ gap: 4 }}>
      <Text style={s.label}>{label}</Text>
      <TextInput placeholderTextColor="#90a3a5" style={s.input} {...props} />
    </View>
  );
}

export function Chip({ label, on, onPress }: { label: string; on: boolean; onPress: () => void }) {
  const t = useTheme();
  return (
    <Pressable accessibilityRole="button" accessibilityState={{ selected: on }} onPress={onPress}
      style={[s.chip, on && { backgroundColor: t.color, borderColor: t.color }]}>
      <Text style={{ color: on ? t.onColor : c.text, fontWeight: "600" }}>{label}</Text>
    </Pressable>
  );
}

export function ErrorText({ error }: { error: unknown }) {
  if (!error) return null;
  return <Text accessibilityRole="alert" style={s.error}>{error instanceof Error ? error.message : String(error)}</Text>;
}
export function Note({ children, tone = "warn" }: { children: ReactNode; tone?: "warn" | "ok" | "bad" }) {
  const map = { warn: [c.warnBg, c.warn], ok: [c.okBg, c.ok], bad: [c.badBg, c.bad] } as const;
  return <Text style={[s.note, { backgroundColor: map[tone][0], color: map[tone][1] }]}>{children}</Text>;
}
export function Badge({ value }: { value?: string | null }) {
  if (!value) return null;
  return <Text style={s.badge}>{value.replaceAll("_", " ")}</Text>;
}
export function Loading() { return <View style={{ padding: 32, alignItems: "center" }}><ActivityIndicator /></View>; }

export function FareLines({ b, title }: { b?: Breakdown | null; title?: string }) {
  if (!b) return null;
  return (
    <View style={{ gap: 4 }}>
      {title && <P muted small>{title}</P>}
      {b.lines.map((l) => <Row key={l.code} style={{ justifyContent: "space-between" }}><P>{l.label}</P><P>{money(l.amountMinor, b.currency)}</P></Row>)}
      {b.taxMinor > 0 && <Row style={{ justifyContent: "space-between" }}><P>Tax</P><P>{money(b.taxMinor, b.currency)}</P></Row>}
      <Row style={{ justifyContent: "space-between", borderTopWidth: 1, borderTopColor: c.line, paddingTop: 4 }}><Text style={s.total}>Total</Text><Text style={s.total}>{money(b.totalMinor, b.currency)}</Text></Row>
    </View>
  );
}

/** Async action with busy/error state. */
export function useAction() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const run = useCallback(async <T,>(fn: () => Promise<T>) => {
    setBusy(true); setError(null);
    try { return await fn(); } catch (e) { setError(e); return undefined; } finally { setBusy(false); }
  }, []);
  return { busy, error, run, setError };
}

/** Polls fn every ms while the app is in the foreground. */
export function usePoll(fn: () => void, ms: number) {
  const ref = useRef(fn);
  ref.current = fn;
  useEffect(() => {
    ref.current();
    const id = setInterval(() => AppState.currentState === "active" && ref.current(), ms);
    const sub = AppState.addEventListener("change", (st) => st === "active" && ref.current());
    return () => { clearInterval(id); sub.remove(); };
  }, [ms]);
}

const s = StyleSheet.create({
  screen: { flex: 1, backgroundColor: c.bg },
  content: { padding: 16, gap: 12, paddingBottom: 48 },
  header: { flexDirection: "row", alignItems: "center", gap: 10, paddingHorizontal: 16, paddingVertical: 12 },
  headerTitle: { fontSize: 18, fontWeight: "800", flexShrink: 1 },
  logo: { width: 32, height: 32, borderRadius: 8, backgroundColor: "#fff" },
  logoMark: { alignItems: "center", justifyContent: "center" },
  h1: { fontSize: 24, fontWeight: "800", color: c.text },
  h2: { fontSize: 18, fontWeight: "700", color: c.text },
  p: { fontSize: 15, color: c.text, lineHeight: 21 },
  card: { backgroundColor: c.surface, borderRadius: 12, padding: 14, gap: 10, borderWidth: 1, borderColor: c.line },
  btn: { minHeight: 48, borderRadius: 10, borderWidth: 1, alignItems: "center", justifyContent: "center", paddingHorizontal: 16 },
  linkBtn: { minHeight: 36, alignSelf: "flex-start", paddingHorizontal: 0 },
  btnText: { fontSize: 16, fontWeight: "700" },
  label: { fontSize: 13, fontWeight: "600", color: c.muted },
  input: { minHeight: 48, borderRadius: 10, borderWidth: 1, borderColor: c.line, backgroundColor: c.surface, paddingHorizontal: 12, fontSize: 16, color: c.text },
  chip: { paddingHorizontal: 14, paddingVertical: 10, borderRadius: 20, borderWidth: 1, borderColor: c.line, backgroundColor: c.surface },
  error: { backgroundColor: c.badBg, color: c.bad, padding: 10, borderRadius: 8, overflow: "hidden" },
  note: { padding: 10, borderRadius: 8, overflow: "hidden", fontSize: 14 },
  badge: { backgroundColor: "#eaf1f0", color: c.text, paddingHorizontal: 8, paddingVertical: 3, borderRadius: 6, overflow: "hidden", fontSize: 12, fontWeight: "600" },
  total: { fontSize: 16, fontWeight: "800", color: c.text },
});
