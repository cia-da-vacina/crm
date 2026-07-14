"use client";

import styled from "styled-components";
import { Heading, Stack, Text, Button } from "@cia-da-vacina/design-system";

const Shell = styled.div`
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: linear-gradient(165deg, #f3f7f5 0%, #e7f0eb 50%, #f7faf8 100%);
`;

export default function OfflinePage() {
  return (
    <Shell>
      <Stack gap={3} style={{ textAlign: "center", maxWidth: 360 }}>
        <Heading as="h1" display>
          Sem conexão
        </Heading>
        <Text muted>
          O CRM precisa de internet para inbox e WhatsApp. Assim que voltar, toque
          em tentar novamente.
        </Text>
        <Button type="button" onClick={() => window.location.reload()}>
          Tentar novamente
        </Button>
      </Stack>
    </Shell>
  );
}
