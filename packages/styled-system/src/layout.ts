import { system } from "./core";

export const layout = system({
  width: { property: "width", scale: "sizes" },
  height: { property: "height", scale: "sizes" },
  minWidth: { property: "minWidth", scale: "sizes" },
  maxWidth: { property: "maxWidth", scale: "sizes" },
  minHeight: { property: "minHeight", scale: "sizes" },
  maxHeight: { property: "maxHeight", scale: "sizes" },
  size: { property: ["width", "height"], scale: "sizes" },
  display: true,
  verticalAlign: true,
  overflow: true,
  overflowX: true,
  overflowY: true,
});

export type LayoutProps = {
  width?: number | string;
  height?: number | string;
  minWidth?: number | string;
  maxWidth?: number | string;
  minHeight?: number | string;
  maxHeight?: number | string;
  size?: number | string;
  display?: string;
  verticalAlign?: string;
  overflow?: string;
  overflowX?: string;
  overflowY?: string;
};
