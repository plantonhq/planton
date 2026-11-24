'use client';

import Link from 'next/link';
import { Stack, Typography } from '@mui/material';
import { WebsiteLogo } from '../WebsiteLogo';
import { headerLinks } from '../../data/navigation';
import { scopedTokens as tokens } from '../../theme/tokens';

/**
 * The desktop header's left side: the logo and the header's links as plain
 * text links. No menus: a stranger sees every destination at once.
 */
export function HeaderLinks() {
  return (
    <Stack
      component="nav"
      aria-label="Main"
      direction="row"
      sx={{
        display: { xs: 'none', md: 'flex' },
        alignItems: 'center',
        gap: 3,
        fontSize: '0.875rem',
      }}
    >
      <WebsiteLogo />
      {headerLinks.map((item) => (
        <Link key={item.href} href={item.href} style={{ textDecoration: 'none', color: 'inherit' }}>
          <Typography
            sx={{
              fontSize: '0.875rem',
              fontWeight: 500,
              color: 'grey.100',
              transition: 'color 150ms ease',
              '&:hover': { color: tokens.text.primary },
            }}
          >
            {item.label}
          </Typography>
        </Link>
      ))}
    </Stack>
  );
}
