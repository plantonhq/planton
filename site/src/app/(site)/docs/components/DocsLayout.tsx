'use client';

import React, { useRef, useState } from 'react';
import RightSidebar from '@/app/(site)/docs/components/RightSidebar';
import { Author } from '@/lib/mdx';
import { IconButton, Typography } from '@mui/material';
import { Stack } from '@mui/material';
import { DocsSidebar } from '@/app/(site)/docs/components/DocsSidebar';
import { SearchBar } from '@/app/(site)/docs/components/SearchBar';
import { Drawer } from '@mui/material';
import {
  Close as CloseIcon,
  FormatListBulleted as ListIcon,
  Search as SearchIcon,
} from '@mui/icons-material';
import { DocItem } from '@/app/(site)/docs/utils/fileSystem';

/**
 * Height of the main site header (fixed, defined in MainLayout as pt-[70px]).
 * Used to position sticky elements below it.
 */
const SITE_HEADER_HEIGHT = '70px';

/**
 * Combined offset for elements that sit below both the site header and docs header.
 * Site header (70px) + docs header (~57px) = 127px.
 */
const BELOW_BOTH_HEADERS = '127px';

interface DocsLayoutProps {
  children: React.ReactNode;
  author?: Author[];
  content?: string;
  structure: DocItem[];
}

export const DocsLayout: React.FC<DocsLayoutProps> = ({ children, author = [], content, structure }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const searchOpenRef = useRef<(() => void) | null>(null);

  const handleSidebarToggle = () => {
    setSidebarOpen(!sidebarOpen);
  };

  const handleSearchOpen = () => {
    searchOpenRef.current?.();
  };

  return (
    <div className="min-h-screen font-inter antialiased">
      {/* Docs header — desktop only, sticks below the main site header */}
      <div
        className="hidden md:sticky md:block z-10 bg-[#0a0a0a] border-b border-[#2a2a2a]"
        style={{ top: SITE_HEADER_HEIGHT }}
      >
        <Stack direction="row" className="items-center justify-between px-4 py-3">
          <Typography variant="h6" className="text-[#b0b0b0] font-semibold text-lg">
            Planton Documentation
          </Typography>
          <SearchBar onOpenRef={searchOpenRef} />
        </Stack>
      </div>

      <div className="flex">
        {/* Left Sidebar — sticks below both headers */}
        <div
          className="hidden md:block sticky flex-shrink-0 w-80"
          style={{ top: BELOW_BOTH_HEADERS, height: `calc(100vh - ${BELOW_BOTH_HEADERS})` }}
        >
          <div className="h-full overflow-y-auto border-r border-[#2a2a2a]">
            <DocsSidebar structure={structure} />
          </div>
        </div>

        {/* Mobile Sidebar */}
        <Drawer
          anchor="left"
          open={sidebarOpen}
          onClose={handleSidebarToggle}
          className="md:hidden"
          PaperProps={{
            className: 'w-80 bg-[#0a0a0a]',
          }}
        >
          <Stack
            direction="row"
            className="items-center justify-between p-4 border-b border-[#2a2a2a]"
          >
            <Typography variant="h6" className="text-[#b0b0b0] font-semibold">
              Documentation
            </Typography>
            <IconButton onClick={handleSidebarToggle} className="text-white">
              <CloseIcon />
            </IconButton>
          </Stack>
          <DocsSidebar structure={structure} onNavigate={() => setSidebarOpen(false)} />
        </Drawer>

        {/* Main Content Area */}
        <div className="flex-1 min-h-screen min-w-0 overflow-x-hidden">
          <div className={`px-4 sm:px-6 md:px-8 lg:px-12 py-4 md:py-8 ${author.length > 0 ? 'max-w-4xl mx-auto' : 'w-full'}`}>
            {/* Mobile docs navigation bar */}
            <div className="md:hidden flex items-center gap-2 mb-4 -mx-1">
              <button
                onClick={handleSidebarToggle}
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 transition-colors"
              >
                <ListIcon fontSize="small" />
                <span className="text-sm font-medium">Documentation menu</span>
              </button>
              <button
                onClick={handleSearchOpen}
                className="flex items-center gap-1.5 px-2.5 py-2 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 transition-colors ml-auto"
                aria-label="Search documentation"
              >
                <SearchIcon sx={{ fontSize: 18 }} />
                <span className="text-sm font-medium">Search</span>
              </button>
            </div>
            {children}
          </div>
        </div>

        {/* Right Sidebar — Table of Contents */}
        <div
          className="hidden xl:block sticky flex-shrink-0 w-80"
          style={{ top: BELOW_BOTH_HEADERS, height: `calc(100vh - ${BELOW_BOTH_HEADERS})` }}
        >
          <div className="h-full overflow-y-auto border-l border-[#2a2a2a]">
            <RightSidebar author={author} content={content} />
          </div>
        </div>
      </div>
    </div>
  );
};
