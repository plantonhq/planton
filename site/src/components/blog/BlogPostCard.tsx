import React from 'react';
import Link from 'next/link';
import { BlogPost } from '@/lib/mdx';
import { formatShortDate } from '@/lib/utils';

interface BlogPostCardProps {
  post: BlogPost;
}

const BlogPostCard: React.FC<BlogPostCardProps> = ({ post }) => {
  return (
    <article className="bg-[#eeeeeb] border border-[#c7c7c4] rounded-lg overflow-hidden hover:border-[#999996] transition-all duration-300">
      {post.featuredImage && (
        <div className="aspect-video overflow-hidden">
          {/* The post's own image, dimensions unknown at build time; images are unoptimized on this static export, so next/image would emit the same tag. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={post.featuredImage}
            alt={post.title}
            className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
          />
        </div>
      )}
      
      <div className="p-6">
        <div className="flex items-center gap-2 mb-3">
          {post.tags.map((tag, index) => (
            <span
              key={index}
              className="px-2 py-1 bg-[#e3e3e0] text-[#454545] text-xs font-medium rounded-full border border-[#999996]"
            >
              {tag}
            </span>
          ))}
        </div>
        
        <h2 className="text-xl font-bold text-[#171717] mb-2 line-clamp-2">
          {post.title}
        </h2>
        
        {post.excerpt && (
          <p className="text-[#454545] mb-4 line-clamp-3">
            {post.excerpt}
          </p>
        )}
        
        <div className="flex items-center justify-between text-sm text-[#454545] mb-4">
          <div className="flex items-center gap-2">
            {post.author.map((author, index) => (
              <span key={index} className="font-medium text-[#171717]">
                {author.name}
              </span>
            ))}
          </div>
          <time dateTime={post.date}>
            {formatShortDate(post.date)}
          </time>
        </div>
        
        <Link
          href={`/blog/${post.slug}`}
          className="inline-flex items-center text-[#171717] hover:text-[#454545] font-medium transition-colors"
        >
          Read more
          <svg
            className="ml-2 w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </Link>
      </div>
    </article>
  );
};

export default BlogPostCard; 