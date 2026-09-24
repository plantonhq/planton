/** Real exported-page acceptance. The test owns time; no testing API ships to
 * visitors. Network calls are mocked and cannot submit a lead or book a meeting. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';

const output = process.env.WORKFLOW_CAPTURES ?? path.join(process.env.HOMEPAGE_CAPTURES ?? os.tmpdir(), 'planton-workflow-review');
fs.mkdirSync(output, { recursive: true });
const server = createPreviewServer();
await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
const base = `http://127.0.0.1:${server.address().port}`;
const browser = await puppeteer.launch({ headless: true });
const page = await browser.newPage();
const errors = [];
const checks = [];
const check = (name, value) => { assert.ok(value, name); checks.push(name); };
page.on('pageerror', error => errors.push(error.message));
await page.setRequestInterception(true);
page.on('request', request => request.url().startsWith(base) || request.url().startsWith('data:')
  ? request.continue()
  : request.respond({ status: 200, contentType: request.resourceType() === 'script' ? 'text/javascript' : 'text/plain', body: '' }));
await page.evaluateOnNewDocument(() => {
  let now = 0, next = 0;
  const nativeFrame = window.requestAnimationFrame.bind(window);
  const painted = () => new Promise(resolve => nativeFrame(() => nativeFrame(resolve)));
  const callbacks = new Map();
  Object.defineProperty(performance, 'now', { value: () => now });
  window.requestAnimationFrame = callback => { callbacks.set(++next, callback); return next; };
  window.cancelAnimationFrame = id => callbacks.delete(id);
  window.advanceWorkflowTime = async milliseconds => {
    await painted();
    now += milliseconds;
    const pending = [...callbacks.values()];
    callbacks.clear();
    pending.forEach(callback => callback(now));
    await painted();
  };
});
// Poll on wall time: requestAnimationFrame belongs to the test clock.
const waitFor = predicate => page.waitForFunction(predicate, { polling: 25 });
const advance = async ms => { await page.evaluate(ms => window.advanceWorkflowTime(ms), ms); };
const capture = async (selector, filename) => {
  await page.$eval('header', e => { e.style.visibility = 'hidden'; });
  await (await page.$(selector)).screenshot({ path: path.join(output, filename) });
  await page.$eval('header', e => { e.style.visibility = ''; });
};
const providers = ['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean'];
const stories = [...providers, 'delivery', 'agents'];
const select = async id => {
  if (providers.includes(id)) {
    await page.click(`[role="tab"]:nth-child(${providers.indexOf(id) + 1})`);
    await waitFor(() => !document.querySelector('[inert]'));
  }
  await page.mouse.move(0, 0);
};
const expose = async selector => {
  await page.$eval(selector, e => window.scrollTo({ top: scrollY + e.getBoundingClientRect().top - 88, behavior: 'instant' }));
  await advance(0);
};
try {
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  for (const width of [320, 390, 768, 1280, 1680]) {
    await page.setViewport({ width, height: 1000 });
    await page.goto(base, { waitUntil: 'networkidle0' });
    await page.evaluate(() => document.fonts.ready);
    for (const id of stories) {
      await select(id);
      const selector = `[data-workflow="${id}"]`;
      check(`${width}/${id}: static completed diagram`, await page.$eval(selector, e => e.dataset.playing === 'false' && [...e.querySelectorAll('[data-node]')].every(n => n.dataset.state === 'complete')));
      check(`${width}/${id}: readable bounds`, await page.$eval(selector, e => {
        const bounds = e.getBoundingClientRect();
        return bounds.left >= 0 && bounds.right <= innerWidth && [...e.querySelectorAll('svg[role="img"] text')].filter(t => t.getBoundingClientRect().width).every(t => {
          const r = t.getBoundingClientRect(); return r.left >= bounds.left && r.right <= bounds.right;
        });
      }));
      check(`${width}/${id}: labels fit cards`, await page.$eval(selector, e => [...e.querySelectorAll('[data-node]')].filter(n => n.getBoundingClientRect().width).every(n => {
        const box = n.querySelector('rect').getBoundingClientRect();
        return [...n.querySelectorAll('text')].every(t => { const r = t.getBoundingClientRect(); return r.left >= box.left && r.right <= box.right && r.bottom <= box.bottom; });
      })));
      if (providers.includes(id)) check(`${width}/${id}: icons embedded`, await page.$eval(selector, e => [...e.querySelectorAll('[data-node] image')].every(i => i.getAttribute('href').startsWith('data:image/svg+xml;base64,'))));
      await capture(selector, `${id}-${width}-static.png`);
    }
    check(`${width}: reduced motion disables cycle`, await page.$eval('[data-architectures]', e => e.dataset.cycling === 'false'));
  }
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'no-preference' }]);
  for (const [width, height] of [[1366, 768], [1440, 900], [1920, 1080]]) {
    await page.setViewport({ width, height });
    await page.goto(base, { waitUntil: 'networkidle0' });
    const heights = [];
    for (const id of stories) {
      await select(id);
      const selector = `[data-workflow="${id}"]`;
      await page.click(`${selector} button[aria-label^="Inspect step 3:"]`);
      await expose(providers.includes(id) ? '[data-architectures]' : selector);
      check(`${width}x${height}/${id}: whole component fits`, await page.$eval(providers.includes(id) ? '[data-architectures]' : selector, e => e.getBoundingClientRect().bottom <= innerHeight));
      if (providers.includes(id)) check(`${width}/${id}: desktop resource labels stay readable`, await page.$eval(`${selector} [data-node] text`, text => Number(text.getAttribute('font-size')) * text.getScreenCTM().a >= 14));
      check(`${width}/${id}: controls fit`, await page.$eval(selector, e => e.scrollWidth <= e.clientWidth));
      if (providers.includes(id)) heights.push(await page.$eval('[data-architectures]', e => e.getBoundingClientRect().height));
      await page.screenshot({ path: path.join(output, `${id}-${width}x${height}-viewport.png`) });
    }
    check(`${width}: provider tabs preserve height`, Math.max(...heights) - Math.min(...heights) < 1);
  }
  await page.setViewport({ width: 1440, height: 1000 });
  await page.goto(base, { waitUntil: 'networkidle0' });
  await expose('[data-architectures]');
  await page.mouse.move(0, 0);
  await waitFor(() => document.querySelector('[data-workflow="aws"]').dataset.playing === 'true');
  for (const [current, next] of providers.map((id, index) => [id, providers[(index + 1) % providers.length]])) {
    await advance(24000);
    check(`${current}: final hold before advance`, await page.$eval('[data-architectures]', (e, current) => e.dataset.selected === current, current));
    await advance(3900);
    check(`${current}: full four-second hold`, await page.$eval('[data-architectures]', (e, current) => e.dataset.selected === current, current));
    await advance(101);
    check(`${current}: advances to ${next}`, await page.$eval('[data-architectures]', (e, next) => e.dataset.selected === next, next));
    await waitFor(() => !document.querySelector('[inert]'));
    await advance(0);
  }
  check('SVG ids are isolated', await page.$$eval('svg [id]', elements => new Set(elements.map(e => e.id)).size === elements.length));
  const awsLogo = await page.$eval('[role="tab"]:first-child img', image => image.getAttribute('src'));
  check('AWS tab uses dark lettering on the light surface', Buffer.from(awsLogo.split(',')[1], 'base64').toString().includes('fill="#232F3E"'));
  check('cycle never moves keyboard focus', await page.evaluate(() => document.activeElement === document.body));
  await page.click('[role="tab"]:nth-child(2)');
  await page.click('[role="tab"]:nth-child(4)');
  await page.click('[role="tab"]:nth-child(5)');
  await waitFor(() => !document.querySelector('[inert]'));
  check('rapid selection leaves one active provider', await page.$eval('[data-architectures]', e => e.dataset.selected === 'digitalocean' && e.querySelectorAll('[data-workflow]').length === 1));
  await select('gcp');
  check('manual selection disables cycle', await page.$eval('[data-architectures]', e => e.dataset.cycling === 'false'));
  await expose('[data-architectures]');
  await advance(29000);
  check('manual story completes and holds', await page.$eval('[data-workflow="gcp"]', e => e.dataset.phase === '5' && e.dataset.playing === 'false'));
  await advance(60000);
  check('manual story does not rotate', await page.$eval('[data-architectures]', e => e.dataset.selected === 'gcp'));

  for (const id of stories) {
    await select(id);
    const selector = `[data-workflow="${id}"]`;
    await expose(selector);
    await page.click(`${selector} button[aria-label^="Replay:"]`);
    await page.mouse.move(0, 0);
    await advance(8500);
    check(`${id}: deterministic replay`, await page.$eval(selector, e => e.dataset.phase === '2'));
    await page.click(`${selector} button[aria-label^="Pause:"]`);
    await page.mouse.move(0, 0);
    const paused = await page.$eval(selector, e => e.innerHTML);
    await advance(15000);
    check(`${id}: pause freezes frame`, paused === await page.$eval(selector, e => e.innerHTML));
    const heights = [];
    for (let phase = 0; phase < 6; phase++) {
      await page.focus(`${selector} button[aria-label^="Inspect step ${phase + 1}:"]`);
      await page.keyboard.press('Enter');
      check(`${id}/${phase}: keyboard phase inspection`, await page.$eval(selector, (e, phase) => e.dataset.phase === String(phase) && e.dataset.playing === 'false', phase));
      heights.push(await page.$eval(selector, e => e.getBoundingClientRect().height));
      if (id === 'agents' && phase === 2) check('agent review blocks execution', await page.$eval(`${selector} [data-node="planton"]`, e => e.dataset.state === 'waiting'));
      if (id === 'delivery' && phase === 3) check('production waits for approval', await page.$eval(`${selector} [data-node="production"]`, e => e.dataset.state === 'waiting'));
      await capture(selector, `${id}-phase-${phase + 1}.png`);
    }
    check(`${id}: notes do not shift layout`, Math.max(...heights) - Math.min(...heights) < 1);
    if (providers.includes(id)) {
      await page.click(`${selector} details > summary`);
      check(`${id}: inventory has supporting resources`, await page.$$eval(`${selector} [data-resource]`, nodes => nodes.length > 6));
      await page.select(`${selector} [data-resource-explorer] select`, id === 'gcp' ? 'app' : id === 'digitalocean' ? 'servers' : 'api');
      check(`${id}: selected dependencies readable`, await page.$eval(`${selector} [aria-live]`, e => e.textContent.includes('Prerequisites:')));
      await capture(`${selector} [data-resource-explorer]`, `${id}-focused-resources.png`);
      await page.click(`${selector} details > summary`);
    }
  }
  await page.goto(base, { waitUntil: 'networkidle0' });
  await expose('[data-architectures]');
  await page.mouse.move(0, 0);
  await advance(4500);
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }));
  await waitFor(() => document.querySelector('[data-workflow="aws"]').dataset.playing === 'false');
  await advance(60000);
  check('offscreen time excluded', await page.$eval('[data-workflow="aws"]', e => e.dataset.phase === '1'));
  await expose('[data-architectures]');
  await page.evaluate(() => { Object.defineProperty(document, 'hidden', { configurable: true, value: true }); document.dispatchEvent(new Event('visibilitychange')); });
  await advance(60000);
  check('hidden browser time excluded', await page.$eval('[data-workflow="aws"]', e => e.dataset.phase === '1'));
  await page.evaluate(() => { delete document.hidden; document.dispatchEvent(new Event('visibilitychange')); });
  await advance(0);
  await page.hover('[data-workflow="aws"] h3');
  await advance(4000);
  check('hover does not interrupt playback', await page.$eval('[data-workflow="aws"]', e => e.dataset.phase === '2'));
  await page.mouse.move(0, 0);
  await page.hover('[role="tab"][aria-selected="true"]');
  await advance(4000);
  check('hovering provider tabs does not interrupt playback', await page.$eval('[data-workflow="aws"]', e => e.dataset.phase === '3'));
  await page.mouse.move(0, 0);
  await page.focus('[role="tab"][aria-selected="true"]');
  check('keyboard focus stops cycling', await page.$eval('[data-architectures]', e => e.dataset.cycling === 'false'));
  await page.keyboard.press('ArrowRight');
  check('arrow key changes tab and focus', await page.evaluate(() => document.activeElement.textContent === 'GCP' && document.activeElement.getAttribute('aria-selected') === 'true'));
  await page.keyboard.press('End');
  check('End selects final provider', await page.$eval('[data-architectures]', e => e.dataset.selected === 'digitalocean'));
  await page.keyboard.press('Home');
  check('Home selects AWS', await page.$eval('[data-architectures]', e => e.dataset.selected === 'aws'));
  await waitFor(() => !document.querySelector('[inert]'));
  await page.click('[data-workflow="aws"] button[aria-label^="Pause:"]');
  await select('azure');
  check('paused state survives tab selection', await page.$eval('[data-workflow="azure"]', e => e.dataset.playing === 'false'));
  await page.click('[data-workflow="azure"] [data-auto-cycle]');
  await page.mouse.move(0, 0);
  check('explicit Auto-Cycle enables rotation', await page.$eval('[data-architectures]', e => e.dataset.cycling === 'true'));
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  await advance(60000);
  check('runtime reduced motion stops rotation', await page.$eval('[data-architectures]', e => e.dataset.cycling === 'false' && e.dataset.selected === 'azure'));
  const markdown = fs.readFileSync('out/index.md', 'utf8');
  check('all provider stories in Markdown', ['AwsBedrockKnowledgeBase', 'KubernetesQdrant', 'AzureContainerAppJob', 'CloudflareQueue', 'DigitalOceanDropletAutoscalePool'].every(kind => markdown.includes(kind)));
  check('no private paths in Markdown', !markdown.includes('/Users/'));
  await page.setJavaScriptEnabled(false);
  await page.goto(base, { waitUntil: 'networkidle0' });
  check('no-JS all seven stories survive', await page.$$eval('[data-workflow]', figures => figures.length === 7 && figures.every(e => e.querySelectorAll('details > ol > li').length === 6)));
  check('no-JS five resource inventories', await page.$$eval('[data-resource-explorer]', items => items.length === 5));
  check('no application errors', errors.length === 0);
  fs.writeFileSync(path.join(output, 'acceptance.json'), JSON.stringify({ passed: checks.length, checks, errors }, null, 2));
  console.log(`PASS: ${checks.length} workflow checks; captures: ${output}`);
} catch (error) {
  console.error('Last checks:', checks.slice(-8), 'Application errors:', errors);
  throw error;
} finally {
  await browser.close();
  await new Promise(resolve => server.close(resolve));
}
