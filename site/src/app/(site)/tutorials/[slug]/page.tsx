import React from 'react';
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
