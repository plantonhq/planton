'use client';

import type { ComponentProps, FC } from 'react';
import { Button, type ButtonProps } from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';
import { scopedTokens as tokens } from '../../theme/tokens';

const btnBaseSx: SxProps<Theme> = {
  px: { xs: 1.5, md: 2.5 },
  py: { xs: 1, md: 1.5 },
  height: { xs: 32, md: 40 },
  fontSize: { xs: '0.75rem', md: '1rem' },
  lineHeight: { xs: 2, md: 1 },
  fontWeight: 500,
  borderRadius: '10px',
  width: 'fit-content',
};

export const ShellButton: FC<ButtonProps & ComponentProps<'a'>> = ({ sx, ...props }) => {
  return <Button sx={[btnBaseSx, ...(Array.isArray(sx) ? sx : [sx])] as SxProps<Theme>} {...props} />;
};

const primaryOverlaySx: SxProps<Theme> = {
  bgcolor: tokens.cta.background,
  color: tokens.cta.text,
  '&:hover': { bgcolor: 'grey.200' },
  transition: 'background-color 150ms ease',
};

export const PrimaryShellButton: FC<ButtonProps & ComponentProps<'a'>> = ({ sx, ...props }) => {
  return (
    <ShellButton
      sx={[primaryOverlaySx, ...(Array.isArray(sx) ? sx : [sx])] as SxProps<Theme>}
      {...props}
    />
  );
};
