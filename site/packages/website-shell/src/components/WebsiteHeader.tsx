'use client';

import { Box, Stack } from '@mui/material';
import { HeaderLinks } from './header/HeaderLinks';
import { MobileNav } from './header/MobileNav';
import { DesktopAuthButtons } from './header/AuthButtons';
import { WebsiteLogo } from './WebsiteLogo';

/**
 * The fixed header every website page shares: the logo and four plain links
 * on the left, Sign In and Start Free on the right. Below the md breakpoint
 * (768px) the links and doors move into a drawer behind a menu button.
 */
export interface WebsiteHeaderProps {
  /** Called when a visitor clicks Start Free, for the site's own analytics. */
  onSelfServiceAction?: () => void;
}

export function WebsiteHeader({ onSelfServiceAction }: WebsiteHeaderProps) {
  return (
    <Stack
      component="header"
      direction="row"
      sx={{
        width: '100%',
        height: { md: 70 },
        position: 'fixed',
        top: 0,
        left: 0,
        zIndex: 30,
        justifyContent: 'space-between',
        alignItems: 'center',
        px: { xs: 2.5, md: 4 },
        py: { xs: 1.5, md: 1.25 },
        bgcolor: 'var(--website-header-background, rgba(10, 10, 10, 0.8))',
        backdropFilter: 'blur(16px)',
        borderBottom: '1px solid var(--website-header-border, rgba(42, 42, 42, 0.2))',
      }}
    >
      {/* Mobile: logo on the left */}
      <Box sx={{ display: { xs: 'block', md: 'none' } }}>
        <WebsiteLogo />
      </Box>

      {/* Desktop: logo and links on the left */}
      <HeaderLinks />

      {/* Desktop: the two doors on the right */}
      <Stack
        direction="row"
        sx={{ display: { xs: 'none', md: 'flex' }, alignItems: 'center', gap: 1.5, fontSize: '0.875rem' }}
      >
        <DesktopAuthButtons onSelfServiceAction={onSelfServiceAction} />
      </Stack>

      {/* Mobile: the drawer's button on the right */}
      <Box sx={{ display: { xs: 'block', md: 'none' } }}>
        <MobileNav onSelfServiceAction={onSelfServiceAction} />
      </Box>
    </Stack>
  );
}
