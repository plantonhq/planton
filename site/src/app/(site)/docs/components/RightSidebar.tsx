'use client';

import React from 'react';
import { Author } from '@/lib/mdx';
import AuthorSection from '@/components/blog/AuthorSection';
import TableOfContents from '@/components/blog/TableOfContents';

interface RightSidebarProps {
  author: Author[];
  content?: string;
}

const RightSidebar: React.FC<RightSidebarProps> = ({ author = [], content }) => {
  return (
    <div className="w-80 h-full flex flex-col">
      {/* Author Section - Only show if authors are present */}
      {author.length > 0 && (
        <>
          <AuthorSection author={author} />
          <div className="border-b border-[#2a2a2a]" />
        </>
      )}
      {/* Table of Contents */}
      {content && <TableOfContents content={content} />}
    </div>
  );
};

export default RightSidebar;