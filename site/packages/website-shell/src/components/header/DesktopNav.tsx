'use client';

import Link from 'next/link';
import { Stack, Typography } from '@mui/material';
import { WebsiteLogo } from '../WebsiteLogo';
import { MegaMenu } from './MegaMenu';
import { distributionIcons, productIcons, resourceIcons, withIcons } from './menu-icons';
import {
  menuProduct,
  menuDistributions,
  menuExplore,
  menuSolutions,
  menuResources,
  homepageUtilities,
} from '../../data/navigation';
import { scopedTokens as tokens } from '../../theme/tokens';

export function DesktopNav({ variant = 'default' }: { variant?: 'default' | 'homepage' }) {
  return (
    <Stack
      direction="row"
      sx={{
        display: { xs: 'none', md: 'flex' },
        alignItems: 'center',
        gap: 3,
        fontSize: '0.875rem',
      }}
    >
      <WebsiteLogo />
      <MegaMenu
        title="Product"
        leftMenu={[{ items: withIcons(menuProduct, productIcons) }]}
        rightMenu={[
          { title: 'Distributions', items: withIcons(menuDistributions, distributionIcons) },
          // The Distributions section is listed above this column, so its index is not repeated here; the footer's Explore group keeps it.
          { title: 'Explore', items: menuExplore.filter((item) => item.href !== '/distributions') },
        ]}
        leftWidth={320}
        rightWidth={240}
      />
      <MegaMenu
        title="Solutions"
        leftMenu={[{ items: menuSolutions }]}
        footerMenu={{ label: 'View all Solutions', href: '/solutions' }}
      />
      <MegaMenu
        title="Resources"
        leftMenu={[
          {
            items: withIcons(
              variant === 'homepage' ? [...menuResources, ...homepageUtilities] : menuResources,
              resourceIcons
            ),
          },
        ]}
        leftWidth={300}
      />
      <Link href="/pricing" style={{ textDecoration: 'none', color: 'inherit' }}>
        <Typography
          sx={{
            fontSize: '0.875rem',
            fontWeight: 500,
            color: 'grey.100',
            transition: 'color 150ms ease',
            '&:hover': { color: tokens.text.primary },
          }}
        >
          Pricing
        </Typography>
      </Link>
    </Stack>
  );
}
