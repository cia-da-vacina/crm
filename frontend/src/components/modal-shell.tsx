"use client";

import { AnimatePresence, motion } from "framer-motion";
import type { ReactNode } from "react";
import styled from "styled-components";
import { modalBackdrop, modalPanel } from "@/lib/motion";

const Overlay = styled(motion.div)`
  position: fixed;
  inset: 0;
  z-index: 400;
  display: grid;
  place-items: center;
  padding: max(16px, env(safe-area-inset-top)) 16px
    max(16px, env(safe-area-inset-bottom));
  background: rgba(8, 43, 35, 0.38);
  backdrop-filter: blur(14px) saturate(1.2);
  -webkit-backdrop-filter: blur(14px) saturate(1.2);
`;

const Panel = styled(motion.div)<{ $maxWidth: string }>`
  width: 100%;
  max-width: ${({ $maxWidth }) => $maxWidth};
  padding: 24px;
  border-radius: ${({ theme }) => theme.radii.lg};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.4) inset,
    0 24px 64px rgba(8, 43, 35, 0.18),
    0 8px 20px rgba(8, 43, 35, 0.1);
  transform-origin: center center;
  will-change: transform, opacity;
`;

type Props = {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  /** Close when clicking the dimmed backdrop. Default true. */
  closeOnBackdrop?: boolean;
  maxWidth?: string;
};

/**
 * Centered modal with an Apple-like open: soft blur backdrop + spring scale settle.
 */
export function ModalShell({
  open,
  onClose,
  children,
  closeOnBackdrop = true,
  maxWidth = "440px",
}: Props) {
  return (
    <AnimatePresence>
      {open ? (
        <Overlay
          key="modal-overlay"
          initial={modalBackdrop.initial}
          animate={modalBackdrop.animate}
          exit={modalBackdrop.exit}
          transition={modalBackdrop.transition}
          onClick={() => {
            if (closeOnBackdrop) onClose();
          }}
        >
          <Panel
            $maxWidth={maxWidth}
            role="dialog"
            aria-modal="true"
            initial={modalPanel.initial}
            animate={modalPanel.animate}
            exit={modalPanel.exit}
            onClick={(e) => e.stopPropagation()}
          >
            {children}
          </Panel>
        </Overlay>
      ) : null}
    </AnimatePresence>
  );
}
