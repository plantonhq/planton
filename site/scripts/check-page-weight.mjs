/**
 * Home's page-weight budget. Home is the page a stranger opens first, often on
 * a phone from a link in a thread, so what it costs to load is part of what it
 * says. This check reads the built export: Home's HTML, and every script and
 * stylesheet that HTML references, each counted raw and gzipped (what the
 * network carries), and every screen Home shows, from the site's own images
 * (content/assets/, which the assets bucket serves). It fails the build when
 * Home outgrows its budget.
 *
 * The budget is Home's measured weight when it was set, plus a stated margin
 * for small changes. A section that needs more raises the budget in the same
 * change, with the reason beside the number; the budget never rises silently.
 *
 * Usage: node scripts/check-page-weight.mjs   (after `next build` has written out/)
 */
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { fileURLToPath } from 'node:url';

const site = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const out = path.join(site, 'out');

/**
 * Gzipped bytes. Set 2026-11-25 for the hero, the first app and the close,
 * measured then (24.1 KB of HTML, 290.1 KB of scripts, 14.6 KB of styles) plus
 * about 10%. The scripts are mostly the site shell every page shares (React,
 * Next and MUI); Home's own sections add little to them.
 *
 * Home is built a section at a time, and each section raises the HTML line by
 * what it measured, never by a guess: "ask your coding agent for
 * infrastructure" fit under the first line (26.3 KB); "run it yourself" took
 * Home to 27.7 KB on 2026-11-26; the open-source section (one agent's walk and
 * the catalog's open-source projects) took it to 31.6 KB the same day, so HTML
 * is that plus about 10%. The deploy run (its words, a screen's record and the
 * folded plan) fit under it at 32.9 KB on 2026-11-27.
 */
const BUDGET = {
  html: 34 * 1024,
  scripts: 320 * 1024,
  styles: 16 * 1024,
};

/**
 * Each screen's image, its largest file (the desktop crop or the phone's),
 * already compressed. Set 2026-11-27 with the first screen (the run, 37 KB at
 * twice its shown width): a screen is a flat console capture, so one well over
 * this is the wrong crop or the wrong format, not a richer picture.
 */
const SCREEN_IMAGE_BUDGET = 300 * 1024;

const gz = (bytes) => zlib.gzipSync(bytes, { level: 9 }).length;
const html = fs.readFileSync(path.join(out, 'index.html'));
const text = html.toString('utf8');

/** A referenced asset's file in the export; the asset prefix (`/_site`) is copied over `out/` after the build. */
const file = (href) => path.join(out, href.replace(/^\/_site(?=\/_next\/)/, ''));
const refs = (pattern) =>
  [...new Set([...text.matchAll(pattern)].map((m) => m[1]))].filter((h) => h.startsWith('/'));
const scripts = refs(/<script[^>]+src="([^"]+)"/g);
const styles = refs(/<link[^>]+rel="stylesheet"[^>]+href="([^"]+)"/g);
const sum = (hrefs) =>
  hrefs.reduce(
    (t, h) => {
      const bytes = fs.readFileSync(file(h));
      return { raw: t.raw + bytes.length, gzip: t.gzip + gz(bytes) };
    },
    { raw: 0, gzip: 0 }
  );

const measured = {
  html: { raw: html.length, gzip: gz(html) },
  scripts: sum(scripts),
  styles: sum(styles),
};
const kb = (n) => `${(n / 1024).toFixed(1)} KB`;
let failed = false;
for (const [part, m] of Object.entries(measured)) {
  const over = BUDGET[part] > 0 && m.gzip > BUDGET[part];
  failed ||= over;
  console.log(
    `${over ? 'FAIL' : 'ok  '} Home ${part}: ${kb(m.gzip)} gzipped (${kb(m.raw)} raw)${BUDGET[part] ? `, budget ${kb(BUDGET[part])}` : ''}`
  );
}
// Home's screens: the images its records name, read from the site's own copy of the bucket.
const ASSETS_URL = 'https://assets.planton.ai/site/';
const screenImages = [
  ...new Set(
    [...text.matchAll(/(?:src|srcSet)="(https:\/\/assets\.planton\.ai\/site\/[^"]+)"/g)].map(
      (m) => m[1]
    )
  ),
];
for (const src of screenImages) {
  const bytes = fs.statSync(path.join(site, 'content/assets', src.slice(ASSETS_URL.length))).size;
  const over = bytes > SCREEN_IMAGE_BUDGET;
  failed ||= over;
  console.log(
    `${over ? 'FAIL' : 'ok  '} Home screen ${path.basename(src)}: ${kb(bytes)}, budget ${kb(SCREEN_IMAGE_BUDGET)}`
  );
}
if (Object.values(BUDGET).some((b) => b === 0)) {
  console.error('FAIL: Home has no page-weight budget set.');
  process.exit(1);
}
if (failed) {
  console.error(
    'FAIL: Home is over its page-weight budget. Make it lighter, or raise the budget here with the reason.'
  );
  process.exit(1);
}
