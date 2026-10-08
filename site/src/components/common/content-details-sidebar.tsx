'use client';

import React from 'react';
import AuthorSection from '@/components/blog/AuthorSection';
import TableOfContents from '@/components/blog/TableOfContents';
import { Author } from '@/lib/mdx';



interface IMdxRightBar {
  author: Author[];
  content: string;
}

const MdxRightBar: React.FC<IMdxRightBar> = ({ author, content }) => {
  return (
    <div className="w-80 h-full flex flex-col">
      {/* Author Section */}
      <AuthorSection author={author} />
      <div className="border-b border-[#c7c7c4]" />
      {/* Table of Contents */}
      <TableOfContents content={content} />
    </div>
  );
};

export { MdxRightBar };
