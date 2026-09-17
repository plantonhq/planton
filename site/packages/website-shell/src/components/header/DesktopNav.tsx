'use client';

import Link from 'next/link';
import { Stack, Typography } from '@mui/material';
import { WebsiteLogo } from '../WebsiteLogo';
import { MegaMenu } from './MegaMenu';
import { distributionIcons, productIcons, resourceIcons, withIcons } from './menu-icons';
import {
  menuProduct,
  menuDistributions,
  menuExplorer,
  menuByUseCases,
  menuBySize,
  menuByRole,
  menuResources,
} from '../../data/navigation';

export function DesktopNav() {
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
          { title: 'Explore', items: menuExplorer },
        ]}
        leftWidth={320}
      />
      <MegaMenu
        title="Solutions"
        leftMenu={[
          { title: 'By Use Case', items: menuByUseCases },
          { title: 'By Size', items: menuBySize },
        ]}
        rightMenu={[{ title: 'By Role', items: menuByRole }]}
        footerMenu={{ label: 'View all Solutions', href: '/solutions' }}
      />
      <MegaMenu
        title="Resources"
        leftMenu={[{ items: withIcons(menuResources, resourceIcons) }]}
        leftWidth={300}
      />
      <Link href="/pricing" style={{ textDecoration: 'none', color: 'inherit' }}>
        <Typography
          sx={{
            fontSize: '0.875rem',
            fontWeight: 500,
            color: 'grey.100',
            transition: 'color 150ms ease',
            '&:hover': { color: '#fff' },
          }}
        >
          Pricing
        </Typography>
      </Link>
    </Stack>
  );
}
