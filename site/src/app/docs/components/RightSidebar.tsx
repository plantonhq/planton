'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import { Author } from '@/lib/mdx';

interface PresetsLinkInfo {
  /** URL path to the presets list page, e.g. "/docs/catalog/aws/documentdb/presets" */
  path: string;
  /** Number of presets available */
  count: number;
}

interface RightSidebarProps {
  author?: Author[];
  content?: string;
  /** When set, renders a pinned "Presets" link section below the TOC. */
  presetsLink?: PresetsLinkInfo;
}

interface Heading {
  id: string;
  text: string;
  level: number;
}

const RightSidebar: React.FC<RightSidebarProps> = ({ author = [], content, presetsLink }) => {
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [activeId, setActiveId] = useState<string>('');

  useEffect(() => {
    if (!content) return;

    // Extract headings from markdown content
    const lines = content.split('\n');
    const extractedHeadings: Heading[] = [];

    lines.forEach((line) => {
      const match = line.match(/^(#{1,6})\s+(.+)$/);
      if (match) {
        const level = match[1].length;
        const text = match[2];
        const id = text
          .toLowerCase()
          .replace(/[^a-z0-9\s-]/g, '')
          .replace(/\s+/g, '-');

        // Only show h2 and h3 in TOC
        if (level === 2 || level === 3) {
          extractedHeadings.push({ id, text, level });
        }
      }
    });

    setHeadings(extractedHeadings);
  }, [content]);

  useEffect(() => {
    // Track active heading based on scroll position
    const handleScroll = () => {
      const headingElements = headings.map((h) => ({
        id: h.id,
        element: document.getElementById(h.id),
      }));

      let currentActiveId = '';
      
      for (const { id, element } of headingElements) {
        if (element) {
          const rect = element.getBoundingClientRect();
          if (rect.top <= 100) {
            currentActiveId = id;
          }
        }
      }

      setActiveId(currentActiveId);
    };

    window.addEventListener('scroll', handleScroll);
    handleScroll(); // Initial check

    return () => window.removeEventListener('scroll', handleScroll);
  }, [headings]);

  const scrollToHeading = (id: string) => {
    const element = document.getElementById(id);
    if (element) {
      const yOffset = -80; // Offset for fixed header
      const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  };

  return (
    <Box className="p-6">
      {/* Table of Contents */}
      {headings.length > 0 && (
        <Box className="mb-6">
          <Typography variant="subtitle2" className="text-muted-foreground font-semibold mb-3 uppercase text-xs">
            On This Page
          </Typography>
          <nav>
            <ul className="space-y-2">
              {headings.map((heading) => (
                <li
                  key={heading.id}
                  className={`${heading.level === 3 ? 'ml-4' : ''}`}
                >
                  <button
                    onClick={() => scrollToHeading(heading.id)}
                    className={`text-sm text-left w-full hover:text-foreground transition-colors ${
                      activeId === heading.id
                        ? 'text-foreground font-medium'
                        : 'text-muted-foreground'
                    }`}
                  >
                    {heading.text}
                  </button>
                </li>
              ))}
            </ul>
          </nav>
        </Box>
      )}

      {/* Presets link */}
      {presetsLink && (
        <Box className="border-t border-border pt-5 mb-5">
          <Typography variant="subtitle2" className="text-muted-foreground font-semibold mb-3 uppercase text-xs">
            Presets
          </Typography>
          <Link
            href={presetsLink.path}
            className="group flex items-start gap-3 p-3 -mx-1 rounded-lg hover:bg-secondary transition-colors"
          >
            <span className="flex-shrink-0 mt-0.5 text-foreground">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="2" y="6" width="20" height="12" rx="2" />
                <path d="M12 12h.01" />
                <path d="M17 12h.01" />
                <path d="M7 12h.01" />
              </svg>
            </span>
            <span className="flex-1 min-w-0">
              <span className="block text-sm font-medium text-foreground transition-colors">
                {presetsLink.count} ready-to-deploy {presetsLink.count === 1 ? 'configuration' : 'configurations'}
              </span>
              <span className="block text-xs text-muted-foreground mt-0.5">
                View presets &rarr;
              </span>
            </span>
          </Link>
        </Box>
      )}

      {/* Author Information */}
      {author && author.length > 0 && (
        <Box className="border-t border-border pt-6">
          <Typography variant="subtitle2" className="text-muted-foreground font-semibold mb-3 uppercase text-xs">
            Author{author.length > 1 ? 's' : ''}
          </Typography>
          <div className="space-y-3">
            {author.map((a, index) => (
              <div key={index} className="flex items-start gap-3">
                {a.image && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={a.image}
                    alt={a.name}
                    className="w-10 h-10 rounded-full"
                  />
                )}
                <div className="flex-1">
                  <Typography className="text-foreground text-sm font-medium">
                    {a.name}
                  </Typography>
                  {a.role && (
                    <Typography className="text-muted-foreground text-xs">
                      {a.role}
                    </Typography>
                  )}
                </div>
              </div>
            ))}
          </div>
        </Box>
      )}
    </Box>
  );
};

export default RightSidebar;

