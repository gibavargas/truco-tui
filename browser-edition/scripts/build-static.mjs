import { readFileSync } from "node:fs";
import { mkdir, cp, readFile, writeFile, rename } from "node:fs/promises";
import { createHash } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as esbuild from "esbuild";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const browserEditionDir = path.resolve(__dirname, "..");
const rootDir = path.resolve(browserEditionDir, "..");
const outdir = process.env.TRUCO_BROWSER_OUTDIR || path.join(browserEditionDir, "dist");

await mkdir(outdir, { recursive: true });
await mkdir(path.join(outdir, "assets"), { recursive: true });

await esbuild.build({
  entryPoints: [path.join(browserEditionDir, "web", "main.ts")],
  bundle: true,
  format: "esm",
  target: ["es2022"],
  minify: true,
  sourcemap: false,
  outfile: path.join(outdir, "assets", "app.js"),
});

await esbuild.build({
  entryPoints: [path.join(browserEditionDir, "web", "styles.css")],
  bundle: true,
  minify: true,
  sourcemap: false,
  outfile: path.join(outdir, "assets", "app.css"),
});

// Compute content hashes for cache-busting
function fileHash(filePath) {
  const buf = readFileSync(filePath);
  return createHash("md5").update(buf).digest("hex").slice(0, 8);
}

const jsHash = fileHash(path.join(outdir, "assets", "app.js"));
const cssHash = fileHash(path.join(outdir, "assets", "app.css"));

const jsName = `app.${jsHash}.js`;
const cssName = `app.${cssHash}.css`;

await rename(path.join(outdir, "assets", "app.js"), path.join(outdir, "assets", jsName));
await rename(path.join(outdir, "assets", "app.css"), path.join(outdir, "assets", cssName));

// Inject hashed filenames into index.html
let html = await readFile(path.join(browserEditionDir, "web", "index.html"), "utf8");
html = html.replace("assets/app.css", `assets/${cssName}`);
html = html.replace("assets/app.js", `assets/${jsName}`);
await writeFile(path.join(outdir, "index.html"), html, "utf8");

const publicDir = path.join(browserEditionDir, "public");
await cp(publicDir, outdir, { recursive: true });

console.log(`Browser frontend bundled to ${path.relative(rootDir, outdir)}`);
console.log(`  CSS: ${cssName}`);
console.log(`  JS:  ${jsName}`);
