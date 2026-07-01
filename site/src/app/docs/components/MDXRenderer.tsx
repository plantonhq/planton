'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeHighlight from 'rehype-highlight';
import matter from 'gray-matter';
import { formatDate } from '@/lib/utils';
import { Author } from '@/lib/mdx';
import { CodeBlock } from '@/app/docs/components/CodeBlock';
import { MermaidBlock } from '@/app/docs/components/MermaidBlock';
import {
  CatalogProviderGrid,
  CatalogProvider,
} from '@/app/docs/components/CatalogProviderGrid';
import 'highlight.js/styles/github-dark.css';

function extractMermaidText(node: React.ReactNode): string {
  if (typeof node === 'string') return node;
  if (typeof node === 'number') return String(node);
  if (!node) return '';
  if (Array.isArray(node)) return node.map(extractMermaidText).join('');
  if (React.isValidElement(node) && node.props) {
    return extractMermaidText((node.props as { children?: React.ReactNode }).children);
  }
  return '';
}

interface MdxMetadata {
  title: string;
  date?: string;
  author?: Author[];
  featuredImage?: string;
  featuredImageType?: string;
  tags?: string[];
  content: string;
}

interface MDXRendererProps {
  mdxContent: string;
  /** When set, the catalog provider grid is rendered after the markdown content. */
  catalogProviders?: CatalogProvider[];
  nextArticle?: {
    title: string;
    excerpt?: string;
    slug: string;
  };
}

// ---------------------------------------------------------------------------
// MarkdownImage — proper React component so it can hold error state.
// Replaces the previous inline arrow function in the ReactMarkdown components
// map.  Detects provider icon images and shows a letter-badge fallback when
// the image fails to load.
//
// react-markdown passes the full set of <img> HTML attributes plus its own
// ExtraProps, so the component accepts React.ImgHTMLAttributes and spreads
// only the subset it cares about.
// ---------------------------------------------------------------------------

const MarkdownImage: React.FC<
  React.ImgHTMLAttributes<HTMLImageElement>
> = (props) => {
  const { src, alt, className, ...rest } = props;
  const [error, setError] = useState(false);

  if (!src || typeof src !== 'string') return null;

  // Provider icons live at /images/providers/{name}.svg (no subdirectory).
  const isProviderIcon =
    src.startsWith('/images/providers/') &&
    src.endsWith('.svg') &&
    src.split('/').length === 4; // exactly /images/providers/foo.svg

  if (error && isProviderIcon) {
    const letter = (alt || '?').charAt(0).toUpperCase();
    return (
      <span
        className="inline-flex items-center justify-center rounded bg-secondary text-sm font-bold text-muted-foreground flex-shrink-0 w-8 h-8"
        aria-label={alt || 'icon'}
      >
        {letter}
      </span>
    );
  }

  const finalClassName =
    className || 'max-w-full h-auto rounded-lg shadow-lg my-6 block';

  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      {...rest}
      src={src}
      alt={alt || ''}
      className={finalClassName}
      onError={() => setError(true)}
    />
  );
};

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
    <div className="mt-12 p-6 rounded-lg bg-secondary border border-border">
      <div className="max-w-none">
        <p className="text-lg text-muted-foreground m-0 font-bold">Next article</p>
        <h3 className="text-xl font-bold text-foreground m-0 my-2">{nextArticle.title}</h3>
        {nextArticle.excerpt && (
          <div className="relative mb-4 min-h-24">
            <div className="text-muted-foreground leading-6">{nextArticle.excerpt}</div>
          </div>
        )}
        <Link
          href={nextArticle.slug}
          className="inline-flex items-center px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground font-semibold rounded-md transition-colors duration-200 hover:translate-y-[-1px] active:translate-y-[1px]"
        >
          Read next article
        </Link>
      </div>
    </div>
  );
};

export const MDXRenderer: React.FC<MDXRendererProps> = ({
  mdxContent,
  catalogProviders,
  nextArticle,
}) => {
  const { data, content } = matter(mdxContent);
  const metadata: MdxMetadata = data as MdxMetadata;

  return (
    <div className="w-full">
      <article>
        {/* Header */}
        <header className="mb-8">
          {/* Date and Author */}
          {(metadata.date || metadata.author) && (
            <div className="flex items-center gap-4 text-muted-foreground mb-6">
              {metadata.date && <time dateTime={metadata.date}>{formatDate(metadata.date)}</time>}
              {metadata.author && (
                <>
                  {metadata.date && <span>•</span>}
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

          {/* Tags */}
          {metadata.tags && (
            <div className="flex gap-2 mb-6">
              {metadata.tags.map((tag, index) => (
                <span
                  key={index}
                  className="px-3 py-1 bg-secondary text-muted-foreground text-sm font-medium rounded-full border border-border"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Featured Image */}
          {metadata.featuredImage && (
            <div className="mb-6">
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
        </header>

        {/* Content */}
        <div className="prose prose-lg max-w-none prose-invert">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeRaw, [rehypeHighlight, { detect: false }]]}
            components={{
              p: ({ children }) => (
                <p className="text-muted-foreground mb-4 leading-relaxed">{children}</p>
              ),
              h1: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h1 id={id} className="text-3xl font-bold text-foreground mt-8 mb-4">
                    {children}
                  </h1>
                );
              },
              h2: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h2 id={id} className="text-2xl font-bold text-foreground mt-6 mb-3">
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
                  <h3 id={id} className="text-xl font-bold text-foreground mt-5 mb-2">
                    {children}
                  </h3>
                );
              },
              h4: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h4 id={id} className="text-lg font-bold text-foreground mt-4 mb-2">
                    {children}
                  </h4>
                );
              },
              h5: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h5 id={id} className="text-base font-bold text-foreground mt-3 mb-2">
                    {children}
                  </h5>
                );
              },
              h6: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h6 id={id} className="text-sm font-bold text-foreground mt-2 mb-1">
                    {children}
                  </h6>
                );
              },
              ul: ({ children }) => (
                <ul className="list-disc list-inside text-muted-foreground mb-4 space-y-2">{children}</ul>
              ),
              ol: ({ children }) => (
                <ol className="list-decimal list-inside text-muted-foreground mb-4 space-y-2">
                  {children}
                </ol>
              ),
              li: ({ children }) => <li className="text-muted-foreground">{children}</li>,
              blockquote: ({ children }) => (
                <blockquote className="border-l-4 border-border pl-4 py-2 my-4 bg-secondary rounded-r text-muted-foreground italic">
                  {children}
                </blockquote>
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
              pre: ({ children }) => {
                const child = React.Children.toArray(children)[0];
                if (
                  React.isValidElement<{ className?: string; children?: React.ReactNode }>(child) &&
                  child.props?.className?.includes('language-mermaid')
                ) {
                  const text = extractMermaidText(child.props.children);
                  if (text) return <MermaidBlock chart={text} />;
                }
                return <CodeBlock>{children}</CodeBlock>;
              },
              a: ({ href, children }) => {
                const isExternal = href?.startsWith('http');
                if (!isExternal && href) {
                  return (
                    <Link
                      href={href}
                      className="text-foreground underline underline-offset-2 hover:text-muted-foreground"
                    >
                      {children}
                    </Link>
                  );
                }
                return (
                  <a
                    href={href}
                    className="text-foreground underline underline-offset-2 hover:text-muted-foreground"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {children}
                  </a>
                );
              },
              img: MarkdownImage,
              table: ({ children }) => (
                <div className="overflow-x-auto my-6">
                  <table className="min-w-full bg-card border border-border rounded-lg">
                    {children}
                  </table>
                </div>
              ),
              thead: ({ children }) => <thead className="bg-secondary">{children}</thead>,
              tbody: ({ children }) => <tbody>{children}</tbody>,
              tr: ({ children }) => <tr className="border-b border-border">{children}</tr>,
              th: ({ children }) => (
                <th className="px-4 py-3 text-left text-foreground font-semibold">{children}</th>
              ),
              td: ({ children }) => <td className="px-4 py-3 text-muted-foreground">{children}</td>,
              hr: () => <hr className="my-8 border-border" />,
            }}
          >
            {content}
          </ReactMarkdown>
        </div>

        {/* Catalog provider grid — rendered only on the catalog index page */}
        {catalogProviders && (
          <CatalogProviderGrid providers={catalogProviders} />
        )}

        {/* Next Article Section */}
        <NextArticle nextArticle={nextArticle} />
      </article>
    </div>
  );
};

