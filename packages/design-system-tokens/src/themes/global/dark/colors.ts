import untitledSemantic from "./colors/index";
import crmOverlay from "./crmOverlay";
import raw from "../../../tokens/global/colors";

function flattenRaw(palette: typeof raw): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [group, scale] of Object.entries(palette)) {
    if (typeof scale === "string") {
      out[group] = scale;
      continue;
    }
    for (const [step, value] of Object.entries(scale as Record<string, string>)) {
      out[`${group}.${step}`] = value;
    }
  }
  return out;
}

/**
 * Final semantic map: Untitled UI + Cia CRM overlay + raw flatten.
 * Overlay wins on conflicting keys (e.g. bg.brand.solid → evergreen).
 */
const colors = {
  ...untitledSemantic,
  ...crmOverlay,
  ...flattenRaw(raw),
} as const;

export default colors;
