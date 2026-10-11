import type { NextConfig } from "next";

// Static export: the Go API embeds `out/` and serves it, so the site and /v1 share one origin.
const config: NextConfig = {
  output: "export",
  trailingSlash: true,
  images: { unoptimized: true },
  // In `next dev`, proxy the API to the deployed server (or a local one via TAXIOS_API).
  ...(process.env.NODE_ENV === "development" && {
    async rewrites() {
      return [{ source: "/v1/:path*", destination: `${process.env.TAXIOS_API ?? "https://taxios.packline.lk"}/v1/:path*` }];
    },
  }),
};
export default config;
