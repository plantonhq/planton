'use client';

import React from 'react';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { MdxRightBar } from '@/components/common';
import { Author } from '@/lib/mdx';

interface MdxContentLayoutProps {
  children: React.ReactNode;
  sectionTitle?: string;
  basePath: string;
  author: Author[];
  content: string;
}

const MdxContentLayout: React.FC<MdxContentLayoutProps> = ({
  children,
  sectionTitle = 'Content',
  basePath,
  author,
  content,
}) => {
  return (
    <div className="flex min-h-screen">
      {/* Left Spacer */}
      <div className="w-64 flex-shrink-0" />

      {/* Main Content */}
      <div className="flex-1 bg-[#f6f6f3]">
        <nav className="px-8 pt-6">
          <Link
            href={basePath}
            className="inline-flex items-center gap-1.5 text-sm text-[#454545] hover:text-[#171717] transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            {sectionTitle}
          </Link>
        </nav>
        {children}
      </div>

      {/* Right Sidebar */}
      <div className="sticky top-16 h-screen w-80 flex-shrink-0">
        <div className="h-full border-l border-[#c7c7c4]">
          <MdxRightBar author={author} content={content} />
        </div>
      </div>

      {/* Right Spacer */}
      <div className="w-64 flex-shrink-0" />
    </div>
  );
};

export { MdxContentLayout };
