/**
 * The two marketing buttons. Primary is true white on black (the one place
 * the site uses pure white, so the call to action is the brightest thing on
 * the screen); secondary is an outline. There is no third button.
 */
import { Button, type ButtonProps } from '@mui/material';
import type { ComponentProps, FC } from 'react';

export const PrimaryButton: FC<ButtonProps & ComponentProps<'a'>> = ({ 
  className, 
  children,
  ...props 
}) => (
  <Button
    className={`
      bg-cta
      hover:bg-gray-200
      text-cta-text font-medium text-sm
      px-5 py-2.5 rounded-lg
      transition-all duration-300
      hover:-translate-y-0.5
      ${className}
    `}
    {...props}
  >
    {children}
  </Button>
);

export const SecondaryButton: FC<ButtonProps & ComponentProps<'a'>> = ({ 
  className, 
  children,
  ...props 
}) => (
  <Button
    className={`
      bg-transparent
      border border-edge-hover
      hover:border-white
      text-white font-medium text-sm
      px-5 py-2.5 rounded-lg
      transition-all duration-300
      hover:bg-white/5
      ${className}
    `}
    {...props}
  >
    {children}
  </Button>
);
