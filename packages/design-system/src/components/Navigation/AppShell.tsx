"use client";

import { useEffect, useId, useState, type ReactNode } from "react";
import styled from "styled-components";
import { MenuIcon, XIcon } from "@cia-da-vacina/icon-system";
import Flex from "../Layout/Flex";
import Box from "../Layout/Box";

const Header = styled.header`
  position: sticky;
  top: 0;
  z-index: ${({ theme }) => theme.zIndices.sticky};
  border-bottom: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  background: ${({ theme }) => theme.colors["nav.bg"]};
  backdrop-filter: blur(12px);
`;

const HeaderInner = styled.div`
  max-width: 960px;
  margin: 0 auto;
  padding: 10px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
`;

const Brand = styled.a`
  font-family: ${({ theme }) => theme.fonts.display};
  font-size: ${({ theme }) => theme.fontSizes.lg};
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  color: ${({ theme }) => theme.colors["text.brand"]};
  letter-spacing: ${({ theme }) => theme.letterSpacings.tight};
  white-space: nowrap;
  text-decoration: none;
`;

const DesktopNav = styled.nav`
  display: none;
  align-items: center;
  gap: 4px;

  @media (min-width: 900px) {
    display: flex;
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

  @media (min-width: 900px) {
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

const Trailing = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
`;

const Overlay = styled.button`
  position: fixed;
  inset: 0;
  z-index: ${({ theme }) => theme.zIndices.overlay ?? 90};
  border: 0;
  margin: 0;
  padding: 0;
  background: ${({ theme }) => theme.colors["bg.overlay"]};
  cursor: pointer;
`;

const Drawer = styled.aside<{ $open: boolean }>`
  position: fixed;
  top: 0;
  left: 0;
  z-index: ${({ theme }) => (theme.zIndices.overlay ?? 90) + 1};
  width: min(88vw, 320px);
  height: 100dvh;
  padding: 16px;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border-right: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  transform: translateX(${({ $open }) => ($open ? "0" : "-105%")});
  transition: transform 180ms ease;
  display: flex;
  flex-direction: column;
  gap: 16px;

  @media (min-width: 900px) {
    display: none;
  }
`;

const DrawerLink = styled.div`
  a,
  & > a {
    display: block;
    width: 100%;
    padding: 12px 14px !important;
    border-radius: ${({ theme }) => theme.radii.md};
    font-size: ${({ theme }) => theme.fontSizes.md} !important;
    text-decoration: none;
  }
`;

const NavLink = styled.a<{ $active?: boolean }>`
  padding: 6px 10px;
  border-radius: ${({ theme }) => theme.radii.sm};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  color: ${({ theme, $active }) =>
    $active ? theme.colors["nav.item.active.text"] : theme.colors["nav.item.text"]};
  background: ${({ theme, $active }) =>
    $active ? theme.colors["nav.item.active.bg"] : "transparent"};
  text-decoration: none;
  transition: background 120ms ease, color 120ms ease;

  &:hover {
    color: ${({ theme }) => theme.colors["nav.item.active.text"]};
  }
`;

const Main = styled.main`
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 16px 16px 56px;
  display: flex;
  flex-direction: column;
  align-items: center;

  @media (min-width: 768px) {
    padding: 20px 20px 56px;
  }
`;

const Content = styled.div`
  width: 100%;
  max-width: 720px;
  min-width: 0;
`;

export type AppShellLink = {
  href: string;
  label: string;
  active?: boolean;
};

export type AppShellProps = {
  brandHref?: string;
  brandLabel?: string;
  links: AppShellLink[];
  trailing?: ReactNode;
  children: ReactNode;
  renderLink?: (link: AppShellLink) => ReactNode;
  contentMaxWidth?: number | string;
};

export default function AppShell({
  brandHref = "/",
  brandLabel = "Cia da Vacina",
  links,
  trailing,
  children,
  renderLink,
  contentMaxWidth = 720,
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

  function renderNavLink(link: AppShellLink) {
    if (renderLink) return renderLink(link);
    return (
      <NavLink href={link.href} $active={link.active}>
        {link.label}
      </NavLink>
    );
  }

  return (
    <Box minHeight="100vh">
      <Header>
        <HeaderInner>
          <Flex alignItems="center" gap={2} style={{ minWidth: 0 }}>
            <MobileToggle
              type="button"
              aria-label={open ? "Fechar menu" : "Abrir menu"}
              aria-expanded={open}
              aria-controls={drawerId}
              onClick={() => setOpen((v) => !v)}
            >
              {open ? <XIcon size="md" /> : <MenuIcon size="md" />}
            </MobileToggle>
            <Brand href={brandHref}>{brandLabel}</Brand>
            <DesktopNav aria-label="Principal">
              {links.map((link) => (
                <span key={link.href}>{renderNavLink(link)}</span>
              ))}
            </DesktopNav>
          </Flex>
          <Trailing>{trailing}</Trailing>
        </HeaderInner>
      </Header>

      {open ? (
        <Overlay type="button" aria-label="Fechar menu" onClick={() => setOpen(false)} />
      ) : null}
      <Drawer id={drawerId} $open={open} aria-hidden={!open}>
        <Flex alignItems="center" justifyContent="space-between">
          <Brand href={brandHref} onClick={() => setOpen(false)}>
            {brandLabel}
          </Brand>
          <MobileToggle
            type="button"
            aria-label="Fechar menu"
            onClick={() => setOpen(false)}
            style={{ display: "inline-flex" }}
          >
            <XIcon size="md" />
          </MobileToggle>
        </Flex>
        <nav aria-label="Mobile">
          <Flex flexDirection="column" gap={1}>
            {links.map((link) => (
              <DrawerLink key={link.href} onClick={() => setOpen(false)}>
                {renderNavLink(link)}
              </DrawerLink>
            ))}
          </Flex>
        </nav>
      </Drawer>

      <Main>
        <Content style={{ maxWidth: contentMaxWidth }}>{children}</Content>
      </Main>
    </Box>
  );
}
