'use client';

import React, { useEffect, useState } from 'react';
import { Search as SearchIcon } from '@mui/icons-material';

interface SearchTriggerProps {
  onClick: () => void;
}

/**
 * A button styled to look like a search input field.
 * Displays a search icon, placeholder text, and a keyboard shortcut badge.
 * Clicking it (or pressing ⌘K) opens the search modal.
 */
export const SearchTrigger: React.FC<SearchTriggerProps> = ({ onClick }) => {
  const [isMac, setIsMac] = useState(false);

  // Detect platform after mount to avoid SSR hydration mismatch
  useEffect(() => {
    setIsMac(navigator.userAgent.includes('Mac'));
  }, []);

  return (
    <button
      type="button"
      onClick={onClick}
      className="
        flex items-center gap-2 w-64
        px-3 py-1.5
        rounded-md
        border border-border
        bg-secondary/50
        text-muted-foreground text-sm
        cursor-pointer
        transition-colors duration-200
        hover:border-ring
        focus-visible:outline-none focus-visible:border-ring
      "
    >
      <SearchIcon sx={{ fontSize: 20, color: 'currentColor' }} />
      <span className="flex-1 text-left">Search documentation...</span>
      <kbd
        className="
          hidden sm:inline-flex
          items-center
          px-1.5 py-0.5
          text-[11px] font-mono leading-none
          text-muted-foreground
          border border-border rounded
        "
      >
        {isMac ? '⌘K' : 'Ctrl K'}
      </kbd>
    </button>
  );
};
