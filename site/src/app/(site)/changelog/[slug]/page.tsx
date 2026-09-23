import React from 'react';
import Link from 'next/link';
import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import matter from 'gray-matter';
import { ArrowLeft } from 'lucide-react';
import {
  getAllChangelogEntries,
  getChangelogContentBySlug,
} from '@/lib/changelog';
import { cleanSlug, formatDate } from '@/lib/utils';
import { ChangelogCategoryBadge, ChangelogMarkdownBody } from '@/components/changelog';
import { PageActions } from '@/components/common/PageActions';
import type { ChangelogCategory } from '@/lib/types-client';

// ---------------------------------------------------------------------------
// Static generation
// ---------------------------------------------------------------------------

export async function generateStaticParams() {
  const entries = getAllChangelogEntries();
  return entries.map((entry) => ({ slug: entry.slug }));
}

// ---------------------------------------------------------------------------
// OpenGraph metadata
// ---------------------------------------------------------------------------

interface PageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const raw = getChangelogContentBySlug(cleanSlug(slug));

  if (!raw) {
    return { title: 'Entry Not Found' };
  }

  const { data } = matter(raw);

  return {
    title: `${data.title} | Planton Changelog`,
    description: data.excerpt || `${data.title} — Planton changelog entry`,
    openGraph: {
      title: data.title,
      description: data.excerpt,
      type: 'article',
      publishedTime: data.date,
    },
  };
}

// ---------------------------------------------------------------------------
// Page component
// ---------------------------------------------------------------------------

export default async function ChangelogEntryPage({ params }: PageProps) {
  const { slug } = await params;
  const cleanSlugValue = cleanSlug(slug);
  const raw = getChangelogContentBySlug(cleanSlugValue);

  if (!raw) {
    notFound();
  }

  const { data, content } = matter(raw);
  const category = data.category as ChangelogCategory;
  const tags: string[] = data.tags || [];
  const authors: { name: string; title?: string }[] = data.author || [];

  return (
    <div className="min-h-screen font-inter antialiased">
      {/* Back navigation */}
      <nav className="max-w-3xl mx-auto px-4 pt-6">
        <Link
          href="/changelog"
          className="inline-flex items-center gap-1.5 text-sm text-[#a0a0a0] hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Changelog
        </Link>
      </nav>

      <article className="max-w-3xl mx-auto px-4 pt-8 pb-16">
        {/* Header */}
        <header className="mb-8 pb-8 border-b border-[#2a2a2a]">
          {/* Category badge */}
          {category && (
            <div className="mb-4">
              <ChangelogCategoryBadge category={category} />
            </div>
          )}

          {/* Date + author */}
          <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-[#a0a0a0] mb-4">
            {data.date && (
              <time dateTime={data.date}>{formatDate(data.date)}</time>
            )}
            {authors.length > 0 && (
              <>
                <span className="hidden sm:inline">·</span>
                <div className="flex gap-2">
                  {authors.map((a, i) => (
                    <span key={i} className="font-medium text-white">
                      {a.name}
                    </span>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Title + page actions */}
          <div className="flex items-start gap-2">
            <h1 className="flex-1 text-3xl md:text-4xl font-bold text-white mb-4">
              {data.title}
            </h1>
            <PageActions
              markdownContent={raw}
              title={data.title}
              path={`/changelog/${cleanSlugValue}`}
              rawPath={`/changelog/${cleanSlugValue}.md`}
            />
          </div>

          {/* Tags */}
          {tags.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => (
                <span
                  key={tag}
                  className="px-2 py-0.5 bg-white/5 text-white/40 border border-white/10 rounded text-xs"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </header>

        {/* Body */}
        <ChangelogMarkdownBody content={content} />
      </article>
    </div>
  );
}
