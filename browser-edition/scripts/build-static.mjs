import { mkdir, cp, readFile, writeFile } from "node:fs/promises";
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

const html = await readFile(path.join(browserEditionDir, "web", "index.html"), "utf8");
await writeFile(path.join(outdir, "index.html"), html, "utf8");

const publicDir = path.join(browserEditionDir, "public");
await cp(publicDir, outdir, { recursive: true });

console.log(`Browser frontend bundled to ${path.relative(rootDir, outdir)}`);
