'use client';

import React from 'react';
import { MDXRenderer } from '@/lib/MDXRenderer';
import { MDXParserClient } from '@/lib/mdx-client';
import type { BlogPost } from '@/lib/mdx';

interface NextArticle {
  title: string;
  excerpt?: string;
  slug: string;
}

interface BlogPostContentProps {
  slug: string;
  post: string;
  allPosts: BlogPost[];
  nextArticle?: NextArticle;
}

export function BlogPostContent({ slug, post, nextArticle }: BlogPostContentProps) {
  const mdxContent = MDXParserClient.reconstructMDX(post);

  return (
    <div className="p-8">
      <MDXRenderer 
        mdxContent={mdxContent} 
        markdownContent={post}
        nextArticle={nextArticle}
        path={slug}
      />
    </div>
  );
}
