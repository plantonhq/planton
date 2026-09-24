/**
 * Internal link gate: every link inside the built export resolves to a live
 * page, and every retired route is whole.
 *
 * It walks out/, reads every HTML file, collects each same-site href (a path
 * starting with "/", or an absolute planton.ai URL), and checks that the
 * target exists as <path>.html, <path>/index.html, or a static file under
 * out/. Anchors, query strings, and asset URLs under the asset prefix are
 * normalized away. A dead link fails the build and names the page it is on
 * and where it points.
 *
 * Retired routes (src/data/retired-routes.ts) are held to three laws, because
 * a forward exists for the outside world and never for our own links:
 *   1. every retired path has a stub in the export, so an old link on any
 *      origin that serves the export directly still lands somewhere;
 *   2. every retired path forwards to a live registered page, never to
 *      another retired path or to nothing;
 *   3. no page in the export links to a retired path; our own links point at
 *      the live page, so a visitor never rides a forward we could have spared.
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
  const registryModule = await import(pathToFileURL(path.join(siteRoot, 'src/data/site-pages.ts')).href);
  const retired = new Set(retiredModule.RETIRED_ROUTES.map((r) => r.from));
  const registered = new Set(registryModule.SITE_PAGES.map((p) => p.path));

  // ---- retired routes are whole --------------------------------------------
  const broken = [];
  for (const r of retiredModule.RETIRED_ROUTES) {
    if (!exists(r.from)) broken.push(`${r.from} is retired but the export has no stub for it (add an eight-line RetiredRoute page at that path)`);
    if (retired.has(r.to)) broken.push(`${r.from} forwards to ${r.to}, which is itself retired; forward straight to the live page`);
    else if (!registered.has(r.to) && !exists(r.to)) broken.push(`${r.from} forwards to ${r.to}, which is neither a registered page nor in the export`);
  }
  if (broken.length) {
    fail(`${broken.length} retired route(s) are not whole:\n  ${broken.join('\n  ')}`);
    return;
  }

  const files = walkHtml(exportDir);
  const dead = new Map(); // target -> Set(pages)
  const stale = new Map(); // retired target -> Set(pages)
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
      if (retired.has(target)) {
        // A retired page's own stub links to its destination, never to itself,
        // so any link to a retired path is a link we should have repointed.
        stale.set(target, (stale.get(target) ?? new Set()).add(routeOf(file)));
        continue;
      }
      if (exists(target)) continue;
      dead.set(target, (dead.get(target) ?? new Set()).add(routeOf(file)));
    }
  }

  if (stale.size) {
    const lines = [...stale.entries()].map(([target, pages]) => `  ${target}  (linked from ${[...pages].slice(0, 4).join(', ')}${pages.size > 4 ? `, +${pages.size - 4} more` : ''})`);
    fail(`${stale.size} internal link target(s) are retired paths; point each link at the live page the retired route forwards to:\n${lines.join('\n')}`);
    return;
  }

  if (dead.size) {
    const lines = [...dead.entries()].map(([target, pages]) => `  ${target}  (linked from ${[...pages].slice(0, 4).join(', ')}${pages.size > 4 ? `, +${pages.size - 4} more` : ''})`);
    fail(`${dead.size} internal link target(s) resolve to nothing in the export:\n${lines.join('\n')}`);
    return;
  }
  console.log(`\u2713 ${GATE}: ${checked} internal links across ${files.length} pages resolve to an exported page or a static file; ${retired.size} retired routes are whole`);
}

main().catch((err) => fail(err.stack ?? String(err)));
