/** Social card from the real hero scene. Poster-only layout hides site chrome,
 * preserves the light page palette, and uses the completed animation state. */
import fs from 'node:fs';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';
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
  await page.addStyleTag({
    content: `
    header, footer, main section:not(:first-child) {display:none!important}
    [data-homepage-appearance] {position:fixed!important;inset:0!important;width:1200px;height:630px;background:var(--hp-canvas)!important}
    main>div {padding:28px!important;max-width:none!important}
    main section:first-child {display:grid!important;grid-template-columns:1fr 1.1fr!important;gap:30px!important;padding:0!important;height:574px!important;max-width:none!important}
    main section:first-child h1 {font-size:48px!important}
    main section:first-child>figure {width:100%!important;max-width:none!important}
    [data-hero-phase] figcaption {display:none!important}
    [class*="intro"] {font-size:16px!important}
  `,
  });
  await page.evaluate(() => {
    const label = document.querySelector('[class*="eyebrow"]');
    label.textContent = 'PLANTON / THE SELF-SERVICE CLOUD PLATFORM';
  });
  fs.mkdirSync('public/_site/images/og', { recursive: true });
  await page.screenshot({ path: 'public/_site/images/og/homepage.png' });
  console.log('Wrote public/_site/images/og/homepage.png (1200 × 630)');
} finally {
  await browser.close();
  await new Promise((resolve) => server.close(resolve));
}
