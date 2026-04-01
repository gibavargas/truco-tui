import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as esbuild from "esbuild";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const wailsDir = path.resolve(__dirname, "..");
const frontendDir = path.join(wailsDir, "frontend");
const outdir = path.join(frontendDir, "dist");
const watch = process.argv.includes("--watch");

const jsBuild = {
  entryPoints: [path.join(frontendDir, "src", "main.ts")],
  bundle: true,
  format: "esm",
  target: ["es2022"],
  minify: !watch,
  sourcemap: watch,
  outfile: path.join(outdir, "assets", "app.js"),
};

const cssBuild = {
  entryPoints: [path.join(frontendDir, "src", "styles.css")],
  bundle: true,
  minify: !watch,
  sourcemap: watch,
  outfile: path.join(outdir, "assets", "app.css"),
};

await mkdir(path.join(outdir, "assets"), { recursive: true });
await syncStatic();

if (watch) {
  const jsContext = await esbuild.context(jsBuild);
  const cssContext = await esbuild.context(cssBuild);
  await Promise.all([jsContext.watch(), cssContext.watch()]);
  console.log("Wails frontend watcher ready");
  process.stdin.resume();
} else {
  await esbuild.build(jsBuild);
  await esbuild.build(cssBuild);
  console.log(`Wails frontend bundled to ${path.relative(wailsDir, outdir)}`);
}

async function syncStatic() {
  await rm(outdir, { recursive: true, force: true });
  await mkdir(path.join(outdir, "assets"), { recursive: true });

  const html = await readFile(path.join(frontendDir, "index.html"), "utf8");
  await writeFile(path.join(outdir, "index.html"), html, "utf8");
  await cp(path.join(frontendDir, "public"), outdir, { recursive: true });
}
