/**
 * Apex-routing guard: every top-level path this site serves must be one the
 * apex domain will actually deliver.
 *
 * planton.ai is shared with the console. A router at the edge hands a declared
 * list of first path segments (and a few exact root files) to this site; every
 * other path goes to the console, which reads an unknown first segment as an
 * organization slug. So a page can build, deploy, and still never be reachable
 * at planton.ai/<path> -- which is exactly what happened to /privacy, /terms,
 * and /refund-policy. And a tenant could register an organization named after
 * a page and shadow it, unless the platform reserves that handle.
 *
 * This guard reads the site's route registry (src/data/site-pages.ts) and
 * the root files the export ships, derives the set of top-level path
 * segments, and checks each against two lists that live in the sibling
 * planton-platform checkout:
 *
 *   1. the router's website list -- `site_roots` and `site_files` in its params
 *      file (infrastructure/desktop/Infra.foundation.InfraStack.foundation-apex-router.yaml),
 *      the one place the website's paths are declared;
 *   2. the reserved handles -- PlatformReservedHandles.java's RESERVED_HANDLES.
 *
 * A missing entry fails the build and names exactly what to add and where.
 * Without the sibling checkout (CI in the open-source repo) the guard skips
 * with a loud notice, the same posture as the displayed-vs-enforced guard;
 * the authoritative run is the local `make build` where both checkouts exist.
 *
 * Usage:  node scripts/check-apex-routing.mjs
 *         PLANTON_PLATFORM_DIR=/path/to/planton-platform node scripts/check-apex-routing.mjs
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { pathToFileURL } from 'node:url';
import { parse as parseYaml } from 'yaml';

const GUARD = 'apex-routing guard';
const siteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const repoRoot = path.resolve(siteRoot, '..');
const platformDir = process.env.PLANTON_PLATFORM_DIR ?? path.resolve(repoRoot, '..', 'planton-platform');

const ROUTER_PARAMS = 'infrastructure/desktop/Infra.foundation.InfraStack.foundation-apex-router.yaml';
const HANDLES = 'product/libs/java/domain/reserved-handles/src/main/java/ai/planton/reservedhandles/PlatformReservedHandles.java';

/** Root files the export ships that the router must hand to the site exactly. */
const ROOT_FILES = ['/sitemap.xml', '/robots.txt', '/llms.txt', '/llms-full.txt', '/favicon.ico'];

/**
 * Segments that are console routes by design and never website pages; a
 * registry path under one of these is a mistake this guard would otherwise
 * misreport as "add to the router's list".
 */
const CONSOLE_ROOTS = new Set(['dashboard', 'login', 'logout', 'signup', 'auth', 'oauth', 'api', 'orgs', 'organization']);

function fail(message) {
  console.error(`\u2717 ${GUARD}: ${message}`);
  process.exitCode = 1;
}

async function loadRegistry() {
  // Node runs the data files directly through type stripping; the files carry
  // .ts extensions on their relative imports for exactly this.
  const mod = await import(pathToFileURL(path.join(siteRoot, 'src/data/site-pages.ts')).href);
  const retired = await import(pathToFileURL(path.join(siteRoot, 'src/data/retired-routes.ts')).href);
  return { pages: mod.SITE_PAGES, retired: retired.RETIRED_ROUTES, unregistered: mod.UNREGISTERED_PREFIXES };
}

/** The first path segment of a route, e.g. "/trust/the-record" -> "trust"; "/" -> null. */
function firstSegment(route) {
  const seg = route.split('/').filter(Boolean)[0];
  return seg ?? null;
}

/** The router's website list: the first segments (`site_roots`) and exact root files (`site_files`) it hands to the site. */
function readRouterList() {
  const file = path.join(platformDir, ROUTER_PARAMS);
  if (!fs.existsSync(file)) return null;
  const params = Object.fromEntries(parseYaml(fs.readFileSync(file, 'utf8')).params.map((p) => [p.name, String(p.value ?? '')]));
  const entries = (name) => (params[name] ?? '').split(',').map((entry) => entry.trim()).filter(Boolean);
  return { roots: new Set(entries('site_roots')), files: new Set(entries('site_files')), file };
}

function readReservedHandles() {
  const file = path.join(platformDir, HANDLES);
  if (!fs.existsSync(file)) return null;
  const text = fs.readFileSync(file, 'utf8');
  const start = text.indexOf('RESERVED_HANDLES = Set.of(');
  const end = text.indexOf(');', start);
  const body = text.slice(start, end);
  return { handles: new Set([...body.matchAll(/"([a-z0-9-]+)"/g)].map((m) => m[1])), file };
}

async function main() {
  if (!fs.existsSync(platformDir)) {
    console.warn(
      `\u26a0 ${GUARD}: sibling planton-platform checkout not found at ${platformDir} -- ` +
        `the router-list and reserved-handle checks are skipped here. The authoritative run is the local build where both checkouts exist.`,
    );
    return;
  }

  const { pages, retired, unregistered } = await loadRegistry();
  const router = readRouterList();
  const reserved = readReservedHandles();
  if (!router) fail(`the router's website list not found at ${path.join(platformDir, ROUTER_PARAMS)}`);
  if (!reserved) fail(`reserved handles not found at ${path.join(platformDir, HANDLES)}`);
  if (process.exitCode) return;

  // The site's top-level segments: registered pages, retired paths (the router must still deliver them
  // to the site so the redirect can happen), and the noindex prefixes served by their own layouts.
  const segments = new Set();
  for (const p of pages) {
    const seg = firstSegment(p.path);
    if (seg) segments.add(seg);
  }
  for (const r of retired) {
    const seg = firstSegment(r.from);
    if (seg) segments.add(seg);
  }
  for (const prefix of unregistered) segments.add(firstSegment(prefix));

  const missingAtEdge = [];
  const missingHandles = [];
  for (const seg of [...segments].sort()) {
    if (CONSOLE_ROOTS.has(seg)) {
      fail(`"/${seg}" is a console route; the site must not register a page under it`);
      continue;
    }
    if (!router.roots.has(seg)) missingAtEdge.push(seg);
    if (!reserved.handles.has(seg)) missingHandles.push(`"${seg}"`);
  }
  for (const file of ROOT_FILES) {
    if (!router.files.has(file)) missingAtEdge.push(file);
  }

  if (missingAtEdge.length) {
    fail(
      `these paths are served by the site but not in the router's website list, so planton.ai hands them to the console. ` +
        `Add each to site_roots (a segment) or site_files (a root file) in ${ROUTER_PARAMS} and apply it before the pages merge:\n  ` +
        missingAtEdge.join('\n  '),
    );
  }
  if (missingHandles.length) {
    fail(
      `these top-level paths are not reserved handles, so a tenant could register an organization with that slug and shadow the page. ` +
        `Add each to RESERVED_HANDLES in ${HANDLES}:\n  ` + missingHandles.join(', '),
    );
  }
  if (!process.exitCode) {
    console.log(`\u2713 ${GUARD}: ${segments.size} top-level paths and ${ROOT_FILES.length} root files are in the router's website list and reserved as handles`);
  }
}

main().catch((err) => {
  fail(err.stack ?? String(err));
});
