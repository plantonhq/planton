import React from 'react';
import type { ChangelogCategory } from '@/lib/types-client';

interface ChangelogCategoryBadgeProps {
  category: ChangelogCategory;
  className?: string;
}

const CATEGORY_CONFIG: Record<
  ChangelogCategory,
  { label: string; bg: string; text: string; border: string }
> = {
  feature: {
    label: 'Feature',
    bg: 'bg-[#e3e3e0]',
    text: 'text-[#171717]',
    border: 'border-[#c7c7c4]',
  },
  improvement: {
    label: 'Improvement',
    bg: 'bg-[#eeeeeb]',
    text: 'text-[#454545]',
    border: 'border-[#c7c7c4]',
  },
  fix: {
    label: 'Fix',
    bg: 'bg-[#eeeeeb]',
    text: 'text-[#454545]',
    border: 'border-[#c7c7c4]',
  },
  breaking: {
    label: 'Breaking',
    bg: 'bg-[#eeeeeb]',
    text: 'text-[#595959]',
    border: 'border-[#c7c7c4]',
  },
};

const ChangelogCategoryBadge: React.FC<ChangelogCategoryBadgeProps> = ({
  category,
  className = '',
}) => {
  const config = CATEGORY_CONFIG[category];

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 text-xs font-medium rounded border ${config.bg} ${config.text} ${config.border} ${className}`}
    >
      {config.label}
    </span>
  );
};

export default ChangelogCategoryBadge;
