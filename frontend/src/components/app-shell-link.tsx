"use client";

import Link from "next/link";
import styled, { keyframes } from "styled-components";
import type { AppShellLink } from "@cia-da-vacina/design-system";
import {
  isPendingForHref,
  useNavigationFeedback,
} from "@/providers/navigation-provider";

const pulse = keyframes`
  0%,
  100% {
    opacity: 0.45;
  }
  50% {
    opacity: 1;
  }
`;

const NavLink = styled(Link)<{ $active?: boolean; $pending?: boolean }>`
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
  transition:
    background 140ms ease,
    color 140ms ease,
    transform 100ms ease;

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
    transition: background 140ms ease;
  }
`;

const Icon = styled.span<{ $pending?: boolean }>`
  display: inline-flex;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  animation: ${({ $pending }) => ($pending ? pulse : "none")} 0.9s ease infinite;

  & > svg {
    width: 18px !important;
    height: 18px !important;
  }
`;

const PendingDot = styled.span`
  margin-left: auto;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: ${({ theme }) => theme.colors["text.brand"]};
  opacity: 0;
  transform: scale(0.6);
  transition:
    opacity 140ms ease,
    transform 140ms ease;
  transition-delay: 90ms;
  pointer-events: none;

  &[data-show="true"] {
    opacity: 1;
    transform: scale(1);
    animation: ${pulse} 0.9s ease infinite;
  }
`;

export function AppShellNextLink(link: AppShellLink) {
  const { pendingHref, begin } = useNavigationFeedback();
  const pending = isPendingForHref(pendingHref, link.href);
  const active = Boolean(link.active) || pending;

  return (
    <NavLink
      href={link.href}
      prefetch
      $active={active}
      $pending={pending}
      aria-current={active ? "page" : undefined}
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) return;
        begin(link.href);
      }}
    >
      {link.icon ? <Icon $pending={pending}>{link.icon}</Icon> : null}
      {link.label}
      <PendingDot data-show={pending ? "true" : "false"} aria-hidden />
    </NavLink>
  );
}
