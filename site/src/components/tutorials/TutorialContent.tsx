'use client';

import React from 'react';
import { MDXRenderer } from '@/lib/MDXRenderer';
import { MDXParserClient } from '@/lib/mdx-client';
import type { Tutorial } from '@/lib/tutorials';

interface NextArticle {
  title: string;
  excerpt?: string;
  slug: string;
}

interface TutorialContentProps {
  slug: string;
  tutorialContent: string;
  allTutorials: Tutorial[];
  nextArticle?: NextArticle;
}

export function TutorialContent({ slug, tutorialContent, nextArticle }: TutorialContentProps) {
  const mdxContent = MDXParserClient.reconstructMDX(tutorialContent);

  return (
    <div className="p-8">
      <MDXRenderer 
        mdxContent={mdxContent} 
        markdownContent={tutorialContent}
        nextArticle={nextArticle}
        path={slug}
      />
    </div>
  );
}
