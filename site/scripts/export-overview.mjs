/** Offline-only output. Versioned videos are deliberately kept outside Git and
 * outside the deletion-managed R2 site/ mirror. --stills is a quick art review. */
import path from 'node:path';
import fs from 'node:fs/promises';
import os from 'node:os';
import { fileURLToPath } from 'node:url';
import { bundle } from '@remotion/bundler';
import { renderMedia, renderStill, selectComposition } from '@remotion/renderer';
import puppeteer from 'puppeteer';
import sharp from 'sharp';
import { OVERVIEW_VIDEO as video } from '../src/data/homepage-video.ts';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const destination = path.resolve(process.argv.find(a => a.startsWith('--output='))?.slice(9) ?? path.join(os.tmpdir(), 'planton-overview-exports', video.version));
await fs.mkdir(destination, { recursive: true });
const bundleDir = await fs.mkdtemp(path.join(os.tmpdir(), 'planton-overview-bundle-'));
try {
  const serveUrl = await bundle({ entryPoint: path.join(root, 'video/workflows.tsx'), outDir: bundleDir, publicDir: path.join(root, 'public') });
  const browserExecutable = process.env.PUPPETEER_EXECUTABLE_PATH ?? puppeteer.executablePath();
  const common = { serveUrl, browserExecutable, chromeMode: 'chrome-for-testing' };
  const composition = await selectComposition({ ...common, id: video.id });
  const samples = [3, 8, 12, 16, 20, 24, 28, 31, 35, 39, 41, 47, 52, 57];
  for (const second of samples) {
    await renderStill({ ...common, composition, frame: second * video.fps, output: path.join(destination, `frame-${second}.png`) });
  }
  const poster = path.join(destination, 'homepage-overview.webp');
  await sharp(path.join(destination, 'frame-3.png')).resize(1280, 720).webp({ quality: 85 }).toFile(poster);
  const tiles = await Promise.all(samples.map(async (second, i) => ({ input: await sharp(path.join(destination, `frame-${second}.png`)).resize(640, 360).png().toBuffer(), left: (i % 2) * 640, top: Math.floor(i / 2) * 360 })));
  await sharp({ create: { width: 1280, height: Math.ceil(samples.length / 2) * 360, channels: 3, background: '#111316' } }).composite(tiles).png().toFile(path.join(destination, 'contact-sheet.png'));
  if (!process.argv.includes('--stills')) {
    for (const [name, scale, crf] of [['1080p', 1, 18], ['720p', 2 / 3, 23]]) {
      const outputLocation = path.join(destination, `${video.id}-${name}.mp4`);
      let last = -1;
      await renderMedia({ ...common, composition, codec: 'h264', pixelFormat: 'yuv420p', colorSpace: 'bt709', crf, scale, concurrency: 3, outputLocation,
        // Remotion finalizes H.264 with fast-start. Explicit output flags make
        // that contract resilient if its muxer defaults change later.
        ffmpegOverride: ({ args }) => [...args.slice(0, -1), '-movflags', '+faststart', args[args.length - 1]],
        onProgress: ({ progress }) => { const step = Math.floor(progress * 10); if (step !== last) { last = step; console.log(`${name}: ${step * 10}%`); } },
      });
      console.log(`${outputLocation}: ${((await fs.stat(outputLocation)).size / 1048576).toFixed(2)} MiB`);
    }
  }
  console.log(`Preview artifacts: ${destination}`);
} finally {
  await fs.rm(bundleDir, { recursive: true, force: true });
}
