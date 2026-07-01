#!/usr/bin/env node

/**
 * Build script that single-sources the site's headline numbers from code, so
 * copy can never drift into a stale hardcoded count.
 *
 * Reads (from the repo root):
 *   - apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto  -> component count (`(kind_meta)` annotations)
 *   - apis/dev/planton/provider/*                                          -> provider count
 *   - charts/<provider>/<chart>/Chart.yaml                                 -> chart count
 *
 * Writes: src/site/stats.generated.ts
 */

import * as fs from "fs";
import * as path from "path";

const scriptDir = __dirname;
const projectRoot = path.join(scriptDir, "../..");
const siteRoot = path.join(scriptDir, "..");

function countKinds(): number {
  const protoPath = path.join(
    projectRoot,
    "apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto",
  );
  const src = fs.readFileSync(protoPath, "utf-8");
  const matches = src.match(/\(kind_meta\)/g);
  return matches ? matches.length : 0;
}

function countProviders(): number {
  const providerDir = path.join(projectRoot, "apis/dev/planton/provider");
  if (!fs.existsSync(providerDir)) return 0;
  return fs
    .readdirSync(providerDir, { withFileTypes: true })
    .filter((d) => d.isDirectory() && !d.name.startsWith("_") && !d.name.startsWith("."))
    .length;
}

function countCharts(): number {
  const chartsDir = path.join(projectRoot, "charts");
  if (!fs.existsSync(chartsDir)) return 0;
  let count = 0;
  for (const provider of fs.readdirSync(chartsDir, { withFileTypes: true })) {
    if (!provider.isDirectory()) continue;
    const providerPath = path.join(chartsDir, provider.name);
    for (const chart of fs.readdirSync(providerPath, { withFileTypes: true })) {
      if (!chart.isDirectory()) continue;
      if (fs.existsSync(path.join(providerPath, chart.name, "Chart.yaml"))) {
        count += 1;
      }
    }
  }
  return count;
}

function main() {
  console.log("📊 Generating stats.generated.ts...");

  const components = countKinds();
  const providers = countProviders();
  const charts = countCharts();

  if (!components || !providers || !charts) {
    throw new Error(
      `Refusing to write empty stats (components=${components}, providers=${providers}, charts=${charts}). Check repo-root paths.`,
    );
  }

  const out = `/**
 * GENERATED FILE — do not edit by hand.
 *
 * Written by scripts/generate-stats.ts (runs as part of \`yarn build\`) from the
 * canonical sources in the repo: the cloud-resource-kind enum, the provider
 * tree, and charts/. This keeps the site's numbers single-sourced from code so
 * they never drift into a stale hardcoded count.
 */
export const generatedStats = {
  /** Annotated cloud-resource kinds (catalog breadth). */
  components: ${components},
  /** Cloud providers in the catalog. */
  providers: ${providers},
  /** Ready-made infra charts (stacks). */
  charts: ${charts},
} as const;

export type GeneratedStats = typeof generatedStats;
`;

  fs.writeFileSync(path.join(siteRoot, "src/site/stats.generated.ts"), out);
  console.log(
    `✅ stats.generated.ts (${components} components, ${providers} providers, ${charts} charts)`,
  );
}

main();
