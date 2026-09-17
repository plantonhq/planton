/**
 * The marketing type scale: a section title, its subtitle, a feature title,
 * and body copy. Four sizes do the work on every page; anything that needs a
 * fifth is a design question, not a new component.
 */
import { Typography, type TypographyProps } from '@mui/material';
import type { FC } from 'react';

export const SectionTitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    variant="h2"
    className={`text-xl md:text-2xl lg:text-3xl font-semibold text-white leading-snug tracking-tight ${className}`}
    {...props}
  />
);

export const SectionSubtitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    className={`text-sm md:text-base text-fg-secondary font-normal mt-3 max-w-2xl ${className}`}
    {...props}
  />
);

export const FeatureTitle: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    variant="h3"
    className={`text-base md:text-lg font-semibold text-white ${className}`}
    {...props}
  />
);

export const BodyText: FC<TypographyProps> = ({ className, ...props }) => (
  <Typography
    className={`text-sm text-fg-body leading-relaxed ${className}`}
    {...props}
  />
);
