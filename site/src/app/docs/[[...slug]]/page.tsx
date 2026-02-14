import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import {
  DocItem,
  getMarkdownContent,
  getDocumentationStructure,
  getNextDocItem,
  generateStaticParamsFromStructure,
  processDocumentationSlug,
} from '@/app/docs/utils/fileSystem';
import { MDXRenderer } from '@/app/docs/components/MDXRenderer';
import { CatalogProvider } from '@/app/docs/components/CatalogProviderGrid';
import { Author, MDXParser } from '@/lib/mdx';
import RightSidebar from '@/app/docs/components/RightSidebar';
import matter from 'gray-matter';

/**
 * Extract catalog provider metadata from the docs structure tree.
 *
 * Walks the top-level "catalog" directory and returns a sorted list of
 * providers with their component counts (derived from child file entries).
 * This data is computed at build time and serialised to the client component.
 */
function extractCatalogProviders(structure: DocItem[]): CatalogProvider[] {
  const catalogDir = structure.find(
    (item) => item.name === 'catalog' && item.type === 'directory',
  );
  if (!catalogDir?.children) return [];

  return catalogDir.children
    .filter((item) => item.type === 'directory')
    .map((provider) => ({
      name: provider.name,
      path: provider.path,
      componentCount:
        provider.children?.filter((child) => child.type === 'file').length ?? 0,
    }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

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

    // Get the documentation structure to find the next item and catalog data
    const allDocs = await getDocumentationStructure();
    const nextDocItem = getNextDocItem(path, allDocs);

    // When rendering the catalog index page, derive provider metadata from
    // the docs structure so the grid is data-driven (never stale).
    const catalogProviders =
      path === 'catalog' ? extractCatalogProviders(allDocs) : undefined;

    const author = (data?.author as unknown as Author[]) || [];

    return (
      <>
        {/* Main Content Area */}
        <div className="flex-1 min-h-screen overflow-x-hidden">
          <div className={`px-4 sm:px-6 lg:px-12 py-8 max-w-full ${author.length > 0 ? 'max-w-4xl mx-auto' : ''}`}>
            <MDXRenderer
              mdxContent={mdxContent}
              catalogProviders={catalogProviders}
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
