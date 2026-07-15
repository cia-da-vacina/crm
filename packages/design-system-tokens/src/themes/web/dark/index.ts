import dark from "../../global/dark";
import borderWidths from "../../../tokens/global/borderWidths";
import fontWeights from "../../../tokens/global/fontWeights";
import lineHeights from "../../../tokens/global/lineHeights";
import breakpoints from "../../../tokens/web/breakpoints";
import fonts from "../../../tokens/web/fonts";
import fontSizes from "../../../tokens/web/fontSizes";
import letterSpacings from "../../../tokens/web/letterSpacings";
import mediaQueries from "../../../tokens/web/mediaQueries";
import radii from "../../../tokens/web/radii";
import space from "../../../tokens/web/space";
import zIndices from "../../../tokens/web/zIndices";
import shadows from "../../../tokens/web/shadows";
import iconSizes from "../../../tokens/web/iconSizes";

const webDark = {
  name: dark.name,
  colors: dark.colors,
  borderWidths,
  breakpoints,
  fonts,
  fontSizes,
  fontWeights,
  letterSpacings,
  lineHeights,
  mediaQueries,
  radii,
  shadows,
  space,
  sizes: space,
  iconSizes,
  zIndices,
};

export type WebDarkTheme = typeof webDark;
export default webDark;
