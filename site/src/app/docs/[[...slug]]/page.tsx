import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import {
  getMarkdownContent,
  getDocumentationStructure,
  getNextDocItem,
  generateStaticParamsFromStructure,
  processDocumentationSlug,
} from '@/app/docs/utils/fileSystem';
import { MDXRenderer } from '@/app/docs/components/MDXRenderer';
import { Author, MDXParser } from '@/lib/mdx';
import RightSidebar from '@/app/docs/components/RightSidebar';
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
      title: `${title} - OpenMCF Documentation`,
      description: data?.description || 'OpenMCF Documentation',
    };
  } catch {
    return {
      title: 'Documentation - OpenMCF',
      description: 'OpenMCF Documentation',
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

    const author = (data?.author as unknown as Author[]) || [];

    return (
      <>
        {/* Main Content Area */}
        <div className="flex-1 min-h-screen overflow-x-hidden">
          <div className={`px-4 sm:px-6 lg:px-12 py-8 max-w-full ${author.length > 0 ? 'max-w-4xl mx-auto' : ''}`}>
            <MDXRenderer
              mdxContent={mdxContent}
              nextArticle={
                nextDocItem
                  ? {
                      title: nextDocItem.pageTitle || nextDocItem.title,
                      excerpt: nextDocItem.excerpt,
                      slug: `/docs/${nextDocItem.slug}`,
                    }
                  : undefined
              }
            />
          </div>
        </div>

        {/* Right Sidebar - Table of contents */}
        <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
          <div className="h-full overflow-y-auto bg-slate-950 border-l border-purple-900/30">
            <RightSidebar author={author} content={content} />
          </div>
        </div>
      </>
    );
  } catch (error) {
    console.error('Error loading documentation:', error);
    notFound();
  }
}
