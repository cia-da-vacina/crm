import styled, { keyframes } from "styled-components";

const spin = keyframes`
  to { transform: rotate(360deg); }
`;

const Ring = styled.div<{ $size: number }>`
  width: ${({ $size }) => $size}px;
  height: ${({ $size }) => $size}px;
  border-radius: 50%;
  border: 2px solid ${({ theme }) => theme.colors["border.default"]};
  border-top-color: ${({ theme }) => theme.colors["bg.brand.solid"]};
  animation: ${spin} 0.7s linear infinite;
`;

export default function Spinner({ size = 20 }: { size?: number }) {
  return <Ring $size={size} role="status" aria-label="Carregando" />;
}
