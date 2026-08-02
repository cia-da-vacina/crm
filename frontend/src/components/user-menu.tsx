"use client";

import { useEffect, useId, useRef, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import styled from "styled-components";
import { Avatar, Text } from "@cia-da-vacina/design-system";
import { ChevronDownIcon, LogoutIcon, Moon01, Sun } from "@cia-da-vacina/icon-system";
import { easeOut } from "@/lib/motion";
import { useAuth } from "@/providers/auth-provider";
import { useThemeMode } from "@/providers/theme-provider";
import type { UserRole } from "@/domain";

const ROLE_LABELS: Record<UserRole, string> = {
  admin: "Administrador",
  manager: "Gerente",
  supervisor: "Supervisor",
  agent: "Atendente",
};

const Trigger = styled(motion.button)`
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
  transition: background 160ms ease, border-color 160ms ease;

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

const Chevron = styled(motion.span)`
  display: none;
  transform-origin: center;

  @media (min-width: 640px) {
    display: inline-flex;
  }
`;

const Panel = styled(motion.div)`
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: ${({ theme }) => theme.zIndices.dropdown};
  width: min(280px, calc(100vw - 24px));
  padding: ${({ theme }) => theme.space[3]};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  transform-origin: top right;
  overflow: hidden;
`;

const Header = styled(motion.div)`
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
  margin-bottom: ${({ theme }) => theme.space[3]};
`;

const HeaderMeta = styled.div`
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  padding-top: 2px;
  line-height: 1.3;
`;

const Divider = styled(motion.div)`
  height: 1px;
  margin: 0 0 ${({ theme }) => theme.space[2]};
  background: ${({ theme }) => theme.colors["border.subtle"]};
`;

const Actions = styled(motion.div)`
  display: flex;
  flex-direction: column;
  gap: 2px;
`;

const MenuAction = styled(motion.button)`
  display: grid;
  grid-template-columns: 20px 1fr;
  align-items: center;
  column-gap: 10px;
  width: 100%;
  margin: 0;
  padding: 8px 10px;
  border: 0;
  border-radius: ${({ theme }) => theme.radii.sm};
  background: transparent;
  color: ${({ theme }) => theme.colors["text.brand"]};
  font-family: ${({ theme }) => theme.fonts.body};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  font-weight: ${({ theme }) => theme.fontWeights.medium};
  text-align: left;
  cursor: pointer;

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  }

  &:focus-visible {
    outline: none;
    box-shadow: ${({ theme }) => theme.shadows.focus};
  }
`;

const MenuIcon = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;

  & > svg {
    width: 16px !important;
    height: 16px !important;
  }
`;

const Wrap = styled.div`
  position: relative;
`;

const panelTransition = {
  duration: 0.22,
  ease: easeOut,
};

const itemVariants = {
  hidden: { opacity: 0, y: 6 },
  show: { opacity: 1, y: 0 },
};

export function UserMenu() {
  const { user, logout } = useAuth();
  const { mode, toggle } = useThemeMode();
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

  return (
    <Wrap ref={wrapRef}>
      <Trigger
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        onClick={() => setOpen((v) => !v)}
        whileTap={{ scale: 0.98 }}
        transition={{ duration: 0.12 }}
      >
        <Avatar name={user.name} size={32} alt="" />
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
            {roleLabel}
          </Text>
        </Meta>
        <Chevron
          animate={{ rotate: open ? 180 : 0 }}
          transition={{ duration: 0.22, ease: easeOut }}
        >
          <ChevronDownIcon size="xs" fill="text.muted" />
        </Chevron>
      </Trigger>

      <AnimatePresence>
        {open ? (
          <Panel
            id={menuId}
            role="menu"
            initial={{ opacity: 0, y: -8, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -6, scale: 0.98 }}
            transition={panelTransition}
          >
            <Header
              variants={itemVariants}
              initial="hidden"
              animate="show"
              transition={{ duration: 0.2, ease: easeOut, delay: 0.03 }}
            >
              <Avatar name={user.name} size={40} alt="" />
              <HeaderMeta>
                <Text fontWeight="semibold" fontSize="sm">
                  {user.name}
                </Text>
                <Text fontSize="xs" muted>
                  {roleLabel}
                </Text>
                <Text
                  fontSize="xs"
                  muted
                  style={{
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                >
                  {user.email}
                </Text>
              </HeaderMeta>
            </Header>

            <Divider
              aria-hidden
              initial={{ scaleX: 0, opacity: 0 }}
              animate={{ scaleX: 1, opacity: 1 }}
              transition={{ duration: 0.22, ease: easeOut, delay: 0.06 }}
              style={{ transformOrigin: "left center" }}
            />

            <Actions
              initial="hidden"
              animate="show"
              variants={{
                hidden: {},
                show: {
                  transition: { staggerChildren: 0.045, delayChildren: 0.08 },
                },
              }}
            >
              <MenuAction
                type="button"
                role="menuitem"
                onClick={() => toggle()}
                variants={itemVariants}
                transition={{ duration: 0.2, ease: easeOut }}
                whileHover={{ x: 2 }}
                whileTap={{ scale: 0.99 }}
              >
                <MenuIcon>
                  {mode === "dark" ? <Sun size="sm" /> : <Moon01 size="sm" />}
                </MenuIcon>
                {mode === "dark" ? "Tema claro" : "Tema escuro"}
              </MenuAction>

              <MenuAction
                type="button"
                role="menuitem"
                onClick={() => {
                  setOpen(false);
                  void logout();
                }}
                variants={itemVariants}
                transition={{ duration: 0.2, ease: easeOut }}
                whileHover={{ x: 2 }}
                whileTap={{ scale: 0.99 }}
              >
                <MenuIcon>
                  <LogoutIcon size="sm" />
                </MenuIcon>
                Sair da conta
              </MenuAction>
            </Actions>
          </Panel>
        ) : null}
      </AnimatePresence>
    </Wrap>
  );
}
