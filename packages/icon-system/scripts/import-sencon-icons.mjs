/**
 * Converts Sencon Icon System (.tsx per icon) into @cia-da-vacina/icon-system
 * createIcon modules + gallery data + barrel exports.
 *
 * Usage: node scripts/import-sencon-icons.mjs
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const SENCON =
  process.env.SENCON_ICON_DIR ||
  "E:\\Projetos\\SenconIconSystem\\src\\components\\Icon";
const OUT_DIR = path.join(ROOT, "src", "icons", "generated");
const GALLERY_DATA = path.join(ROOT, "gallery", "icons-data.js");
const BARREL = path.join(ROOT, "src", "icons", "generated", "index.ts");

const ATTR_MAP = {
  "fill-rule": "fillRule",
  "clip-rule": "clipRule",
  "stroke-width": "strokeWidth",
  "stroke-linecap": "strokeLinecap",
  "stroke-linejoin": "strokeLinejoin",
  "stroke-miterlimit": "strokeMiterlimit",
  "stroke-dasharray": "strokeDasharray",
  "stroke-dashoffset": "strokeDashoffset",
  "stroke-opacity": "strokeOpacity",
  "fill-opacity": "fillOpacity",
  "clip-path": "clipPath",
  "font-size": "fontSize",
  "font-family": "fontFamily",
  "font-weight": "fontWeight",
  "text-anchor": "textAnchor",
  class: "className",
  "xlink:href": "xlinkHref",
};

function toJsxAttrs(attrs) {
  return attrs
    .replace(/\s([a-zA-Z_:][-a-zA-Z0-9_:.]*)=(["'])(.*?)\2/g, (_, name, q, val) => {
      const lower = name.toLowerCase();
      if (
        lower === "stroke" &&
        (val === "inherit" || val === "currentColor")
      ) {
        return "";
      }
      if (
        lower === "fill" &&
        (val === "inherit" || val === "currentColor" || val === "none")
      ) {
        // keep fill="none" as explicit none for compound shapes; inherit → drop
        if (val === "none") return ` fill="none"`;
        return "";
      }
      if (lower === "stroke-width" || lower === "strokewidth") {
        return ""; // parent Svg controls weight
      }
      if (lower === "stroke-linecap" || lower === "strokelinecap") return "";
      if (lower === "stroke-linejoin" || lower === "strokelinejoin") return "";
      const mapped = ATTR_MAP[name] || ATTR_MAP[lower] || name;
      // boolean-ish
      if (val === "true" || val === "false") {
        return ` ${mapped}={${val}}`;
      }
      // numeric
      if (/^-?\d+(\.\d+)?$/.test(val)) {
        return ` ${mapped}={${val}}`;
      }
      const escaped = val.replace(/\\/g, "\\\\").replace(/`/g, "\\`");
      return ` ${mapped}="${escaped}"`;
    })
    .replace(/\s+/g, " ")
    .trim();
}

function htmlInnerToJsx(inner) {
  // self-closing void-ish tags already as <path .../>
  return inner
    .replace(/<!--[\s\S]*?-->/g, "")
    .replace(
      /<([a-zA-Z][a-zA-Z0-9]*)([^>]*)\/>/g,
      (_, tag, attrs) => `<${tag} ${toJsxAttrs(attrs)} />`,
    )
    .replace(
      /<([a-zA-Z][a-zA-Z0-9]*)([^>]*)>([\s\S]*?)<\/\1>/g,
      (_, tag, attrs, children) => {
        const jsxChildren = htmlInnerToJsx(children);
        const a = toJsxAttrs(attrs);
        return a
          ? `<${tag} ${a}>${jsxChildren}</${tag}>`
          : `<${tag}>${jsxChildren}</${tag}>`;
      },
    )
    .trim();
}

function extractSvg(source) {
  const m = source.match(/<svg\b([^>]*)>([\s\S]*?)<\/svg>/i);
  if (!m) return null;
  const attrStr = m[1];
  const viewBox =
    (attrStr.match(/viewBox=(["'])(.*?)\1/i) || [])[2] || "0 0 24 24";
  const inner = m[2].trim();
  return { viewBox, inner, jsx: htmlInnerToJsx(inner) };
}

function isValidName(name) {
  return /^[A-Z][A-Za-z0-9]*$/.test(name);
}

fs.rmSync(OUT_DIR, { recursive: true, force: true });
fs.mkdirSync(OUT_DIR, { recursive: true });

const files = fs
  .readdirSync(SENCON)
  .filter((f) => f.endsWith(".tsx") && f !== "index.ts" && f !== "index.tsx");

const gallery = [];
const exports = [];
let ok = 0;
let fail = 0;

for (const file of files) {
  const name = path.basename(file, ".tsx");
  if (!isValidName(name)) {
    console.warn("skip invalid name", name);
    fail++;
    continue;
  }
  const full = path.join(SENCON, file);
  // skip Native subtree if somehow present as file
  if (fs.statSync(full).isDirectory()) continue;

  const source = fs.readFileSync(full, "utf8");
  const svg = extractSvg(source);
  if (!svg || !svg.jsx) {
    console.warn("no svg", name);
    fail++;
    continue;
  }

  const code = `import { createIcon } from "../../createIcon";

export const ${name} = createIcon(
  "${name}",
  <>
    ${svg.jsx}
  </>,
);
`;

  fs.writeFileSync(path.join(OUT_DIR, `${name}.tsx`), code);

  // gallery paths as raw HTML (browser SVG)
  const galleryInner = svg.inner
    .replace(/\sstroke="inherit"/gi, "")
    .replace(/\sfill="inherit"/gi, "")
    .replace(/\sstrokeWidth="2"/gi, "")
    .replace(/\sstroke-width="2"/gi, "");

  gallery.push({ name, viewBox: svg.viewBox, paths: galleryInner });
  exports.push(name);
  ok++;
}

exports.sort((a, b) => a.localeCompare(b));
gallery.sort((a, b) => a.name.localeCompare(b.name));

const barrel = `${exports
  .map((n) => `export { ${n} } from "./${n}";`)
  .join("\n")}\n`;
fs.writeFileSync(BARREL, barrel);

fs.writeFileSync(
  GALLERY_DATA,
  `/* auto-generated — do not edit */\nwindow.CIA_ICONS = ${JSON.stringify(gallery)};\n`,
);

console.log(`imported ${ok} icons (${fail} failed) → ${OUT_DIR}`);
console.log(`gallery data → ${GALLERY_DATA}`);
