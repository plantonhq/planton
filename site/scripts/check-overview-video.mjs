/** Native-player acceptance. No real analytics or lead traffic. CI uses a tiny
 * generated H.264 fixture; set OVERVIEW_VIDEO_FILE to exercise the actual film. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import puppeteer from 'puppeteer';
import sharp from 'sharp';
import { RenderInternals } from '@remotion/renderer';
import { createPreviewServer } from './preview-homepage.mjs';
import { OVERVIEW_VIDEO as video } from '../src/data/homepage-video.ts';

const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'planton-video-acceptance-'));
let fixture = process.env.OVERVIEW_VIDEO_FILE;
if (!fixture) {
  fixture = path.join(temporary, 'fixture.mp4');
  const frame = path.join(temporary, 'fixture.png');
  await sharp('public/_site/images/product/homepage-overview.webp').resize(320, 180).png().toFile(frame);
  const ffmpeg = RenderInternals.getExecutablePath({ type: 'ffmpeg', indent: false, logLevel: 'error', binariesDirectory: null });
  const generated = spawnSync(ffmpeg, ['-v', 'error', '-loop', '1', '-framerate', '30', '-i', frame, '-t', '60', '-an', '-c:v', 'libx264', '-pix_fmt', 'yuv420p', '-movflags', '+faststart', fixture], { cwd: path.dirname(ffmpeg), encoding: 'utf8' });
  assert.equal(generated.status, 0, generated.stderr);
}
const media = fs.readFileSync(fixture);
const captures = process.env.HOMEPAGE_CAPTURES ?? path.join(os.tmpdir(), 'planton-overview-review');
fs.mkdirSync(captures, { recursive: true });
const server = createPreviewServer();
await new Promise(r => server.listen(0, '127.0.0.1', r));
const base = `http://127.0.0.1:${server.address().port}`;
const browser = await puppeteer.launch({ headless: true });
const page = await browser.newPage();
const errors = [], checks = [];
let mediaRequests = 0, failMedia = false;
const check = (name, value) => { assert.ok(value, name); checks.push(name); };
page.on('pageerror', e => errors.push(e.message));
await page.setRequestInterception(true);
page.on('request', request => {
  if (request.url().startsWith(video.base)) {
    mediaRequests++;
    if (failMedia) return request.abort('failed');
    const match = /bytes=(\d+)-(\d*)/.exec(request.headers().range ?? '');
    const start = match ? Number(match[1]) : 0, end = match?.[2] ? Math.min(Number(match[2]), media.length - 1) : media.length - 1;
    return request.respond({ status: match ? 206 : 200, contentType: 'video/mp4', headers: { 'accept-ranges': 'bytes', ...(match ? { 'content-range': `bytes ${start}-${end}/${media.length}` } : {}) }, body: media.subarray(start, end + 1) });
  }
  if (request.url().startsWith(base) || request.url().startsWith('data:')) return request.continue();
  return request.respond({ status: 200, contentType: request.resourceType() === 'script' ? 'text/javascript' : 'text/plain', body: '' });
});
const selector = '#homepage-overview-video';
const events = () => page.evaluate(() => Array.from(window.dataLayer ?? [], x => Array.from(x)).filter(x => x[0] === 'event' && x[1].startsWith('overview_video_')));
async function playByKeyboard(noScript = false) {
  await page.focus(selector);
  await page.keyboard.press('Space');
  const playing = () => page.$eval(selector, v => !v.paused && v.currentTime > 0);
  if (noScript) {
    // Chromium suppresses document timers as well as scripts. Poll through CDP
    // from Node so the no-JS assertion observes actual native media playback.
    for (let attempt = 0; attempt < 50; attempt++) {
      if (await playing()) return;
      await new Promise(r => setTimeout(r, 100));
    }
    assert.fail('No-JavaScript playback did not start');
  } else await page.waitForFunction(() => { const v = document.querySelector('#homepage-overview-video'); return !v.paused && v.currentTime > 0; });
}
try {
  for (const [width, height] of [[320, 844], [390, 844], [768, 1024], [1366, 768], [1440, 900], [1920, 1080]]) {
    await page.setViewport({ width, height });
    await page.goto(base, { waitUntil: 'networkidle0' });
    await page.$eval(selector, v => v.scrollIntoView({ block: 'center' }));
    const state = await page.$eval(selector, v => {
      const r = v.getBoundingClientRect();
      return { ratio: r.width / r.height, preload: v.preload, paused: v.paused, controls: v.controls, auto: v.autoplay, loop: v.loop, ready: v.readyState, overflow: document.documentElement.scrollWidth > innerWidth };
    });
    check(`${width}: stable 16:9 and no horizontal overflow`, Math.abs(state.ratio - 16 / 9) < 0.02 && !state.overflow);
    check(`${width}: native controls with no eager media`, state.preload === 'none' && state.controls && state.paused && !state.auto && !state.loop && state.ready === 0 && mediaRequests === 0);
    await page.screenshot({ path: path.join(captures, `overview-${width}.png`) });
  }
  check('video connects hero to product explanation', await page.evaluate(() => {
    const section = document.querySelector('#homepage-overview-video').closest('section');
    return Boolean(section.previousElementSibling.querySelector('#homepage-title') && section.nextElementSibling.querySelector('#overview-title'));
  }));
  await page.click('section[aria-labelledby="overview-video-title"] summary');
  check('six transcript chapters are readable', await page.$$eval('section[aria-labelledby="overview-video-title"] details li', nodes => nodes.length === 6 && nodes.every(n => n.getBoundingClientRect().height > 0)));
  await page.click('section[aria-labelledby="overview-video-title"] summary');
  await page.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'reduce' }]);
  check('reduced motion does not start the video', await page.$eval(selector, v => v.paused));
  await playByKeyboard();
  check('keyboard play fetches and decodes media', mediaRequests > 0 && await page.$eval(selector, v => v.videoWidth > 0));
  await page.$eval(selector, v => { v.textTracks[0].mode = 'showing'; });
  await page.waitForFunction(() => document.querySelector('#homepage-overview-video').textTracks[0].cues?.length === 6);
  check('English captions are available', await page.$eval(selector, v => v.textTracks[0].language === 'en' && v.textTracks[0].cues[0].text.includes('coordination')));
  await page.keyboard.press('Space');
  check('keyboard pause works', await page.$eval(selector, v => v.paused));
  await page.$eval(selector, v => { v.currentTime = 42; });
  await page.waitForFunction(() => { const v = document.querySelector('#homepage-overview-video'); return !v.seeking && v.readyState >= 2 && v.currentTime >= 42; });
  check('seeking preserves paused state', await page.$eval(selector, v => v.paused));
  await page.$eval(selector, v => v.requestFullscreen());
  check('native video enters fullscreen', await page.evaluate(() => document.fullscreenElement?.tagName === 'VIDEO'));
  await page.evaluate(() => document.exitFullscreen());
  await page.$eval(selector, v => { v.currentTime = 59.7; });
  await playByKeyboard();
  await page.waitForFunction(() => document.querySelector('#homepage-overview-video').ended);
  await playByKeyboard();
  check('replay starts from the beginning', await page.$eval(selector, v => v.currentTime < 2));
  await page.$eval(selector, v => v.pause());
  const recorded = await events();
  check('one start and one completion per page view', recorded.filter(e => e[1] === 'overview_video_start').length === 1 && recorded.filter(e => e[1] === 'overview_video_complete').length === 1);
  check('analytics contain only video identity', recorded.every(e => JSON.stringify(e[2]) === JSON.stringify({ video_id: video.id, video_version: video.version })));

  failMedia = true;
  await page.goto(base, { waitUntil: 'networkidle0' });
  // A fresh URL avoids Chromium reusing its in-memory media cache from the
  // successful playback above; this request must reach the failure fixture.
  await page.$eval(selector, v => { v.src += '?failure-fixture=1'; });
  await page.focus(selector);
  await page.keyboard.press('Space');
  await page.waitForSelector('section[aria-labelledby="overview-video-title"] [role="alert"]');
  check('network failure offers direct video and transcript', await page.$eval('[role="alert"]', e => e.textContent.includes('transcript') && e.querySelector('a').href.endsWith('720p.mp4')));

  failMedia = false;
  await page.setJavaScriptEnabled(false);
  const before = mediaRequests;
  await page.goto(base, { waitUntil: 'networkidle0' });
  check('no-JavaScript player still defers download', before === mediaRequests);
  await page.click('section[aria-labelledby="overview-video-title"] summary');
  check('no-JavaScript transcript expands', await page.$eval('section[aria-labelledby="overview-video-title"] details', d => d.open));
  await playByKeyboard(true);
  check('no-JavaScript native playback works', await page.$eval(selector, v => v.currentTime > 0));
  check('no browser exceptions', errors.length === 0);
  console.log(`PASS: ${checks.length} overview-video checks (${process.env.OVERVIEW_VIDEO_FILE ? 'actual film' : 'generated fixture'}).`);
  console.log(`Screenshots: ${captures}`);
} catch (error) {
  console.error('Player state:', await page.$eval(selector, v => ({ paused: v.paused, time: v.currentTime, ready: v.readyState, error: v.error?.code, source: v.currentSrc })), 'Completed checks:', checks);
  throw error;
} finally {
  await browser.close();
  await new Promise(r => server.close(r));
  fs.rmSync(temporary, { recursive: true, force: true });
}
