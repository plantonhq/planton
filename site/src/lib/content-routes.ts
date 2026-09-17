import { getAllBlogPosts, getAllTutorials } from '@/lib/mdx';
import { getAllChangelogEntries } from '@/lib/changelog';
import { generateStaticParamsFromStructure, getDocumentationStructure } from '@/app/(site)/docs/utils/fileSystem';

/**
 * The routes generated from content folders (docs, blog, changelog,
 * tutorials), enumerated the same way their pages' generateStaticParams do,
 * so the sitemap can never list a content page that does not exist or miss
 * one that does. The index pages of these sections are registered in
 * src/data/site-pages.ts like any other page; this module lists their
 * children.
 */
export interface ContentRoute {
  path: string;
  lastModified?: Date;
}

export async function contentRoutes(): Promise<ContentRoute[]> {
  const routes: ContentRoute[] = [];

  const docs = generateStaticParamsFromStructure(await getDocumentationStructure());
  for (const { slug } of docs) {
    if (slug.length === 0) continue; // the docs index is a registered page
    routes.push({ path: `/docs/${slug.join('/')}` });
  }

  for (const post of getAllBlogPosts()) {
    routes.push({ path: `/blog/${post.slug}`, lastModified: post.date ? new Date(post.date) : undefined });
  }
  for (const tutorial of getAllTutorials()) {
    routes.push({ path: `/tutorials/${tutorial.slug}`, lastModified: tutorial.date ? new Date(tutorial.date) : undefined });
  }
  for (const entry of getAllChangelogEntries()) {
    routes.push({ path: `/changelog/${entry.slug}`, lastModified: entry.date ? new Date(entry.date) : undefined });
  }

  return routes;
}
