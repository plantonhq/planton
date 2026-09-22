/**
 * The responsive grid every card row uses: one column on phones, two on
 * tablets, up to four on desktop. Pages pick a column count, never a
 * breakpoint.
 */
import { Box } from '@mui/material';
import type { FC, ReactNode } from 'react';

interface GridProps {
  children: ReactNode;
  cols?: 1 | 2 | 3 | 4;
  gap?: 'sm' | 'md' | 'lg';
  className?: string;
}

/**
 * Five cards on a three-wide page leave a hole; this lays cards out as a
 * centered wrap instead, so the last row centers itself. Use it when the
 * count is not a multiple of the column count.
 */
export const CenteredCards: FC<{ children: ReactNode; className?: string }> = ({ children, className = '' }) => (
  <Box className={`flex flex-wrap justify-center gap-6 ${className} [&>*]:basis-full md:[&>*]:basis-[calc(50%-0.75rem)] lg:[&>*]:basis-[calc(33.333%-1rem)]`}>{children}</Box>
);

export const Grid: FC<GridProps> = ({ 
  children, 
  cols = 3, 
  gap = 'md',
  className = '' 
}) => {
  const colsClass = {
    1: 'grid-cols-1',
    2: 'grid-cols-1 md:grid-cols-2',
    3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  };
  
  const gapClass = {
    sm: 'gap-4',
    md: 'gap-6',
    lg: 'gap-8',
  };
  
  return (
    <Box className={`grid ${colsClass[cols]} ${gapClass[gap]} ${className}`}>
      {children}
    </Box>
  );
};
