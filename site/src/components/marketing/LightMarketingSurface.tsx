 'use client';
import { ThemeProvider } from '@mui/material/styles';
import { lightVariables } from '@/theme/homepage';
import { lightWebsiteTheme } from '@/theme/light-website';
import type { ReactNode } from 'react';
export function LightMarketingSurface({children}:{children:ReactNode}) {
  return <ThemeProvider theme={lightWebsiteTheme}><div data-marketing-appearance="light" style={{...lightVariables,color:lightWebsiteTheme.palette.text.primary,background:lightWebsiteTheme.palette.background.default,minHeight:'100vh'}}>{children}</div></ThemeProvider>;
}
