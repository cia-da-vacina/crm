import type { ButtonHTMLAttributes, ReactNode } from "react";
import styled from "styled-components";

const List = styled.div`
  display: flex;
  flex-direction: column;
  border: 1px solid ${({ theme }) => theme.colors["border.default"]};
  border-radius: ${({ theme }) => theme.radii.md};
  background: ${({ theme }) => theme.colors["bg.surface"]};
  overflow: hidden;

  /* Works with direct rows or wrappers (e.g. motion.div around DataListRow). */
  & > * + * {
    border-top: 1px solid ${({ theme }) => theme.colors["border.default"]};
  }
`;

const Row = styled.button<{ $interactive?: boolean }>`
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  width: 100%;
  margin: 0;
  padding: 12px 14px;
  border: 0;
  background: transparent;
  text-align: left;
  font: inherit;
  color: inherit;
  cursor: ${({ $interactive }) => ($interactive ? "pointer" : "default")};
  transition: background 140ms ease, transform 120ms ease;

  ${({ $interactive, theme }) =>
    $interactive
      ? `
    &:hover {
      background: ${theme.colors["bg.surface.muted"]};
    }
    &:active {
      transform: scale(0.995);
    }
    &:focus-visible {
      outline: none;
      box-shadow: inset 0 0 0 2px ${theme.colors["border.focus"]};
    }
  `
      : ""}
`;

const Leading = styled.div`
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
`;

const Trailing = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
`;

export type DataListProps = {
  children: ReactNode;
};

export function DataList({ children }: DataListProps) {
  return <List role="list">{children}</List>;
}

export type DataListRowProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  leading: ReactNode;
  trailing?: ReactNode;
  interactive?: boolean;
};

export function DataListRow({
  leading,
  trailing,
  interactive = true,
  type = "button",
  ...rest
}: DataListRowProps) {
  return (
    <Row
      as={interactive ? "button" : "div"}
      type={interactive ? type : undefined}
      $interactive={interactive}
      role="listitem"
      {...rest}
    >
      <Leading>{leading}</Leading>
      {trailing ? <Trailing>{trailing}</Trailing> : null}
    </Row>
  );
}
