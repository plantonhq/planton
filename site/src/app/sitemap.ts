import type { MetadataRoute } from 'next';
import { INDEXED_PAGES, SITE } from '@/data/site-pages';
import { contentRoutes } from '@/lib/content-routes';

/**
 * The sitemap, generated from the route registry and the content folders at
 * build time. It cannot list a page that does not exist, cannot omit one the
 * registry knows, and never lists a retired path or a noindex surface,
 * because none of those are in the two lists it reads.
 */
export const dynamic = 'force-static';

const PRIORITY: Record<string, number> = {
  home: 1,
  product: 0.9,
  trust: 0.9,
  solutions: 0.8,
  pricing: 0.9,
  content: 0.7,
  company: 0.4,
  legal: 0.3,
  standalone: 0.3,
};

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const pages: MetadataRoute.Sitemap = INDEXED_PAGES.map((page) => ({
    url: `${SITE.url}${page.path === '/' ? '' : page.path}`,
    changeFrequency: page.group === 'content' ? 'weekly' : 'monthly',
    priority: PRIORITY[page.group] ?? 0.5,
  }));
  const content: MetadataRoute.Sitemap = (await contentRoutes()).map((route) => ({
    url: `${SITE.url}${route.path}`,
    lastModified: route.lastModified,
    changeFrequency: 'monthly',
    priority: 0.6,
  }));
  return [...pages, ...content];
}
