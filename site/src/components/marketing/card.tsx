/**
 * Cards: the bordered surface marketing pages group facts into. A card is
 * separated from the canvas by its border, never by a contrasting fill, and
 * lifts one step on hover. FeatureCard and MetricCard are the two fixed
 * layouts pages reuse.
 */
import { Box, Typography } from '@mui/material';
import type { FC, ReactNode } from 'react';
import { BodyText, FeatureTitle } from './typography';

interface CardProps {
  children?: ReactNode;
  className?: string;
  hover?: boolean;
  gradient?: boolean;
}

export const Card: FC<CardProps> = ({ 
  children, 
  className = '', 
  hover = true,
}) => (
  <Box
    className={`
      rounded-xl
      bg-card border border-edge
      ${hover ? 'hover:border-edge-hover hover:bg-raised transition-all duration-300' : ''}
      p-5 md:p-6
      ${className}
    `}
  >
    {children}
  </Box>
);

export const FeatureCard: FC<Omit<CardProps, 'children'> & { icon?: ReactNode; title: string; description: string }> = ({
  icon,
  title,
  description,
  className = '',
}) => (
  <Card className={`${className}`}>
    {icon && (
      <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center mb-3 text-white">
        {icon}
      </Box>
    )}
    <FeatureTitle className="mb-2">{title}</FeatureTitle>
    <BodyText>{description}</BodyText>
  </Card>
);
interface MetricCardProps {
  value: string;
  label: string;
  sublabel?: string;
  className?: string;
}

export const MetricCard: FC<MetricCardProps> = ({
  value,
  label,
  sublabel,
  className = '',
}) => (
  <Card className={`text-center p-4 ${className}`}>
    <Typography className="text-xl md:text-2xl font-bold text-white">
      {value}
    </Typography>
    <Typography className="text-xs text-fg-secondary mt-1">
      {label}
    </Typography>
    {sublabel && (
      <Typography className="text-xs text-fg-muted mt-1">
        {sublabel}
      </Typography>
    )}
  </Card>
);
