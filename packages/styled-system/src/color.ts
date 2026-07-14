import { system } from "./core";

export const color = system({
  color: { property: "color", scale: "colors" },
  bg: { property: "backgroundColor", scale: "colors" },
  backgroundColor: { property: "backgroundColor", scale: "colors" },
  opacity: true,
  fill: { property: "fill", scale: "colors" },
  stroke: { property: "stroke", scale: "colors" },
});

export type ColorProps = {
  color?: string;
  bg?: string;
  backgroundColor?: string;
  opacity?: number | string;
  fill?: string;
  stroke?: string;
};
