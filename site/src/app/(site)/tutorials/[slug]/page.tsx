import React from 'react';
import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { getTutorialContentBySlug, getAllTutorials, getNextTutorial, Author } from '@/lib/mdx';
import { cleanSlug } from '@/lib/utils';
import { MdxContentLayout } from '@/components/common';
import { TutorialContent } from '@/components/tutorials/TutorialContent';
import matter from 'gray-matter';

interface TutorialPageProps {
  params: Promise<{ slug: string }>;
}

export async function generateStaticParams() {
  const tutorials = getAllTutorials();
  const params = tutorials.map((tutorial) => ({
    slug: tutorial.slug,
  }));

  return params;
}

/**
 * Every post carries its own title and description from its frontmatter;
 * before this, all of them shared the site's default title.
 */
export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const raw = getTutorialContentBySlug(cleanSlug(slug));
  if (!raw) return { title: 'Not Found' };
  const { data } = matter(raw);
  const description = data.excerpt || data.description;
  return {
    title: `${data.title} | Planton Tutorials`,
    description,
    alternates: { canonical: `https://planton.ai/tutorials/${cleanSlug(slug)}` },
    openGraph: { title: data.title, description, type: 'article', publishedTime: data.date },
  };
}

export default async function TutorialPage({ params }: TutorialPageProps) {
  const { slug } = await params;

  // Strip .md extensions from slug to handle both clean routes and .md routes
  const cleanSlugValue = cleanSlug(slug);

  const tutorialContent = getTutorialContentBySlug(cleanSlugValue);
  const { data } = matter(tutorialContent);

  if (!tutorialContent) {
    notFound();
  }

  const allTutorials = getAllTutorials();

  // Get next tutorial data on the server side
  const nextTutorial = getNextTutorial(cleanSlugValue, 'date-desc', allTutorials); // Pass allTutorials to avoid duplicate calls

  return (
    <MdxContentLayout
      author={data?.author as unknown as Author[]}
      content={tutorialContent}
      sectionTitle="Tutorials"
      basePath="/tutorials"
    >
      <TutorialContent
        slug={cleanSlugValue}
        tutorialContent={tutorialContent}
        allTutorials={allTutorials}
        nextArticle={
          nextTutorial
            ? {
                title: nextTutorial.title,
                excerpt: nextTutorial.excerpt,
                slug: `/tutorials/${nextTutorial.slug}`,
              }
            : undefined
        }
      />
    </MdxContentLayout>
  );
}
