/**
 * A hairline that fades at both ends. The one gradient the design system
 * allows, because it is a border, not a fill.
 */
import { Box } from '@mui/material';
import type { FC } from 'react';

export const Divider: FC<{ className?: string }> = ({ className = '' }) => (
  <Box className={`w-full h-px bg-gradient-to-r from-transparent via-edge to-transparent ${className}`} />
);
