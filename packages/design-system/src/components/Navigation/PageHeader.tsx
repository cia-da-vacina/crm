import type { ReactNode } from "react";
import styled from "styled-components";
import Heading from "../Typography/Heading";
import Text from "../Typography/Text";
import Flex from "../Layout/Flex";
import Stack from "../Layout/Stack";

const Wrap = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px 16px;
  margin-bottom: 20px;
`;

const Actions = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
`;

export type PageHeaderProps = {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  eyebrow?: ReactNode;
};

export default function PageHeader({
  title,
  description,
  actions,
  eyebrow,
}: PageHeaderProps) {
  return (
    <Wrap>
      <Stack gap={1} style={{ minWidth: 0, flex: "1 1 220px" }}>
        {eyebrow ? (
          <Text fontSize="xs" muted style={{ letterSpacing: "0.04em", textTransform: "uppercase" }}>
            {eyebrow}
          </Text>
        ) : null}
        <Heading as="h1" style={{ fontSize: "1.65rem", lineHeight: 1.2 }}>
          {title}
        </Heading>
        {description ? (
          <Text muted fontSize="sm">
            {description}
          </Text>
        ) : null}
      </Stack>
      {actions ? <Actions>{actions}</Actions> : null}
    </Wrap>
  );
}

export { Flex };
