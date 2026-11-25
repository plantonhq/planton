/** Social card from the real hero scene. Poster-only layout hides site chrome,
 * preserves the light page palette, and uses the completed animation state.
 * The header is hidden, so the card carries the product's name itself: the
 * header's own logo mark beside SITE.name, as text, above the hero. */
import fs from 'node:fs';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';
import { SITE } from '../src/data/site-pages.ts';
const server = createPreviewServer();
await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
const browser = await puppeteer.launch({ headless: true });
try {
  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 630, deviceScaleFactor: 1 });
  await page.setRequestInterception(true);
  page.on('request', (r) =>
    r.url().startsWith('http://127.0.0.1:') || r.url().startsWith('data:')
      ? r.continue()
      : r.respond({ status: 200, body: '' })
  );
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  await page.goto(`http://127.0.0.1:${server.address().port}`, { waitUntil: 'networkidle0' });
  await page.evaluate(() => document.fonts.ready);
  await page.evaluate((name) => {
    const mark = document.querySelector('header a[aria-label="Planton home"] svg')?.cloneNode(true);
    const brand = document.createElement('div');
    brand.setAttribute('data-og-brand', '');
    if (mark) brand.append(mark);
    brand.append(document.createTextNode(name));
    document.querySelector('main > div')?.prepend(brand);
  }, SITE.name);
  await page.addStyleTag({
    content: `
    header, footer, main section:not(:first-of-type) {display:none!important}
    [data-homepage-appearance] {position:fixed!important;inset:0!important;width:1200px;height:630px;background:var(--hp-canvas)!important}
    main>div {padding:28px!important;max-width:none!important}
    [data-og-brand] {display:flex;align-items:center;gap:10px;height:30px;margin-bottom:14px;font-size:22px;font-weight:500;letter-spacing:-0.02em;color:var(--hp-ink)}
    [data-og-brand] svg {width:26px;height:26px}
    main section:first-of-type {display:grid!important;grid-template-columns:1fr 1.1fr!important;gap:30px!important;padding:0!important;height:530px!important;max-width:none!important}
    main section:first-of-type h1 {font-size:48px!important}
    main section:first-of-type>figure {width:100%!important;max-width:none!important}
    [data-hero-phase] figcaption {display:none!important}
    [class*="intro"] {font-size:16px!important}
  `,
  });
  fs.mkdirSync('public/_site/images/og', { recursive: true });
  await page.screenshot({ path: 'public/_site/images/og/homepage.png' });
  console.log('Wrote public/_site/images/og/homepage.png (1200 × 630)');
} finally {
  await browser.close();
  await new Promise((resolve) => server.close(resolve));
}
