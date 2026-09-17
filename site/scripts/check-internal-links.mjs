/**
 * Internal link gate: every link inside the built export resolves to a page
 * the export ships or to a declared retired route.
 *
 * It walks out/, reads every HTML file, collects each same-site href (a path
 * starting with "/", or an absolute planton.ai URL), and checks that the
 * target exists as <path>.html, <path>/index.html, a static file under out/,
 * or an entry in src/data/retired-routes.ts. Anchors, query strings, and
 * asset URLs under the asset prefix are normalized away. A dead link fails
 * the build and names the page it is on and where it points.
 *
 * Runs after `next build` in the build script. Also usable alone:
 *   node scripts/check-internal-links.mjs
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const GATE = 'internal link gate';
const siteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const exportDir = path.join(siteRoot, 'out');

/** Paths the console serves on the shared domain; a link to them is not a dead link. */
const CONSOLE_PATHS = new Set(['/signup', '/login', '/logout', '/dashboard', '/cloud-catalog', '/license/buy', '/license/buy#evaluation']);

function fail(message) {
  console.error(`\u2717 ${GATE}: ${message}`);
  process.exitCode = 1;
}

function walkHtml(dir, out = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (['_next', '_site', '_pagefind'].includes(path.relative(exportDir, full).split(path.sep)[0])) continue;
      walkHtml(full, out);
    } else if (entry.name.endsWith('.html')) {
      out.push(full);
    }
  }
  return out;
}

function routeOf(file) {
  const rel = '/' + path.relative(exportDir, file).replace(/\\/g, '/').replace(/\.html$/, '');
  return rel.replace(/\/index$/, '') || '/';
}

function exists(target) {
  const clean = target.replace(/\/+$/, '') || '/';
  if (clean === '/') return fs.existsSync(path.join(exportDir, 'index.html'));
  const candidates = [
    path.join(exportDir, `${clean}.html`),
    path.join(exportDir, clean, 'index.html'),
    path.join(exportDir, clean),
  ];
  return candidates.some((c) => fs.existsSync(c) && fs.statSync(c).isFile());
}

async function main() {
  if (!fs.existsSync(exportDir)) {
    fail(`no static export at ${exportDir}; run the build first`);
    return;
  }
  const retiredModule = await import(pathToFileURL(path.join(siteRoot, 'src/data/retired-routes.ts')).href);
  const retired = new Set(retiredModule.RETIRED_ROUTES.map((r) => r.from));

  const files = walkHtml(exportDir);
  const dead = new Map(); // target -> Set(pages)
  let checked = 0;
  for (const file of files) {
    const html = fs.readFileSync(file, 'utf8');
    for (const m of html.matchAll(/href="([^"]+)"/g)) {
      let href = m[1].replace(/&amp;/g, '&');
      if (href.startsWith('https://planton.ai')) href = href.slice('https://planton.ai'.length) || '/';
      if (!href.startsWith('/') || href.startsWith('//')) continue;
      const target = href.split('#')[0].split('?')[0] || '/';
      if (target.startsWith('/_site/') || target.startsWith('/_next/')) continue;
      if (/\.(xml|txt|json|pdf|png|jpg|jpeg|svg|ico|md|css|js|woff2?)$/.test(target)) {
        if (!fs.existsSync(path.join(exportDir, target))) {
          dead.set(target, (dead.get(target) ?? new Set()).add(routeOf(file)));
        }
        checked += 1;
        continue;
      }
      if (CONSOLE_PATHS.has(target) || CONSOLE_PATHS.has(href.split('?')[0])) continue;
      checked += 1;
      if (exists(target) || retired.has(target)) continue;
      dead.set(target, (dead.get(target) ?? new Set()).add(routeOf(file)));
    }
  }

  if (dead.size) {
    const lines = [...dead.entries()].map(([target, pages]) => `  ${target}  (linked from ${[...pages].slice(0, 4).join(', ')}${pages.size > 4 ? `, +${pages.size - 4} more` : ''})`);
    fail(`${dead.size} internal link target(s) resolve to nothing in the export and are not declared retired routes:\n${lines.join('\n')}`);
    return;
  }
  console.log(`\u2713 ${GATE}: ${checked} internal links across ${files.length} pages resolve to an exported page, a static file, or a retired route`);
}

main().catch((err) => fail(err.stack ?? String(err)));
