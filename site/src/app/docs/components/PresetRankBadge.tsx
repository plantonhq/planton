'use client';

import React from 'react';

interface PresetRankBadgeProps {
  rank: string;
}

/**
 * Small badge displaying a preset rank number (e.g., "01", "02").
 * Uses a neutral surface consistent with the monochrome design system.
 */
export const PresetRankBadge: React.FC<PresetRankBadgeProps> = ({ rank }) => {
  return (
    <span className="inline-flex items-center justify-center w-8 h-8 rounded-md bg-secondary border border-border text-foreground text-xs font-bold flex-shrink-0">
      {rank}
    </span>
  );
};
