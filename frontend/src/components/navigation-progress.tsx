"use client";

import styled, { keyframes } from "styled-components";
import { useNavigationFeedback } from "@/providers/navigation-provider";

const slide = keyframes`
  0% {
    transform: translateX(-40%) scaleX(0.35);
    opacity: 0.55;
  }
  50% {
    transform: translateX(20%) scaleX(0.55);
    opacity: 1;
  }
  100% {
    transform: translateX(110%) scaleX(0.35);
    opacity: 0.55;
  }
`;

const Track = styled.div<{ $visible: boolean }>`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 500;
  height: 2px;
  pointer-events: none;
  overflow: hidden;
  opacity: ${({ $visible }) => ($visible ? 1 : 0)};
  transition: opacity 160ms ease;
  transition-delay: ${({ $visible }) => ($visible ? "90ms" : "0ms")};
  background: transparent;
`;

const Bar = styled.div`
  height: 100%;
  width: 40%;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    rgba(15, 107, 76, 0) 0%,
    rgba(15, 107, 76, 0.95) 45%,
    rgba(56, 178, 120, 0.95) 100%
  );
  box-shadow: 0 0 12px rgba(15, 107, 76, 0.45);
  animation: ${slide} 1.05s cubic-bezier(0.32, 0.72, 0, 1) infinite;
  transform-origin: left center;
`;

/** Thin top progress cue while client-side route transitions are in flight. */
export function NavigationProgress() {
  const { isNavigating } = useNavigationFeedback();
  return (
    <Track $visible={isNavigating} aria-hidden={!isNavigating}>
      <Bar />
    </Track>
  );
}
