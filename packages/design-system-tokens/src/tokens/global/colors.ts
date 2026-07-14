/**
 * Raw palette — Cia da Vacina
 * Direction: clinical trust (deep evergreen) + soft clinic light,
 * warm amber for pipeline urgency, coral for loss/alert.
 * Avoids purple SaaS / cream-terracotta / neon glow clichés.
 */
const colors = {
  base: {
    white: "#FFFFFF",
    black: "#0B1210",
  },
  evergreen: {
    25: "#F3FBF6",
    50: "#E7F6EE",
    100: "#CDEADB",
    200: "#9DD4B8",
    300: "#6BB892",
    400: "#3D9A72",
    500: "#1F7F58",
    600: "#0F6B4C",
    700: "#0B5540",
    800: "#0A4032",
    900: "#082B23",
    950: "#041915",
  },
  mist: {
    25: "#F7FAF8",
    50: "#F0F5F2",
    100: "#E2EBE6",
    200: "#C8D6CE",
    300: "#A8B9B0",
    400: "#84978C",
    500: "#677A70",
    600: "#4F6158",
    700: "#3C4B44",
    800: "#2A3631",
    900: "#1B2420",
    950: "#101612",
  },
  sand: {
    50: "#FBF7F0",
    100: "#F3EBDD",
    200: "#E6D5B8",
    300: "#D4B98A",
    400: "#C19A5C",
    500: "#A97B3A",
  },
  amber: {
    50: "#FFF8EB",
    100: "#FEEFC6",
    200: "#FBD889",
    300: "#F5BC4B",
    400: "#E99A1A",
    500: "#C8780D",
    600: "#9A5A0A",
    700: "#73440C",
  },
  coral: {
    50: "#FFF1F0",
    100: "#FFE0DD",
    200: "#FFC5BF",
    300: "#F9958C",
    400: "#E85F52",
    500: "#D04032",
    600: "#B12E22",
    700: "#8F241B",
  },
  sky: {
    50: "#F0F7FC",
    100: "#DCECF8",
    200: "#B7D7F0",
    300: "#84B8E3",
    400: "#4A94CE",
    500: "#2D78B5",
    600: "#1F5F94",
  },
} as const;

export default colors;
