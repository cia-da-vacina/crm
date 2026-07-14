"use client";

import { usePathname } from "next/navigation";
import { AppShell } from "@cia-da-vacina/design-system";
import { useAuth } from "@/contexts/auth-context";
import { AppShellNextLink } from "./app-shell-link";
import { UserMenu } from "./user-menu";

const baseRoutes = [
  { href: "/inbox", label: "Inbox" },
  { href: "/dashboard", label: "Dashboard" },
  { href: "/follow-ups", label: "Follow-ups" },
  { href: "/pops", label: "POPs" },
  { href: "/settings/whatsapp", label: "WhatsApp" },
];

export function AppNav({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { user } = useAuth();

  const routes = [
    ...baseRoutes.slice(0, 4),
    ...(user?.role === "admin" ? [{ href: "/users", label: "Usuários" }] : []),
    ...baseRoutes.slice(4),
  ];

  return (
    <AppShell
      brandHref="/inbox"
      links={routes.map((r) => ({
        ...r,
        active: pathname.startsWith(r.href),
      }))}
      renderLink={AppShellNextLink}
      trailing={<UserMenu />}
    >
      {children}
    </AppShell>
  );
}
