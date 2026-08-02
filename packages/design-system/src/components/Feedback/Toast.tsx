"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import styled, { keyframes } from "styled-components";

type ToastTone = "default" | "success" | "danger";

type ToastItem = {
  id: string;
  message: string;
  tone: ToastTone;
};

type ToastContextValue = {
  push: (message: string, tone?: ToastTone) => void;
};

const ToastContext = createContext<ToastContextValue | null>(null);

const rise = keyframes`
  from { opacity: 0; transform: translateY(10px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
`;

const Viewport = styled.div`
  position: fixed;
  right: 16px;
  bottom: 16px;
  z-index: 1200;
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: min(360px, calc(100vw - 24px));
  pointer-events: none;
`;

const ToastCard = styled.div<{ $tone: ToastTone }>`
  pointer-events: auto;
  padding: 12px 14px;
  border-radius: ${({ theme }) => theme.radii.md};
  color: ${({ theme }) => theme.colors["toast.text"]};
  background: ${({ theme, $tone }) =>
    $tone === "success"
      ? theme.colors["toast.success.bg"]
      : $tone === "danger"
        ? theme.colors["bg.danger.solid"]
        : theme.colors["toast.bg"]};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  font-size: ${({ theme }) => theme.fontSizes.sm};
  font-weight: ${({ theme }) => theme.fontWeights.medium};
  animation: ${rise} 220ms cubic-bezier(0.22, 1, 0.36, 1);
`;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const push = useCallback((message: string, tone: ToastTone = "default") => {
    const id = crypto.randomUUID();
    setItems((prev) => [...prev, { id, message, tone }]);
    window.setTimeout(() => {
      setItems((prev) => prev.filter((t) => t.id !== id));
    }, 2800);
  }, []);

  const value = useMemo(() => ({ push }), [push]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      {mounted
        ? createPortal(
            <Viewport aria-live="polite">
              {items.map((t) => (
                <ToastCard key={t.id} $tone={t.tone} role="status">
                  {t.message}
                </ToastCard>
              ))}
            </Viewport>,
            document.body,
          )
        : null}
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    return {
      push: () => {
        /* no-op outside provider */
      },
    };
  }
  return ctx;
}
