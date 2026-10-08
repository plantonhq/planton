import React from 'react';
import type { Metadata } from 'next';
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

/**
 * Every post carries its own title, description and share image from its
 * frontmatter. The featured image is the picture a link card shows on
 * LinkedIn, Medium, Slack and X, so a post shared anywhere is recognizable.
 */
export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const raw = getBlogPostContentBySlug(cleanSlug(slug));
  if (!raw) return { title: 'Not Found' };
  const { data } = matter(raw);
  const description = data.excerpt || data.description;
  const images = data.featuredImage ? [data.featuredImage] : undefined;
  return {
    title: `${data.title} | Planton Blog`,
    description,
    alternates: { canonical: `https://planton.ai/blog/${cleanSlug(slug)}` },
    openGraph: { title: data.title, description, type: 'article', publishedTime: data.date, images },
    twitter: { card: images ? 'summary_large_image' : 'summary', title: data.title, description, images },
  };
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
