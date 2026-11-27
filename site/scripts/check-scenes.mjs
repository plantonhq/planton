/**
 * The scenes guard: every screen of the product the site shows is a crop of a
 * real, reproducible scene, and its image is the site's own.
 *
 * A screen record (`kind: 'screen'`, src/data/artifact.ts) names the demo
 * universe scene it was cropped from (`demo-universe/<company>/company.yaml`,
 * `scenes:`). The scene is the source of the picture: when the console
 * changes, the universe's capture refreshes the scene and the crop is redone
 * from it, so a screen with no scene behind it is a picture nobody can
 * refresh, and a page that shows the product must never hold one.
 *
 * Two laws:
 *
 * 1. THE IMAGE -- every screen's image (and its phone crop) is a file under
 *    content/assets/, at the path the assets bucket serves it from
 *    (https://assets.planton.ai/site/<path>). The site's images are
 *    committed, because the bucket's sync mirrors this folder and deletes
 *    what it lacks. Always enforced.
 *
 * 2. THE SCENE -- every screen's scene id exists in the demo universe, read
 *    live from the sibling planton-platform checkout. Without that checkout
 *    (the open-source repository's CI) this law is skipped with a notice, as
 *    the displayed-vs-enforced guard skips its platform laws; the
 *    authoritative run is the local make build beside the platform checkout.
 *
 * A page that adds screens adds its data module to SCREEN_SOURCES.
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { parse } from 'yaml';

const GUARD = 'check-scenes';
const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const ASSETS_URL = 'https://assets.planton.ai/site/';

/** The data modules whose records hold screens. */
const SCREEN_SOURCES = ['src/data/homepage.ts'];

/** Every screen record in a module's exports, at any depth. */
function screensIn(value, found = []) {
  if (Array.isArray(value)) value.forEach((v) => screensIn(v, found));
  else if (value && typeof value === 'object') {
    if (value.kind === 'screen') found.push(value);
    else Object.values(value).forEach((v) => screensIn(v, found));
  }
  return found;
}

const screens = [];
for (const source of SCREEN_SOURCES) {
  const mod = await import(path.join(repoRoot, source));
  screens.push(...screensIn(mod).map((screen) => ({ source, screen })));
}

const failures = [];

// Law 1: the image is the site's own.
for (const { source, screen } of screens) {
  for (const src of [screen.src, screen.phone?.src].filter(Boolean)) {
    if (!src.startsWith(ASSETS_URL)) {
      failures.push(`${source}: ${screen.scene}'s image ${src} is not served from ${ASSETS_URL}`);
      continue;
    }
    const file = path.join(repoRoot, 'content/assets', src.slice(ASSETS_URL.length));
    if (!fs.existsSync(file))
      failures.push(
        `${source}: ${screen.scene}'s image is missing: ${path.relative(repoRoot, file)}`
      );
  }
}

// Law 2: the scene exists in the demo universe.
const platformDir =
  process.env.PLANTON_PLATFORM_DIR ?? path.resolve(repoRoot, '..', '..', 'planton-platform');
const universeDir = path.join(platformDir, 'demo-universe');
if (!fs.existsSync(universeDir)) {
  console.warn(
    `⚠ ${GUARD}: sibling planton-platform checkout not found at ${platformDir} -- ` +
      `SKIPPING the scene law (the image law still enforced). ` +
      `The authoritative run is the local make build beside the platform checkout.`
  );
} else {
  const ids = new Set();
  for (const company of fs.readdirSync(universeDir)) {
    const file = path.join(universeDir, company, 'company.yaml');
    if (!fs.existsSync(file)) continue;
    for (const scene of parse(fs.readFileSync(file, 'utf8'))?.scenes ?? []) ids.add(scene.id);
  }
  for (const { source, screen } of screens) {
    if (!ids.has(screen.scene))
      failures.push(`${source}: no demo universe scene is named ${screen.scene}`);
  }
}

if (failures.length) {
  console.error(
    `✗ ${GUARD}: ${failures.length} screen(s) with no source:\n  ${failures.join('\n  ')}`
  );
  process.exit(1);
}
console.log(
  `✓ ${GUARD}: ${screens.length} screen(s), each a scene's crop and the site's own image`
);
