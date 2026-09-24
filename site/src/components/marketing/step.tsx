/**
 * A numbered step in a how-it-works row, with the connector line between
 * steps on desktop. With an icon, the tile shows the icon and a "Step N"
 * badge; without one, the tile shows the number itself and the badge is not
 * repeated.
 */
import { Box, Typography } from '@mui/material';
import type { FC, ReactNode } from 'react';
import { Badge } from './badge';

interface StepProps {
  number: number;
  title: string;
  description: string;
  icon?: ReactNode;
  isLast?: boolean;
}

export const Step: FC<StepProps> = ({
  number,
  title,
  description,
  icon,
  isLast = false,
}) => (
  <Box className="flex-1 relative">
    {!isLast && (
      <Box className="hidden lg:block absolute top-12 left-[calc(50%+40px)] w-[calc(100%-80px)] h-0.5 bg-white/10" />
    )}
    
    <Box className="flex flex-col items-center text-center">
      <Box className="w-14 h-14 rounded-xl bg-white/10 border border-white/10 flex items-center justify-center mb-3 text-white">
        {icon ?? <span className="text-lg font-semibold">{number}</span>}
      </Box>

      {icon ? <Badge className="mb-2">Step {number}</Badge> : null}
      
      <Typography className="text-base font-semibold text-white mb-1.5">
        {title}
      </Typography>
      
      <Typography className="text-sm text-fg-secondary max-w-xs leading-relaxed">
        {description}
      </Typography>
    </Box>
  </Box>
);
