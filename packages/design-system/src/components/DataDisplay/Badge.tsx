import styled, { css } from "styled-components";
import type { ReactNode } from "react";
import {
  stageLabels,
  type PipelineStage,
} from "@cia-da-vacina/design-system-tokens";

export type BadgeTone =
  | "neutral"
  | "brand"
  | "success"
  | "warning"
  | "danger"
  | "info"
  | "ai"
  | "human";

const toneStyles: Record<BadgeTone, ReturnType<typeof css>> = {
  neutral: css`
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
    color: ${({ theme }) => theme.colors["text.secondary"]};
  `,
  brand: css`
    background: ${({ theme }) => theme.colors["bg.brand.subtle"]};
    color: ${({ theme }) => theme.colors["text.brand"]};
  `,
  success: css`
    background: ${({ theme }) => theme.colors["stage.fechado.bg"]};
    color: ${({ theme }) => theme.colors["stage.fechado.text"]};
  `,
  warning: css`
    background: ${({ theme }) => theme.colors["bg.warning.subtle"]};
    color: ${({ theme }) => theme.colors["text.warning"]};
  `,
  danger: css`
    background: ${({ theme }) => theme.colors["bg.danger.subtle"]};
    color: ${({ theme }) => theme.colors["text.danger"]};
  `,
  info: css`
    background: ${({ theme }) => theme.colors["bg.info.subtle"]};
    color: ${({ theme }) => theme.colors["stage.em_atendimento.text"]};
  `,
  ai: css`
    background: ${({ theme }) => theme.colors["mode.ai.bg"]};
    color: ${({ theme }) => theme.colors["mode.ai.text"]};
  `,
  human: css`
    background: ${({ theme }) => theme.colors["mode.human.bg"]};
    color: ${({ theme }) => theme.colors["mode.human.text"]};
  `,
};

const stageTone: Record<PipelineStage, BadgeTone> = {
  em_atendimento: "info",
  em_negociacao: "brand",
  aguardando_fechamento: "warning",
  fechado: "success",
  nao_fechado: "danger",
};

const Pill = styled.span<{ $tone: BadgeTone }>`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: ${({ theme }) => theme.radii.sm};
  font-size: ${({ theme }) => theme.fontSizes.xs};
  font-weight: ${({ theme }) => theme.fontWeights.medium};
  line-height: ${({ theme }) => theme.lineHeights.snug};
  white-space: nowrap;
  ${({ $tone }) => toneStyles[$tone]}

  & > svg {
    width: 12px !important;
    height: 12px !important;
    max-width: 12px !important;
    max-height: 12px !important;
    flex-shrink: 0;
  }
`;

export type BadgeProps = {
  tone?: BadgeTone;
  children: ReactNode;
};

export function Badge({ tone = "neutral", children }: BadgeProps) {
  return <Pill $tone={tone}>{children}</Pill>;
}

export type StageBadgeProps = {
  stage: PipelineStage;
};

export function StageBadge({ stage }: StageBadgeProps) {
  return <Badge tone={stageTone[stage]}>{stageLabels[stage]}</Badge>;
}

