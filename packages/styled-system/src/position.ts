import { system } from "./core";

export const position = system({
  position: true,
  zIndex: { property: "zIndex", scale: "zIndices" },
  top: { property: "top", scale: "space" },
  right: { property: "right", scale: "space" },
  bottom: { property: "bottom", scale: "space" },
  left: { property: "left", scale: "space" },
});

export type PositionProps = {
  position?: string;
  zIndex?: number | string;
  top?: number | string;
  right?: number | string;
  bottom?: number | string;
  left?: number | string;
};
