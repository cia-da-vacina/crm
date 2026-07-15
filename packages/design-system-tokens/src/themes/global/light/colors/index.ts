import background from "./background";
import border from "./border";
import button from "./button";
import foreground from "./foreground";
import icons from "./icons";
import text from "./text";
import utility from "./utility";

export default {
  inherit: `inherit`,
  transparent: `transparent`,
  ...background,
  ...text,
  ...border,
  ...foreground,
  ...button,
  ...icons,
  ...utility,
};
