"use client";

import Link from "next/link";
import styled from "styled-components";
import { useNavigationFeedback } from "@/providers/navigation-provider";

const BrandLink = styled(Link)`
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
  min-width: 0;
  padding: 6px 8px 14px;
  text-decoration: none;
  color: inherit;
`;

const BrandMark = styled.img`
  width: 40px;
  height: 40px;
  object-fit: contain;
  flex-shrink: 0;
  border-radius: ${({ theme }) => theme.radii.md};
`;

const BrandText = styled.div`
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
`;

const BrandTitle = styled.div`
  font-family: ${({ theme }) => theme.fonts.display};
  font-size: ${({ theme }) => theme.fontSizes.md};
  font-weight: ${({ theme }) => theme.fontWeights.semibold};
  color: ${({ theme }) => theme.colors["text.brand"]};
  letter-spacing: ${({ theme }) => theme.letterSpacings.tight};
  line-height: 1.15;
`;

const BrandSub = styled.span`
  font-size: 11px;
  font-weight: ${({ theme }) => theme.fontWeights.bold};
  letter-spacing: 0.08em;
  text-transform: uppercase;
  line-height: 1.2;
  color: ${({ theme }) => theme.colors["text.brand"]};
`;

type Props = {
  href: string;
  label: string;
  subLabel?: string;
  logoSrc?: string;
};

export function AppShellBrandLink({ href, label, subLabel, logoSrc }: Props) {
  const { begin } = useNavigationFeedback();
  return (
    <BrandLink
      href={href}
      prefetch
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) return;
        begin(href);
      }}
    >
      {logoSrc ? <BrandMark src={logoSrc} alt="" /> : null}
      <BrandText>
        <BrandTitle>{label}</BrandTitle>
        {subLabel ? <BrandSub>{subLabel}</BrandSub> : null}
      </BrandText>
    </BrandLink>
  );
}
