/** Visual and behavioral acceptance for the homepage story. External traffic is blocked. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';
const output =
  process.env.EXPERIENCE_CAPTURES ??
  path.join(process.env.HOMEPAGE_CAPTURES ?? os.tmpdir(), 'planton-story-review');
fs.mkdirSync(output, { recursive: true });
const server = createPreviewServer();
await new Promise((r) => server.listen(0, '127.0.0.1', r));
const base = `http://127.0.0.1:${server.address().port}`,
  browser = await puppeteer.launch({ headless: true }),
  page = await browser.newPage();
const checks = [],
  errors = [];
const check = (name, value) => {
  assert.ok(value, name);
  checks.push(name);
};
page.on('pageerror', (e) => errors.push(e.message));
await page.setRequestInterception(true);
page.on('request', (r) =>
  r.url().startsWith(base) || r.url().startsWith('data:')
    ? r.continue()
    : r.respond({ status: 200, body: '' })
);
await page.evaluateOnNewDocument(() => {
  let now = 0,
    next = 0;
  const callbacks = new Map(),
    frame = requestAnimationFrame.bind(window),
    paint = () => new Promise((r) => frame(() => frame(r)));
  Object.defineProperty(performance, 'now', { value: () => now });
  window.requestAnimationFrame = (cb) => {
    callbacks.set(++next, cb);
    return next;
  };
  window.cancelAnimationFrame = (id) => callbacks.delete(id);
  window.advance = async (ms) => {
    await paint();
    now += ms;
    const pending = [...callbacks.values()];
    callbacks.clear();
    pending.forEach((cb) => cb(now));
    await paint();
  };
});
const advance = (ms) => page.evaluate((ms) => window.advance(ms), ms);
const overflow = () => page.evaluate(() => document.documentElement.scrollWidth <= innerWidth);
try {
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  for (const [width, height] of [
    [320, 844],
    [390, 844],
    [768, 1024],
    [1366, 768],
    [1440, 900],
    [1920, 1080],
  ]) {
    await page.setViewport({ width, height });
    await page.goto(base, { waitUntil: 'networkidle0' });
    await page.evaluate(() => document.fonts.ready);
    check(`${width}: no page overflow`, await overflow());
    check(
      `${width}: hero complete with reduced motion`,
      await page.$eval('[data-hero-time]', (e) => Number(e.dataset.heroTime) >= 16)
    );
    check(
      `${width}: hero labels stay inside canvas`,
      await page.$eval('[data-hero-time]', (e) => {
        const box = e.getBoundingClientRect();
        return [...e.querySelectorAll('svg text')]
          .filter((t) => t.getBoundingClientRect().width)
          .every((t) => {
            const r = t.getBoundingClientRect();
            return r.left >= box.left && r.right <= box.right;
          });
      })
    );
    if (width >= 1366)
      check(
        `${width}: complete hero and CTA above fold`,
        await page.$eval('main section', (e) => e.getBoundingClientRect().bottom <= innerHeight)
      );
    check(
      `${width}: architecture image is uncropped`,
      await page.$eval('#living-architecture img', (e) => {
        const { width, height } = e.getBoundingClientRect();
        return (
          Math.abs(width / height - 2676 / 1708) < 0.01 &&
          getComputedStyle(e).objectFit === 'contain'
        );
      })
    );
    // Trigger native lazy loading before full-page and section evidence captures.
    await page.$eval('#living-architecture img', (e) => e.scrollIntoView());
    await page.waitForFunction(
      () => {
        const image = document.querySelector('#living-architecture img');
        return image.complete && image.naturalWidth === 2676;
      },
      { polling: 25 }
    );
    await page.$eval('#living-architecture img', (e) => e.decode());
    check(`${width}: complete product export loads`, true);
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({ path: path.join(output, `hero-${width}x${height}.png`) });
    await page.screenshot({ path: path.join(output, `homepage-${width}.png`), fullPage: true });
    if (width === 1366 || width === 390) {
      await page.$eval('header', (e) => {
        e.style.visibility = 'hidden';
      });
      await page.$eval('#living-architecture', (e) => e.scrollIntoView());
      await (
        await page.$('#living-architecture')
      ).screenshot({ path: path.join(output, `product-${width}.png`) });
      await (
        await page.$('#controls')
      ).screenshot({ path: path.join(output, `controls-${width}.png`) });
      await page.$eval('header', (e) => {
        e.style.visibility = '';
      });
    }
  }
  await page.setViewport({ width: 1366, height: 768 });
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'no-preference' }]);
  await page.goto(base, { waitUntil: 'networkidle0' });
  check(
    'homepage primary navigation is demo-led',
    await page.$eval(
      'header a[href="/book-demo"]',
      (e) => getComputedStyle(e).backgroundColor !== 'rgba(0, 0, 0, 0)'
    )
  );
  check(
    'homepage utilities removed from action row',
    await page.$$eval(
      'header a',
      (es) => !es.some((e) => e.textContent === 'Download' || e.textContent === 'Discord')
    )
  );
  check(
    'proof precedes explanations',
    await page.evaluate(() => {
      const ids = [...document.querySelectorAll('main section')].map((e) =>
        e.getAttribute('aria-labelledby')
      );
      return (
        ids.indexOf('proof-title') < ids.indexOf('infrastructure-title') &&
        ids.indexOf('architecture-title') < ids.indexOf('delivery-title')
      );
    })
  );
  await advance(4500);
  check('hero advances', await page.$eval('[data-hero-phase]', (e) => e.dataset.heroPhase === '1'));
  await page.hover('[data-hero-phase]');
  await advance(4000);
  check(
    'hover does not pause hero',
    await page.$eval('[data-hero-phase]', (e) => e.dataset.heroPhase === '2')
  );
  await page.click('[data-hero-phase] button[aria-label^="Pause"]');
  const paused = await page.$eval('[data-hero-time]', (e) => e.dataset.heroTime);
  await advance(5000);
  check(
    'explicit pause freezes hero',
    (await page.$eval('[data-hero-time]', (e) => e.dataset.heroTime)) === paused
  );
  await page.click('[data-hero-phase] button[aria-label^="Play"]');
  await advance(12000);
  check(
    'hero finishes once',
    await page.$eval('[data-hero-time]', (e) => Number(e.dataset.heroTime) === 20)
  );
  await advance(40000);
  check(
    'completed hero holds',
    await page.$eval('[data-hero-time]', (e) => Number(e.dataset.heroTime) === 20)
  );
  await page.click('[data-hero-phase] button[aria-label^="Replay"]');
  await advance(4500);
  check(
    'replay restarts a finished clock',
    await page.$eval(
      '[data-hero-time]',
      (e) => Number(e.dataset.heroTime) >= 4 && Number(e.dataset.heroTime) < 5
    )
  );
  await page.$eval('[data-hero-phase] button[aria-label^="Pause"]', (e) => e.click());
  await page.screenshot({ path: path.join(output, 'hero-animated-1366.png') });
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  // matchMedia change delivery is asynchronous; wait for the observable state
  // rather than racing the React update immediately after CDP emulation.
  await page.waitForFunction(() => Number(document.querySelector('[data-hero-time]').dataset.heroTime) === 20, { polling: 100 });
  check(
    'runtime reduced motion completes hero',
    await page.$eval('[data-hero-time]', (e) => Number(e.dataset.heroTime) === 20)
  );
  await page.evaluate(() => {
    document.documentElement.style.zoom = '2';
  });
  check('200 percent zoom has no horizontal overflow', await overflow());
  await page.screenshot({ path: path.join(output, 'zoom-200.png') });
  await page.evaluate(() => {
    document.documentElement.style.zoom = '';
  });
  for (const width of [320, 1366]) {
    await page.setViewport({ width, height: 768 });
    const opener = '#living-architecture figcaption a';
    const url = page.url();
    await page.click(opener);
    await page.waitForSelector('dialog[open]');
    check(`${width}: diagram opens without navigation`, page.url() === url);
    check(
      `${width}: viewer fills viewport`,
      await page.$eval('dialog', (e) => {
        const r = e.getBoundingClientRect();
        return Math.abs(r.width - innerWidth) < 1 && Math.abs(r.height - innerHeight) < 1;
      })
    );
    check(
      `${width}: close receives initial focus`,
      await page.$eval(
        '[aria-label="Close full-screen diagram"]',
        (e) => e === document.activeElement
      )
    );
    check(
      `${width}: modal locks page scroll`,
      await page.evaluate(() => document.body.style.overflow === 'hidden')
    );
    await page.$eval('dialog img', (e) => e.decode());
    await page.screenshot({ path: path.join(output, `viewer-${width}.png`) });
    await page.click('[aria-label="Zoom in"]');
    check(
      `${width}: zoom creates scrollable inspection surface`,
      await page.$eval(
        '[aria-label^="Architecture diagram."]',
        (e) => e.scrollWidth > e.clientWidth || e.scrollHeight > e.clientHeight
      )
    );
    for (let i = 0; i < 8; i++) {
      await page.keyboard.press('Tab');
      check(
        `${width}: modal contains keyboard focus ${i}`,
        await page.evaluate(() => Boolean(document.activeElement.closest('dialog')))
      );
    }
    await page.keyboard.press('Escape');
    await page.waitForSelector('dialog', { hidden: true });
    check(
      `${width}: escape restores focus`,
      await page.$eval(opener, (e) => e === document.activeElement)
    );
    check(
      `${width}: page scroll restored`,
      await page.evaluate(() => document.body.style.overflow !== 'hidden')
    );
    await page.click('#living-architecture figure > a');
    await page.waitForSelector('dialog[open]');
    check(
      `${width}: reopening resets zoom`,
      await page.$eval('[aria-label="Zoom out"]', (e) => e.disabled)
    );
    await page.click('[aria-label="Close full-screen diagram"]');
    await page.waitForSelector('dialog', { hidden: true });
  }
  await page.setJavaScriptEnabled(false);
  await page.goto(base, { waitUntil: 'networkidle0' });
  check(
    'no JS hero has all product roles',
    await page.$eval('[data-hero-time]', (e) =>
      ['Platform Engineers', 'Developers', 'Infrastructure + Delivery', 'Your Cloud'].every((t) =>
        e.textContent.includes(t)
      )
    )
  );
  check(
    'no JS has all product annotations',
    await page.$eval('main', (e) =>
      e.textContent.includes('Planton Platform and the Planton Operator')
    )
  );
  check(
    'no JS has all cloud stories',
    await page.$$eval('[data-workflow]', (es) =>
      ['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean'].every((id) =>
        es.some((e) => e.dataset.workflow === id)
      )
    )
  );
  check('no runtime errors', errors.length === 0);
  fs.writeFileSync(
    path.join(output, 'acceptance.json'),
    JSON.stringify({ checks, errors }, null, 2)
  );
  console.log(`PASS: ${checks.length} experience checks. Captures: ${output}`);
} finally {
  await browser.close();
  await new Promise((r) => server.close(r));
}
