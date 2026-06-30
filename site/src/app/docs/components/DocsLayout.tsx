'use client';

import React, { useState } from 'react';
import RightSidebar from '@/app/docs/components/RightSidebar';
import { Author } from '@/lib/mdx';
import { IconButton, Drawer, Stack } from '@mui/material';
import { DocsSidebar } from '@/app/docs/components/DocsSidebar';
import { DocsHeader } from '@/app/docs/components/DocsHeader';
import { Close as CloseIcon } from '@mui/icons-material';
import Image from 'next/image';
import Link from 'next/link';

interface DocsLayoutProps {
  children: React.ReactNode;
  author?: Author[];
  content?: string;
}

export const DocsLayout: React.FC<DocsLayoutProps> = ({ children, author = [], content }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const handleSidebarToggle = () => {
    setSidebarOpen(!sidebarOpen);
  };

  return (
    <div className="min-h-screen font-sans antialiased bg-slate-950">
      {/* Header - Dedicated component matching landing page */}
      <DocsHeader onMenuToggle={handleSidebarToggle} />

      <div className="flex pt-16">
        {/* Left Sidebar - Sticky, independently scrollable */}
        <div className="hidden md:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
          <div className="h-full overflow-y-auto bg-slate-950 border-r border-purple-900/30">
            <DocsSidebar />
          </div>
        </div>

        {/* Mobile Sidebar */}
        <Drawer
          anchor="left"
          open={sidebarOpen}
          onClose={handleSidebarToggle}
          className="md:hidden"
          PaperProps={{
            className: 'w-80 bg-slate-950',
          }}
        >
          <Stack
            direction="row"
            className="items-center justify-between p-4 border-b border-purple-900/30"
          >
            <Link href="/" className="flex items-center gap-2">
              <Image 
                src="/icon.png" 
                alt="Planton logo" 
                width={32} 
                height={32} 
                className="h-8 w-auto object-contain" 
              />
              <Image src="/text-logo.svg" alt="Planton" width={136} height={28} className="h-7 w-auto object-contain" />
            </Link>
            <IconButton onClick={handleSidebarToggle} className="text-white">
              <CloseIcon />
            </IconButton>
          </Stack>
          <DocsSidebar onNavigate={() => setSidebarOpen(false)} />
        </Drawer>

        {/* Main Content Area - Auto-expands when no right sidebar, shrinks when sidebar present */}
        <div className="flex-1 min-h-screen overflow-x-hidden">
          <div className={`px-4 sm:px-6 lg:px-12 py-8 max-w-full ${author.length > 0 ? 'max-w-4xl mx-auto' : ''}`}>
            {children}
          </div>
        </div>

        {/* Right Sidebar - Always render for table of contents */}
        <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
          <div className="h-full overflow-y-auto bg-slate-950 border-l border-purple-900/30">
            <RightSidebar author={author} content={content} />
          </div>
        </div>
      </div>
    </div>
  );
};

