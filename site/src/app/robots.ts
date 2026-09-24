import type { MetadataRoute } from 'next';
import { SITE } from '@/data/site-pages';

/**
 * Everything may be crawled. Surfaces that must not be indexed (decks, the
 * investor pages) say so in their own metadata, because a Disallow here would
 * stop crawlers from ever reading that directive and leave shared links
 * indexable.
 */
export const dynamic = 'force-static';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: { userAgent: '*', allow: '/' },
    sitemap: `${SITE.url}/sitemap.xml`,
  };
}
