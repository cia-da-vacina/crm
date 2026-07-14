"use client";

import { useEffect, useState } from "react";
import styled from "styled-components";
import { Button, Flex, Text } from "@cia-da-vacina/design-system";

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

const Banner = styled.div`
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(12px + env(safe-area-inset-bottom, 0px));
  z-index: ${({ theme }) => theme.zIndices.toast};
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  box-shadow: ${({ theme }) => theme.shadows.lg};

  @media (min-width: 900px) {
    left: auto;
    right: 20px;
    width: 360px;
  }
`;

export function InstallPrompt() {
  const [deferred, setDeferred] = useState<BeforeInstallPromptEvent | null>(
    null,
  );
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const dismissed = sessionStorage.getItem("pwa-install-dismissed");
    if (dismissed) return;

    function onBeforeInstall(e: Event) {
      e.preventDefault();
      setDeferred(e as BeforeInstallPromptEvent);
      setVisible(true);
    }

    window.addEventListener("beforeinstallprompt", onBeforeInstall);
    return () =>
      window.removeEventListener("beforeinstallprompt", onBeforeInstall);
  }, []);

  if (!visible || !deferred) return null;

  return (
    <Banner role="dialog" aria-label="Instalar aplicativo">
      <Text fontSize="sm" style={{ flex: 1 }}>
        Instale o CRM Cia da Vacina no celular para acesso rápido.
      </Text>
      <Flex gap={1} flexShrink={0}>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            sessionStorage.setItem("pwa-install-dismissed", "1");
            setVisible(false);
          }}
        >
          Agora não
        </Button>
        <Button
          size="sm"
          onClick={async () => {
            await deferred.prompt();
            await deferred.userChoice;
            setVisible(false);
            setDeferred(null);
          }}
        >
          Instalar
        </Button>
      </Flex>
    </Banner>
  );
}
