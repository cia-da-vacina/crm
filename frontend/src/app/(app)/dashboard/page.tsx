"use client";

import { useQuery } from "@tanstack/react-query";
import {
  Box,
  Flex,
  Heading,
  Spinner,
  Stack,
  Text,
} from "@cia-da-vacina/design-system";
import { ChartIcon } from "@cia-da-vacina/icon-system";
import { useAuth } from "@/contexts/auth-context";
import { api } from "@/lib/api";
import { STAGE_LABELS, type PipelineStage } from "@/lib/types";

function Kpi({ label, value, hint }: { label: string; value: string | number; hint?: string }) {
  return (
    <Box
      p={3}
      bg="bg.surface"
      borderWidth="hairline"
      borderStyle="solid"
      borderColor="border.default"
      borderRadius="md"
      minWidth="140px"
      flex={1}
    >
      <Text fontSize="xs" muted>
        {label}
      </Text>
      <Heading as="h2" style={{ marginTop: 8 }}>
        {value}
      </Heading>
      {hint && (
        <Text fontSize="xs" muted style={{ marginTop: 4 }}>
          {hint}
        </Text>
      )}
    </Box>
  );
}

export default function DashboardPage() {
  const { activeUnitId, units } = useAuth();
  const { data, isLoading } = useQuery({
    queryKey: ["dashboard", activeUnitId],
    queryFn: () => api.dashboard(activeUnitId ?? undefined),
    enabled: Boolean(activeUnitId),
  });

  const unitName = units.find((u) => u.id === activeUnitId)?.name ?? "Unidade";

  return (
    <Stack gap={3} width="100%">
      <Stack gap={1} style={{ textAlign: "center" }} alignItems="center">
        <Flex gap={2} alignItems="center" justifyContent="center">
          <ChartIcon size="lg" fill="text.brand" />
          <Heading as="h1" display>
            Dashboard
          </Heading>
        </Flex>
        <Text muted fontSize="sm">
          {unitName} · consolidado das 5 unidades
        </Text>
      </Stack>

      {isLoading || !data ? (
        <Flex gap={2} alignItems="center">
          <Spinner />
          <Text muted>Carregando indicadores…</Text>
        </Flex>
      ) : (
        <>
          <Flex gap={3} flexWrap="wrap">
            <Kpi label="Abertas" value={data.open_conversations} />
            <Kpi label="IA em triagem" value={data.ai_triage} />
            <Kpi label="Com humano" value={data.human} />
            <Kpi
              label="Conversão"
              value={`${data.conversion_rate}%`}
              hint={`${data.closed} fechados / ${data.not_closed} não fechados`}
            />
            <Kpi label="Follow-ups abertos" value={data.awaiting_followup} />
          </Flex>

          <Stack gap={3}>
            <Heading as="h3">Funil por etapa (unidade)</Heading>
            <Flex gap={2} flexWrap="wrap">
              {(Object.keys(data.by_stage) as PipelineStage[]).map((stage) => (
                <Box
                  key={stage}
                  px={2}
                  py={2}
                  bg="bg.surface"
                  borderWidth="hairline"
                  borderStyle="solid"
                  borderColor="border.default"
                  borderRadius="sm"
                  minWidth="120px"
                >
                  <Text fontSize="xs" muted>
                    {STAGE_LABELS[stage]}
                  </Text>
                  <Text fontWeight="semibold" fontSize="xl">
                    {data.by_stage[stage]}
                  </Text>
                </Box>
              ))}
            </Flex>
          </Stack>

          <Stack gap={3}>
            <Heading as="h3">Por unidade (consolidado)</Heading>
            <Box
              borderWidth="hairline"
              borderStyle="solid"
              borderColor="border.default"
              borderRadius="md"
              bg="bg.surface"
              overflow="hidden"
            >
              {data.units.map((u) => (
                <Flex
                  key={u.unit_id}
                  justifyContent="space-between"
                  px={4}
                  py={3}
                  borderBottom="1px solid"
                  borderBottomColor="border.subtle"
                  gap={3}
                >
                  <Text fontWeight="medium">{u.unit_name}</Text>
                  <Text muted fontSize="sm">
                    abertas {u.open} · fechadas {u.closed} · conversão {u.conversion_rate}%
                  </Text>
                </Flex>
              ))}
            </Box>
          </Stack>
        </>
      )}
    </Stack>
  );
}
