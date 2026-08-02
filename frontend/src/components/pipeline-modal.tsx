"use client";

import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import {
  Button,
  Flex,
  Heading,
  SelectField,
  Stack,
  Text,
  TextField,
} from "@cia-da-vacina/design-system";
import { ModalShell } from "@/components/modal-shell";
import { lossReasonsService } from "@/services";
import { PIPELINE_STAGES, STAGE_LABELS, type PipelineStage } from "@/domain";

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
    queryFn: () => lossReasonsService.list(),
    enabled: open,
  });

  useEffect(() => {
    if (!open) return;
    setStage(current);
    setReasonCode("");
    setReasonText("");
    setError(null);
  }, [open, current]);

  return (
    <ModalShell open={open} onClose={onClose} closeOnBackdrop={!saving}>
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
                ...(reasons ?? []).map((r) => ({
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
          <Button variant="secondary" disabled={saving} onClick={onClose}>
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
    </ModalShell>
  );
}
