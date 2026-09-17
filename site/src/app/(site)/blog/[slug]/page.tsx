import React from 'react';
import { notFound } from 'next/navigation';
import { getBlogPostContentBySlug, getAllBlogPosts, getNextBlogPost, Author } from '@/lib/mdx';
import { cleanSlug } from '@/lib/utils';
import { MdxContentLayout } from '@/components/common';
import { BlogPostContent } from '@/components/blog/BlogPostContent';
import matter from 'gray-matter';

interface BlogPostPageProps {
  params: Promise<{ slug: string }>;
}

export async function generateStaticParams() {
  const posts = getAllBlogPosts();
  const params = posts.map((post) => ({
    slug: post.slug,
  }));

  return params;
}

export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params;

  // Strip .md extensions from slug to handle both clean routes and .md routes
  const cleanSlugValue = cleanSlug(slug);

  const content = getBlogPostContentBySlug(cleanSlugValue);
  const { data } = matter(content);

  if (!content) {
    notFound();
  }

  const allPosts = getAllBlogPosts();

  // Get next post data on the server side
  const nextPost = getNextBlogPost(cleanSlugValue, allPosts, 'date-desc'); // Default sort for static generation

  return (
    <MdxContentLayout
      author={data?.author as unknown as Author[]}
      content={content}
      sectionTitle="Blog"
      basePath="/blog"
    >
      <BlogPostContent
        slug={cleanSlugValue}
        post={content}
        allPosts={allPosts}
        nextArticle={
          nextPost
            ? {
                title: nextPost.title,
                excerpt: nextPost.excerpt,
                slug: `/blog/${nextPost.slug}`,
              }
            : undefined
        }
      />
    </MdxContentLayout>
  );
}
