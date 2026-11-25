/**
 * llms.txt: the site for agents, generated from the same data the pages
 * render from.
 *
 * An agent reading planton.ai/llms.txt gets the exact sentences a person
 * reads on the page, because both come from src/data: the story's chapters,
 * the route registry's titles and descriptions, the platform
 * statistics, and the prices. Nothing here is hand-written prose about the
 * product; it is the data, projected as Markdown.
 *
 * Writes into out/ after `next build`:
 *   llms.txt        the curated index: the story in order, then every page by group, then the docs
 *   llms-full.txt   the index plus the full text of every marketing page and every docs page
 *   <route>.md      one Markdown file beside each registered marketing page's HTML
 *
 * Coverage check: every exported HTML route must be a registered page, a
 * content page under a content index, or under a prefix the registry
 * declares deliberately unregistered. Anything else fails the build,
 * so a new page cannot ship without appearing here.
 *
 * Usage:  node --experimental-strip-types scripts/generate-llms.mjs
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import matter from 'gray-matter';
import { HERO } from '../src/data/homepage-experience.ts';

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
        const route =
          (
            '/' +
            path
              .relative(exportDir, full)
              .replace(/\\/g, '/')
              .replace(/\.html$/, '')
          ).replace(/\/index$/, '') || '/';
        if (['/404', '/_not-found'].includes(route)) continue;
        routes.add(route);
      }
    }
  };
  walk(exportDir);
  return [...routes].sort();
}

// ---------------------------------------------------------------------------
// Content pages (docs, changelog) from their markdown
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
    const rel = path
      .relative(dir, file)
      .replace(/\\/g, '/')
      .replace(/\.mdx?$/, '');
    const route = `/${folder}/${rel}`.replace(/\/index$/, '');
    return {
      route,
      title: data.title ?? rel,
      description: data.description ?? data.excerpt ?? '',
      body: content.trim(),
    };
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

function pageMarkdown(
  page,
  {
    story,
    stats,
    site,
    desktop,
    desktopDownload,
    homepage,
    security,
  }
) {
  const lines = [
    `# ${page.title}`,
    '',
    page.description,
    '',
    `Canonical URL: ${site.url}${page.path === '/' ? '' : page.path}`,
    '',
  ];
  if (page.path === '/') {
    // Home, section by section, in page order, from the same records the page renders.
    const h = homepage.HOMEPAGE;
    const f = h.firstApp;
    const prompt = homepage.SETUP_PROMPT_EXAMPLE;
    lines.push(`## ${h.headline}`, '', h.intro, '', h.caption, '', HERO.description, '');
    lines.push(`## ${f.title}`, '', f.intro, '', `${f.promptLabel}:`, '', '```text');
    for (const line of prompt.lines) lines.push(line.replace('{code}', prompt.code));
    lines.push('```', '');
    for (const side of [f.you, f.agent]) {
      lines.push(`### ${side.title}`, '');
      for (const step of side.steps) lines.push(`- ${step}`);
      lines.push('');
    }
    lines.push(`## ${h.close.title}`, '', h.close.text, '', h.close.note, '', `[Book a Demo](${site.url}/book-demo)`, '');
    return lines.join('\n');
  }
  if (page.path === '/security') {
    // Security, answer by answer, from the records the page renders: each
    // answer, the real thing behind it, and how to check it.
    const sec = security.SECURITY;
    lines.push(`## ${sec.headline}`, '', sec.lede, '');
    for (const section of sec.sections) {
      lines.push(`## ${section.title}`, '');
      for (const p of section.paragraphs) lines.push(p, '');
      const a = section.artifact;
      if (a?.kind === 'code') lines.push(`${a.label}:`, '', `\`\`\`${a.language}`, ...a.lines, '```', '');
      if (a?.kind === 'table') {
        lines.push(`${a.label}:`, '', `| ${a.columns.join(' | ')} |`, `|${a.columns.map(() => '---').join('|')}|`);
        for (const row of a.rows) lines.push(`| ${row.join(' | ')} |`);
        lines.push('');
      }
      const c = section.check;
      if (c) {
        const href = c.href && (c.href.startsWith('http') ? c.href : `${site.url}${c.href}`);
        lines.push(href ? `[${c.text}](${href})` : c.text, '');
        if (c.command) lines.push('```shell', c.command, '```', '');
      }
    }
    lines.push(`## ${sec.more.title}`, '');
    for (const doc of sec.more.docs) lines.push(`- [${doc.label}](${site.url}${doc.href}.md)`);
    lines.push('', `${sec.more.contact.text} ${sec.more.contact.address}`);
    return lines.join('\n').trim() + '\n';
  }
  if (page.chapters?.length) {
    lines.push('## What this page says', '');
    for (const id of page.chapters) {
      const c = story.chapter(id);
      lines.push(chapterMarkdown(c), '');
    }
  }
  // The desktop pages state their own record: the landing's sections in the
  // reader's order, and the download page's first-launch beats and follow-ons
  // (the platform-neutral parts; installers and commands are the download data's).
  if (page.path === desktopDownload.DESKTOP_LANDING_PATH) {
    const d = desktop.DESKTOP_LANDING;
    lines.push(`## ${d.twoWays.title}`, '');
    for (const w of d.twoWays.ways) lines.push(`- **${w.title}.** ${w.trade} ${w.body}`);
    lines.push('', d.twoWays.turn, '', d.twoWays.concession, '');
    lines.push(
      `## ${d.ask.title}`,
      '',
      d.ask.lede,
      '',
      `**${d.ask.agent.title}.** ${d.ask.agent.body}`,
      ''
    );
    lines.push(`**${d.ask.addsTitle}.**`, '');
    for (const a of d.ask.adds) lines.push(`- ${a}`);
    lines.push('');
    for (const section of [d.runs, d.infraHub, d.serviceHub]) {
      lines.push(`## ${section.title}`, '', section.lede, '');
      for (const t of section.tiles ?? section.points) lines.push(`- **${t.title}.** ${t.body}`);
      lines.push('');
    }
    lines.push(
      `- Measured: ${d.serviceHub.measured.figures.map((f) => `${f.value} ${f.label}`).join(', ')}. ${d.serviceHub.measured.provenance}`,
      ''
    );
    lines.push(
      `## ${d.everywhere.title}`,
      '',
      d.everywhere.body,
      '',
      d.everywhere.exit,
      '',
      `## ${d.whyFree.title}`,
      '',
      ...d.whyFree.body.flatMap((p) => [p, ''])
    );
  }
  if (page.path === desktopDownload.DESKTOP_DOWNLOAD_PATH) {
    const d = desktop.DESKTOP_DOWNLOAD;
    lines.push(`## ${d.afterInstall.title}`, '', d.afterInstall.lede, '');
    d.afterInstall.beats.forEach((b, i) => lines.push(`${i + 1}. **${b.title}.** ${b.body}`));
    lines.push('', `## ${d.afterInstall.thenTitle}`, '', d.afterInstall.thenLede, '');
    for (const tab of d.followOns(desktopDownload.DESKTOP_PLATFORMS[0]))
      lines.push(`- **${tab.label}.** \`${tab.commands.join('; ')}\` ${tab.description}`);
    lines.push('');
  }
  if (page.path === '/') {
    lines.push(
      '## By the numbers',
      '',
      `- ${stats.PLATFORM_COUNTS.catalogKinds} catalog kinds across ${stats.PLATFORM_COUNTS.providers} providers`,
      `- ${stats.PLATFORM_COUNTS.controls} technical controls in ${stats.PLATFORM_COUNTS.controlCategories} categories`,
      `- Counted from the open-source repository on ${stats.PLATFORM_COUNTS.countedOn}`,
      ''
    );
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
  const homepage = await load('src/data/homepage.ts');
  const stats = await load('src/data/platform-stats.ts');
  const pricing = await load('src/data/pricing.ts');
  const desktop = await load('src/data/desktop.ts');
  const desktopDownload = await load('src/data/desktop-download.ts');
  const security = await load('src/data/security.ts');

  const site = registry.SITE;
  const pages = registry.SITE_PAGES;
  const contentIndexes = pages.filter((p) => p.group === 'content').map((p) => p.path);

  // ---- coverage -----------------------------------------------------------
  const registered = new Set(pages.map((p) => p.path));
  const unaccounted = [];
  for (const route of exportedRoutes()) {
    if (registered.has(route)) continue;
    if (contentIndexes.some((idx) => route.startsWith(`${idx}/`))) continue;
    if (
      registry.UNREGISTERED_PREFIXES.some(
        (prefix) => route === prefix || route.startsWith(`${prefix}/`)
      )
    )
      continue;
    unaccounted.push(route);
  }
  if (unaccounted.length) {
    fail(
      `${unaccounted.length} exported page(s) are not in the route registry and not under a declared unregistered prefix, so llms.txt and the sitemap would not know them. ` +
        `Register each in src/data/site-pages.ts, or delete the page:\n  ${unaccounted.join('\n  ')}`
    );
    return;
  }

  // ---- per-page markdown ------------------------------------------------
  const marketing = pages.filter((p) => p.index !== false && p.group !== 'content');
  // Everything a page's markdown may quote: the story and the records the pages render from.
  const data = {
    story,
    stats,
    site,
    desktop,
    desktopDownload,
    homepage,
    security,
  };
  for (const page of marketing) {
    const target =
      page.path === '/'
        ? path.join(exportDir, 'index.md')
        : path.join(exportDir, `${page.path.slice(1)}.md`);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, pageMarkdown(page, data));
  }

  // ---- llms.txt -----------------------------------------------------------
  const docs = [
    ...contentFiles('docs'),
    ...contentFiles('changelog'),
  ];
  // The index lists pages under the groups the registry names, in the
  // registry's order. A marketing page whose group has no heading would be
  // built, exported, and silently absent from the index, so it fails here.
  const headings = registry.PAGE_GROUP_HEADINGS;
  const unlisted = marketing.filter((p) => p.path !== '/' && !headings[p.group]);
  if (unlisted.length) {
    fail(
      `${unlisted.length} page(s) belong to a group PAGE_GROUP_HEADINGS (src/data/site-pages.ts) does not list, so llms.txt would omit them:\n  ${unlisted.map((p) => `${p.path} (${p.group})`).join('\n  ')}`
    );
    return;
  }
  const groups = Object.entries(headings).map(([group, label]) => [label, group]);
  const index = [];
  index.push(`# ${site.name}`, '');
  index.push(`> ${registry.sitePage('/').description}`, '');
  index.push(homepage.HOMEPAGE.intro, '');
  index.push(`## ${homepage.HOMEPAGE.firstApp.title}`, '', homepage.HOMEPAGE.firstApp.intro, '');
  index.push('## Who it is for', '', story.chapter('who-it-is-for').claim, '');
  index.push(
    '## Start',
    '',
    `- Hosted free tier: ${pricing.FREE_TIER_SEATS} seats, no card. Team plan per seat. Self-hosted community edition free for up to ${pricing.COMMUNITY_SEAT_LIMIT} seats; ${pricing.EVALUATION_DAYS}-day evaluation key. Planton Desktop free for individuals, including commercial use.`,
    `- Pricing: ${site.url}/pricing`,
    `- [${registry.AGENT_PLAYBOOK.title}](${site.url}${registry.AGENT_PLAYBOOK.path}): ${registry.AGENT_PLAYBOOK.description}`,
    ''
  );
  index.push('## Pages', '');
  index.push(
    `- [${registry.sitePage('/').title}](${site.url}/index.md): ${registry.sitePage('/').description}`
  );
  for (const [label, group] of groups) {
    const rows = marketing.filter((p) => p.group === group);
    if (!rows.length) continue;
    index.push('', `### ${label}`, '');
    for (const p of rows) index.push(`- [${p.title}](${site.url}${p.path}.md): ${p.description}`);
  }
  index.push('', '## Documentation', '');
  for (const d of docs)
    index.push(
      `- [${d.title}](${site.url}${d.route}.md)${d.description ? `: ${d.description}` : ''}`
    );
  index.push(
    '',
    '## Optional',
    '',
    `- [Sitemap](${site.url}/sitemap.xml)`,
    `- [Full text](${site.url}/llms-full.txt)`,
    ''
  );
  fs.writeFileSync(path.join(exportDir, 'llms.txt'), index.join('\n'));

  // ---- llms-full.txt ------------------------------------------------------
  const full = [index.join('\n'), '', '---', ''];
  for (const page of marketing) full.push(pageMarkdown(page, data), '---', '');
  for (const d of docs)
    full.push(
      `# ${d.title}`,
      '',
      d.description,
      '',
      `Canonical URL: ${site.url}${d.route}`,
      '',
      d.body,
      '',
      '---',
      ''
    );
  fs.writeFileSync(path.join(exportDir, 'llms-full.txt'), full.join('\n'));

  console.log(
    `\u2713 ${GENERATOR}: llms.txt (${marketing.length} pages, ${docs.length} documents), llms-full.txt, and ${marketing.length} per-page markdown files written; every exported route is registered, content, or declared unregistered`
  );
}

main().catch((err) => fail(err.stack ?? String(err)));
