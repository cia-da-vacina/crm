import styled from "styled-components";
import type { ReactNode } from "react";
import Box from "../Layout/Box";

const Panel = styled(Box)`
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  overflow: hidden;
  width: 100%;
`;

export default function ConversationList({ children }: { children: ReactNode }) {
  return <Panel>{children}</Panel>;
}
