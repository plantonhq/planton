/**
 * The assets bucket, served from this checkout: a page's image under
 * https://assets.planton.ai/site/ is the committed file at the same path
 * under content/assets/ (the bucket mirrors that folder). The captures and the
 * browser checks answer those requests from here, so they judge the image the
 * branch holds, before it is uploaded and without the network.
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const ASSETS_URL = 'https://assets.planton.ai/site/';
const assetsDir = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'content/assets'
);
const TYPES = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
  '.svg': 'image/svg+xml',
};

/** The response for a request to the assets bucket, or undefined when the URL is not one. */
export function siteAsset(url) {
  if (!url.startsWith(ASSETS_URL)) return undefined;
  const file = path.join(assetsDir, decodeURIComponent(url.slice(ASSETS_URL.length).split('?')[0]));
  if (!file.startsWith(assetsDir + path.sep) || !fs.existsSync(file))
    return { status: 404, contentType: 'text/plain', body: 'not in content/assets' };
  return {
    status: 200,
    contentType: TYPES[path.extname(file)] ?? 'application/octet-stream',
    body: fs.readFileSync(file),
  };
}
