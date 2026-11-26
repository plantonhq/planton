/** Browser acceptance tests. All external traffic is mocked; never creates a lead or meeting. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';

const out = process.env.HOMEPAGE_CAPTURES ?? '/private/tmp/planton-homepage-review';
fs.mkdirSync(out, { recursive: true });
const server = createPreviewServer();
await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
const base = `http://127.0.0.1:${server.address().port}`;
const browser = await puppeteer.launch({ headless: true });
const results = [];
const check = (name, condition) => {
  assert.ok(condition, name);
  results.push(name);
};
const sdk = fs.readFileSync('node_modules/@calcom/embed-core/dist/embed/embed.js', 'utf8');
let mode = 'success',
  submissions = 0,
  lastPayload;
const page = await browser.newPage();
const errors = [];
page.on('pageerror', (e) => errors.push(e.message));
await page.setRequestInterception(true);
page.on('request', (request) => {
  const url = request.url();
  if (url.startsWith(base) || url.startsWith('data:') || url.startsWith('blob:'))
    return request.continue();
  if (url === 'https://webhooks.planton.ai/demo') {
    if (request.method() === 'OPTIONS')
      return request.respond({
        status: 204,
        headers: {
          'access-control-allow-origin': '*',
          'access-control-allow-methods': 'POST, OPTIONS',
          'access-control-allow-headers': 'content-type',
        },
      });
    submissions++;
    lastPayload = JSON.parse(request.postData());
    if (mode === 'network') return request.abort('failed');
    return request.respond({
      status: mode === 'failure' ? 500 : 200,
      headers: { 'access-control-allow-origin': '*' },
      contentType: 'application/json',
      body: JSON.stringify(mode === 'failure' ? { error: 'mock failure' } : { ok: true }),
    });
  }
  if (url === 'https://app.cal.com/embed/embed.js')
    return request.respond({ status: 200, contentType: 'text/javascript', body: sdk });
  if (url.includes('cal.com'))
    return request.respond({
      status: 200,
      contentType: 'text/html',
      body: '<!doctype html><html><body>Mock scheduling surface</body></html>',
    });
  return request.respond({
    status: 200,
    contentType: request.resourceType() === 'script' ? 'text/javascript' : 'text/plain',
    body: '',
  });
});
const events = async () =>
  page.evaluate(() =>
    Array.from(window.dataLayer ?? [], (x) => Array.from(x)).filter((x) => x[0] === 'event')
  );
try {
  for (const width of [320, 390, 768, 1280, 1680]) {
    await page.setViewport({ width, height: 900, deviceScaleFactor: 1 });
    await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
    await page.goto(base, { waitUntil: 'networkidle0' });
    await page.evaluate(() => document.fonts.ready);
    check(
      `${width}: light-only homepage`,
      (await page.$eval(
        '[data-homepage-appearance]',
        (e) => e.dataset.homepageAppearance === 'light'
      )) && !(await page.$('[aria-label="Light appearance"]'))
    );
    check(`${width}: one h1`, await page.$$eval('h1', (es) => es.length === 1));
    const fits = () =>
      page.evaluate(() => {
        const w = innerWidth;
        return [...document.querySelectorAll('main *')]
          .filter((e) => !(e instanceof SVGElement))
          .every((e) => {
            const r = e.getBoundingClientRect();
            if (r.width !== 0 && (r.left < -1 || r.right > w + 1)) return false;
            // Text can run past a box that itself fits (a long unbroken word); only a box that
            // scrolls by design (a code panel), or sits inside one, may hold more than it shows.
            for (let n = e; n && n !== document.body; n = n.parentElement) {
              if (/auto|scroll/.test(getComputedStyle(n).overflowX)) return true;
            }
            return e.clientWidth === 0 || e.scrollWidth <= e.clientWidth + 1;
          });
      });
    check(`${width}: no overflowing content`, await fits());
    // The clouds the page offers, read from its own picker, so a cloud is never added without its checks.
    const clouds = await page.$$eval('#infrastructure input[name="home-cloud"]', (inputs) =>
      inputs.map((input) => input.value)
    );
    // The picker is one row of segments, or two by two when four don't fit: never a ragged row.
    check(
      `${width}: the cloud picker is one row or a full two-by-two grid`,
      await page.$$eval('#infrastructure input[name="home-cloud"]', (inputs) => {
        const tops = [
          ...new Set(inputs.map((i) => Math.round(i.closest('label').getBoundingClientRect().top))),
        ];
        if (tops.length === 1) return true;
        const lefts = [
          ...new Set(
            inputs.map((i) => Math.round(i.closest('label').getBoundingClientRect().left))
          ),
        ];
        return inputs.length === 4 && tops.length === 2 && lefts.length === 2;
      })
    );
    // Every cloud's run fits the narrowest phone, not only the default one.
    if (width === 320) {
      const chosen = await page.$eval(
        '#infrastructure input[name="home-cloud"]:checked',
        (i) => i.value
      );
      for (const cloud of clouds) {
        await page.click(`#infrastructure input[value="${cloud}"]`);
        check(`320: no overflowing content with ${cloud} chosen`, await fits());
      }
      // Leave the choice as it was, so the checks below read the default cloud's run.
      await page.click(`#infrastructure input[value="${chosen}"]`);
    }
    // Reduced motion shows the diagram's finished state, and its labels stay on its canvas.
    check(
      `${width}: hero diagram complete with reduced motion`,
      await page.$eval('[data-hero-time]', (e) => Number(e.dataset.heroTime) >= 16)
    );
    check(
      `${width}: hero labels stay inside the diagram`,
      await page.$eval('[data-hero-time]', (e) => {
        const box = e.getBoundingClientRect();
        return [...e.querySelectorAll('svg text')]
          .filter((t) => t.getBoundingClientRect().width)
          .every((t) => {
            const r = t.getBoundingClientRect();
            return r.left >= box.left - 1 && r.right <= box.right + 1;
          });
      })
    );
    check(
      `${width}: the first app follows the hero`,
      await page.$eval('#first-app-title', (e) => e.textContent.includes('running app'))
    );
    check(
      `${width}: infrastructure follows the first app`,
      await page.evaluate(() => {
        const first = document.querySelector('#first-app');
        const infra = document.querySelector('#infrastructure');
        return (
          !!first &&
          !!infra &&
          first.compareDocumentPosition(infra) === Node.DOCUMENT_POSITION_FOLLOWING
        );
      })
    );
    // What the agent wrote is a real file to read, with its check beside it: never a copy button.
    check(
      `${width}: the agent's manifest, its check, its module and its test, no copy button`,
      await page.$eval(
        '#infrastructure',
        (e) =>
          e.textContent.includes('kind: GcpCloudSql') &&
          e.textContent.includes('manifest is valid') &&
          !!e.querySelector('a[href*="/catalog/gcp/gcpcloudsql/iac"]') &&
          !!e.querySelector('a[href*="/catalog/gcp/gcpcloudsql/e2e"]') &&
          !e.querySelector('button')
      )
    );
    // Open source on your own cluster: the agent's one command with the CLI's own open-source-engine line
    // and the order it worked out (operators first), what each service answered, the installer that is the
    // section's way in, and never a $secret/ reference (signed out, the CLI refuses them) or a copy button.
    check(
      `${width}: open source follows infrastructure`,
      await page.evaluate(() => {
        const infra = document.querySelector('#infrastructure');
        const oss = document.querySelector('#open-source');
        return (
          !!infra &&
          !!oss &&
          infra.compareDocumentPosition(oss) === Node.DOCUMENT_POSITION_FOLLOWING
        );
      })
    );
    check(
      `${width}: open source shows the open-source engine, the operator first, the services answering, no copy button`,
      await page.$eval('#open-source', (e) => {
        // What a reader copies is exactly what the CLI printed: no inserted character anywhere.
        const text = e.textContent;
        if (text.includes('\u2060')) return false;
        return (
          text.includes('running with the Planton open-source engine') &&
          /deploy order: KubernetesCloudNativePgOperator\//.test(text) &&
          text.includes('All 6 resources deployed') &&
          text.includes('PONG') &&
          text.includes('Cluster in healthy state') &&
          text.includes('planton.ai/install.sh') &&
          !text.includes('$secret/') &&
          !e.querySelector('button')
        );
      })
    );
    // Run it yourself is the last section before the close: the two commands a real install ran, the
    // status that cluster printed, the docs, no copy button, and never what an install left unconfigured.
    check(
      `${width}: run it yourself follows open source and precedes the close`,
      await page.evaluate(() => {
        const oss = document.querySelector('#open-source');
        const own = document.querySelector('#run-it-yourself');
        const close = document.querySelector('[aria-labelledby="close-title"]');
        const after = (a, b) => a.compareDocumentPosition(b) === Node.DOCUMENT_POSITION_FOLLOWING;
        return !!oss && !!own && !!close && after(oss, own) && after(own, close);
      })
    );
    check(
      `${width}: run it yourself shows the pinned commands, the Ready status and the docs, no copy button`,
      await page.$eval(
        '#run-it-yourself',
        (e) =>
          e.textContent.includes('charts/planton-operator') &&
          e.textContent.includes('platform.spec.version=v') &&
          /Ready\s+v\d+\.\d+\.\d+\s+Community/.test(e.textContent) &&
          !!e.querySelector('a[href="/docs/self-hosting"]') &&
          !/NotConfigured/.test(e.textContent) &&
          !e.querySelector('button')
      )
    );
    // The visitor picks a cloud once: the section shows that cloud's run, and the choice survives a reload.
    if (width === 1280) {
      const showsKind = (kind) =>
        page.$eval('#infrastructure', (e, k) => e.textContent.includes(`kind: ${k}`), kind);
      await page.click('#infrastructure input[value="aws"]');
      check('picking AWS shows the AWS run', await showsKind('AwsRdsInstance'));
      await page.reload({ waitUntil: 'networkidle0' });
      check('the chosen cloud is remembered', await showsKind('AwsRdsInstance'));
      await page.click('#infrastructure input[value="azure"]');
      check('picking Azure shows the Azure run', await showsKind('AzurePostgresqlFlexibleServer'));
      await page.click('#infrastructure input[value="cloudflare"]');
      check('picking Cloudflare shows the Cloudflare run', await showsKind('CloudflareWorker'));
      await page.click('#infrastructure input[value="gcp"]');
      check('picking Google Cloud shows its run again', await showsKind('GcpCloudSql'));
    }
    // The example prompt is words to read, never a prompt to copy: its code is a marked placeholder.
    check(
      `${width}: setup prompt shows a placeholder code and no copy button`,
      await page.$eval(
        '[aria-labelledby="setup-prompt-label"]',
        (e) =>
          e.textContent.includes('<her one-time code>') &&
          !!e.querySelector('strong') &&
          !e.closest('figure').querySelector('button')
      )
    );
    check(
      `${width}: two demo CTAs`,
      await page.$$eval('main a[href="/book-demo"]', (es) => es.length === 2)
    );
    // Start Free leads every door pair, so each demo link has a Start Free before it.
    check(
      `${width}: Start Free before every demo CTA`,
      await page.$$eval('main a[href="/book-demo"]', (es) =>
        es.every((demo) => demo.previousElementSibling?.getAttribute('href')?.endsWith('/signup'))
      )
    );
    if (width !== 320) {
      await page.screenshot({ path: path.join(out, `home-${width}-light.png`), fullPage: true });
      await (
        await page.$('main section')
      ).screenshot({ path: path.join(out, `hero-${width}-light.png`) });
    }
    check(
      `${width}: readable body contrast`,
      await page.evaluate(() => {
        const root = document.querySelector('[data-homepage-appearance]'),
          s = getComputedStyle(root);
        const rgb = (v) => {
          const el = document.createElement('span');
          el.style.color = v;
          root.append(el);
          const values = getComputedStyle(el)
            .color.match(/[\d.]+/g)
            .slice(0, 3)
            .map(Number);
          el.remove();
          return values;
        };
        const lum = (v) =>
          rgb(v)
            .map((n) => {
              n /= 255;
              return n <= 0.04045 ? n / 12.92 : ((n + 0.055) / 1.055) ** 2.4;
            })
            .reduce((sum, n, i) => sum + n * [0.2126, 0.7152, 0.0722][i], 0);
        const a = lum(s.getPropertyValue('--hp-canvas')),
          b = lum(s.getPropertyValue('--hp-secondary'));
        return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05) >= 4.5;
      })
    );
  }
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'no-preference' }]);
  await page.setViewport({ width: 1280, height: 900 });
  await page.goto(base, { waitUntil: 'networkidle0' });
  // One header everywhere: plain links, Sign In, and Start Free; no demo door and no menus.
  check('header has no demo door', !(await page.$('header a[href="/book-demo"]')));
  check('header has no menus', !(await page.$('header [aria-haspopup="true"]')));
  await page.screenshot({ path: path.join(out, 'navigation-1280-light.png') });
  await page.setViewport({ width: 390, height: 844 });
  await page.click('[aria-label="Open navigation"]');
  await page.waitForSelector('.MuiDrawer-paper');
  await page.waitForFunction(
    () => getComputedStyle(document.querySelector('.MuiDrawer-paper')).transform === 'none'
  );
  check(
    'light mobile drawer portal',
    await page.$eval(
      '.MuiDrawer-paper',
      (e) => getComputedStyle(e).backgroundColor === 'rgb(238, 238, 235)'
    )
  );
  await page.screenshot({ path: path.join(out, 'navigation-390-light.png') });
  await page.click('[aria-label="Close navigation"]');
  await page.waitForSelector('.MuiDrawer-root', { hidden: true });
  await page.setViewport({ width: 1280, height: 900 });
  // Both door pairs report their own placement.
  for (const [index, location] of ['hero', 'close'].entries()) {
    await page.goto(base, { waitUntil: 'networkidle0' });
    for (const [selector, event] of [
      ['main a[href="/book-demo"]', 'demo_cta_click'],
      ['main a[href$="/signup"]', 'self_service_click'],
    ]) {
      await page.evaluate(
        (selector, i) =>
          document
            .querySelectorAll(selector)
            [i].addEventListener('click', (e) => e.preventDefault()),
        selector,
        index
      );
      await (await page.$$(selector))[index].click();
      check(
        `${location}: ${event} attribution`,
        (await events()).some((e) => e[1] === event && e[2].location === location)
      );
    }
  }
  // The self-hosting docs door reports its own placement too.
  await page.goto(base, { waitUntil: 'networkidle0' });
  await page.$eval('#run-it-yourself a[href="/docs/self-hosting"]', (e) =>
    e.addEventListener('click', (event) => event.preventDefault())
  );
  await page.click('#run-it-yourself a[href="/docs/self-hosting"]');
  check(
    'self-hosted: self_service_click attribution',
    (await events()).some(
      (e) =>
        e[1] === 'self_service_click' &&
        e[2].location === 'self-hosted' &&
        e[2].door === 'selfHostedDocs'
    )
  );
  await page.goto(base, { waitUntil: 'networkidle0' });
  await page.$eval('header a[href$="/signup"]', (e) =>
    e.addEventListener('click', (event) => event.preventDefault())
  );
  await page.click('header a[href$="/signup"]');
  check(
    'navigation: self_service_click attribution',
    (await events()).some((e) => e[1] === 'self_service_click' && e[2].location === 'navigation')
  );
  await page.goto(base + '/book-demo', { waitUntil: 'networkidle0' });
  check(
    'demo uses light surface',
    await page.$eval(
      '[data-marketing-appearance]',
      (e) => e.dataset.marketingAppearance === 'light'
    )
  );
  await page.screenshot({ path: path.join(out, 'demo-1280.png'), fullPage: true });
  await page.click('button[type="submit"]');
  check('empty form makes no request', submissions === 0);
  check(
    'all required fields show errors',
    await page.$$eval('[aria-invalid="true"]', (es) => es.length === 6)
  );
  await page.type('#firstName', 'Sample');
  await page.type('#lastName', 'Visitor');
  await page.type('#workEmail', 'visitor@example.invalid');
  await page.type('#company', 'Example');
  await page.select('#jobTitle', 'CTO');
  await page.select('#companySize', '1-10');
  mode = 'failure';
  await page.click('button[type="submit"]');
  await page.waitForSelector('[role="alert"]');
  check(
    'server failure retains values',
    await page.$eval('#workEmail', (e) => e.value === 'visitor@example.invalid')
  );
  check('failure does not show scheduler', !(await page.$('iframe')));
  mode = 'network';
  await page.click('button[type="submit"]');
  await page.waitForFunction(() =>
    document.querySelector('[role="alert"]')?.textContent.includes('submit your request')
  );
  check(
    'network failure is recoverable',
    await page.$eval('button[type="submit"]', (e) => !e.disabled)
  );
  mode = 'success';
  await page.click('button[type="submit"]');
  await page.waitForSelector('iframe');
  await page.waitForFunction(() =>
    document.querySelector('h1').textContent.includes('Choose a time')
  );
  check(
    'payload schema preserved',
    Object.keys(lastPayload).sort().join(',') ===
      'company,companySize,firstName,jobTitle,lastName,workEmail'
  );
  check('calendar light theme', await page.$eval('iframe', (e) => e.src.includes('theme=light')));
  check(
    'calendar prefilled',
    await page.$eval(
      'iframe',
      (e) => e.src.includes('visitor%40example.invalid') && e.src.includes('Sample')
    )
  );
  check(
    'focus moves to scheduler heading',
    await page.$eval('h1', (e) => e === document.activeElement)
  );
  check(
    'form success is not a booking',
    !(await events()).some((e) => e[1] === 'demo_booking_confirmed')
  );
  await page.screenshot({ path: path.join(out, 'demo-scheduler-1280.png'), fullPage: true });
  const fire = async (status, paymentRequired = false) =>
    page.evaluate(
      (status, paymentRequired) =>
        window.dispatchEvent(
          new CustomEvent('CAL:60min:bookingSuccessfulV2', {
            detail: { data: { status, paymentRequired, uid: 'fixture-only' } },
          })
        ),
      status,
      paymentRequired
    );
  await fire('PENDING');
  await fire('ACCEPTED', true);
  check(
    'pending/unpaid bookings not counted',
    !(await events()).some((e) => e[1] === 'demo_booking_confirmed')
  );
  await fire('ACCEPTED');
  await fire('ACCEPTED');
  check(
    'confirmed booking counted once',
    (await events()).filter((e) => e[1] === 'demo_booking_confirmed').length === 1
  );
  check(
    'confirmation shown',
    await page.evaluate(() => document.body.textContent.includes('Your meeting is booked.'))
  );
  const analytics = await events();
  for (const event of ['demo_form_start', 'demo_form_submit_success', 'demo_scheduler_view'])
    check(`${event}: emitted once`, analytics.filter((e) => e[1] === event).length === 1);
  check(
    'no personal information in analytics',
    !JSON.stringify(analytics).match(/Visitor|Sample|example.invalid|fixture-only/)
  );
  await page.evaluate(() =>
    window.dispatchEvent(new CustomEvent('CAL:60min:linkFailed', { detail: { data: {} } }))
  );
  await page.waitForSelector('p[role="alert"]');
  check(
    'calendar failure offers fallback',
    await page.$eval('p[role="alert"] a', (e) => e.href === 'https://cal.com/swarup-donepudi/60min')
  );
  await page.setViewport({ width: 390, height: 844 });
  await page.goto(base + '/book-demo', { waitUntil: 'networkidle0' });
  await page.screenshot({ path: path.join(out, 'demo-390.png'), fullPage: true });
  // Every public page shares Home's light appearance (MarketingShell).
  for (const route of ['/docs/coding-agents', '/pricing']) {
    await page.goto(base + route, { waitUntil: 'networkidle0' });
    check(
      `${route}: shares the light appearance`,
      await page.$eval(
        '[data-marketing-appearance]',
        (e) => e.dataset.marketingAppearance === 'light'
      )
    );
  }
  check('no application JavaScript errors', errors.length === 0);
  fs.writeFileSync(
    path.join(out, 'acceptance.json'),
    JSON.stringify(
      { passed: results.length, checks: results, submissions, externalTraffic: 'mocked', errors },
      null,
      2
    )
  );
  console.log(
    `PASS: ${results.length} checks; captures: ${out}; ${submissions} mocked requests, zero real leads or bookings.`
  );
} finally {
  await browser.close();
  await new Promise((resolve) => server.close(resolve));
}
