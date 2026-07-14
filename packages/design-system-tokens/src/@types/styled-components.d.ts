import "styled-components";
import type { WebLightTheme } from "../themes/web/light";

declare module "styled-components" {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  export interface DefaultTheme extends WebLightTheme {}
}
