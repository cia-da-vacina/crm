import type { NextConfig } from "next";
import fs from "fs";
import path from "path";
import withPWAInit from "@ducanh2912/next-pwa";

/**
 * Resolve a package from this workspace or the monorepo root (npm hoists
 * deps to the workspace root, so `frontend/node_modules/X` may not exist).
 */
function resolvePackage(name: string): string {
  const candidates = [
    path.resolve(__dirname, "node_modules", name),
    path.resolve(__dirname, "..", "node_modules", name),
  ];
  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) return candidate;
  }
  return candidates[0];
}

const withPWA = withPWAInit({
  dest: "public",
  register: true,
  disable: process.env.NODE_ENV === "development",
  fallbacks: {
    document: "/~offline",
  },
});

const nextConfig: NextConfig = {
  // Required for slim Docker images (copies only traced deps into .next/standalone).
  output: "standalone",
  // Trace files from the monorepo root so workspace packages resolve correctly.
  outputFileTracingRoot: path.join(__dirname, ".."),
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
      // Deduplicate ThemeProvider across package boundaries (workspaces hoist).
      "styled-components": resolvePackage("styled-components"),
    };
    return config;
  },
};

export default withPWA(nextConfig);
