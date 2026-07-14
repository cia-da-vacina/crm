import type { NextConfig } from "next";
import path from "path";
import withPWAInit from "@ducanh2912/next-pwa";

const repoRoot = path.resolve(__dirname, "..");
const frontendModules = path.resolve(__dirname, "node_modules");
const packagesRoot = path.join(repoRoot, "packages");

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
      "styled-components": path.join(frontendModules, "styled-components"),
      "@cia-da-vacina/design-system": path.join(packagesRoot, "design-system"),
      "@cia-da-vacina/design-system-tokens": path.join(
        packagesRoot,
        "design-system-tokens",
      ),
      "@cia-da-vacina/icon-system": path.join(packagesRoot, "icon-system"),
      "@cia-da-vacina/styled-system": path.join(packagesRoot, "styled-system"),
    };
    return config;
  },
};

export default withPWA(nextConfig);
