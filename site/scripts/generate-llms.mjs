/**
 * llms.txt: the site for agents, generated from the same data the pages
 * render from.
 *
 * An agent reading planton.ai/llms.txt gets the exact sentences a person
 * reads on the page, because both come from src/data: the story's chapters,
 * the route registry's titles and descriptions, the personas, the platform
 * statistics, and the prices. Nothing here is hand-written prose about the
 * product; it is the data, projected as Markdown.
 *
 * Writes into out/ after `next build`:
 *   llms.txt        the curated index: the story in order, then every page by group, then the docs
 *   llms-full.txt   the index plus the full text of every marketing page and every docs page
 *   <route>.md      one Markdown file beside each registered marketing page's HTML
 *
 * Coverage check: every exported HTML route must be a registered page, a
 * content page under a content index, a retired route, or under a prefix the
 * registry declares deliberately unregistered. Anything else fails the build,
 * so a new page cannot ship without appearing here.
 *
 * Usage:  node --experimental-strip-types scripts/generate-llms.mjs
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import matter from 'gray-matter';

const GENERATOR = 'llms generator';
const siteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const exportDir = path.join(siteRoot, 'out');
const publicDir = path.join(siteRoot, 'public');

function fail(message) {
  console.error(`\u2717 ${GENERATOR}: ${message}`);
  process.exitCode = 1;
}

async function load(rel) {
  return import(pathToFileURL(path.join(siteRoot, rel)).href);
}

// ---------------------------------------------------------------------------
// Export routes (for the coverage check)
// ---------------------------------------------------------------------------

function exportedRoutes() {
  const routes = new Set();
  const walk = (dir) => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      const top = path.relative(exportDir, full).split(path.sep)[0];
      if (entry.isDirectory()) {
        if (['_next', '_site', '_pagefind'].includes(top)) continue;
        walk(full);
      } else if (entry.name.endsWith('.html')) {
        const route = ('/' + path.relative(exportDir, full).replace(/\\/g, '/').replace(/\.html$/, '')).replace(/\/index$/, '') || '/';
        if (['/404', '/_not-found'].includes(route)) continue;
        routes.add(route);
      }
    }
  };
  walk(exportDir);
  return [...routes].sort();
}

// ---------------------------------------------------------------------------
// Content pages (docs, blog, tutorials, changelog) from their markdown
// ---------------------------------------------------------------------------

function contentFiles(folder) {
  const dir = path.join(publicDir, folder);
  if (!fs.existsSync(dir)) return [];
  const files = [];
  const walk = (d) => {
    for (const entry of fs.readdirSync(d, { withFileTypes: true })) {
      if (entry.name.startsWith('_')) continue;
      const full = path.join(d, entry.name);
      if (entry.isDirectory()) walk(full);
      else if (/\.mdx?$/.test(entry.name)) files.push(full);
    }
  };
  walk(dir);
  return files.sort().map((file) => {
    const raw = fs.readFileSync(file, 'utf8');
    const { data, content } = matter(raw);
    const rel = path.relative(dir, file).replace(/\\/g, '/').replace(/\.mdx?$/, '');
    const route = `/${folder}/${rel}`.replace(/\/index$/, '');
    return { route, title: data.title ?? rel, description: data.description ?? data.excerpt ?? '', body: content.trim() };
  });
}

// ---------------------------------------------------------------------------
// Markdown projections
// ---------------------------------------------------------------------------

function chapterMarkdown(chapter) {
  const lines = [`### ${chapter.number}. ${chapter.title}`, '', chapter.claim, ''];
  for (const p of chapter.proof) lines.push(`- ${p}`);
  return lines.join('\n');
}

function pageMarkdown(page, story, personas, stats, site) {
  const lines = [`# ${page.title}`, '', page.description, '', `Canonical URL: ${site.url}${page.path === '/' ? '' : page.path}`, ''];
  if (page.chapters?.length) {
    lines.push('## What this page says', '');
    for (const id of page.chapters) {
      const c = story.chapter(id);
      lines.push(chapterMarkdown(c), '');
    }
  }
  if (page.group === 'solutions') {
    lines.push('## Who Planton is for', '');
    for (const p of personas) lines.push(`- **${p.name}.** ${p.who} ${p.wall}`);
    lines.push('');
  }
  if (page.path === '/' || page.group === 'trust') {
    lines.push('## By the numbers', '', `- ${stats.PLATFORM_COUNTS.componentKinds} component kinds across ${stats.PLATFORM_COUNTS.providers} providers`, `- ${stats.PLATFORM_COUNTS.infraCharts} Infra Charts`, `- ${stats.PLATFORM_COUNTS.controls} technical controls in ${stats.PLATFORM_COUNTS.controlCategories} categories, ${stats.PLATFORM_COUNTS.frameworkCrosswalks} framework crosswalks`, `- Counted from the open-source repository on ${stats.PLATFORM_COUNTS.countedOn}`, '');
  }
  return lines.join('\n').trim() + '\n';
}

async function main() {
  if (!fs.existsSync(exportDir)) {
    fail(`no static export at ${exportDir}; run next build first`);
    return;
  }
  const registry = await load('src/data/site-pages.ts');
  const story = await load('src/data/story.ts');
  const personasModule = await load('src/data/personas.ts');
  const stats = await load('src/data/platform-stats.ts');
  const pricing = await load('src/data/pricing.ts');
  const retiredModule = await load('src/data/retired-routes.ts');

  const site = registry.SITE;
  const pages = registry.SITE_PAGES;
  const personas = personasModule.PERSONAS;
  const retired = new Set(retiredModule.RETIRED_ROUTES.map((r) => r.from));
  const contentIndexes = pages.filter((p) => p.group === 'content').map((p) => p.path);

  // ---- coverage -----------------------------------------------------------
  const registered = new Set(pages.map((p) => p.path));
  const unaccounted = [];
  for (const route of exportedRoutes()) {
    if (registered.has(route) || retired.has(route)) continue;
    if (contentIndexes.some((idx) => route.startsWith(`${idx}/`))) continue;
    if (registry.UNREGISTERED_PREFIXES.some((prefix) => route === prefix || route.startsWith(`${prefix}/`))) continue;
    unaccounted.push(route);
  }
  if (unaccounted.length) {
    fail(
      `${unaccounted.length} exported page(s) are not in the route registry, not retired, and not under a declared unregistered prefix, so llms.txt and the sitemap would not know them. ` +
        `Register each in src/data/site-pages.ts (or retire it in retired-routes.ts):\n  ${unaccounted.join('\n  ')}`,
    );
    return;
  }

  // ---- per-page markdown ------------------------------------------------
  const marketing = pages.filter((p) => p.index !== false && p.group !== 'content');
  for (const page of marketing) {
    const target = page.path === '/' ? path.join(exportDir, 'index.md') : path.join(exportDir, `${page.path.slice(1)}.md`);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, pageMarkdown(page, story, personas, stats, site));
  }

  // ---- llms.txt -----------------------------------------------------------
  const docs = [...contentFiles('docs'), ...contentFiles('tutorials'), ...contentFiles('blog'), ...contentFiles('changelog')];
  const groups = [
    ['Trust', 'trust'],
    ['Product', 'product'],
    ['Solutions', 'solutions'],
    ['Pricing', 'pricing'],
    ['Company', 'company'],
    ['Legal', 'legal'],
  ];
  const index = [];
  index.push(`# ${site.name}`, '');
  index.push(`> ${registry.sitePage('/').description}`, '');
  index.push(story.STORY_SPINE, '');
  index.push('## The story, in order', '');
  for (const c of story.SITE_CHAPTERS) index.push(`- ${c.number}. ${c.title}: ${c.claim}`);
  index.push('');
  index.push('## Who it is for', '');
  for (const p of personas) index.push(`- ${p.name}: ${p.who}`);
  index.push('');
  index.push('## Start', '', `- Hosted free tier: ${pricing.FREE_TIER_SEATS} seats, no card. Team plan per seat. Self-hosted community edition free for up to ${pricing.COMMUNITY_SEAT_LIMIT} seats; ${pricing.EVALUATION_DAYS}-day evaluation key. Planton Desktop free for individuals, including commercial use.`, `- Pricing: ${site.url}/pricing`, '');
  index.push('## Pages', '');
  index.push(`- [${registry.sitePage('/').title}](${site.url}/index.md): ${registry.sitePage('/').description}`);
  for (const [label, group] of groups) {
    const rows = marketing.filter((p) => p.group === group);
    if (!rows.length) continue;
    index.push('', `### ${label}`, '');
    for (const p of rows) index.push(`- [${p.title}](${site.url}${p.path}.md): ${p.description}`);
  }
  index.push('', '## Documentation', '');
  for (const d of docs) index.push(`- [${d.title}](${site.url}${d.route}.md)${d.description ? `: ${d.description}` : ''}`);
  index.push('', '## Optional', '', `- [Sitemap](${site.url}/sitemap.xml)`, `- [Full text](${site.url}/llms-full.txt)`, '');
  fs.writeFileSync(path.join(exportDir, 'llms.txt'), index.join('\n'));

  // ---- llms-full.txt ------------------------------------------------------
  const full = [index.join('\n'), '', '---', ''];
  for (const page of marketing) full.push(pageMarkdown(page, story, personas, stats, site), '---', '');
  for (const d of docs) full.push(`# ${d.title}`, '', d.description, '', `Canonical URL: ${site.url}${d.route}`, '', d.body, '', '---', '');
  fs.writeFileSync(path.join(exportDir, 'llms-full.txt'), full.join('\n'));

  console.log(`\u2713 ${GENERATOR}: llms.txt (${marketing.length} pages, ${docs.length} documents), llms-full.txt, and ${marketing.length} per-page markdown files written; every exported route is registered, retired, content, or declared unregistered`);
}

main().catch((err) => fail(err.stack ?? String(err)));
