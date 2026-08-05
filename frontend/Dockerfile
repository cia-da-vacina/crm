# syntax=docker/dockerfile:1

# CRM Cia da Vacina — frontend (Next.js) image.
# Build context MUST be the monorepo root (workspaces + design-system packages).
#
#   docker compose --env-file .env.docker.example up --build -d
#   docker build -f frontend/Dockerfile -t cia-vacina-frontend .

ARG NODE_VERSION=20

# ─── Base ────────────────────────────────────────────────────────────────────
FROM node:${NODE_VERSION}-alpine AS base
RUN apk add --no-cache libc6-compat
WORKDIR /app

# ─── Dependencies ────────────────────────────────────────────────────────────
FROM base AS deps
COPY package.json package-lock.json ./
COPY frontend/package.json ./frontend/
COPY packages/styled-system/package.json ./packages/styled-system/
COPY packages/design-system-tokens/package.json ./packages/design-system-tokens/
COPY packages/icon-system/package.json ./packages/icon-system/
COPY packages/design-system/package.json ./packages/design-system/
# Workspaces only — never git+ssh. Ignore lifecycle scripts (no src yet in this stage).
RUN npm ci --ignore-scripts

# ─── Build ───────────────────────────────────────────────────────────────────
FROM base AS builder
# npm workspaces hoist deps to root — no per-package node_modules required
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ARG USE_MOCKS=true
ARG COOKIE_SECURE=false
ARG API_URL=http://backend:8080/api/v1

ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV USE_MOCKS=$USE_MOCKS
ENV COOKIE_SECURE=$COOKIE_SECURE
ENV API_URL=$API_URL

RUN npm run build

# ─── Runtime ─────────────────────────────────────────────────────────────────
FROM base AS runner

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup --system --gid 1001 nodejs \
  && adduser --system --uid 1001 nextjs

RUN apk add --no-cache wget

# Standalone output (outputFileTracingRoot = monorepo root) lands as:
#   .next/standalone/frontend/server.js
COPY --from=builder /app/frontend/public ./frontend/public
COPY --from=builder --chown=nextjs:nodejs /app/frontend/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/frontend/.next/static ./frontend/.next/static

USER nextjs
EXPOSE 3000

# Runtime env (USE_MOCKS / API_URL / COOKIE_SECURE) comes from compose / -e.
CMD ["node", "frontend/server.js"]
