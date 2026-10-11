import type { ExpoConfig } from "expo/config";

// One codebase, many apps (D-022):
//   APP_VARIANT   passenger | driver
//   TAXIOS_TENANT optional operator slug — pins a per-operator store build (no operator code screen)
//   TAXIOS_APP_NAME / TAXIOS_BUNDLE_ID   the operator's store identity for that build
// Everything else about the brand (colour, logo, contacts) is loaded at runtime from the API.
const variant = process.env.APP_VARIANT === "driver" ? "driver" : "passenger";
const tenant = process.env.TAXIOS_TENANT ?? "";
const name = process.env.TAXIOS_APP_NAME ?? (variant === "driver" ? "Taxios Driver" : "Taxios");
const bundle = process.env.TAXIOS_BUNDLE_ID ?? (variant === "driver" ? "lk.taxios.driver" : "lk.taxios.rider");

export default (): ExpoConfig => ({
  name,
  slug: variant === "driver" ? "taxios-driver" : "taxios-rider",
  scheme: variant === "driver" ? "taxiosdriver" : "taxios",
  version: "0.1.0",
  orientation: "portrait",
  platforms: ["ios", "android"],
  userInterfaceStyle: "light",
  ios: {
    bundleIdentifier: bundle,
    supportsTablet: false,
    infoPlist: { ITSAppUsesNonExemptEncryption: false },
  },
  android: { package: bundle },
  plugins: [
    "expo-secure-store",
    ["expo-location", {
      locationWhenInUsePermission: variant === "driver"
        ? "Your location is shared with your operator while you are online or on a trip, so nearby rides can be offered to you."
        : "Your location is used once to fill in your pickup address. You can also type it instead.",
    }],
    ["expo-image-picker", {
      photosPermission: "Choose a photo of your licence, ID or insurance to send to your operator for review.",
      cameraPermission: "Take a photo of your licence, ID or insurance to send to your operator for review.",
    }],
  ],
  extra: { variant, tenant, apiUrl: process.env.TAXIOS_API_URL ?? "https://taxios.packline.lk" },
});
