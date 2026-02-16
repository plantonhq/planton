#!/usr/bin/env node

/**
 * Build script to generate the documentation structure JSON file.
 *
 * Scans: public/docs/ (markdown files with YAML frontmatter)
 * Outputs: public/docs-structure.json
 *
 * This replaces the API route at /api/docs/structure which doesn't work
 * with static exports served by `serve -s` (extensionless files are not
 * recognised, so the server falls back to index.html).
 */

import * as fs from 'fs';
import * as path from 'path';
import matter from 'gray-matter';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface DocItem {
  name: string;
  type: 'file' | 'directory';
  path: string;
  children?: DocItem[];
  title?: string;
  description?: string;
  icon?: string;
  category?: string;
  order?: number;
  badge?: string;
  isExternal?: boolean;
  externalUrl?: string;
  hasIndex?: boolean;
  excerpt?: string;
  componentName?: string;
}

// ---------------------------------------------------------------------------
// Icon maps (mirrors fileSystem.ts)
// ---------------------------------------------------------------------------

const iconMap: Record<string, string> = {
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

const categoryIcons: Record<string, string> = {
  docs: '📚',
  concepts: '💡',
  'deployment-components': '📦',
  deployment: '🚀',
  monitoring: '📊',
  security: '🔒',
};

function resolveIcon(
  metaIcon: string | undefined,
  type: 'file' | 'directory',
  name: string,
  category?: string,
): string {
  if (metaIcon) {
    const mapped = iconMap[metaIcon];
    if (mapped) return mapped;
  }
  return getDefaultIcon(type, name, category);
}

function getDefaultIcon(type: string, name: string, category?: string): string {
  const nameLower = name.toLowerCase();
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
  if (category && categoryIcons[category]) return categoryIcons[category];
  return type === 'directory' ? '📁' : '📄';
}

// ---------------------------------------------------------------------------
// Excerpt generation (mirrors utils.ts generateExcerptFromContent)
// ---------------------------------------------------------------------------

function generateExcerptFromContent(content: string, maxLength = 500): string {
  const stripped = content
    .replace(/^---[\s\S]*?---/, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/<[^>]*>/g, '')
    .replace(/^[-*_]{3,}$/gm, '')
    .replace(/^>\s+/gm, '')
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    .replace(/_{1,2}([^_]+)_{1,2}/g, '$1')
    .replace(/~~([^~]+)~~/g, '$1')
    .replace(/\|.*\|/g, '')
    .replace(/\n\s*\n/g, '\n')
    .replace(/\s+/g, ' ')
    .trim();

  if (stripped.length <= maxLength) return stripped;
  const truncated = stripped.substring(0, maxLength);
  const lastSpace = truncated.lastIndexOf(' ');
  if (lastSpace > maxLength * 0.8) return truncated.substring(0, lastSpace) + '...';
  return truncated + '...';
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatName(name: string): string {
  return name
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (l) => l.toUpperCase())
    .replace(/\s+/g, ' ')
    .trim();
}

// ---------------------------------------------------------------------------
// Structure builder (mirrors fileSystem.ts buildStructure)
// ---------------------------------------------------------------------------

function buildStructure(dirPath: string, relativePath = ''): DocItem[] {
  if (!fs.existsSync(dirPath)) return [];

  const items = fs.readdirSync(dirPath);
  const structure: DocItem[] = [];

  for (const item of items) {
    const fullPath = path.join(dirPath, item);
    const stat = fs.statSync(fullPath);
    const itemRelativePath = path.join(relativePath, item);

    if (stat.isDirectory()) {
      // In the catalog section, skip "presets" subdirectories inside
      // component directories so they don't appear in the sidebar.
      // Preset pages are still generated via the full structure in fileSystem.ts.
      const relParts = relativePath.split('/').filter(Boolean);
      if (item === 'presets' && relParts.length === 3 && relParts[0] === 'catalog') {
        continue;
      }

      const children = buildStructure(fullPath, itemRelativePath);
      const indexFiles = ['index.md', 'README.md'];
      const hasIndex = indexFiles.some((f) => fs.existsSync(path.join(fullPath, f)));

      if (children.length > 0 || hasIndex) {
        let metadata: Record<string, unknown> = {};

        for (const indexFile of indexFiles) {
          const indexPath = path.join(fullPath, indexFile);
          if (fs.existsSync(indexPath)) {
            try {
              const { data } = matter(fs.readFileSync(indexPath, 'utf-8'));
              metadata = data;
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
          title: (metadata.title as string) || formatName(item),
          description: metadata.description as string | undefined,
          icon: resolveIcon(metadata.icon as string | undefined, 'directory', item, category),
          category,
          order: (metadata.order as number) || 0,
          badge: metadata.badge as string | undefined,
          isExternal: (metadata.isExternal as boolean) || false,
          externalUrl: metadata.externalUrl as string | undefined,
          hasIndex,
          excerpt: '',
          componentName: metadata.componentName as string | undefined,
        });
      }
    } else if (item.endsWith('.md')) {
      if (
        !item.startsWith('prompt.') &&
        !item.startsWith('response.') &&
        !item.includes('.not-good.') &&
        !['index.md', 'README.md'].includes(item)
      ) {
        try {
          const fileContent = fs.readFileSync(fullPath, 'utf-8');
          const { data } = matter(fileContent);
          const category = relativePath.split('/')[0] || 'general';

          structure.push({
            name: item.replace(/\.md$/i, ''),
            type: 'file',
            path: itemRelativePath.replace(/\.md$/i, ''),
            title: (data.title as string) || formatName(item.replace(/\.md$/i, '')),
            description: data.description as string | undefined,
            icon: resolveIcon(
              data.icon as string | undefined,
              'file',
              item.replace(/\.md$/i, ''),
              category,
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
          const category = relativePath.split('/')[0] || 'general';
          structure.push({
            name: item.replace(/\.md$/i, ''),
            type: 'file',
            path: itemRelativePath.replace(/\.md$/i, ''),
            title: formatName(item.replace(/\.md$/i, '')),
            icon: getDefaultIcon('file', item.replace(/\.md$/i, ''), category),
            category,
            order: 0,
          });
        }
      }
    }
  }

  return structure.sort((a, b) => {
    if (a.order !== b.order) return (a.order || 0) - (b.order || 0);
    if (a.type !== b.type) return a.type === 'directory' ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main() {
  const docsDir = path.join(process.cwd(), 'public/docs');
  const outputFile = path.join(process.cwd(), 'public/docs-structure.json');

  console.log('📄 Generating docs-structure.json...');

  const structure = buildStructure(docsDir);
  fs.writeFileSync(outputFile, JSON.stringify(structure));

  // Count items recursively
  const countItems = (items: DocItem[]): number =>
    items.reduce((n, item) => n + 1 + (item.children ? countItems(item.children) : 0), 0);

  const total = countItems(structure);
  const sizeKB = Math.round(fs.statSync(outputFile).size / 1024);

  console.log(`✅ Generated docs-structure.json (${total} items, ${sizeKB} KB)`);
}

main();
