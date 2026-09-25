/** Validate deliverable encoding and fast-start before publishing. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { RenderInternals } from '@remotion/renderer';
import { OVERVIEW_VIDEO as video } from '../src/data/homepage-video.ts';
const destination = process.argv[2];
assert.ok(destination, 'Pass the export directory');
const ffprobe = RenderInternals.getExecutablePath({ type: 'ffprobe', indent: false, logLevel: 'error', binariesDirectory: null });
for (const [name, width, height] of [['1080p', 1920, 1080], ['720p', 1280, 720]]) {
  const filename = path.join(destination, `${video.id}-${name}.mp4`);
  const result = spawnSync(ffprobe, ['-v', 'error', '-show_format', '-show_streams', '-of', 'json', filename], { cwd: path.dirname(ffprobe), encoding: 'utf8' });
  assert.equal(result.status, 0, result.stderr);
  const info = JSON.parse(result.stdout);
  const silent = process.argv.includes('--silent');
  assert.equal(info.streams.length, silent ? 1 : 2);
  const stream = info.streams.find(s => s.codec_type === 'video');
  if (!silent) {
    const audio = info.streams.find(s => s.codec_type === 'audio');
    assert.equal(audio.codec_name, 'aac');
    assert.equal(audio.channels, 2);
    assert.ok(Math.abs(Number(audio.duration) - video.duration) < 0.05);
  }
  assert.equal(stream.codec_name, 'h264');
  assert.equal(stream.pix_fmt, 'yuv420p');
  assert.equal(stream.width, width);
  assert.equal(stream.height, height);
  assert.equal(stream.avg_frame_rate, '30/1');
  assert.equal(Number(info.format.duration), video.duration);
  const bytes = fs.readFileSync(filename), atoms = [];
  for (let offset = 0; offset < bytes.length;) {
    const size32 = bytes.readUInt32BE(offset), type = bytes.toString('ascii', offset + 4, offset + 8);
    const size = size32 === 1 ? Number(bytes.readBigUInt64BE(offset + 8)) : size32 || bytes.length - offset;
    atoms.push(type);
    assert.ok(size >= 8, 'Valid top-level MP4 atom');
    offset += size;
  }
  assert.ok(atoms.indexOf('moov') >= 0 && atoms.indexOf('moov') < atoms.indexOf('mdat'), 'Fast-start metadata precedes media');
  console.log(`PASS ${name}: ${width}×${height}, 30fps, 60s, H.264 ${silent ? 'picture master' : '+ stereo AAC'}, fast-start, ${(bytes.length / 1048576).toFixed(2)} MiB`);
}
