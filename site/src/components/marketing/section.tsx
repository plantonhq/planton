/**
 * A full-width marketing section on the canvas, with the content column every
 * page shares. Sections are the unit the landing page composes; a page never
 * sets its own horizontal padding or maximum width.
 */
import { Box } from '@mui/material';
import type { FC, ReactNode } from 'react';

interface SectionProps {
  children: ReactNode;
  className?: string;
  id?: string;
  variant?: 'default' | 'dark' | 'gradient';
}

export const Section: FC<SectionProps> = ({ 
  children, 
  className = '', 
  id,
  variant: _variant = 'default' 
}) => {
  return (
    <Box 
      component="section" 
      id={id}
      className={`w-full py-12 md:py-16 px-4 md:px-8 overflow-x-hidden bg-canvas ${className}`}
    >
      <Box className="max-w-7xl mx-auto overflow-hidden">
        {children}
      </Box>
    </Box>
  );
};
