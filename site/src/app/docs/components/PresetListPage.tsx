'use client';

import React, { useState, useCallback, useRef } from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import {
  ArrowLeft,
  ChevronDown,
  ExternalLink,
  Link as LinkIcon,
  Check,
} from 'lucide-react';
import { PresetRankBadge } from './PresetRankBadge';
import { PresetYamlViewer } from './PresetYamlViewer';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface PresetEntry {
  slug: string;
  rank: string;
  title: string;
  excerpt: string;
  yamlContent: string;
  mdContent: string;
}

interface PresetListPageProps {
  componentTitle: string;
  componentSlug: string;
  provider: string;
  presets: PresetEntry[];
  /** Base path for links, e.g. "/docs/catalog/aws/documentdb/presets" */
  basePath: string;
  /** Path to the component catalog page, e.g. "/docs/catalog/aws/documentdb" */
  catalogPath: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export const PresetListPage: React.FC<PresetListPageProps> = ({
  componentTitle,
  presets,
  basePath,
  catalogPath,
}) => {
  const [expandedSlug, setExpandedSlug] = useState<string | null>(null);
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const toggleEntry = useCallback(
    (slug: string) => {
      setExpandedSlug(expandedSlug === slug ? null : slug);
    },
    [expandedSlug],
  );

  const copyLink = useCallback(
    async (slug: string, e: React.MouseEvent) => {
      e.stopPropagation();
      const url = `${window.location.origin}${basePath}/${slug}`;
      await navigator.clipboard.writeText(url);
      setCopiedSlug(slug);
      if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current);
      copyTimeoutRef.current = setTimeout(() => setCopiedSlug(null), 2000);
    },
    [basePath],
  );

  const openInNewPage = useCallback(
    (slug: string, e: React.MouseEvent) => {
      e.stopPropagation();
      window.open(`${basePath}/${slug}`, '_blank');
    },
    [basePath],
  );

  return (
    <div className="w-full">
      {/* Back navigation */}
      <Link
        href={catalogPath}
        className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        {componentTitle}
      </Link>

      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground mb-2">
          {componentTitle} Presets
        </h1>
        <p className="text-muted-foreground">
          Ready-to-deploy configuration presets for {componentTitle}. Each preset
          is a complete manifest you can copy, customize, and deploy.
        </p>
        <span className="inline-block mt-3 px-3 py-1 rounded-full text-xs font-medium bg-secondary text-muted-foreground border border-border">
          {presets.length} {presets.length === 1 ? 'preset' : 'presets'}
        </span>
      </div>

      {/* Accordion entries */}
      <div className="space-y-1">
        {presets.map((preset) => {
          const isExpanded = expandedSlug === preset.slug;
          const isCopied = copiedSlug === preset.slug;

          return (
            <article key={preset.slug} className="relative">
              {/* Clickable header */}
              <div
                onClick={() => toggleEntry(preset.slug)}
                className="relative py-5 px-5 -mx-5 rounded-lg hover:bg-secondary/40 transition-all duration-200 cursor-pointer group"
              >
                {/* Row 1: rank + title + action icons */}
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <PresetRankBadge rank={preset.rank} />
                    <h3 className="text-lg font-semibold text-foreground group-hover:text-foreground transition-colors truncate">
                      {preset.title}
                    </h3>
                  </div>

                  {/* Action icons (visible on hover) */}
                  <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                    <button
                      onClick={(e) => copyLink(preset.slug, e)}
                      className="p-1.5 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
                      title="Copy link"
                    >
                      {isCopied ? (
                        <Check className="w-4 h-4 text-success" />
                      ) : (
                        <LinkIcon className="w-4 h-4" />
                      )}
                    </button>
                    <button
                      onClick={(e) => openInNewPage(preset.slug, e)}
                      className="p-1.5 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
                      title="Open in new page"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </button>
                    <div
                      className={`p-1.5 text-muted-foreground transition-transform duration-200 ${
                        isExpanded ? 'rotate-180' : ''
                      }`}
                    >
                      <ChevronDown className="w-4 h-4" />
                    </div>
                  </div>
                </div>

                {/* Row 2: excerpt (only when collapsed) */}
                {!isExpanded && (
                  <p className="text-muted-foreground text-sm leading-relaxed mt-2 ml-11">
                    {preset.excerpt}
                  </p>
                )}
              </div>

              {/* Expanded content */}
              {isExpanded && (
                <div className="px-5 -mx-5 pb-6 pt-2">
                  {/* YAML viewer */}
                  <PresetYamlViewer yamlContent={preset.yamlContent} />

                  {/* Rendered description markdown */}
                  <div className="prose prose-lg max-w-none prose-invert">
                    <ReactMarkdown
                      remarkPlugins={[remarkGfm]}
                      rehypePlugins={[rehypeRaw]}
                      components={{
                        // Strip the H1 since the title is already in the header
                        h1: () => null,
                        p: ({ children }) => (
                          <p className="text-muted-foreground mb-4 leading-relaxed">
                            {children}
                          </p>
                        ),
                        h2: ({ children }) => (
                          <h2 className="text-lg font-bold text-foreground mt-6 mb-2">
                            {children}
                          </h2>
                        ),
                        h3: ({ children }) => (
                          <h3 className="text-base font-bold text-foreground mt-4 mb-2">
                            {children}
                          </h3>
                        ),
                        ul: ({ children }) => (
                          <ul className="list-disc list-inside text-muted-foreground mb-4 space-y-1">
                            {children}
                          </ul>
                        ),
                        li: ({ children }) => (
                          <li className="text-muted-foreground">{children}</li>
                        ),
                        strong: ({ children }) => (
                          <strong className="text-foreground font-semibold">
                            {children}
                          </strong>
                        ),
                        code: ({ children, className }) => {
                          const isInline = !className;
                          if (isInline) {
                            return (
                              <code className="bg-secondary text-foreground px-1.5 py-0.5 rounded text-sm">
                                {children}
                              </code>
                            );
                          }
                          return <code className={className}>{children}</code>;
                        },
                        table: ({ children }) => (
                          <div className="overflow-x-auto my-4">
                            <table className="min-w-full bg-card border border-border rounded-lg text-sm">
                              {children}
                            </table>
                          </div>
                        ),
                        thead: ({ children }) => (
                          <thead className="bg-secondary">{children}</thead>
                        ),
                        tbody: ({ children }) => <tbody>{children}</tbody>,
                        tr: ({ children }) => (
                          <tr className="border-b border-border">
                            {children}
                          </tr>
                        ),
                        th: ({ children }) => (
                          <th className="px-3 py-2 text-left text-foreground font-semibold">
                            {children}
                          </th>
                        ),
                        td: ({ children }) => (
                          <td className="px-3 py-2 text-muted-foreground">
                            {children}
                          </td>
                        ),
                      }}
                    >
                      {preset.mdContent}
                    </ReactMarkdown>
                  </div>

                  {/* Link to full detail page */}
                  <Link
                    href={`${basePath}/${preset.slug}`}
                    className="inline-flex items-center gap-1.5 mt-4 text-sm text-foreground hover:text-muted-foreground transition-colors"
                  >
                    Open full page
                    <ExternalLink className="w-3.5 h-3.5" />
                  </Link>
                </div>
              )}
            </article>
          );
        })}
      </div>
    </div>
  );
};
