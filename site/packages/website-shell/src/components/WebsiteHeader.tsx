'use client';

import { Box, Stack } from '@mui/material';
import { DesktopNav } from './header/DesktopNav';
import { MobileNav } from './header/MobileNav';
import { DesktopAuthButtons } from './header/AuthButtons';
import { DesktopDownloadLink } from './header/DownloadLink';
import { DiscordButton } from './shared/DiscordButton';
import { WebsiteLogo } from './WebsiteLogo';

/**
 * Full-width fixed header bar that renders the marketing website's
 * navigation chrome. Automatically switches between desktop mega-menu
 * and mobile hamburger drawer at the md breakpoint (768px).
 *
 * Mobile: logo on the left, hamburger on the right (matches stigmer.ai).
 * Desktop: full mega-menu nav on the left, actions on the right.
 */
export interface WebsiteHeaderProps {
  variant?: 'default' | 'homepage';
  onPrimaryAction?: () => void;
  onSelfServiceAction?: () => void;
}
export function WebsiteHeader({
  variant = 'default',
  onPrimaryAction,
  onSelfServiceAction,
}: WebsiteHeaderProps) {
  return (
    <>
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

        {/* Desktop: full nav on the left */}
        <DesktopNav variant={variant} />

        {/* Desktop: right-side actions */}
        <Stack
          direction="row"
          sx={{
            display: { xs: 'none', md: 'flex' },
            alignItems: 'center',
            gap: 1.5,
            fontSize: '0.875rem',
          }}
        >
          {variant === 'default' && (
            <>
              <DiscordButton compact />
              <DesktopDownloadLink />
            </>
          )}
          <DesktopAuthButtons
            variant={variant}
            onPrimaryAction={onPrimaryAction}
            onSelfServiceAction={onSelfServiceAction}
          />
        </Stack>

        {/* Mobile: hamburger on the right */}
        <Box sx={{ display: { xs: 'block', md: 'none' } }}>
          <MobileNav
            variant={variant}
            onPrimaryAction={onPrimaryAction}
            onSelfServiceAction={onSelfServiceAction}
          />
        </Box>
      </Stack>
    </>
  );
}
