/**
 * A number with its label, for the proof strip. Numbers come from
 * `@/data/platform-stats`; this component never carries one.
 */
import { Box, Typography } from '@mui/material';
import type { FC } from 'react';

interface MetricProps {
  value: string;
  label: string;
  className?: string;
}

export const Metric: FC<MetricProps> = ({ value, label, className = '' }) => (
  <Box className={`text-center ${className}`}>
    <Typography className="text-2xl md:text-3xl font-bold text-white">
      {value}
    </Typography>
    <Typography className="text-xs md:text-sm text-fg-secondary mt-1">
      {label}
    </Typography>
  </Box>
);
