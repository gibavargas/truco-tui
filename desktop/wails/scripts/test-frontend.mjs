import { build } from "esbuild";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const workdir = process.cwd();
const tempDir = mkdtempSync(path.join(tmpdir(), "truco-wails-frontend-test-"));
const outfile = path.join(tempDir, "runtime-state.test.mjs");

await build({
  entryPoints: [path.join(workdir, "frontend/src/runtime-state.test.ts")],
  outfile,
  bundle: true,
  format: "esm",
  platform: "node",
  sourcemap: "inline",
  target: "node20",
});

const result = spawnSync(process.execPath, ["--test", outfile], {
  stdio: "inherit",
});

process.exit(result.status ?? 1);
