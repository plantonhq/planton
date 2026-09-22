'use client';

import { CssBaseline, ThemeProvider as MUIThemeProvider } from '@mui/material';
import { AppRouterCacheProvider } from '@mui/material-nextjs/v16-appRouter';
import { websiteTheme } from '@planton/website-shell/theme';
import type { PropsWithChildren } from 'react';

/**
 * The one MUI theme on the site. It is the website-shell package's theme,
 * the same object the console mounts when it renders the marketing header
 * and footer, so a component looks identical wherever the shell is. The
 * shell's own provider nests the same theme inside the site group; nesting
 * an identical theme is a no-op, and standalone surfaces (decks, the demo)
 * get the theme without the header and footer.
 *
 * CssBaseline lives here and only here: the shell omits it on purpose so a
 * host application owns global resets.
 */
export function ThemeProvider({ children }: PropsWithChildren) {
  return (
    <AppRouterCacheProvider options={{ enableCssLayer: true }}>
      <MUIThemeProvider theme={websiteTheme}>
        <CssBaseline />
        {children}
      </MUIThemeProvider>
    </AppRouterCacheProvider>
  );
}
