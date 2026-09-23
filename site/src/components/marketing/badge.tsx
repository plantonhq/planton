/**
 * A pill label. Neutral by default; the semantic variants exist for a fact
 * that is a status (a step number is neutral, "included" is success). The
 * `purple` variant is a legacy name that renders neutral and is kept only
 * until its last caller is rebuilt.
 */
import { Box } from '@mui/material';
import type { FC, ReactNode } from 'react';

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'purple';
  className?: string;
}

export const Badge: FC<BadgeProps> = ({ 
  children, 
  variant = 'default',
  className = '' 
}) => {
  const variantClasses = {
    default: 'bg-edge text-fg-secondary border-edge-hover',
    success: 'bg-ok/10 text-ok border-ok/30',
    warning: 'bg-warn/10 text-warn border-warn/30',
    purple: 'bg-edge text-fg-secondary border-edge-hover',
  };
  
  return (
    <Box
      component="span"
      className={`
        inline-flex items-center
        px-3 py-1.5 rounded-full
        text-xs md:text-sm font-medium
        border
        ${variantClasses[variant]}
        ${className}
      `}
    >
      {children}
    </Box>
  );
};
