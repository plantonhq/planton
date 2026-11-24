'use client';

import type { PropsWithChildren } from 'react';
import { Box } from '@mui/material';
import type { Theme } from '@mui/material/styles';
import { WebsiteThemeProvider } from '../providers/WebsiteThemeProvider';
import { WebsiteHeader } from './WebsiteHeader';
import { WebsiteFooter } from './WebsiteFooter';

interface WebsiteShellProps extends PropsWithChildren {
  /**
   * Override the default website theme. Create one with
   * `createWebsiteTheme(overrides)` from `@planton/website-shell/theme`.
   */
  theme?: Theme;
  /** Called when a visitor clicks the header's Start Free, for the site's own analytics. */
  onSelfServiceAction?: () => void;
}

/**
 * Complete website chrome wrapper: header + content area + footer, the same
 * on every page, wrapped in the website theme.
 *
 * When used inside the console, the nested `WebsiteThemeProvider`
 * overrides the console's theme locally — MUI's ThemeProvider nesting
 * is designed for exactly this use case.
 *
 * The content area is offset by 70px to account for the fixed header.
 */
export function WebsiteShell({ theme, children, onSelfServiceAction }: WebsiteShellProps) {
  return (
    <WebsiteThemeProvider theme={theme}>
      <Box sx={{ minHeight: '100%', bgcolor: 'background.default' }}>
        <WebsiteHeader onSelfServiceAction={onSelfServiceAction} />
        <Box component="div" sx={{ pt: '70px' }}>
          {children}
        </Box>
        <WebsiteFooter />
      </Box>
    </WebsiteThemeProvider>
  );
}
