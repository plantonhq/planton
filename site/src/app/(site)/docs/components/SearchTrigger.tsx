'use client';

import React, { useSyncExternalStore } from 'react';
import { Search as SearchIcon } from '@mui/icons-material';

const subscribeNoop = () => () => {};
const getIsMac = () => navigator.userAgent.includes('Mac');
const getIsMacServer = () => false;

interface SearchTriggerProps {
  onClick: () => void;
}

/**
 * A button styled to look like a search input field.
 * Displays a search icon, placeholder text, and a keyboard shortcut badge.
 * Clicking it (or pressing Cmd+K) opens the search modal.
 */
export const SearchTrigger: React.FC<SearchTriggerProps> = ({ onClick }) => {
  const isMac = useSyncExternalStore(subscribeNoop, getIsMac, getIsMacServer);

  return (
    <button
      type="button"
      onClick={onClick}
      className="
        flex items-center gap-2 w-64
        px-3 py-1.5
        rounded-md
        border border-[#c7c7c4]
        bg-transparent
        text-[#616161] text-sm
        cursor-pointer
        transition-colors duration-200
        hover:border-[#999996] hover:text-[#454545]
        focus-visible:outline-none focus-visible:border-white
      "
    >
      <SearchIcon sx={{ fontSize: 20, color: '#666666', flexShrink: 0 }} />
      <span className="flex-1 text-left">Search documentation...</span>
      <kbd
        className="
          hidden sm:inline-flex
          items-center
          px-1.5 py-0.5
          text-[11px] font-mono leading-none
          text-[#616161]
          border border-[#c7c7c4] rounded
        "
      >
        {isMac ? '\u2318K' : 'Ctrl K'}
      </kbd>
    </button>
  );
};
