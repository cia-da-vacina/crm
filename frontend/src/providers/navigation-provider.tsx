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
import { usePathname, useRouter } from "next/navigation";

type NavigationFeedbackValue = {
  /** Href currently navigating to, if any. */
  pendingHref: string | null;
  /** Mark a destination as in-flight (for Link onClick). */
  begin: (href: string) => void;
  /** push + immediate pending feedback. */
  navigate: (href: string) => void;
  /** True while any in-app navigation is waiting on the new route. */
  isNavigating: boolean;
};

const NavigationFeedbackContext = createContext<NavigationFeedbackValue | null>(
  null,
);

function normalizeHref(href: string): string {
  const path = href.split("?")[0]?.split("#")[0] ?? href;
  return path.endsWith("/") && path.length > 1 ? path.slice(0, -1) : path;
}

function isSameDestination(pathname: string, href: string): boolean {
  const target = normalizeHref(href);
  const current = normalizeHref(pathname);
  return current === target;
}

export function NavigationFeedbackProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [pendingHref, setPendingHref] = useState<string | null>(null);

  useEffect(() => {
    setPendingHref(null);
  }, [pathname]);

  // Safety net if navigation is cancelled or stalls.
  useEffect(() => {
    if (!pendingHref) return;
    const timer = window.setTimeout(() => setPendingHref(null), 12_000);
    return () => window.clearTimeout(timer);
  }, [pendingHref]);

  const begin = useCallback(
    (href: string) => {
      if (isSameDestination(pathname, href)) return;
      setPendingHref(normalizeHref(href));
    },
    [pathname],
  );

  const navigate = useCallback(
    (href: string) => {
      if (isSameDestination(pathname, href)) return;
      begin(href);
      router.push(href);
    },
    [begin, pathname, router],
  );

  const value = useMemo<NavigationFeedbackValue>(
    () => ({
      pendingHref,
      begin,
      navigate,
      isNavigating: pendingHref != null,
    }),
    [pendingHref, begin, navigate],
  );

  return (
    <NavigationFeedbackContext.Provider value={value}>
      {children}
    </NavigationFeedbackContext.Provider>
  );
}

export function useNavigationFeedback(): NavigationFeedbackValue {
  const ctx = useContext(NavigationFeedbackContext);
  if (!ctx) {
    throw new Error("useNavigationFeedback must be used within NavigationFeedbackProvider");
  }
  return ctx;
}

/** Soft match for nav items: pending target equals or is nested under the item href. */
export function isPendingForHref(pendingHref: string | null, href: string): boolean {
  if (!pendingHref) return false;
  const item = normalizeHref(href);
  return pendingHref === item || pendingHref.startsWith(`${item}/`);
}
