import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';
import { DOCS_DIRECTORY } from '@/lib/constants';
import { generateExcerptFromContent } from '@/lib/utils';

export interface DocItem {
  name: string;
  type: 'file' | 'directory';
  path: string;
  children?: DocItem[];
  // Enhanced properties for dynamic sidebar
  title?: string;
  pageTitle?: string; // Full page title from the first # heading (may differ from sidebar title)
  description?: string;
  icon?: string;
  category?: string;
  order?: number;
  badge?: string; // For "Popular", "Beta", etc.
  isExternal?: boolean;
  externalUrl?: string;
  hasIndex?: boolean; // For directories with index files
  excerpt?: string;
  componentName?: string; // Original component directory name (for icon resolution in catalog)
}

export interface MarkdownContent {
  content: string;
  data: {
    title?: string;
    description?: string;
    icon?: string;
    category?: string;
    order?: number;
    badge?: string;
    isExternal?: boolean;
    externalUrl?: string;
    [key: string]: string | string[] | number | boolean | undefined;
  };
  isMdx?: boolean;
}

// Icon mapping for different content types
const iconMap: Record<string, string> = {
  // Default icons
  'chart-line': '📊',
  flag: '🚩',
  eye: '👁️',
  gear: '⚙️',
  users: '👥',
  database: '🗄️',
  code: '💻',
  rocket: '🚀',
  book: '📚',
  docs: '📖',
  platform: '🏢',
  cloud: '☁️',
  guide: '🗺️',
  tutorial: '🎓',
  api: '🔌',
  sdk: '🛠️',
  integration: '🔗',
  deployment: '🚀',
  monitoring: '📈',
  security: '🔒',
  performance: '⚡',
  lightbulb: '💡',
  package: '📦',
};

// Category-based icon mapping
const categoryIcons: Record<string, string> = {
  docs: '📚',
  concepts: '💡',
  'deployment-components': '📦',
  deployment: '🚀',
  monitoring: '📊',
  security: '🔒',
};

// Resolve an icon value coming from frontmatter
function resolveIcon(
  metaIcon: string | undefined,
  type: 'file' | 'directory',
  name: string,
  category?: string
): string {
  if (metaIcon) {
    const mapped = iconMap[metaIcon];
    if (mapped) {
      return mapped;
    }
  }
  return getDefaultIcon(type, name, category);
}

// Default icons for different file types
const getDefaultIcon = (type: string, name: string, category?: string): string => {
  // Check if there's a specific icon in the name
  const nameLower = name.toLowerCase();

  // Check for specific patterns in the name
  if (nameLower.includes('api')) return iconMap['api'];
  if (nameLower.includes('sdk')) return iconMap['sdk'];
  if (nameLower.includes('guide')) return iconMap['guide'];
  if (nameLower.includes('tutorial')) return iconMap['tutorial'];
  if (nameLower.includes('integration')) return iconMap['integration'];
  if (nameLower.includes('deployment')) return iconMap['deployment'];
  if (nameLower.includes('monitoring')) return iconMap['monitoring'];
  if (nameLower.includes('security')) return iconMap['security'];
  if (nameLower.includes('performance')) return iconMap['performance'];
  if (nameLower.includes('cloud')) return iconMap['cloud'];

  // Check category-based icons
  if (category && categoryIcons[category]) {
    return categoryIcons[category];
  }

  // Default icons based on type
  return type === 'directory' ? '📁' : '📄';
};

export async function getMarkdownContent(filePath: string): Promise<string> {
  // Try different file extensions and paths (.md only)
  const possiblePaths = [
    path.join(DOCS_DIRECTORY, `${filePath}.md`),
    path.join(DOCS_DIRECTORY, filePath, 'index.md'),
    path.join(DOCS_DIRECTORY, filePath, 'README.md'),
  ];

  for (const candidatePath of possiblePaths) {
    if (fs.existsSync(candidatePath)) {
      return fs.readFileSync(candidatePath, 'utf-8');
    }
  }

  // If no markdown file found, try to find any .md file in the directory
  const dirPath = path.join(DOCS_DIRECTORY, filePath);
  if (fs.existsSync(dirPath) && fs.statSync(dirPath).isDirectory()) {
    const files = fs.readdirSync(dirPath);
    const mdLikeFile = files.find((file) => file.endsWith('.md'));
    if (mdLikeFile) {
      return fs.readFileSync(path.join(dirPath, mdLikeFile), 'utf-8');
    }
  }

  throw new Error(`No markdown file found for path: ${filePath}`);
}

export async function getDocumentationStructure(): Promise<DocItem[]> {
  return buildStructure(DOCS_DIRECTORY);
}

function buildStructure(dirPath: string, relativePath: string = ''): DocItem[] {
  if (!fs.existsSync(dirPath)) {
    return [];
  }

  const items = fs.readdirSync(dirPath);
  const structure: DocItem[] = [];

  for (const item of items) {
    const fullPath = path.join(dirPath, item);
    const stat = fs.statSync(fullPath);
    const itemRelativePath = path.join(relativePath, item);

    if (stat.isDirectory()) {
      const children = buildStructure(fullPath, itemRelativePath);

      // Check if this directory has an index file (needed even when children is empty,
      // e.g. a component directory that has been converted from flat file to directory layout)
      const indexFiles = ['index.md', 'README.md'];
      const hasIndex = indexFiles.some((indexFile) => fs.existsSync(path.join(fullPath, indexFile)));

      if (children.length > 0 || hasIndex) {
        // Try to get metadata and page title from index/README (.md only)
        let metadata: MarkdownContent['data'] = {};
        let dirPageTitle: string | undefined;

        for (const indexFile of indexFiles) {
          const indexPath = path.join(fullPath, indexFile);
          if (fs.existsSync(indexPath)) {
            try {
              const fileContent = fs.readFileSync(indexPath, 'utf-8');
              const { data, content: mdContent } = matter(fileContent);
              metadata = data;
              // Extract # heading for full page title
              const headingMatch = mdContent.match(/^#\s+(.+)$/m);
              dirPageTitle = headingMatch ? headingMatch[1].trim() : undefined;
              break;
            } catch (error) {
              console.warn(`Failed to parse metadata from ${indexPath}:`, error);
            }
          }
        }

        const category = relativePath.split('/')[0] || item;

        structure.push({
          name: item,
          type: 'directory',
          path: itemRelativePath,
          children,
          title: metadata.title || formatName(item),
          pageTitle: dirPageTitle,
          description: metadata.description,
          icon: resolveIcon(metadata.icon as string | undefined, 'directory', item, category),
          category,
          order: metadata.order || 0,
          badge: metadata.badge,
          isExternal: (metadata.isExternal as boolean) || false,
          externalUrl: metadata.externalUrl as string | undefined,
          hasIndex,
          excerpt: '' // Directories don't have content to generate excerpts from
        });
      }
    } else if (item.endsWith('.md')) {
      // Skip certain files that are not meant for documentation
      // Also skip index.md and README.md as they represent directory content
      if (
        !item.startsWith('prompt.') &&
        !item.startsWith('response.') &&
        !item.includes('.not-good.') &&
        !['index.md', 'README.md'].includes(item)
      ) {
        try {
          const fileContent = fs.readFileSync(fullPath, 'utf-8');
          const { data, content: mdContent } = matter(fileContent);
          const category = relativePath.split('/')[0] || 'general';

          // Extract the first # heading from content for use as the full page title.
          // This may differ from frontmatter title (which is the sidebar label).
          const headingMatch = mdContent.match(/^#\s+(.+)$/m);
          const pageTitle = headingMatch ? headingMatch[1].trim() : undefined;

          structure.push({
            name: item.replace(/\.md$/i, ''),
            type: 'file',
            path: itemRelativePath.replace(/\.md$/i, ''),
            title: (data.title as string) || formatName(item.replace(/\.md$/i, '')),
            pageTitle,
            description: data.description as string | undefined,
            icon: resolveIcon(
              data.icon as string | undefined,
              'file',
              item.replace(/\.md$/i, ''),
              category
            ),
            category,
            order: (data.order as number) || 0,
            badge: data.badge as string | undefined,
            isExternal: (data.isExternal as boolean) || false,
            externalUrl: data.externalUrl as string | undefined,
            excerpt: generateExcerptFromContent(fileContent),
            componentName: data.componentName as string | undefined,
          });
        } catch (error) {
          console.warn(`Failed to parse metadata from ${fullPath}:`, error);
          // Fallback without metadata
          const category = relativePath.split('/')[0] || 'general';
          structure.push({
            name: item.replace(/\.md$/i, ''),
            type: 'file',
            path: itemRelativePath.replace(/\.md$/i, ''),
            title: formatName(item.replace(/\.md$/i, '')),
            icon: getDefaultIcon('file', item.replace(/\.md$/i, ''), category),
            category,
            order: 0
          });
        }
      }
    }
  }

  // Sort by order first, then by type, then by name
  return structure.sort((a, b) => {
    // First by order
    if (a.order !== b.order) {
      return (a.order || 0) - (b.order || 0);
    }
    // Then by type (directories first)
    if (a.type !== b.type) {
      return a.type === 'directory' ? -1 : 1;
    }
    // Finally by name
    return a.name.localeCompare(b.name);
  });
}

export function getDocPathFromSlug(slug: string[]): string {
  return slug.join('/');
}

export function getSlugFromPath(filePath: string): string[] {
  return filePath.split('/').filter(Boolean);
}

function formatName(name: string): string {
  // Convert kebab-case or snake_case to Title Case
  return name
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (l) => l.toUpperCase())
    .replace(/\s+/g, ' ')
    .trim();
}

// Function to get the next documentation item
export function getNextDocItem(
  currentPath: string,
  structure: DocItem[]
): { title: string; pageTitle: string; excerpt: string; slug: string } | null {
  // Flatten the structure into reading order.
  // Directories with an index page are included before their children
  // so that section transitions navigate to the section index first.
  const flattenItems = (items: DocItem[]): DocItem[] => {
    const result: DocItem[] = [];
    for (const item of items) {
      // Skip presets in catalog section from reading order — preset pages
      // are accessed via the right sidebar TOC link, not "Next article".
      if (item.name === 'presets' && item.path.startsWith('catalog/')) continue;

      if (item.type === 'file') {
        result.push(item);
      } else if (item.type === 'directory') {
        if (item.hasIndex) {
          result.push(item);
        }
        if (item.children) {
          result.push(...flattenItems(item.children));
        }
      }
    }
    return result;
  };

  const allItems = flattenItems(structure);
  const currentIndex = allItems.findIndex((item) => item.path === currentPath);

  if (currentIndex === -1 || currentIndex === allItems.length - 1) {
    return null;
  }

  const nextDocItem = allItems[currentIndex + 1];

  return {
    title: nextDocItem.title || '',
    pageTitle: nextDocItem.pageTitle || nextDocItem.title || '',
    excerpt: nextDocItem.excerpt || '',
    slug: nextDocItem.path,
  };
}

/**
 * Generates static params for documentation routes based on the documentation structure.
 * This function can be reused across different documentation pages that need to generate
 * static routes for markdown files and directories.
 *
 * @param structure - The documentation structure to generate params from
 * @returns Array of slug parameters for static generation
 */
export function generateStaticParamsFromStructure(structure: DocItem[]): { slug: string[] }[] {
  const params: { slug: string[] }[] = [];

  // Add the root docs path
  params.push({ slug: [] });

  const addPaths = (items: DocItem[], currentPath: string[] = []) => {
    items.forEach((item) => {
      if (item.type === 'file') {
        // Add the clean route (without .md extension)
        params.push({ slug: [...currentPath, item.name] });
      } else if (item.type === 'directory') {
        // If directory has an index file, add a path for the directory itself
        if (item.hasIndex) {
          params.push({ slug: [...currentPath, item.name] });
        }
        // Recursively add paths for children
        addPaths(item.children || [], [...currentPath, item.name]);
      }
    });
  };

  addPaths(structure);

  return params;
}

/**
 * Processes documentation slug parameters to handle markdown extensions and clean paths.
 * This function can be reused across different documentation pages that need to process
 * slug parameters and handle both clean routes and .md extension routes.
 *
 * @param slug - The slug array from the route parameters
 * @returns Object containing processed slug information
 */
export function processDocumentationSlug(slug: string[] = []) {
  // Strip .md extensions from slug parts to handle both clean routes and .md routes
  const cleanSlug = slug.map((part) => part.replace(/\.md$/i, ''));
  const path = cleanSlug.join('/');

  return {
    originalSlug: slug,
    cleanSlug,
    path: path || 'index',
  };
}

