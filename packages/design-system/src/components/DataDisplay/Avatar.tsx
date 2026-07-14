import styled from "styled-components";

export type AvatarProps = {
  name: string;
  size?: number;
  /** Optional photo URL; falls back to initials. */
  src?: string | null;
  alt?: string;
};

const Circle = styled.div<{ $size: number }>`
  width: ${({ $size }) => $size}px;
  height: ${({ $size }) => $size}px;
  border-radius: ${({ theme }) => theme.radii.round};
  background: ${({ theme }) => theme.colors["bg.brand.subtle"]};
  color: ${({ theme }) => theme.colors["text.brand"]};
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: ${({ $size }) => Math.max(11, Math.round($size * 0.38))}px;
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: inset 0 0 0 1px ${({ theme }) => theme.colors["border.subtle"]};
`;

const Photo = styled.img`
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
`;

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("");
}

export default function Avatar({
  name,
  size = 36,
  src,
  alt,
}: AvatarProps) {
  return (
    <Circle $size={size} aria-hidden={!alt}>
      {src ? <Photo src={src} alt={alt ?? name} /> : initials(name)}
    </Circle>
  );
}
