import type { ReactNode } from "react";
import styled from "styled-components";

const Bar = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 12px;
  margin-bottom: 14px;
  padding: 10px 12px;
  background: ${({ theme }) => theme.colors["bg.surface"]};
  border: 1px solid ${({ theme }) => theme.colors["border.subtle"]};
  border-radius: ${({ theme }) => theme.radii.md};
`;

const Slot = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
`;

export type ToolbarProps = {
  leading?: ReactNode;
  trailing?: ReactNode;
  children?: ReactNode;
};

export default function Toolbar({ leading, trailing, children }: ToolbarProps) {
  return (
    <Bar>
      <Slot>{leading ?? children}</Slot>
      {trailing ? <Slot>{trailing}</Slot> : null}
    </Bar>
  );
}
