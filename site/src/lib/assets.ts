/**
 * Where the site's static files are served from.
 *
 * planton.ai shares its domain with the console: the edge passes a fixed list
 * of path prefixes through to this static site and sends everything else,
 * including `/_next`, to the console. So the site's bundles and images live
 * under one prefix that is on that list, `/_site` (next.config.ts sets it as
 * the assetPrefix, and the Makefile copies the bundle there after the build).
 *
 * Components build public URLs with `asset()` and never type the prefix, so
 * the day the apex routing no longer needs it, this is the one line that
 * changes. Images belong on the asset CDN (assets.planton.ai) per the media
 * rule; `asset()` is for what must be same-origin.
 */
export const PUBLIC_ASSET_PREFIX = '/_site';

/** A same-origin URL for a file under public/_site/, from its path inside that folder. */
export function asset(pathInsideSite: string): string {
  const clean = pathInsideSite.startsWith('/') ? pathInsideSite : `/${pathInsideSite}`;
  return `${PUBLIC_ASSET_PREFIX}${clean}`;
}
