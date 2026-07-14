const STYLE_PROPS = new Set([
  "m", "mt", "mr", "mb", "ml", "mx", "my",
  "p", "pt", "pr", "pb", "pl", "px", "py",
  "margin", "marginTop", "marginRight", "marginBottom", "marginLeft",
  "padding", "paddingTop", "paddingRight", "paddingBottom", "paddingLeft",
  "gap", "rowGap", "columnGap",
  "color", "bg", "backgroundColor", "opacity", "fill", "stroke",
  "width", "height", "minWidth", "maxWidth", "minHeight", "maxHeight", "size",
  "display", "verticalAlign", "overflow", "overflowX", "overflowY",
  "alignItems", "alignContent", "justifyItems", "justifyContent",
  "flexWrap", "flexDirection", "flex", "flexGrow", "flexShrink", "flexBasis",
  "justifySelf", "alignSelf", "order",
  "border", "borderWidth", "borderStyle", "borderColor", "borderRadius",
  "borderTop", "borderRight", "borderBottom", "borderLeft",
  "borderTopWidth", "borderTopStyle", "borderTopColor",
  "borderRightWidth", "borderRightStyle", "borderRightColor",
  "borderBottomWidth", "borderBottomStyle", "borderBottomColor",
  "borderLeftWidth", "borderLeftStyle", "borderLeftColor",
  "position", "zIndex", "top", "right", "bottom", "left",
  "fontFamily", "fontSize", "fontWeight", "lineHeight", "letterSpacing",
  "textAlign", "fontStyle", "textTransform", "textDecoration",
  "boxShadow", "textShadow",
  "variant", "cursor",
]);

export function shouldForwardProp(prop: string): boolean {
  return !STYLE_PROPS.has(prop);
}

export const shouldForwardPropDiv = shouldForwardProp;
