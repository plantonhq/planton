'use client';

import React, { useState, useCallback, useEffect } from 'react';

interface HeadingWithAnchorProps {
  id: string;
  level: 1 | 2 | 3 | 4 | 5 | 6;
  className: string;
  children: React.ReactNode;
}

// Link icon SVG component
const LinkIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    strokeWidth={2}
    stroke="currentColor"
    className="w-5 h-5"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244"
    />
  </svg>
);

// Checkmark icon SVG component
const CheckIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    strokeWidth={2.5}
    stroke="currentColor"
    className="w-5 h-5 text-[#10b981]"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="M4.5 12.75l6 6 9-13.5"
    />
  </svg>
);

export const HeadingWithAnchor: React.FC<HeadingWithAnchorProps> = ({
  id,
  level,
  className,
  children,
}) => {
  const [copied, setCopied] = useState(false);

  // Reset copied state after 2 seconds
  useEffect(() => {
    if (copied) {
      const timer = setTimeout(() => setCopied(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [copied]);

  const handleAnchorClick = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();

      // Update URL hash
      window.history.pushState(null, '', `#${id}`);

      // Scroll to element smoothly
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }

      // Copy full URL to clipboard
      const fullUrl = `${window.location.origin}${window.location.pathname}#${id}`;
      navigator.clipboard.writeText(fullUrl).then(() => {
        setCopied(true);
      });
    },
    [id]
  );

  const anchorLink = (
    <a
      href={`#${id}`}
      onClick={handleAnchorClick}
      className={`absolute -left-6 top-1/2 -translate-y-1/2 transition-opacity duration-200 p-1 hidden md:inline-block ${
        copied ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
      } ${copied ? '' : 'text-[#616161] hover:text-[#171717]'}`}
      aria-label={`Link to ${typeof children === 'string' ? children : 'this section'}`}
    >
      {copied ? <CheckIcon /> : <LinkIcon />}
    </a>
  );

  // scroll-mt-36 (144px) accounts for main site header (70px) + docs header (~57px) + breathing room
  const headingClassName = `group relative scroll-mt-36 ${className}`;

  switch (level) {
    case 1:
      return <h1 id={id} className={headingClassName}>{anchorLink}{children}</h1>;
    case 2:
      return <h2 id={id} className={headingClassName}>{anchorLink}{children}</h2>;
    case 3:
      return <h3 id={id} className={headingClassName}>{anchorLink}{children}</h3>;
    case 4:
      return <h4 id={id} className={headingClassName}>{anchorLink}{children}</h4>;
    case 5:
      return <h5 id={id} className={headingClassName}>{anchorLink}{children}</h5>;
    case 6:
      return <h6 id={id} className={headingClassName}>{anchorLink}{children}</h6>;
    default:
      return <h2 id={id} className={headingClassName}>{anchorLink}{children}</h2>;
  }
};

// Helper function to generate ID from heading content
export const generateHeadingId = (children: React.ReactNode): string => {
  return (
    children
      ?.toString()
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-') || ''
  );
};

