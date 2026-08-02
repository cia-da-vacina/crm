"use client";

import { usePathname } from "next/navigation";
import styled from "styled-components";
import { AppShell } from "@cia-da-vacina/design-system";
import {
  BuildingIcon,
  Calendar as CalendarIcon,
  ChartIcon,
  ClockIcon,
  InboxIcon,
  MessageIcon,
  SettingsIcon,
  UsersIcon,
} from "@cia-da-vacina/icon-system";
import { useAuth } from "@/providers/auth-provider";
import { useNavigationFeedback } from "@/providers/navigation-provider";
import { AppShellNextLink } from "./app-shell-link";
import { AppShellBrandLink } from "./app-shell-brand";
import { UnitSwitcher } from "./unit-switcher";
import { UserMenu } from "./user-menu";

const Content = styled.div<{ $pending?: boolean }>`
  width: 100%;
  transition: opacity 160ms ease;
  opacity: ${({ $pending }) => ($pending ? 0.62 : 1)};
  pointer-events: ${({ $pending }) => ($pending ? "none" : "auto")};
`;

const baseRoutes = [
  { href: "/inbox", label: "Inbox", group: "Atendimento", icon: <InboxIcon size="sm" /> },
  {
    href: "/engagements",
    label: "Interações",
    group: "Atendimento",
    icon: <MessageIcon size="sm" />,
  },
  { href: "/follow-ups", label: "Follow-ups", group: "Atendimento", icon: <ClockIcon size="sm" /> },
  { href: "/dashboard", label: "Dashboard", group: "Gestão", icon: <ChartIcon size="sm" /> },
  {
    href: "/campaigns",
    label: "Agenda",
    group: "Gestão",
    icon: <CalendarIcon size="sm" />,
  },
  { href: "/pops", label: "POPs", group: "Gestão", icon: <MessageIcon size="sm" /> },
];

const adminRoutes = [
  { href: "/units", label: "Unidades", group: "Admin", icon: <BuildingIcon size="sm" /> },
  { href: "/users", label: "Usuários", group: "Admin", icon: <UsersIcon size="sm" /> },
  {
    href: "/settings/meta",
    label: "Canais Meta",
    group: "Admin",
    icon: <SettingsIcon size="sm" />,
  },
];

export function AppNav({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { user } = useAuth();
  const { isNavigating } = useNavigationFeedback();

  const routes = [...baseRoutes, ...(user?.role === "admin" ? adminRoutes : [])];

  return (
    <AppShell
      brandHref="/inbox"
      brandLabel="Cia da Vacina"
      brandSubLabel="CRM"
      brandLogoSrc="/favicon.svg"
      links={routes.map((r) => ({
        ...r,
        active: pathname.startsWith(r.href),
      }))}
      renderLink={AppShellNextLink}
      renderBrand={(brand) => <AppShellBrandLink {...brand} />}
      leading={<UnitSwitcher />}
      trailing={<UserMenu />}
    >
      <Content $pending={isNavigating}>{children}</Content>
    </AppShell>
  );
}
