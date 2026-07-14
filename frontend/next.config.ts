import type { NextConfig } from "next";
import path from "path";
import withPWAInit from "@ducanh2912/next-pwa";

const frontendModules = path.resolve(__dirname, "node_modules");

const withPWA = withPWAInit({
  dest: "public",
  register: true,
  skipWaiting: true,
  disable: process.env.NODE_ENV === "development",
  fallbacks: {
    document: "/~offline",
  },
});

const nextConfig: NextConfig = {
  compiler: {
    styledComponents: true,
  },
  transpilePackages: [
    "@cia-da-vacina/design-system",
    "@cia-da-vacina/design-system-tokens",
    "@cia-da-vacina/icon-system",
    "@cia-da-vacina/styled-system",
  ],
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      // Deduplicate ThemeProvider across package boundaries.
      "styled-components": path.join(frontendModules, "styled-components"),
    };
    return config;
  },
};

export default withPWA(nextConfig);
