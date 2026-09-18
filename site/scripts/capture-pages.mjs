/**
 * Design-review captures of the built static export.
 *
 * A page is not done when it builds; it is done when an independent review of
 * its pixels finds nothing above minor. This script produces those pixels the
 * same way every time: it serves out/ itself (no external server to start),
 * opens each scene in headless Chromium at the review widths, and writes one
 * full-page PNG per scene. Fixed viewports, a fixed user agent per scene, and
 * a settled network mean any pixel difference between two runs is a change in
 * the page -- which is what makes a before/after grade honest.
 *
 * Scenes that depend on the browser's platform also ASSERT the page did the
 * right thing: a Windows user agent must promote the Windows card, and so on.
 * A failed assertion fails the run, so the harness is a test as well as a
 * camera.
 *
 * Usage (after `make build`):
 *   node scripts/capture-pages.mjs --out /path/to/captures
 *   node scripts/capture-pages.mjs --out ... --only download-1280,download-windows
 *   node scripts/capture-pages.mjs --out ... --tag before        # <scene>-before-dark.png
 *   node scripts/capture-pages.mjs --out ... --publish-og        # also refresh public/_site/images/og/*.png
 *   node scripts/capture-pages.mjs --out ... --base http://localhost:4175   # an already-running server
 *   node scripts/capture-pages.mjs --out ... --export /path/to/other/out   # serve a different export (a main build)
 *   node scripts/capture-pages.mjs --out ... --compare /path/to/before      # after capturing, diff against a prior set
 *
 * Determinism. Pages animate (a typing hero, framer-motion reveals), and a
 * pixel comparison of two runs is only honest when both runs stopped the clock
 * at the same instant. Every scene therefore renders under Chromium's virtual
 * time: the page loads, then exactly VIRTUAL_TIME_BUDGET_MS of virtual time
 * elapses (timers and animation frames included) before the screenshot, no
 * matter how fast or slow the machine is. Two builds of the same page produce
 * byte-identical PNGs; a difference is a change in the page.
 *
 * --compare reads <before>/<scene>-dark.png for every scene captured (the
 * before set is captured with no --tag), diffs pixel by pixel, writes
 * <scene>-diff.png beside the after set for any mismatch, and exits 1 when
 * any scene differs. This is the proof behind a "zero visual change" commit.
 *
 * The website has one theme (dark), so every file is <scene>-dark.png; a
 * reviewer asking for the light pair is told the surface has none.
 */

import fs from 'node:fs';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const puppeteer = require('puppeteer');
const { PNG } = require('pngjs');
const pixelmatch = (await import('pixelmatch')).default;

/** Virtual milliseconds every scene runs before its screenshot; long enough for every entrance animation to finish. */
const VIRTUAL_TIME_BUDGET_MS = 20000;

const args = process.argv.slice(2);
const arg = (name, fallback) => {
  const i = args.indexOf(`--${name}`);
  return i >= 0 ? args[i + 1] : fallback;
};
const flag = (name) => args.includes(`--${name}`);

const OUT_DIR = arg('out');
if (!OUT_DIR) {
  console.error('usage: capture-pages.mjs --out <dir> [--only a,b] [--tag before] [--publish-og] [--base URL]');
  process.exit(2);
}
const ONLY = arg('only', '')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);
const TAG = arg('tag', '');
const PUBLISH_OG = flag('publish-og');
const COMPARE_DIR = arg('compare');

const siteRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const exportDir = path.resolve(arg('export', path.join(siteRoot, 'out')));
const ogDir = path.join(siteRoot, 'public/_site/images/og');

// The frozen user-agent strings browsers actually send; the page's detector
// reads these when Client Hints are absent (Firefox, Safari) and the hint
// platform when present (Chromium). Both paths are exercised below.
const UA = {
  mac: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15',
  windows: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36',
  linux: 'Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0',
  iphone: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1',
};

/**
 * One row per scene. `expectTab` is the platform tab the page must have
 * selected on load (asserted); `og` marks a 1200x630 first-screen capture
 * that is also the page's Open Graph image when --publish-og is given;
 * `ogTitleSize` overrides the poster's headline size for a long headline.
 */
const SCENES = [
  // The marketing pages a "zero visual change" commit is proven against: the
  // landing page, the Product, Distributions, Trust, and Solutions groups,
  // pricing, and the decks that ride the deck engine. Two review widths for
  // a group's first page, one for its siblings.
  { name: 'landing-1680', route: '/', width: 1680, ua: UA.mac },
  { name: 'landing-1280', route: '/', width: 1280, ua: UA.mac },
  { name: 'landing-phone', route: '/', width: 390, ua: UA.iphone },
  { name: 'product-1280', route: '/product', width: 1280, ua: UA.mac },
  { name: 'infra-hub-1680', route: '/product/infra-hub', width: 1680, ua: UA.mac },
  { name: 'infra-hub-1280', route: '/product/infra-hub', width: 1280, ua: UA.mac },
  { name: 'service-hub-1280', route: '/product/service-hub', width: 1280, ua: UA.mac },
  { name: 'coding-agents-1280', route: '/product/coding-agents', width: 1280, ua: UA.mac },
  { name: 'cli-1280', route: '/product/cli', width: 1280, ua: UA.mac },
  { name: 'catalog-1280', route: '/product/catalog', width: 1280, ua: UA.mac },
  { name: 'import-1280', route: '/product/import', width: 1280, ua: UA.mac },
  { name: 'open-source-1280', route: '/product/open-source', width: 1280, ua: UA.mac },
  { name: 'distributions-1280', route: '/distributions', width: 1280, ua: UA.mac },
  { name: 'hosted-1680', route: '/distributions/hosted', width: 1680, ua: UA.mac },
  { name: 'hosted-1280', route: '/distributions/hosted', width: 1280, ua: UA.mac },
  { name: 'self-hosted-1280', route: '/distributions/self-hosted', width: 1280, ua: UA.mac },
  { name: 'solutions-1680', route: '/solutions', width: 1680, ua: UA.mac },
  { name: 'solutions-1280', route: '/solutions', width: 1280, ua: UA.mac },
  { name: 'platform-engineer-1680', route: '/solutions/platform-engineer', width: 1680, ua: UA.mac },
  { name: 'platform-engineer-1280', route: '/solutions/platform-engineer', width: 1280, ua: UA.mac },
  { name: 'platform-engineer-phone', route: '/solutions/platform-engineer', width: 390, ua: UA.iphone },
  { name: 'engineering-leader-1680', route: '/solutions/engineering-leader', width: 1680, ua: UA.mac },
  { name: 'engineering-leader-1280', route: '/solutions/engineering-leader', width: 1280, ua: UA.mac },
  { name: 'it-consultancy-1280', route: '/solutions/it-consultancy', width: 1280, ua: UA.mac },
  { name: 'startup-founder-1280', route: '/solutions/startup-founder', width: 1280, ua: UA.mac },
  { name: 'security-leader-1280', route: '/solutions/security-and-governance-leader', width: 1280, ua: UA.mac },
  // The persona decks: each cover, and the roadmap slide of one deck by its
  // hash so the disclosure line is in a capture. The engine reads the hash
  // on load, inside the virtual clock.
  { name: 'deck-platform-engineer-1280', route: '/decks/platform-engineer', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-platform-engineer-rules-1280', route: '/decks/platform-engineer#your-rules-hold', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-platform-engineer-next-1280', route: '/decks/platform-engineer#what-is-next', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-engineering-leader-1280', route: '/decks/engineering-leader', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-it-consultancy-1280', route: '/decks/it-consultancy', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-startup-founder-1280', route: '/decks/startup-founder', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-security-leader-1280', route: '/decks/security-and-governance-leader', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'pricing-1680', route: '/pricing', width: 1680, ua: UA.mac },
  { name: 'pricing-1280', route: '/pricing', width: 1280, ua: UA.mac },
  { name: 'trust-1680', route: '/trust', width: 1680, ua: UA.mac },
  { name: 'trust-1280', route: '/trust', width: 1280, ua: UA.mac },
  { name: 'trust-verified-1680', route: '/trust/verified-before-deploy', width: 1680, ua: UA.mac },
  { name: 'trust-verified-1280', route: '/trust/verified-before-deploy', width: 1280, ua: UA.mac },
  { name: 'trust-rules-1280', route: '/trust/rules-and-approvals', width: 1280, ua: UA.mac },
  { name: 'trust-record-1280', route: '/trust/the-record', width: 1280, ua: UA.mac },
  { name: 'trust-posture-1280', route: '/trust/security-posture', width: 1280, ua: UA.mac },
  { name: 'trust-keys-1280', route: '/trust/your-cloud-your-keys', width: 1280, ua: UA.mac },
  { name: 'deck-sep-1280', route: '/meets/sep', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-nirav-1280', route: '/meets/nirav', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-clear-route-1280', route: '/meets/clear-route', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'deck-rahul-gulati-1280', route: '/meets/rahul-gulati', width: 1280, height: 800, ua: UA.mac, viewportOnly: true },
  { name: 'download-1680', route: '/features/desktop/download', width: 1680, ua: UA.mac, expectTab: 'macOS' },
  { name: 'download-1280', route: '/features/desktop/download', width: 1280, ua: UA.mac, expectTab: 'macOS' },
  { name: 'download-windows', route: '/features/desktop/download', width: 1280, ua: UA.windows, uaPlatform: 'Windows', expectTab: 'Windows' },
  { name: 'download-linux', route: '/features/desktop/download', width: 1280, ua: UA.linux, expectTab: 'Linux' },
  { name: 'download-phone', route: '/features/desktop/download', width: 390, ua: UA.iphone, expectTab: 'macOS' },
  { name: 'download-og', route: '/features/desktop/download', width: 1200, height: 630, ua: UA.mac, og: 'download.png' },
  { name: 'desktop-1680', route: '/features/desktop', width: 1680, ua: UA.mac },
  { name: 'desktop-1280', route: '/features/desktop', width: 1280, ua: UA.mac },
  { name: 'desktop-phone', route: '/features/desktop', width: 390, ua: UA.iphone },
  { name: 'desktop-og', route: '/features/desktop', width: 1200, height: 630, ua: UA.mac, og: 'desktop.png', ogTitleSize: 44 },
];

// ---------------------------------------------------------------------------
// A static server for out/: enough of one to render the export the way GitHub
// Pages does (route -> route.html or route/index.html; /_site/_next assets).
// ---------------------------------------------------------------------------
const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.woff2': 'font/woff2',
  '.woff': 'font/woff',
  '.ico': 'image/x-icon',
  '.txt': 'text/plain',
  '.xml': 'application/xml',
};

function resolveFile(urlPath) {
  const clean = decodeURIComponent(urlPath.split('?')[0]).replace(/\/+$/, '') || '/';
  const candidates =
    clean === '/'
      ? [path.join(exportDir, 'index.html')]
      : [path.join(exportDir, clean), path.join(exportDir, `${clean}.html`), path.join(exportDir, clean, 'index.html')];
  for (const c of candidates) {
    if (c.startsWith(exportDir) && fs.existsSync(c) && fs.statSync(c).isFile()) return c;
  }
  return null;
}

function serveExport() {
  if (!fs.existsSync(exportDir)) {
    console.error(`no static export at ${exportDir}; run \`make build\` first`);
    process.exit(2);
  }
  const server = http.createServer((req, res) => {
    const file = resolveFile(req.url ?? '/');
    if (!file) {
      res.writeHead(404);
      res.end('not found');
      return;
    }
    res.writeHead(200, { 'content-type': MIME[path.extname(file)] ?? 'application/octet-stream' });
    fs.createReadStream(file).pipe(res);
  });
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => resolve({ server, base: `http://127.0.0.1:${server.address().port}` }));
  });
}

// ---------------------------------------------------------------------------

/**
 * Put the page on a virtual clock with a budget of VIRTUAL_TIME_BUDGET_MS.
 * Timers, animation frames, and JS-driven transitions all run on it, and the
 * clock pauses while network fetches are pending, so the page reaches the
 * same virtual instant on every run regardless of the machine. `finish()`
 * waits for the budget to run out, then freezes the clock for the screenshot.
 */
async function startVirtualTime(page) {
  const client = await page.createCDPSession();
  const expired = new Promise((resolve) => client.once('Emulation.virtualTimeBudgetExpired', resolve));
  await client.send('Emulation.setVirtualTimePolicy', {
    policy: 'pauseIfNetworkFetchesPending',
    budget: VIRTUAL_TIME_BUDGET_MS,
    maxVirtualTimeTaskStarvationCount: 1_000_000,
  });
  const started = Date.now();
  return {
    async finish() {
      const outcome = await Promise.race([
        expired.then(() => 'expired'),
        new Promise((r) => setTimeout(() => r('timed-out'), 90000)),
      ]);
      await client.send('Emulation.setVirtualTimePolicy', { policy: 'pause' });
      await client.detach();
      if (outcome !== 'expired') {
        throw new Error(`virtual time did not reach its budget in ${Date.now() - started}ms of real time; the capture is not deterministic`);
      }
    },
  };
}

/** Scroll the document in viewport-sized steps (real time), then return to the top. */
async function scrollThrough(page) {
  await page.evaluate(async () => {
    // Half a viewport per step, and a real pause at each, so every
    // IntersectionObserver (framer-motion's whileInView) is delivered and its
    // entrance started before the next step; otherwise a reveal can be left
    // at opacity 0 when the clock freezes.
    const step = Math.max(200, Math.floor(window.innerHeight / 2));
    const settle = () => new Promise((r) => setTimeout(r, 80));
    for (let y = 0; y < document.body.scrollHeight; y += step) {
      window.scrollTo(0, y);
      await settle();
    }
    window.scrollTo(0, document.body.scrollHeight);
    await settle();
    window.scrollTo(0, 0);
    await settle();
  });
  await page.evaluate(() => document.fonts.ready);
}

/** Diff one capture against its counterpart in the before set; returns the mismatch count or null when no before exists. */
function compareWithBefore(scene, afterFile) {
  const beforeFile = path.join(COMPARE_DIR, `${scene.name}-dark.png`);
  if (!fs.existsSync(beforeFile)) return { status: 'no-before' };
  const before = PNG.sync.read(fs.readFileSync(beforeFile));
  const after = PNG.sync.read(fs.readFileSync(afterFile));
  if (before.width !== after.width || before.height !== after.height) {
    return { status: 'size', detail: `${before.width}x${before.height} -> ${after.width}x${after.height}` };
  }
  const diff = new PNG({ width: before.width, height: before.height });
  const mismatched = pixelmatch(before.data, after.data, diff.data, before.width, before.height, { threshold: 0.1 });
  if (mismatched > 0) {
    const diffFile = path.join(OUT_DIR, `${scene.name}-diff.png`);
    fs.writeFileSync(diffFile, PNG.sync.write(diff));
    return { status: 'differs', detail: `${mismatched} px, see ${path.relative(process.cwd(), diffFile)}` };
  }
  return { status: 'identical' };
}

async function capture(browser, base, scene) {
  const page = await browser.newPage();
  await page.setViewport({ width: scene.width, height: scene.height ?? 900, deviceScaleFactor: 1 });
  await page.setUserAgent(scene.ua);
  if (scene.uaPlatform) {
    // Chromium exposes navigator.userAgentData; give the page the hint a real
    // Chromium on that platform would send, so the Client Hints path is tested.
    await page.evaluateOnNewDocument((platform) => {
      Object.defineProperty(navigator, 'userAgentData', { value: { platform }, configurable: true });
    }, scene.uaPlatform);
  }
  await page.emulateMediaFeatures([
    { name: 'prefers-color-scheme', value: 'dark' },
    { name: 'prefers-reduced-motion', value: 'reduce' },
  ]);
  await page.goto(base + scene.route, { waitUntil: 'networkidle0', timeout: 45000 });
  // CSS keyframe animations run on the compositor thread and ignore the virtual
  // clock (a pulsing cursor was the one moving pixel between two identical
  // builds). Stop them at their resting value; JS-driven motion still runs
  // under virtual time and settles deterministically.
  await page.addStyleTag({ content: '*, *::before, *::after { animation: none !important; }' });
  // Full-page screenshots reveal the whole document at once, which fires every
  // scroll-triggered entrance mid-frame. Walk the page first (real time) so
  // those observers fire in order, then let virtual time finish what they
  // started. A component that starts a long JS timeline on mount (the v4
  // hero's typing loop) is offset by however long hydration took, which can
  // differ between two builds; a page must reach its resting frame without
  // depending on wall-clock phase (honor prefers-reduced-motion) to be
  // comparable here.
  if (!scene.og && !scene.viewportOnly) await scrollThrough(page);
  const clock = await startVirtualTime(page);
  await clock.finish();
  if (scene.og) {
    // A share image is a poster, not a viewport: no site chrome, no clipped
    // card, the page's own headline block centred in the 1200x630 frame.
    await page.addStyleTag({
      content: `
        header, main [class*="sticky"] { display: none !important; }
        main { padding-top: 0 !important; }
        main section:first-of-type { min-height: 630px; display: flex; align-items: center; padding: 0 !important; }
        main section:first-of-type .mb-10 { margin-bottom: 0 !important; }
        main section:first-of-type h1 { font-size: ${scene.ogTitleSize ?? 64}px !important; line-height: 1.15 !important; margin-bottom: 24px !important; max-width: 1000px !important; }
        main section:first-of-type h1 + p { font-size: 24px !important; line-height: 1.4 !important; max-width: 820px !important; }
        main section:first-of-type h1 + p + p { font-size: 16px !important; margin-top: 24px !important; }
        main section:first-of-type a { text-decoration: none !important; }
        main section:first-of-type img, main section:first-of-type div:has(> img) { display: none !important; }
        main section:first-of-type div:has(> a), main section:first-of-type div:has(> code) { display: none !important; }
        [role="tablist"], div:has(> [role="tabpanel"]) { display: none !important; }
      `,
    });
  }
  let verdict = null;
  if (scene.expectTab) {
    const selected = await page.$eval('[role="tab"][aria-selected="true"]', (el) => el.textContent?.trim()).catch(() => null);
    verdict = selected === scene.expectTab ? `ok (${selected})` : `FAIL expected ${scene.expectTab}, got ${selected}`;
  }

  const suffix = TAG ? `-${TAG}` : '';
  const file = path.join(OUT_DIR, `${scene.name}${suffix}-dark.png`);
  await page.screenshot({ path: file, fullPage: !scene.og && !scene.viewportOnly });
  if (scene.og && PUBLISH_OG) {
    fs.mkdirSync(ogDir, { recursive: true });
    fs.copyFileSync(file, path.join(ogDir, scene.og));
  }
  await page.close();
  return { file, verdict };
}

async function main() {
  const scenes = ONLY.length ? SCENES.filter((s) => ONLY.includes(s.name)) : SCENES;
  if (!scenes.length) {
    console.error(`no scenes match --only ${ONLY.join(',')}; known: ${SCENES.map((s) => s.name).join(', ')}`);
    process.exit(2);
  }
  fs.mkdirSync(OUT_DIR, { recursive: true });

  const external = arg('base');
  const served = external ? null : await serveExport();
  const base = external ?? served.base;
  const browser = await puppeteer.launch({ headless: true, args: ['--no-sandbox'] });

  let failed = 0;
  let differs = 0;
  try {
    for (const scene of scenes) {
      const { file, verdict } = await capture(browser, base, scene);
      if (verdict?.startsWith('FAIL')) failed += 1;
      let comparison = '';
      if (COMPARE_DIR) {
        const result = compareWithBefore(scene, file);
        comparison = `  [${result.status}${result.detail ? `: ${result.detail}` : ''}]`;
        if (result.status === 'differs' || result.status === 'size') differs += 1;
      }
      console.log(`${path.relative(process.cwd(), file)}${verdict ? `  ${verdict}` : ''}${comparison}`);
    }
  } finally {
    await browser.close();
    served?.server.close();
  }
  if (failed) {
    console.error(`${failed} scene assertion(s) failed`);
    process.exit(1);
  }
  if (differs) {
    console.error(`${differs} scene(s) differ from ${COMPARE_DIR}; this is not a zero-visual-change build`);
    process.exit(1);
  }
  if (COMPARE_DIR) console.log(`all ${scenes.length} scenes identical to ${COMPARE_DIR}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
