import "styled-components";
import type { WebLightTheme } from "@cia-da-vacina/design-system-tokens";

declare module "styled-components" {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  export interface DefaultTheme extends WebLightTheme {}
}
