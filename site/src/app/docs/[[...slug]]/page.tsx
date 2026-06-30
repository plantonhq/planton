import fs from 'fs';
import nodePath from 'path';
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
import { DOCS_DIRECTORY } from '@/lib/constants';
import { MDXRenderer } from '@/app/docs/components/MDXRenderer';
import { CatalogProvider } from '@/app/docs/components/CatalogProviderGrid';
import { PresetListPage, PresetEntry } from '@/app/docs/components/PresetListPage';
import { PresetDetailPage } from '@/app/docs/components/PresetDetailPage';
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
      title: provider.title || provider.name.toUpperCase(),
      path: provider.path,
      componentCount:
        provider.children?.length ?? 0,
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

/**
 * Detect whether the current catalog page has a co-located presets/ directory
 * and return link info for the right sidebar.  Returns undefined when no
 * presets exist for this page.
 */
function detectPresetsLink(
  docPath: string,
): { path: string; count: number } | undefined {
  // Only applies to catalog component pages: catalog/{provider}/{component}
  const parts = docPath.split('/');
  if (parts.length !== 3 || parts[0] !== 'catalog') return undefined;

  const presetsIndexPath = nodePath.join(
    DOCS_DIRECTORY,
    docPath,
    'presets',
    'index.md',
  );

  if (!fs.existsSync(presetsIndexPath)) return undefined;

  // Count preset files (*.yaml) in the presets directory
  const presetsDir = nodePath.join(DOCS_DIRECTORY, docPath, 'presets');
  const presetCount = fs
    .readdirSync(presetsDir)
    .filter((f) => f.endsWith('.yaml')).length;

  if (presetCount === 0) return undefined;

  return {
    path: `/docs/${docPath}/presets`,
    count: presetCount,
  };
}

/**
 * Load preset YAML + MD content for the preset-list page.
 * Called at build time (server component) so filesystem access is safe.
 */
function loadPresetsForListPage(
  presetsDir: string,
  presetsMeta: Array<{ slug: string; rank: string; title: string; excerpt: string }>,
): PresetEntry[] {
  return presetsMeta.map((p) => {
    const yamlPath = nodePath.join(presetsDir, `${p.slug}.yaml`);
    const mdPath = nodePath.join(presetsDir, `${p.slug}.md`);

    const yamlContent = fs.existsSync(yamlPath)
      ? fs.readFileSync(yamlPath, 'utf-8')
      : '';

    // Read the MD and strip frontmatter for the expanded accordion body
    let mdContent = '';
    if (fs.existsSync(mdPath)) {
      const raw = fs.readFileSync(mdPath, 'utf-8');
      const { content } = matter(raw);
      mdContent = content;
    }

    return {
      slug: p.slug,
      rank: p.rank,
      title: p.title,
      excerpt: p.excerpt,
      yamlContent,
      mdContent,
    };
  });
}

export default async function DocsPage({ params }: { params: DocsParams }) {
  const { slug = [] } = await params;
  const { path: docPath } = processDocumentationSlug(slug);

  try {
    const content = await getMarkdownContent(docPath);
    const { data, content: mdBody } = matter(content);

    // -----------------------------------------------------------------
    // Preset list page (type: "preset-list")
    // -----------------------------------------------------------------
    if (data.type === 'preset-list') {
      const presetsDir = nodePath.join(DOCS_DIRECTORY, docPath);
      const presetsMeta = (data.presets || []) as Array<{
        slug: string;
        rank: string;
        title: string;
        excerpt: string;
      }>;

      const presets = loadPresetsForListPage(presetsDir, presetsMeta);
      const basePath = `/docs/${docPath}`;
      const catalogPath = `/docs/catalog/${data.provider}/${data.componentSlug}`;

      return (
        <>
          <div className="flex-1 min-h-screen overflow-x-hidden">
            <div className="px-4 sm:px-6 lg:px-12 py-8 max-w-full">
              <PresetListPage
                componentTitle={data.componentTitle as string}
                componentSlug={data.componentSlug as string}
                provider={data.provider as string}
                presets={presets}
                basePath={basePath}
                catalogPath={catalogPath}
              />
            </div>
          </div>

          {/* Right Sidebar */}
          <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
            <div className="h-full overflow-y-auto bg-slate-950 border-l border-purple-900/30">
              <RightSidebar content={content} />
            </div>
          </div>
        </>
      );
    }

    // -----------------------------------------------------------------
    // Preset detail page (type: "preset")
    // -----------------------------------------------------------------
    if (data.type === 'preset') {
      const presetSlug = data.presetSlug as string;
      const presetsDir = nodePath.join(DOCS_DIRECTORY, nodePath.dirname(docPath));
      const yamlPath = nodePath.join(presetsDir, `${presetSlug}.yaml`);
      const yamlContent = fs.existsSync(yamlPath)
        ? fs.readFileSync(yamlPath, 'utf-8')
        : '';

      const basePath = `/docs/${nodePath.dirname(docPath)}`;

      return (
        <>
          <div className="flex-1 min-h-screen overflow-x-hidden">
            <div className="px-4 sm:px-6 lg:px-12 py-8 max-w-full">
              <PresetDetailPage
                title={data.title as string}
                rank={data.rank as string}
                presetSlug={presetSlug}
                componentSlug={data.componentSlug as string}
                componentTitle={data.componentTitle as string}
                provider={data.provider as string}
                yamlContent={yamlContent}
                mdContent={mdBody}
                basePath={basePath}
              />
            </div>
          </div>

          {/* Right Sidebar */}
          <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
            <div className="h-full overflow-y-auto bg-slate-950 border-l border-purple-900/30">
              <RightSidebar content={content} />
            </div>
          </div>
        </>
      );
    }

    // -----------------------------------------------------------------
    // Standard documentation page (default)
    // -----------------------------------------------------------------
    const mdxContent = MDXParser.reconstructMDX(content);

    // Get the documentation structure to find the next item and catalog data
    const allDocs = await getDocumentationStructure();
    const nextDocItem = getNextDocItem(docPath, allDocs);

    // When rendering the catalog index page, derive provider metadata from
    // the docs structure so the grid is data-driven (never stale).
    const catalogProviders =
      docPath === 'catalog' ? extractCatalogProviders(allDocs) : undefined;

    const author = (data?.author as unknown as Author[]) || [];

    // Check if this catalog page has presets (for right sidebar link)
    const presetsLink = detectPresetsLink(docPath);

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
            <RightSidebar author={author} content={content} presetsLink={presetsLink} />
          </div>
        </div>
      </>
    );
  } catch (error) {
    console.error('Error loading documentation:', error);
    notFound();
  }
}
