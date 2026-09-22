import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import {
  getMarkdownContent,
  getDocumentationStructure,
  getNextDocItem,
  generateStaticParamsFromStructure,
  processDocumentationSlug,
  resolveDocFilePath,
} from '@/app/(site)/docs/utils/fileSystem';
import { MDXRenderer } from '@/lib/MDXRenderer';
import { Author, MDXParser } from '@/lib/mdx';
import { DocsLayout } from '@/app/(site)/docs/components/DocsLayout';
import matter from 'gray-matter';
type DocsParams = Promise<{ slug?: string[] }>;

export async function generateMetadata({ params }: { params: DocsParams }): Promise<Metadata> {
  const { slug = [] } = await params;
  const { path } = processDocumentationSlug(slug);

  try {
    const content = await getMarkdownContent(path);
    const { data } = matter(content);
    const title = data?.title || slug[slug.length - 1] || 'Documentation';

    return {
      title: `${title} - Planton Documentation`,
      description: data?.description || 'Planton Documentation',
    };
  } catch {
    return {
      title: 'Documentation - Planton',
      description: 'Planton Documentation',
    };
  }
}

export async function generateStaticParams() {
  const structure = await getDocumentationStructure();
  return generateStaticParamsFromStructure(structure);
}

export default async function DocsPage({ params }: { params: DocsParams }) {
  const { slug = [] } = await params;
  const { path } = processDocumentationSlug(slug);

  try {
    const content = await getMarkdownContent(path);
    const { data } = matter(content);
    const mdxContent = MDXParser.reconstructMDX(content);

    // Get the documentation structure to find the next item
    const allDocs = await getDocumentationStructure();
    const nextDocItem = getNextDocItem(path, allDocs);

    // Resolve the actual file path for "Open Raw" (e.g. "platform/index.md" not "platform.md")
    const resolvedFilePath = resolveDocFilePath(path);

    // Normal rendering with full layout
    return (
      <DocsLayout author={data?.author as unknown as Author[]} content={content} structure={allDocs}>
        <MDXRenderer
          mdxContent={mdxContent}
          markdownContent={content}
          title={data?.title}
          nextArticle={
            nextDocItem
              ? {
                  title: nextDocItem.title,
                  excerpt: nextDocItem.excerpt,
                  slug: `/docs/${nextDocItem.slug}`,
                }
              : undefined
          }
          path={`/docs/${path}`}
          rawPath={`/docs/${resolvedFilePath}`}
        />
      </DocsLayout>
    );
  } catch (error) {
    console.error('Error loading documentation:', error);
    notFound();
  }
}
