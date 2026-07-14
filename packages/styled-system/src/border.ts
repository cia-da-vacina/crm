import { system } from "./core";

export const border = system({
  border: true,
  borderWidth: { property: "borderWidth", scale: "borderWidths" },
  borderStyle: true,
  borderColor: { property: "borderColor", scale: "colors" },
  borderRadius: { property: "borderRadius", scale: "radii" },
  borderTop: true,
  borderRight: true,
  borderBottom: true,
  borderLeft: true,
  borderTopWidth: { property: "borderTopWidth", scale: "borderWidths" },
  borderTopStyle: true,
  borderTopColor: { property: "borderTopColor", scale: "colors" },
  borderRightWidth: { property: "borderRightWidth", scale: "borderWidths" },
  borderRightStyle: true,
  borderRightColor: { property: "borderRightColor", scale: "colors" },
  borderBottomWidth: { property: "borderBottomWidth", scale: "borderWidths" },
  borderBottomStyle: true,
  borderBottomColor: { property: "borderBottomColor", scale: "colors" },
  borderLeftWidth: { property: "borderLeftWidth", scale: "borderWidths" },
  borderLeftStyle: true,
  borderLeftColor: { property: "borderLeftColor", scale: "colors" },
});

export type BorderProps = {
  border?: string;
  borderWidth?: number | string;
  borderStyle?: string;
  borderColor?: string;
  borderRadius?: number | string;
  borderTop?: string;
  borderRight?: string;
  borderBottom?: string;
  borderLeft?: string;
  borderTopWidth?: number | string;
  borderTopStyle?: string;
  borderTopColor?: string;
  borderRightWidth?: number | string;
  borderRightStyle?: string;
  borderRightColor?: string;
  borderBottomWidth?: number | string;
  borderBottomStyle?: string;
  borderBottomColor?: string;
  borderLeftWidth?: number | string;
  borderLeftStyle?: string;
  borderLeftColor?: string;
};
