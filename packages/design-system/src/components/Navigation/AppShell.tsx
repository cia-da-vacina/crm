"use client";

import { useEffect, useId, useState, type ReactNode } from "react";
import styled from "styled-components";
import { MenuIcon, XIcon } from "@cia-da-vacina/icon-system";
import Flex from "../Layout/Flex";
import Text from "../Typography/Text";

const Shell = styled.div`
  display: flex;
  min-height: 100dvh;
  background: ${({ theme }) => theme.colors["bg.canvas"]};
`;

const Sidebar = styled.aside<{ $open: boolean }>`
  position: fixed;
  inset: 0 auto 0 0;
  z-index: ${({ theme }) => (theme.zIndices.overlay ?? 90) + 1};
  width: min(88vw, 260px);
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px 12px;
  background: ${({ theme }) => theme.colors["nav.sidebar.bg"]};
  border-right: 1px solid ${({ theme }) => theme.colors["nav.sidebar.border"]};
  box-shadow: ${({ theme, $open }) => ($open ? theme.shadows.lg : "none")};
  transform: translateX(${({ $open }) => ($open ? "0" : "-105%")});
  transition: transform 220ms cubic-bezier(0.22, 1, 0.36, 1),
    box-shadow 220ms ease;

  @media (min-width: 960px) {
    position: sticky;
    top: 0;
    height: 100dvh;
    transform: none;
    flex-shrink: 0;
    width: 248px;
    box-shadow: none;
  }
`;

const Overlay = styled.button`
  position: fixed;
  inset: 0;
  z-index: ${({ theme }) => theme.zIndices.overlay ?? 90};
  border: 0;
  margin: 0;
  padding: 0;
  background: rgba(15, 28, 22, 0.28);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  cursor: pointer;
  animation: cvOverlayIn 180ms cubic-bezier(0.22, 1, 0.36, 1);

  @keyframes cvOverlayIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @media (min-width: 960px) {
    display: none;
  }
`;

const BrandLink = styled.a`
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
  min-width: 0;
  padding: 6px 8px 14px;
  text-decoration: none;
  color: inherit;
`;

const BrandMark = styled.img`
  width: 40px;
  height: 40px;
  object-fit: contain;
  flex-shrink: 0;
  border-radius: ${({ theme }) => theme.radii.md};
`;

const BrandText = styled.div`
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
`;

const BrandTitle = styled.div`
  font-family: ${({ theme }) => theme.fonts.display};
  font-size: ${({ theme }) => theme.fontSizes.md};
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  color: ${({ theme }) => theme.colors["text.brand"]};
  letter-spacing: ${({ theme }) => theme.letterSpacings.tight};
  line-height: 1.15;
`;

const BrandSub = styled.span`
  font-size: 11px;
  font-weight: ${({ theme }) => theme.fontWeights.bold};
  letter-spacing: 0.08em;
  text-transform: uppercase;
  line-height: 1.2;
  color: ${({ theme }) => theme.colors["text.brand"]};
`;

const NavSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 12px;
`;

const NavSectionLabel = styled.div`
  padding: 8px 10px 4px;
  font-size: 10px;
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: ${({ theme }) => theme.colors["text.muted"]};
`;

const NavItem = styled.a<{ $active?: boolean }>`
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: ${({ theme }) => theme.radii.md};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  font-weight: ${({ theme, $active }) =>
    $active ? theme.fontWeights.semibold : theme.fontWeights.medium};
  color: ${({ theme, $active }) =>
    $active ? theme.colors["nav.item.active.text"] : theme.colors["nav.item.text"]};
  background: ${({ theme, $active }) =>
    $active ? theme.colors["nav.item.active.bg"] : "transparent"};
  text-decoration: none;
  transition: background 160ms ease, color 160ms ease, transform 120ms ease;

  &:hover {
    background: ${({ theme, $active }) =>
      $active ? theme.colors["nav.item.active.bg"] : theme.colors["bg.surface.muted"]};
    color: ${({ theme }) => theme.colors["nav.item.active.text"]};
  }

  &:active {
    transform: scale(0.985);
  }

  &::before {
    content: "";
    position: absolute;
    left: 0;
    top: 8px;
    bottom: 8px;
    width: 3px;
    border-radius: 999px;
    background: ${({ theme, $active }) =>
      $active ? theme.colors["text.brand"] : "transparent"};
    transition: background 160ms ease;
  }
`;

const NavIcon = styled.span`
  display: inline-flex;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  opacity: 0.9;

  & > svg {
    width: 18px !important;
    height: 18px !important;
  }
`;

const Column = styled.div`
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
`;

const TopBar = styled.header`
  position: sticky;
  top: 0;
  z-index: ${({ theme }) => theme.zIndices.sticky};
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 56px;
  padding: 8px 16px;
  background: ${({ theme }) => theme.colors["nav.bg"]};
  border-bottom: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  backdrop-filter: blur(12px);

  @media (min-width: 960px) {
    padding: 8px 24px;
  }
`;

const MobileToggle = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  margin: 0;
  padding: 0;
  border: 1px solid transparent;
  border-radius: ${({ theme }) => theme.radii.sm};
  background: transparent;
  color: ${({ theme }) => theme.colors["text.brand"]};
  cursor: pointer;

  @media (min-width: 960px) {
    display: none;
  }

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  }

  &:focus-visible {
    outline: none;
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

const SidebarClose = styled(MobileToggle)`
  @media (min-width: 960px) {
    display: none !important;
  }
`;

const TopLeading = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
`;

const TopTrailing = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
`;

const Main = styled.main`
  flex: 1;
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  padding: 20px 16px 48px;
  min-width: 0;

  @media (min-width: 960px) {
    padding: 24px 28px 56px;
  }
`;

export type AppShellLink = {
  href: string;
  label: string;
  active?: boolean;
  icon?: ReactNode;
  group?: string;
};

export type AppShellProps = {
  brandHref?: string;
  brandLabel?: string;
  brandSubLabel?: string;
  brandLogoSrc?: string;
  links: AppShellLink[];
  leading?: ReactNode;
  trailing?: ReactNode;
  children: ReactNode;
  renderLink?: (link: AppShellLink) => ReactNode;
  renderBrand?: (props: {
    href: string;
    label: string;
    subLabel?: string;
    logoSrc?: string;
  }) => ReactNode;
  /** @deprecated ignored — content is full-bleed up to 1440px */
  contentMaxWidth?: number | string;
};

export default function AppShell({
  brandHref = "/",
  brandLabel = "Cia da Vacina",
  brandSubLabel = "CRM",
  brandLogoSrc,
  links,
  leading,
  trailing,
  children,
  renderLink,
  renderBrand,
}: AppShellProps) {
  const [open, setOpen] = useState(false);
  const drawerId = useId();

  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = prev;
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const groups = links.reduce<Record<string, AppShellLink[]>>((acc, link) => {
    const key = link.group ?? "Menu";
    (acc[key] ??= []).push(link);
    return acc;
  }, {});

  function renderNavLink(link: AppShellLink) {
    if (renderLink) return renderLink(link);
    return (
      <NavItem href={link.href} $active={link.active}>
        {link.icon ? <NavIcon>{link.icon}</NavIcon> : null}
        {link.label}
      </NavItem>
    );
  }

  const sidebar = (
    <Sidebar id={drawerId} $open={open} aria-hidden={!open}>
      <Flex alignItems="flex-start" justifyContent="space-between" gap={2}>
        {renderBrand ? (
          <div onClick={() => setOpen(false)} style={{ minWidth: 0, flex: 1 }}>
            {renderBrand({
              href: brandHref,
              label: brandLabel,
              subLabel: brandSubLabel,
              logoSrc: brandLogoSrc,
            })}
          </div>
        ) : (
          <BrandLink href={brandHref} onClick={() => setOpen(false)}>
            {brandLogoSrc ? <BrandMark src={brandLogoSrc} alt="" /> : null}
            <BrandText>
              <BrandTitle>{brandLabel}</BrandTitle>
              {brandSubLabel ? <BrandSub>{brandSubLabel}</BrandSub> : null}
            </BrandText>
          </BrandLink>
        )}
        <SidebarClose
          type="button"
          aria-label="Fechar menu"
          onClick={() => setOpen(false)}
        >
          <XIcon size="md" />
        </SidebarClose>
      </Flex>

      <nav aria-label="Principal" style={{ flex: 1, overflowY: "auto" }}>
        {Object.entries(groups).map(([group, groupLinks]) => (
          <NavSection key={group}>
            <NavSectionLabel>{group}</NavSectionLabel>
            {groupLinks.map((link) => (
              <div key={link.href} onClick={() => setOpen(false)}>
                {renderNavLink(link)}
              </div>
            ))}
          </NavSection>
        ))}
      </nav>

      <Text fontSize="xs" muted style={{ padding: "8px 10px" }}>
        Atendimento multicanal
      </Text>
    </Sidebar>
  );

  return (
    <Shell>
      {open ? (
        <Overlay type="button" aria-label="Fechar menu" onClick={() => setOpen(false)} />
      ) : null}
      {sidebar}
      <Column>
        <TopBar>
          <TopLeading>
            <MobileToggle
              type="button"
              aria-label={open ? "Fechar menu" : "Abrir menu"}
              aria-expanded={open}
              aria-controls={drawerId}
              onClick={() => setOpen((v) => !v)}
            >
              {open ? <XIcon size="md" /> : <MenuIcon size="md" />}
            </MobileToggle>
            {leading}
          </TopLeading>
          <TopTrailing>{trailing}</TopTrailing>
        </TopBar>
        <Main>{children}</Main>
      </Column>
    </Shell>
  );
}
