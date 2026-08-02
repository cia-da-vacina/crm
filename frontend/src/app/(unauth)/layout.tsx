/** Minimal shell for public pages (login, etc.) — no AppShell/nav chrome. */
export default function UnauthLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
