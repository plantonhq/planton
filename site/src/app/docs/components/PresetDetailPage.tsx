'use client';

import React from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import { PresetRankBadge } from './PresetRankBadge';
import { PresetYamlViewer } from './PresetYamlViewer';
import { ArrowLeft, FileCode, FileText } from 'lucide-react';

interface PresetDetailPageProps {
  title: string;
  rank: string;
  presetSlug: string;
  componentSlug: string;
  componentTitle: string;
  provider: string;
  yamlContent: string;
  /** Raw markdown body (frontmatter already stripped by the caller). */
  mdContent: string;
  /** Base path for raw file links, e.g. "/docs/catalog/aws/documentdb/presets" */
  basePath: string;
}

/**
 * Detail page for a single preset.  Shows:
 * 1. Back navigation to the presets list
 * 2. Rank badge + title
 * 3. Action buttons (Raw YAML, Raw MD)
 * 4. Full YAML manifest viewer with copy
 * 5. Rendered description markdown
 */
export const PresetDetailPage: React.FC<PresetDetailPageProps> = ({
  title,
  rank,
  presetSlug,
  componentTitle,
  yamlContent,
  mdContent,
  basePath,
}) => {
  const rawYamlUrl = `${basePath}/${presetSlug}.yaml`;
  const rawMdUrl = `${basePath}/${presetSlug}.md`;

  return (
    <div className="w-full">
      {/* Back navigation */}
      <Link
        href={basePath}
        className="inline-flex items-center gap-1.5 text-sm text-gray-400 hover:text-purple-400 transition-colors mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        {componentTitle} Presets
      </Link>

      {/* Header row: rank + title + actions */}
      <div className="flex flex-col sm:flex-row sm:items-start gap-4 mb-8">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <PresetRankBadge rank={rank} />
          <h1 className="text-2xl font-bold text-white truncate">{title}</h1>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <a
            href={rawYamlUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-slate-800 border border-slate-700/50 text-gray-300 hover:text-white hover:bg-slate-700 transition-all"
          >
            <FileCode className="w-3.5 h-3.5" />
            Raw YAML
          </a>
          <a
            href={rawMdUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-slate-800 border border-slate-700/50 text-gray-300 hover:text-white hover:bg-slate-700 transition-all"
          >
            <FileText className="w-3.5 h-3.5" />
            Raw Markdown
          </a>
        </div>
      </div>

      {/* YAML manifest */}
      <PresetYamlViewer yamlContent={yamlContent} />

      {/* Description body */}
      <div className="prose prose-lg max-w-none prose-invert">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeRaw]}
          components={{
            // Strip the H1 since we already show the title above
            h1: () => null,
            p: ({ children }) => (
              <p className="text-gray-300 mb-4 leading-relaxed">{children}</p>
            ),
            h2: ({ children }) => {
              const id = children
                ?.toString()
                .toLowerCase()
                .replace(/[^a-z0-9\s-]/g, '')
                .replace(/\s+/g, '-');
              return (
                <h2 id={id} className="text-xl font-bold text-white mt-8 mb-3">
                  {children}
                </h2>
              );
            },
            h3: ({ children }) => {
              const id = children
                ?.toString()
                .toLowerCase()
                .replace(/[^a-z0-9\s-]/g, '')
                .replace(/\s+/g, '-');
              return (
                <h3 id={id} className="text-lg font-bold text-white mt-5 mb-2">
                  {children}
                </h3>
              );
            },
            ul: ({ children }) => (
              <ul className="list-disc list-inside text-gray-300 mb-4 space-y-2">{children}</ul>
            ),
            ol: ({ children }) => (
              <ol className="list-decimal list-inside text-gray-300 mb-4 space-y-2">{children}</ol>
            ),
            li: ({ children }) => <li className="text-gray-300">{children}</li>,
            strong: ({ children }) => (
              <strong className="text-white font-semibold">{children}</strong>
            ),
            code: ({ children, className }) => {
              const isInline = !className;
              if (isInline) {
                return (
                  <code className="bg-slate-800/60 text-sky-300 px-1.5 py-0.5 rounded text-sm">
                    {children}
                  </code>
                );
              }
              return <code className={className}>{children}</code>;
            },
            a: ({ href, children }) => (
              <a
                href={href}
                className="text-purple-400 hover:text-purple-300 underline"
                target={href?.startsWith('http') ? '_blank' : undefined}
                rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
              >
                {children}
              </a>
            ),
            table: ({ children }) => (
              <div className="overflow-x-auto my-6">
                <table className="min-w-full bg-slate-900 border border-purple-900/30 rounded-lg">
                  {children}
                </table>
              </div>
            ),
            thead: ({ children }) => <thead className="bg-purple-900/20">{children}</thead>,
            tbody: ({ children }) => <tbody>{children}</tbody>,
            tr: ({ children }) => <tr className="border-b border-purple-900/30">{children}</tr>,
            th: ({ children }) => (
              <th className="px-4 py-3 text-left text-white font-semibold">{children}</th>
            ),
            td: ({ children }) => <td className="px-4 py-3 text-gray-300">{children}</td>,
            hr: () => <hr className="my-8 border-purple-900/30" />,
          }}
        >
          {mdContent}
        </ReactMarkdown>
      </div>
    </div>
  );
};
