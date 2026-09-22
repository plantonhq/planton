import type { Metadata } from 'next';
import { SITE, sitePage } from '@/data/site-pages';
import { retiredRoute } from '@/data/retired-routes';

/**
 * The metadata for a registered page, from the route registry. A page's
 * title, description, canonical URL, Open Graph card, and index/noindex
 * decision are all facts about the route, so they live in one record and
 * every page reads them the same way:
 *
 *   export const metadata = pageMetadata('/trust/the-record');
 *
 * A page with a genuinely page-specific card (its own Open Graph image, a
 * JSON-LD block) spreads this and adds to it; it never restates the title.
 */
export function pageMetadata(path: string, overrides: Metadata = {}): Metadata {
  const page = sitePage(path);
  const url = `${SITE.url}${page.path === '/' ? '' : page.path}`;
  const title = `${page.title}${SITE.titleSuffix}`;
  const indexable = page.index !== false;
  return {
    title,
    description: page.description,
    alternates: { canonical: url },
    robots: indexable ? undefined : { index: false, follow: false },
    openGraph: {
      type: 'website',
      siteName: SITE.name,
      url,
      title,
      description: page.description,
      images: [{ url: SITE.defaultOgImage, width: 1200, height: 630 }],
    },
    ...overrides,
  };
}

/**
 * The metadata for a retired path: a title that says so, no indexing, and a
 * canonical pointing at the page that answers for it, so a crawler that still
 * holds the old URL transfers what it knows to the new one.
 */
export function retiredRouteMetadata(from: string): Metadata {
  const route = retiredRoute(from);
  return {
    title: `Moved${SITE.titleSuffix}`,
    robots: { index: false, follow: true },
    alternates: { canonical: `${SITE.url}${route.to}` },
  };
}
