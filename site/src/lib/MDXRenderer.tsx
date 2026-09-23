'use client';

import React from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeHighlight from 'rehype-highlight';
import matter from 'gray-matter';
import { formatDate } from '@/lib/utils';
import { Author } from '@/lib/types-client';
import { PageActions } from '@/components/common/PageActions';
import CloudflareVideo, { getEmbedInfoFromUrl } from '@/components/media/CloudflareVideo';
import { CodeBlock, MermaidDiagram } from '@/components/common';
import { attribute, contentChildren, elementChildren, isElement, mermaidSource, textOf } from '@/lib/hast';
import { HeadingWithAnchor, generateHeadingId } from '@/components/docs';
import {
  HEADING_H1_CLASSES,
  HEADING_H2_CLASSES,
  HEADING_H3_CLASSES,
  HEADING_H4_CLASSES,
  HEADING_H5_CLASSES,
  HEADING_H6_CLASSES,
  LINK_CLASSES,
  TAG_CLASSES,
  BLOCKQUOTE_CLASSES,
  BLOCKQUOTE_WARNING_CLASSES,
  INLINE_CODE_CLASSES,
  PARAGRAPH_CLASSES,
  LIST_CLASSES,
  NEXT_ARTICLE_BUTTON_CLASSES,
  NEXT_ARTICLE_CARD_CLASSES,
  TABLE_WRAPPER_CLASSES,
  TABLE_CLASSES,
  TABLE_HEAD_CLASSES,
  TABLE_ROW_CLASSES,
  TABLE_HEADER_CLASSES,
  TABLE_CELL_CLASSES,
  HR_CLASSES,
} from '@/theme/docs';

interface MdxMetadata {
  title: string;
  date: string;
  author: Author[];
  featuredImage?: string;
  featuredImageType?: string;
  featuredVideo?: string;
  tags: string[];
  content: string;
  hideCopyMarkdown?: boolean;
  hideViewMarkdown?: boolean;
}

interface MDXRendererProps {
  mdxContent: string;
  markdownContent?: string;
  title?: string;
  nextArticle?: {
    title: string;
    excerpt?: string;
    slug: string;
  };
  path: string;
  /** Resolved path to the actual .md file on disk (e.g. "/docs/platform/index.md"). Used for "Open Raw". */
  rawPath?: string;
}

// NextArticle component for navigation
interface NextArticleProps {
  nextArticle?: {
    title: string;
    excerpt?: string;
    slug: string;
  };
}

const NextArticle: React.FC<NextArticleProps> = ({ nextArticle }) => {
  if (!nextArticle) return null;

  return (
    <div className={NEXT_ARTICLE_CARD_CLASSES}>
      <div className="max-w-none">
        <p className="text-base md:text-lg text-[#666] m-0 font-bold">Next article</p>
        <h3 className="text-lg md:text-xl font-bold text-white m-0 my-2">{nextArticle.title}</h3>
        {nextArticle.excerpt && (
          <div className="relative mb-4 min-h-24">
            <div className="text-[#a0a0a0] leading-6 excerpt-text">{nextArticle.excerpt}</div>
            <div className="excerpt-gradient" />
          </div>
        )}
        <a
          href={nextArticle.slug}
          className={NEXT_ARTICLE_BUTTON_CLASSES}
        >
          Read next article
        </a>
      </div>
    </div>
  );
};

export const MDXRenderer: React.FC<MDXRendererProps> = ({
  mdxContent,
  markdownContent,
  title,
  nextArticle,
  path,
  rawPath,
}) => {
  const { data, content: rawContent } = matter(mdxContent);
  const metadata: MdxMetadata = data as MdxMetadata;

  // Strip leading h1 from the body when it duplicates the frontmatter title.
  // Many markdown files start with `# Title` that repeats the frontmatter —
  // rendering both creates a visual stutter.
  const content = metadata.title
    ? rawContent.replace(
        new RegExp(`^\\s*#\\s+${metadata.title.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*\n`),
        '',
      )
    : rawContent;

  // Configuration for page actions
  const hideCopyMarkdown = metadata?.hideCopyMarkdown ?? false;
  const hideViewMarkdown = metadata?.hideViewMarkdown ?? false;
  const shouldShowActions = markdownContent && !(hideCopyMarkdown && hideViewMarkdown);

  return (
    <div className="w-full">
      <article data-pagefind-body>
        {/* Header */}
        <header className="mb-6 md:mb-8">
          {/* Title row with page actions */}
          {metadata.title && (
            <div className="flex items-start gap-2">
              <h1 className="flex-1 text-2xl sm:text-3xl md:text-4xl font-bold text-[#b0b0b0] mb-3 md:mb-4">
                {metadata.title}
              </h1>
              {shouldShowActions && (
                <PageActions
                  markdownContent={markdownContent}
                  title={title || metadata.title}
                  path={path}
                  rawPath={rawPath}
                  hideCopyMarkdown={hideCopyMarkdown}
                  hideViewMarkdown={hideViewMarkdown}
                />
              )}
            </div>
          )}

          {/* Date and Author */}
          {(metadata.date || metadata.author) && (
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-[#a0a0a0] text-sm md:text-base mb-4 md:mb-6">
              {metadata.date && <time dateTime={metadata.date}>{formatDate(metadata.date)}</time>}
              {metadata.author && (
                <>
                  {metadata.date && <span className="hidden sm:inline">•</span>}
                  <div className="flex gap-2">
                    {metadata.author.map((author, index) => (
                      <span key={index} className="font-medium">
                        {author.name}
                      </span>
                    ))}
                  </div>
                </>
              )}
            </div>
          )}

          {/* Tags — wraps on narrow screens */}
          {metadata.tags && (
            <div className="flex flex-wrap gap-2 mb-4 md:mb-6">
              {metadata.tags.map((tag, index) => (
                <span
                  key={index}
                  className={TAG_CLASSES}
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Featured Image */}
          {metadata.featuredImage && (
            <div className="mb-6">
              {/* The image and its dimensions come from the document's frontmatter, unknown at build time; images are unoptimized on this static export, so next/image would emit the same tag. */}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={metadata.featuredImage}
                alt={metadata.title}
                className={`w-full rounded-lg shadow-lg ${
                  metadata.featuredImageType === 'full'
                    ? 'h-96 object-cover'
                    : 'max-h-96 object-contain'
                }`}
              />
            </div>
          )}

          {/* Featured Video */}
          {metadata.featuredVideo && (
            <div className="mb-6">
              <CloudflareVideo url={metadata.featuredVideo} title={metadata.title} />
            </div>
          )}
        </header>

        {/* Content — prose scales up from base on mobile to lg on desktop */}
        <div className="prose max-w-none prose-invert md:prose-lg tracking-tight">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeRaw, rehypeHighlight]}
            components={{
              p: ({ children, node }) => {
                // A paragraph that is only a bare link to a video becomes the embed.
                const content = contentChildren(node);
                const only = content.length === 1 ? content[0] : undefined;
                const href = isElement(only) && only.tagName === 'a' ? attribute(only, 'href') : undefined;
                if (only && href) {
                  const linkText = textOf(only);
                  const shouldEmbed = !linkText || linkText.trim() === href.trim();
                  if (shouldEmbed && getEmbedInfoFromUrl(href)) {
                    return <CloudflareVideo url={href} title={metadata.title} />;
                  }
                }
                return <p className={PARAGRAPH_CLASSES}>{children}</p>;
              },
              h1: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={1} className={HEADING_H1_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              h2: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={2} className={HEADING_H2_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              h3: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={3} className={HEADING_H3_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              h4: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={4} className={HEADING_H4_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              h5: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={5} className={HEADING_H5_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              h6: ({ children }) => (
                <HeadingWithAnchor id={generateHeadingId(children)} level={6} className={HEADING_H6_CLASSES}>
                  {children}
                </HeadingWithAnchor>
              ),
              ul: ({ children }) => (
                <ul className={`list-disc list-inside ${LIST_CLASSES}`}>{children}</ul>
              ),
              ol: ({ children }) => (
                <ol className={`list-decimal list-inside ${LIST_CLASSES}`}>
                  {children}
                </ol>
              ),
              li: ({ children }) => <li className="text-[#a0a0a0]">{children}</li>,
              strong: ({ children }) => <strong className="font-semibold text-[#b0b0b0]">{children}</strong>,
              blockquote: ({ children, node }) => {
                // Detect callout type from the HAST tree. Blockquotes
                // starting with **Tip:**, **Note:**, **Warning:** etc.
                // get the appropriate treatment. Warning/Caution get a
                // semantic red border; all others get the neutral style.
                const firstPara = elementChildren(node).find((c) => c.tagName === 'p');
                const firstChild = elementChildren(firstPara)[0];
                const lead = firstChild?.tagName === 'strong' ? textOf(firstChild).toLowerCase() : '';
                const isWarning = lead.startsWith('warning') || lead.startsWith('caution');

                return (
                  <blockquote
                    className={
                      isWarning ? BLOCKQUOTE_WARNING_CLASSES : BLOCKQUOTE_CLASSES
                    }
                  >
                    {children}
                  </blockquote>
                );
              },
              code: ({ children, className }) => {
                // Fenced code blocks receive hljs/language-* classes from
                // rehype-highlight. Only apply inline-code styling to true
                // inline code (which has no className).
                if (className) {
                  return <code className={className}>{children}</code>;
                }
                return (
                  <code className={INLINE_CODE_CLASSES}>
                    {children}
                  </code>
                );
              },
              pre: ({ children, node }) => {
                // A fenced mermaid block renders as a diagram.
                const mermaid = mermaidSource(node);
                if (mermaid !== undefined) return <MermaidDiagram chart={mermaid} />;
                return <CodeBlock>{children}</CodeBlock>;
              },
              a: ({ href, children }) => {
                const isExternal = href?.startsWith('http://') || href?.startsWith('https://');
                if (isExternal) {
                  return (
                    <a
                      href={href}
                      className={LINK_CLASSES}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {children}
                    </a>
                  );
                }
                return (
                  <Link
                    href={href || '#'}
                    className={LINK_CLASSES}
                  >
                    {children}
                  </Link>
                );
              },
              img: ({ src, alt }) => {
                if (!src) return null;
                // Avoid wrapping with a div to prevent <div> inside <p> which breaks hydration.
                // The image is the document's own, dimensions unknown at build time; images are
                // unoptimized on this static export, so next/image would emit the same tag.
                return (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={src}
                    alt={alt || ''}
                    className="max-w-full h-auto rounded-lg shadow-lg my-4 md:my-6 block"
                  />
                );
              },
              table: ({ children }) => (
                <div className={TABLE_WRAPPER_CLASSES}>
                  <table className={TABLE_CLASSES}>
                    {children}
                  </table>
                </div>
              ),
              thead: ({ children }) => <thead className={TABLE_HEAD_CLASSES}>{children}</thead>,
              tbody: ({ children }) => <tbody>{children}</tbody>,
              tr: ({ children }) => <tr className={TABLE_ROW_CLASSES}>{children}</tr>,
              th: ({ children }) => (
                <th className={TABLE_HEADER_CLASSES}>{children}</th>
              ),
              td: ({ children }) => <td className={TABLE_CELL_CLASSES}>{children}</td>,
              hr: () => <hr className={HR_CLASSES} />,
            }}
          >
            {content}
          </ReactMarkdown>
        </div>

        {/* Next Article Section */}
        <NextArticle nextArticle={nextArticle} />
      </article>
    </div>
  );
};
