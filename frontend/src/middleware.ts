import { NextResponse, type NextRequest } from "next/server";
import { COOKIE_ACCESS, COOKIE_REFRESH } from "@/server/cookie-names";

/** Non-`/login` paths that never require an authenticated session. */
const PUBLIC_PATHS = new Set(["/~offline"]);

/** Auth BFF endpoints the client must be able to call while unauthenticated. */
const PUBLIC_API_PREFIXES = [
  "/api/auth/login",
  "/api/auth/refresh",
  "/api/auth/session",
];

/** Route groups that require an authenticated session. */
const PROTECTED_PREFIXES = [
  "/inbox",
  "/dashboard",
  "/follow-ups",
  "/pops",
  "/users",
  "/settings",
  "/engagements",
  "/customers",
  "/campaigns",
  "/units",
];

function isPublicPath(pathname: string): boolean {
  if (PUBLIC_PATHS.has(pathname)) return true;
  if (PUBLIC_API_PREFIXES.some((prefix) => pathname.startsWith(prefix))) return true;
  if (pathname.startsWith("/manifest")) return true;
  if (pathname.startsWith("/icons/")) return true;
  if (pathname === "/favicon.ico" || pathname === "/favicon.svg") return true;
  return false;
}

function isProtectedPath(pathname: string): boolean {
  return PROTECTED_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );
}

export function middleware(request: NextRequest): NextResponse {
  const { pathname } = request.nextUrl;

  const hasSession = Boolean(
    request.cookies.get(COOKIE_ACCESS)?.value || request.cookies.get(COOKIE_REFRESH)?.value,
  );

  if (pathname === "/login") {
    return hasSession
      ? NextResponse.redirect(new URL("/inbox", request.url))
      : NextResponse.next();
  }

  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  if (isProtectedPath(pathname) && !hasSession) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("from", pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon\\.ico|favicon\\.svg|icons/|manifest|sw\\.js|workbox-|swe-worker-).*)",
  ],
};
