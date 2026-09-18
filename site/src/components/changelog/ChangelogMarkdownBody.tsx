'use client';

import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeHighlight from 'rehype-highlight';
import MermaidDiagram from '@/components/common/MermaidDiagram';
import { mermaidSource } from '@/lib/hast';
import {
  HEADING_H1_CLASSES,
  HEADING_H2_CLASSES,
  HEADING_H3_CLASSES,
  HEADING_H4_CLASSES,
} from '@/theme/docs';

interface ChangelogMarkdownBodyProps {
  /** Markdown body content (no frontmatter). */
  content: string;
  className?: string;
}

/**
 * Renders markdown content with planton.ai dark-theme styling and Mermaid
 * diagram support. Used for both the inline-expand view on the changelog list
 * page and the standalone detail page.
 */
const ChangelogMarkdownBody: React.FC<ChangelogMarkdownBodyProps> = ({
  content,
  className = '',
}) => {
  return (
    <div className={`prose prose-invert max-w-none ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw, rehypeHighlight]}
        components={{
          // ---- Headings ----
          h1: ({ children, ...props }) => (
            <h1 className={HEADING_H1_CLASSES} {...props}>{children}</h1>
          ),
          h2: ({ children, ...props }) => (
            <h2 className={HEADING_H2_CLASSES} {...props}>{children}</h2>
          ),
          h3: ({ children, ...props }) => (
            <h3 className={HEADING_H3_CLASSES} {...props}>{children}</h3>
          ),
          h4: ({ children, ...props }) => (
            <h4 className={HEADING_H4_CLASSES} {...props}>{children}</h4>
          ),

          // ---- Body text ----
          p: ({ children, ...props }) => (
            <p className="text-[#a0a0a0] mb-4 leading-relaxed" {...props}>
              {children}
            </p>
          ),
          strong: ({ children, ...props }) => (
            <strong className="font-bold text-white" {...props}>
              {children}
            </strong>
          ),
          em: ({ children, ...props }) => (
            <em className="italic text-[#d4d4d4]" {...props}>
              {children}
            </em>
          ),

          // ---- Links ----
          a: ({ children, href, ...props }) => (
            <a
              href={href}
              className="text-white hover:text-white/70 underline decoration-white/20"
              target={href?.startsWith('http') ? '_blank' : undefined}
              rel={
                href?.startsWith('http')
                  ? 'noopener noreferrer'
                  : undefined
              }
              {...props}
            >
              {children}
            </a>
          ),

          // ---- Lists ----
          ul: ({ children, ...props }) => (
            <ul
              className="list-disc pl-6 space-y-2 mb-4 text-[#a0a0a0]"
              {...props}
            >
              {children}
            </ul>
          ),
          ol: ({ children, ...props }) => (
            <ol
              className="list-decimal pl-6 space-y-2 mb-4 text-[#a0a0a0]"
              {...props}
            >
              {children}
            </ol>
          ),
          li: ({ children, ...props }) => (
            <li className="text-[#a0a0a0]" {...props}>
              {children}
            </li>
          ),

          // ---- Blockquote ----
          blockquote: ({ children, ...props }) => (
            <blockquote
              className="border-l-2 border-[#3a3a3a] pl-4 py-3 my-5 bg-[#111] rounded-r text-[#a0a0a0] italic"
              {...props}
            >
              {children}
            </blockquote>
          ),

          // ---- Code ----
          code: ({ children, className: codeClassName, ...props }) => (
            <code
              className={`bg-[#2a2a2a] text-white rounded text-sm break-words ${codeClassName || ''}`}
              {...props}
            >
              {children}
            </code>
          ),
          pre: ({ children, node }) => {
            // A fenced mermaid block renders as a diagram.
            const mermaid = mermaidSource(node);
            if (mermaid !== undefined) return <MermaidDiagram chart={mermaid} />;
            return (
              <pre className="bg-[#1a1a1a] p-4 rounded-lg overflow-x-auto my-4 border border-[#2a2a2a]">
                {children}
              </pre>
            );
          },

          // ---- Tables ----
          table: ({ children, ...props }) => (
            <div className="overflow-x-auto my-4 border border-[#2a2a2a] rounded-lg">
              <table
                className="min-w-full divide-y divide-[#2a2a2a]"
                {...props}
              >
                {children}
              </table>
            </div>
          ),
          thead: ({ children, ...props }) => (
            <thead className="bg-[#111]" {...props}>
              {children}
            </thead>
          ),
          th: ({ children, ...props }) => (
            <th
              className="px-4 py-2 text-left text-sm font-semibold text-white"
              {...props}
            >
              {children}
            </th>
          ),
          td: ({ children, ...props }) => (
            <td
              className="px-4 py-2 border-t border-[#2a2a2a] text-[#a0a0a0]"
              {...props}
            >
              {children}
            </td>
          ),

          // ---- Misc ----
          // The image is the entry's own, dimensions unknown at build time; images are
          // unoptimized on this static export, so next/image would emit the same tag.
          img: ({ src, alt, ...props }) =>
            src ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={src}
                alt={alt || ''}
                className="rounded-lg my-4 max-w-full h-auto border border-[#2a2a2a]"
                {...props}
              />
            ) : null,
          hr: (props) => (
            <hr className="my-8 border-[#2a2a2a]" {...props} />
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};

export default ChangelogMarkdownBody;
