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
        border border-purple-500/30
        bg-transparent
        text-white/50 text-sm
        cursor-pointer
        transition-colors duration-200
        hover:border-purple-500/50
        focus-visible:outline-none focus-visible:border-purple-500/80
      "
    >
      <SearchIcon sx={{ fontSize: 20, color: 'rgba(156, 163, 175, 1)' }} />
      <span className="flex-1 text-left">Search documentation...</span>
      <kbd
        className="
          hidden sm:inline-flex
          items-center
          px-1.5 py-0.5
          text-[11px] font-mono leading-none
          text-white/40
          border border-purple-500/30 rounded
        "
      >
        {isMac ? '⌘K' : 'Ctrl K'}
      </kbd>
    </button>
  );
};
