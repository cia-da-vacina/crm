import type { CSSObject } from "styled-components";

export type Theme = Record<string, unknown>;

// Loose props typing so interpolations work with styled-components generics.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type StyleFn = (props: any) => CSSObject | undefined;

export function get(
  obj: unknown,
  path: string | number,
  fallback?: unknown,
): unknown {
  if (obj == null) return fallback;
  if (typeof path === "number") {
    const arr = obj as Record<number, unknown>;
    return arr[path] !== undefined ? arr[path] : fallback;
  }
  const key = String(path);
  // Prefer exact keys (semantic tokens like "border.default", "bg.brand.solid")
  if (typeof obj === "object" && key in (obj as Record<string, unknown>)) {
    return (obj as Record<string, unknown>)[key];
  }
  const parts = key.split(".");
  let current: unknown = obj;
  for (const part of parts) {
    if (current == null || typeof current !== "object") return fallback;
    current = (current as Record<string, unknown>)[part];
  }
  return current !== undefined ? current : fallback;
}

type PropConfig = {
  property?: string | string[];
  scale?: string;
  transform?: (value: unknown, scale: unknown, props: Record<string, unknown>) => unknown;
  defaultScale?: unknown;
};

type SystemConfig = Record<string, true | PropConfig>;

function toCssValue(value: unknown): string | number {
  if (typeof value === "number" || typeof value === "string") return value;
  return String(value);
}

export function system(config: SystemConfig): StyleFn {
  const entries = Object.entries(config);

  return (props) => {
    const theme = (props.theme ?? {}) as Theme;
    const result: CSSObject = {};

    for (const [prop, conf] of entries) {
      const raw = props[prop];
      if (raw === undefined || raw === null) continue;

      const cfg: PropConfig = conf === true ? { property: prop } : conf;
      const scale = cfg.scale
        ? get(theme, cfg.scale, cfg.defaultScale)
        : undefined;

      const transformed = cfg.transform
        ? cfg.transform(raw, scale, props)
        : get(scale, raw as string | number, raw);

      const properties = Array.isArray(cfg.property)
        ? cfg.property
        : [cfg.property ?? prop];

      for (const property of properties) {
        if (!property) continue;
        (result as Record<string, string | number>)[property] =
          toCssValue(transformed);
      }
    }

    return Object.keys(result).length ? result : undefined;
  };
}

export function compose(...parsers: StyleFn[]): StyleFn {
  return (props) => {
    const out: CSSObject = {};
    for (const parser of parsers) {
      const partial = parser(props);
      if (partial) Object.assign(out, partial);
    }
    return Object.keys(out).length ? out : undefined;
  };
}

export function variant(opts: {
  key?: string;
  prop?: string;
  variants?: Record<string, CSSObject>;
  scale?: string;
}): StyleFn {
  const prop = opts.prop ?? "variant";
  return (props) => {
    const value = props[prop] as string | undefined;
    if (!value) return undefined;
    if (opts.variants?.[value]) return opts.variants[value];
    if (opts.scale) {
      return get(props.theme, `${opts.scale}.${value}`) as CSSObject | undefined;
    }
    if (opts.key) {
      return get(props.theme, `${opts.key}.${value}`) as CSSObject | undefined;
    }
    return undefined;
  };
}
