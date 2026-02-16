'use client';

import React from 'react';

interface PresetRankBadgeProps {
  rank: string;
}

/**
 * Small badge displaying a preset rank number (e.g., "01", "02").
 * Uses a purple-tinted background consistent with the site design system.
 */
export const PresetRankBadge: React.FC<PresetRankBadgeProps> = ({ rank }) => {
  return (
    <span className="inline-flex items-center justify-center w-8 h-8 rounded-md bg-purple-900/50 border border-purple-700/40 text-purple-300 text-xs font-bold flex-shrink-0">
      {rank}
    </span>
  );
};
