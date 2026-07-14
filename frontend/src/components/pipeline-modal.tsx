"use client";

import { useQuery } from "@tanstack/react-query";
import { useState, type MouseEvent } from "react";
import {
  Box,
  Button,
  Flex,
  Heading,
  SelectField,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { api } from "@/lib/api";
import { PIPELINE_STAGES, STAGE_LABELS, type PipelineStage } from "@/lib/types";

type Props = {
  open: boolean;
  current: PipelineStage;
  onClose: () => void;
  onConfirm: (payload: {
    stage: PipelineStage;
    reason_code?: string;
    reason_text?: string;
  }) => Promise<void>;
};

export function PipelineModal({ open, current, onClose, onConfirm }: Props) {
  const [stage, setStage] = useState<PipelineStage>(current);
  const [reasonCode, setReasonCode] = useState("");
  const [reasonText, setReasonText] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: reasons } = useQuery({
    queryKey: ["loss-reasons"],
    queryFn: () => api.listLossReasons(),
    enabled: open,
  });

  if (!open) return null;

  return (
    <Box
      position="fixed"
      top={0}
      left={0}
      right={0}
      bottom={0}
      style={{
        background: "rgba(8,43,35,0.45)",
        zIndex: 400,
        display: "grid",
        placeItems: "center",
        padding: 16,
      }}
      onClick={onClose}
    >
      <Box
        bg="bg.surface"
        borderRadius="lg"
        p={4}
        borderWidth="hairline"
        borderStyle="solid"
        borderColor="border.default"
        width="100%"
        maxWidth="440px"
        boxShadow="lg"
        onClick={(e: MouseEvent) => e.stopPropagation()}
      >
        <Stack gap={4}>
          <Heading as="h3">Mover no pipeline</Heading>
          <SelectField
            label="Etapa"
            value={stage}
            onChange={(e) => setStage(e.target.value as PipelineStage)}
            options={PIPELINE_STAGES.map((s) => ({
              value: s,
              label: STAGE_LABELS[s],
            }))}
          />
          {stage === "nao_fechado" && (
            <>
              <SelectField
                label="Motivo da não conversão"
                value={reasonCode}
                onChange={(e) => setReasonCode(e.target.value)}
                options={[
                  { value: "", label: "Selecione…" },
                  ...(reasons?.items ?? []).map((r) => ({
                    value: r.code,
                    label: r.label,
                  })),
                ]}
              />
              <TextField
                label="Detalhe (opcional)"
                value={reasonText}
                onChange={(e) => setReasonText(e.target.value)}
              />
            </>
          )}
          {error && <Text color="text.danger">{error}</Text>}
          <Flex gap={2} justifyContent="flex-end">
            <Button variant="secondary" onClick={onClose}>
              Cancelar
            </Button>
            <Button
              disabled={saving}
              onClick={async () => {
                setError(null);
                setSaving(true);
                try {
                  await onConfirm({
                    stage,
                    reason_code: reasonCode || undefined,
                    reason_text: reasonText || undefined,
                  });
                  onClose();
                } catch (e) {
                  setError(e instanceof Error ? e.message : "Falha ao salvar");
                } finally {
                  setSaving(false);
                }
              }}
            >
              {saving ? "Salvando…" : "Confirmar"}
            </Button>
          </Flex>
        </Stack>
      </Box>
    </Box>
  );
}
