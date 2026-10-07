'use client';

import React from 'react';
import Link from 'next/link';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import { HeadingWithAnchor, generateHeadingId } from '@/components/docs';
import {
  HEADING_H1_CLASSES,
  HEADING_H2_CLASSES,
  HEADING_H3_CLASSES,
  HEADING_H4_CLASSES,
  LINK_CLASSES,
} from '@/theme/docs';

interface LegalContentProps {
  content: string;
}

export function LegalContent({ content }: LegalContentProps) {
  return (
    <div className="min-h-screen py-12 md:py-20">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="prose max-w-none md:prose-lg">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeRaw]}
            components={{
              p: ({ children }) => (
                <p className="text-[#454545] mb-4 leading-relaxed">{children}</p>
              ),
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
              ul: ({ children }) => (
                <ul className="list-disc list-inside text-[#454545] mb-4 space-y-2">{children}</ul>
              ),
              ol: ({ children }) => (
                <ol className="list-decimal list-inside text-[#454545] mb-4 space-y-2">{children}</ol>
              ),
              li: ({ children }) => <li className="text-[#454545]">{children}</li>,
              blockquote: ({ children }) => (
                <blockquote className="border-l-2 border-[#999996] pl-4 py-3 my-5 bg-[#eeeeeb] rounded-r text-[#454545] italic">
                  {children}
                </blockquote>
              ),
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
                  <Link href={href || '#'} className={LINK_CLASSES}>
                    {children}
                  </Link>
                );
              },
              table: ({ children }) => (
                <div className="overflow-x-auto my-6 -mx-4 px-4 sm:mx-0 sm:px-0">
                  <table className="min-w-full bg-[#eeeeeb] border border-[#c7c7c4] rounded-lg">
                    {children}
                  </table>
                </div>
              ),
              thead: ({ children }) => <thead className="bg-[#eeeeeb]">{children}</thead>,
              tbody: ({ children }) => <tbody>{children}</tbody>,
              tr: ({ children }) => <tr className="border-b border-[#c7c7c4]">{children}</tr>,
              th: ({ children }) => (
                <th className="px-4 py-3 text-left text-[#171717] font-semibold text-sm">{children}</th>
              ),
              td: ({ children }) => (
                <td className="px-4 py-3 text-[#454545] text-sm">{children}</td>
              ),
              hr: () => <hr className="my-8 border-[#c7c7c4]" />,
              strong: ({ children }) => <strong className="text-[#171717] font-semibold">{children}</strong>,
              em: ({ children }) => <em className="text-[#454545]">{children}</em>,
            }}
          >
            {content}
          </ReactMarkdown>
        </div>
      </div>
    </div>
  );
}
