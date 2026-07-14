import styled from "styled-components";
import type { PipelineStage } from "@cia-da-vacina/design-system-tokens";
import { BotIcon, UserIcon } from "@cia-da-vacina/icon-system";
import Avatar from "../DataDisplay/Avatar";
import { Badge, StageBadge } from "../DataDisplay/Badge";
import Flex from "../Layout/Flex";
import Stack from "../Layout/Stack";
import Text from "../Typography/Text";

export type ConversationRowProps = {
  contactName: string;
  preview: string;
  stage: PipelineStage;
  mode: "ai_triage" | "human";
  aiSummary?: string | null;
  timestamp: string;
  onClick?: () => void;
  selected?: boolean;
};

const Row = styled.button<{ $selected?: boolean }>`
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: ${({ theme }) => theme.space[3]};
  padding: ${({ theme }) => `${theme.space[3]} ${theme.space[4]}`};
  border: 0;
  border-bottom: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  background: ${({ theme, $selected }) =>
    $selected ? theme.colors["bg.brand.subtle"] : theme.colors["bg.surface"]};
  text-align: left;
  cursor: pointer;
  transition: background 120ms ease;

  &:hover {
    background: ${({ theme }) => theme.colors["bg.surface.muted"]};
  }

  &:last-child {
    border-bottom: 0;
  }
`;

export default function ConversationRow({
  contactName,
  preview,
  stage,
  mode,
  aiSummary,
  timestamp,
  onClick,
  selected,
}: ConversationRowProps) {
  return (
    <Row type="button" onClick={onClick} $selected={selected}>
      <Flex gap={2} alignItems="center" minWidth={0} flex={1}>
        <Avatar name={contactName} size={32} />
        <Stack gap={1} minWidth={0} flex={1}>
          <Flex alignItems="center" gap={1} flexWrap="wrap">
            <Text fontWeight="semibold" fontSize="sm">
              {contactName}
            </Text>
            <StageBadge stage={stage} />
            <Badge tone={mode === "ai_triage" ? "ai" : "human"}>
              {mode === "ai_triage" ? (
                <>
                  <BotIcon size="xs" /> IA
                </>
              ) : (
                <>
                  <UserIcon size="xs" /> Humano
                </>
              )}
            </Badge>
          </Flex>
          <Text
            muted
            fontSize="sm"
            style={{
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {preview}
          </Text>
          {aiSummary && (
            <Text fontSize="xs" color="text.muted">
              Resumo IA: {aiSummary}
            </Text>
          )}
        </Stack>
      </Flex>
      <Text fontSize="xs" color="text.muted" style={{ flexShrink: 0 }}>
        {timestamp}
      </Text>
    </Row>
  );
}
