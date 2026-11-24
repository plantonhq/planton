'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Divider, Stack, Typography, IconButton } from '@mui/material';
import { DensityMedium, Close } from '@mui/icons-material';
import { WebsiteLogo } from '../WebsiteLogo';
import { ShellDrawer } from './styled';
import { MobileAuthButtons } from './AuthButtons';
import { headerLinks } from '../../data/navigation';
import { scopedTokens as tokens } from '../../theme/tokens';

/** The header's links and doors in a drawer, below the md breakpoint. */
export function MobileNav({ onSelfServiceAction }: { onSelfServiceAction?: () => void }) {
  const [open, setOpen] = useState(false);
  const toggleDrawer = () => setOpen((prev) => !prev);

  return (
    <>
      <IconButton
        aria-label="Open navigation"
        onClick={toggleDrawer}
        size="small"
        sx={{ color: 'text.primary' }}
      >
        <DensityMedium fontSize="small" />
      </IconButton>

      {/* Opens from the side the menu button sits on, so the drawer covers it
          and its own close button is the only control on screen. */}
      <ShellDrawer anchor="right" open={open} onClose={toggleDrawer}>
        <Stack
          sx={{ gap: 3.5 }}
          onClick={(e) => {
            if ((e.target as HTMLElement).closest('a')) setOpen(false);
          }}
        >
          <Stack direction="row" sx={{ alignItems: 'center', gap: 4, justifyContent: 'space-between' }}>
            <WebsiteLogo />
            <IconButton aria-label="Close navigation" onClick={toggleDrawer} sx={{ color: 'text.primary' }}>
              <Close />
            </IconButton>
          </Stack>

          <Stack component="nav" aria-label="Main" sx={{ gap: 3 }}>
            {headerLinks.map((item) => (
              <Link key={item.href} href={item.href} style={{ textDecoration: 'none', color: 'inherit' }}>
                <Typography sx={{ color: 'text.primary', fontSize: '1rem', fontWeight: 500 }}>
                  {item.label}
                </Typography>
              </Link>
            ))}
          </Stack>

          <Divider sx={{ borderColor: tokens.edge.default }} />

          <Stack sx={{ gap: 1.5 }}>
            <MobileAuthButtons onSelfServiceAction={onSelfServiceAction} />
          </Stack>
        </Stack>
      </ShellDrawer>
    </>
  );
}
