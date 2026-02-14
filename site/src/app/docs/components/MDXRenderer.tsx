'use client';

import React from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import rehypeHighlight from 'rehype-highlight';
import matter from 'gray-matter';
import { formatDate } from '@/lib/utils';
import { Author } from '@/lib/mdx';
import { CodeBlock } from '@/app/docs/components/CodeBlock';
import 'highlight.js/styles/github-dark.css';

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
  nextArticle?: {
    title: string;
    excerpt?: string;
    slug: string;
  };
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
    <div className="mt-12 p-6 rounded-lg bg-purple-900/20 border border-purple-900/30">
      <div className="max-w-none">
        <p className="text-lg text-gray-400 m-0 font-bold">Next article</p>
        <h3 className="text-xl font-bold text-white m-0 my-2">{nextArticle.title}</h3>
        {nextArticle.excerpt && (
          <div className="relative mb-4 min-h-24">
            <div className="text-gray-300 leading-6">{nextArticle.excerpt}</div>
          </div>
        )}
        <Link
          href={nextArticle.slug}
          className="inline-flex items-center px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white font-semibold rounded-md transition-colors duration-200 hover:translate-y-[-1px] active:translate-y-[1px]"
        >
          Read next article
        </Link>
      </div>
    </div>
  );
};

export const MDXRenderer: React.FC<MDXRendererProps> = ({
  mdxContent,
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
            <div className="flex items-center gap-4 text-gray-300 mb-6">
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
                  className="px-3 py-1 bg-purple-900/30 text-purple-200 text-sm font-medium rounded-full border border-purple-700/30"
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
            rehypePlugins={[rehypeRaw, rehypeHighlight]}
            components={{
              p: ({ children }) => (
                <p className="text-gray-300 mb-4 leading-relaxed">{children}</p>
              ),
              h1: ({ children }) => {
                const id = children
                  ?.toString()
                  .toLowerCase()
                  .replace(/[^a-z0-9\s-]/g, '')
                  .replace(/\s+/g, '-');

                return (
                  <h1 id={id} className="text-3xl font-bold text-white mt-8 mb-4">
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
                  <h2 id={id} className="text-2xl font-bold text-white mt-6 mb-3">
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
                  <h3 id={id} className="text-xl font-bold text-white mt-5 mb-2">
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
                  <h4 id={id} className="text-lg font-bold text-white mt-4 mb-2">
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
                  <h5 id={id} className="text-base font-bold text-white mt-3 mb-2">
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
                  <h6 id={id} className="text-sm font-bold text-white mt-2 mb-1">
                    {children}
                  </h6>
                );
              },
              ul: ({ children }) => (
                <ul className="list-disc list-inside text-gray-300 mb-4 space-y-2">{children}</ul>
              ),
              ol: ({ children }) => (
                <ol className="list-decimal list-inside text-gray-300 mb-4 space-y-2">
                  {children}
                </ol>
              ),
              li: ({ children }) => <li className="text-gray-300">{children}</li>,
              blockquote: ({ children }) => (
                <blockquote className="border-l-4 border-purple-500 pl-4 py-2 my-4 bg-purple-900/20 rounded-r text-gray-300 italic">
                  {children}
                </blockquote>
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
              pre: ({ children }) => (
                <CodeBlock>{children}</CodeBlock>
              ),
              a: ({ href, children }) => {
                const isExternal = href?.startsWith('http');
                if (!isExternal && href) {
                  return (
                    <Link
                      href={href}
                      className="text-purple-400 hover:text-purple-300 underline"
                    >
                      {children}
                    </Link>
                  );
                }
                return (
                  <a
                    href={href}
                    className="text-purple-400 hover:text-purple-300 underline"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {children}
                  </a>
                );
              },
              img: ({ src, alt, className }) => {
                if (!src) return null;
                // Preserve custom classes if provided (for provider icons, etc.)
                const finalClassName = className || "max-w-full h-auto rounded-lg shadow-lg my-6 block";
                return (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={src}
                    alt={alt || ''}
                    className={finalClassName}
                  />
                );
              },
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
            {content}
          </ReactMarkdown>
        </div>

        {/* Next Article Section */}
        <NextArticle nextArticle={nextArticle} />
      </article>
    </div>
  );
};

