"use client";

import Link from "next/link";
import styled from "styled-components";
import type { AppShellLink } from "@cia-da-vacina/design-system";

const NavLink = styled(Link)<{ $active?: boolean }>`
  display: inline-block;
  padding: 8px 12px;
  border-radius: ${({ theme }) => theme.radii.md};
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

export function AppShellNextLink(link: AppShellLink) {
  return (
    <NavLink href={link.href} $active={link.active}>
      {link.label}
    </NavLink>
  );
}
