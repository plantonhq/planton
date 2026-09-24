/** Render a social-ready MP4 from the website's scene. Generated media stays out
 * of Git and outside R2's mirrored prefix until deliberately published. */
import path from 'node:path';
import fs from 'node:fs/promises';
import os from 'node:os';
import { fileURLToPath } from 'node:url';
import { bundle } from '@remotion/bundler';
import { renderMedia, renderStill, selectComposition } from '@remotion/renderer';
import puppeteer from 'puppeteer';
import { WORKFLOWS } from '../src/data/workflow-explainers.ts';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const [id = 'aws', destination = path.join(os.tmpdir(), `planton-${id}.mp4`)] =
  process.argv.slice(2);
if (id !== 'hero' && !(id in WORKFLOWS))
  throw new Error(`Choose hero, ${Object.keys(WORKFLOWS).join(', ')}.`);
const outputLocation = path.resolve(destination);
await fs.mkdir(path.dirname(outputLocation), { recursive: true });
const bundleDir = await fs.mkdtemp(path.join(os.tmpdir(), 'planton-workflow-bundle-'));
try {
  const serveUrl = await bundle({
    entryPoint: path.join(root, 'video/workflows.tsx'),
    outDir: bundleDir,
  });
  const browserExecutable = process.env.PUPPETEER_EXECUTABLE_PATH ?? puppeteer.executablePath();
  const composition = await selectComposition({
    serveUrl,
    id,
    browserExecutable,
    chromeMode: 'chrome-for-testing',
  });
  await renderStill({
    composition,
    serveUrl,
    browserExecutable,
    chromeMode: 'chrome-for-testing',
    frame: 18 * composition.fps,
    output: outputLocation.replace(/\.mp4$/i, '') + '-poster.png',
  });
  await renderMedia({
    composition,
    serveUrl,
    browserExecutable,
    chromeMode: 'chrome-for-testing',
    codec: 'h264',
    pixelFormat: 'yuv420p',
    crf: 18,
    concurrency: 2,
    outputLocation,
  });
  const { size } = await fs.stat(outputLocation);
  console.log(
    `${outputLocation}: ${(size / 1024 / 1024).toFixed(2)} MiB, ${composition.width}×${composition.height}, ${composition.durationInFrames / composition.fps}s, H.264`
  );
} finally {
  await fs.rm(bundleDir, { recursive: true, force: true });
}
