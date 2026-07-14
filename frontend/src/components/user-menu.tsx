"use client";

import { useEffect, useId, useRef, useState } from "react";
import styled from "styled-components";
import {
  Avatar,
  Box,
  Button,
  Flex,
  SelectField,
  Text,
} from "@cia-da-vacina/design-system";
import { ChevronDownIcon, LogoutIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import type { UserRole } from "@/lib/types";

const ROLE_LABELS: Record<UserRole, string> = {
  admin: "Administrador",
  manager: "Gerente",
  supervisor: "Supervisor",
  agent: "Atendente",
};

function avatarUrl(seed: string) {
  return `https://api.dicebear.com/9.x/notionists/svg?seed=${encodeURIComponent(seed)}&backgroundColor=d4ece4`;
}

const Trigger = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 4px;
  margin: 0;
  border: 1px solid transparent;
  border-radius: ${({ theme }) => theme.radii.md};
  background: transparent;
  cursor: pointer;
  font-family: ${({ theme }) => theme.fonts.body};
  color: ${({ theme }) => theme.colors["text.primary"]};
  transition: background 120ms ease, border-color 120ms ease;

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
    border-color: ${({ theme }) => theme.colors["border.subtle"]};
  }

  &:focus-visible {
    outline: none;
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }

  &[aria-expanded="true"] {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
    border-color: ${({ theme }) => theme.colors["border.default"]};
  }
`;

const Meta = styled.div`
  display: none;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.15;
  min-width: 0;

  @media (min-width: 640px) {
    display: flex;
  }
`;

const Chevron = styled.span`
  display: none;

  @media (min-width: 640px) {
    display: inline-flex;
  }
`;

const Panel = styled.div`
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: ${({ theme }) => theme.zIndices.dropdown};
  width: min(260px, calc(100vw - 24px));
  padding: ${({ theme }) => theme.space[3]};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  box-shadow: ${({ theme }) => theme.shadows.lg};
`;

const Wrap = styled.div`
  position: relative;
`;

export function UserMenu() {
  const { user, logout, units, activeUnitId, setUnit } = useAuth();
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  if (!user) return null;

  const roleLabel = ROLE_LABELS[user.role] ?? user.role;
  const photo = avatarUrl(user.email || user.id);
  const activeUnit = units.find((u) => u.id === activeUnitId);

  return (
    <Wrap ref={wrapRef}>
      <Trigger
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        onClick={() => setOpen((v) => !v)}
      >
        <Avatar name={user.name} src={photo} size={32} alt="" />
        <Meta>
          <Text
            fontSize="xs"
            fontWeight="semibold"
            style={{
              maxWidth: 120,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {user.name}
          </Text>
          <Text
            fontSize="xs"
            muted
            style={{
              maxWidth: 120,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {activeUnit?.name ?? roleLabel}
          </Text>
        </Meta>
        <Chevron>
          <ChevronDownIcon size="xs" fill="text.muted" />
        </Chevron>
      </Trigger>

      {open ? (
        <Panel id={menuId} role="menu">
          <Flex gap={3} alignItems="center" mb={3}>
            <Avatar name={user.name} src={photo} size={44} alt="" />
            <Flex flexDirection="column" style={{ minWidth: 0, lineHeight: 1.25 }}>
              <Text fontWeight="semibold" fontSize="sm">
                {user.name}
              </Text>
              <Text fontSize="xs" muted>
                {roleLabel}
              </Text>
              <Text fontSize="xs" muted style={{ overflow: "hidden", textOverflow: "ellipsis" }}>
                {user.email}
              </Text>
            </Flex>
          </Flex>

          {units.length > 0 ? (
            <Box mb={3}>
              <Text as="label" htmlFor="user-menu-unit" fontSize="xs" muted>
                Unidade ativa
              </Text>
              <Box mt={1}>
                <SelectField
                  id="user-menu-unit"
                  fieldSize="sm"
                  fullWidth
                  appearance="default"
                  value={activeUnitId ?? ""}
                  onChange={(e) => setUnit(e.target.value)}
                  options={units.map((u) => ({ value: u.id, label: u.name }))}
                />
              </Box>
            </Box>
          ) : null}

          <Box
            height="1px"
            bg="border.subtle"
            mb={2}
            aria-hidden
          />

          <Button
            variant="ghost"
            size="sm"
            fullWidth
            leftIcon={<LogoutIcon size="sm" />}
            onClick={() => {
              setOpen(false);
              void logout();
            }}
          >
            Sair da conta
          </Button>
        </Panel>
      ) : null}
    </Wrap>
  );
}
