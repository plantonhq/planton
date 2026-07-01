'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { IconButton, Drawer, Stack } from '@mui/material';
import { Menu as MenuIcon, Close as CloseIcon } from '@mui/icons-material';
import { SiteHeader } from '@/components/chrome';
import { PlantonMark, Wordmark } from '@/components/brand';
import { DocsSidebar } from '@/app/docs/components/DocsSidebar';
import { SearchBar } from '@/app/docs/components/SearchBar';

/**
 * Docs route layout: the shared SiteHeader (with a search slot + mobile menu),
 * a sticky left sidebar, and the page content. Monochrome, token-driven.
 */
export default function DocsPageLayout({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const toggle = () => setSidebarOpen((v) => !v);

  return (
    <div className="min-h-screen bg-background">
      <SiteHeader
        leading={
          <IconButton onClick={toggle} size="small" className="md:hidden" sx={{ color: 'text.primary' }}>
            <MenuIcon />
          </IconButton>
        }
        slot={<div className="hidden md:block"><SearchBar /></div>}
      />

      <div className="flex pt-16">
        {/* Left sidebar — sticky, independently scrollable */}
        <div className="hidden md:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
          <div className="h-full overflow-y-auto border-r border-border">
            <DocsSidebar />
          </div>
        </div>

        {/* Mobile sidebar */}
        <Drawer
          anchor="left"
          open={sidebarOpen}
          onClose={toggle}
          className="md:hidden"
          PaperProps={{ className: 'w-80 bg-background' }}
        >
          <Stack direction="row" className="items-center justify-between p-4 border-b border-border">
            <Link href="/" className="flex items-center gap-2.5 text-foreground">
              <PlantonMark size={22} />
              <Wordmark />
            </Link>
            <IconButton onClick={toggle} sx={{ color: 'text.primary' }}>
              <CloseIcon />
            </IconButton>
          </Stack>
          <DocsSidebar onNavigate={() => setSidebarOpen(false)} />
        </Drawer>

        {children}
      </div>
    </div>
  );
}
