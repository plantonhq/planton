/** Repeatable local lab evidence, not field Core Web Vitals. No external requests. */
import fs from 'node:fs';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';
const server = createPreviewServer();
await new Promise((r) => server.listen(0, '127.0.0.1', r));
const browser = await puppeteer.launch({ headless: true });
const runs = [];
try {
  for (const width of [1366, 390])
    for (let run = 0; run < 3; run++) {
      const page = await browser.newPage();
      await page.setViewport({ width, height: width === 390 ? 844 : 768 });
      await page.setRequestInterception(true);
      page.on('request', (r) =>
        r.url().startsWith('http://127.0.0.1:') || r.url().startsWith('data:')
          ? r.continue()
          : r.respond({ status: 200, body: '' })
      );
      await page.evaluateOnNewDocument(() => {
        window.lab = { lcp: 0, cls: 0, interactions: [] };
        new PerformanceObserver((l) => {
          for (const e of l.getEntries()) window.lab.lcp = e.startTime;
        }).observe({ type: 'largest-contentful-paint', buffered: true });
        new PerformanceObserver((l) => {
          for (const e of l.getEntries()) if (!e.hadRecentInput) window.lab.cls += e.value;
        }).observe({ type: 'layout-shift', buffered: true });
        new PerformanceObserver((l) => {
          for (const e of l.getEntries())
            if (e.interactionId) window.lab.interactions.push(e.duration);
        }).observe({ type: 'event', durationThreshold: 16, buffered: true });
      });
      await page.goto(`http://127.0.0.1:${server.address().port}`, { waitUntil: 'networkidle0' });
      await page.evaluate(() => document.fonts.ready);
      const initial = await page.evaluate(() => ({
        ...window.lab,
        jsBytes: performance
          .getEntriesByType('resource')
          .filter((r) => r.initiatorType === 'script')
          .reduce((n, r) => n + r.decodedBodySize, 0),
      }));
      await page.click('a[href="#how-it-works"]');
      const faq = await page.$('section[aria-labelledby=faq-title] details summary');
      await faq.scrollIntoView();
      await faq.click();
      await new Promise((r) => setTimeout(r, 100));
      runs.push({
        width,
        run,
        ...initial,
        interactionMax: await page.evaluate(() => Math.max(0, ...window.lab.interactions)),
      });
      await page.close();
    }
  fs.writeFileSync(
    process.argv[2] ?? '/tmp/homepage-performance.json',
    JSON.stringify(
      {
        environment:
          'Local headless Chrome; warm OS cache, no network/CPU throttling, external scripts blocked. Event Timing sample is not field INP.',
        runs,
      },
      null,
      2
    )
  );
  console.log(JSON.stringify(runs));
} finally {
  await browser.close();
  await new Promise((r) => server.close(r));
}
