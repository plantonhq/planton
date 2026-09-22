'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Divider, Stack, Typography } from '@mui/material';
import { DensityMedium, Close } from '@mui/icons-material';
import { WebsiteLogo } from '../WebsiteLogo';
import { ShellDrawer } from './styled';
import { MenuAccordion } from './MenuAccordion';
import { MegaMenuItem } from './MegaMenuItem';
import { distributionIcons, productIcons, resourceIcons, withIcons } from './menu-icons';
import { MobileAuthButtons } from './AuthButtons';
import { MobileDownloadLink } from './DownloadLink';
import { DiscordButton } from '../shared/DiscordButton';
import {
  menuProduct,
  menuDistributions,
  menuExplore,
  menuSolutions,
  menuResources,
} from '../../data/navigation';
import { tokens } from '../../theme/tokens';

const dividerSx = { borderColor: tokens.edge.default } as const;

export function MobileNav() {
  const [open, setOpen] = useState(false);
  const [expandedPanel, setExpandedPanel] = useState<string | false>(false);

  const handlePanelChange = (panel: string) => (_: React.SyntheticEvent, isExpanded: boolean) => {
    setExpandedPanel(isExpanded ? panel : false);
  };

  const toggleDrawer = () => setOpen((prev) => !prev);

  return (
    <>
      <DensityMedium
        fontSize="small"
        onClick={toggleDrawer}
        sx={{ position: 'relative', zIndex: 10, cursor: 'pointer' }}
      />

      <ShellDrawer open={open} onClose={toggleDrawer}>
        <Stack
          sx={{ gap: 3.5 }}
          onClick={(e) => {
            if ((e.target as HTMLElement).closest('a')) setOpen(false);
          }}
        >
          <Stack direction="row" sx={{ alignItems: 'center', gap: 4, justifyContent: 'space-between' }}>
            <WebsiteLogo />
            <Close onClick={toggleDrawer} sx={{ cursor: 'pointer' }} />
          </Stack>

          <Stack sx={{ gap: 4 }}>
            {/* Product */}
            <MenuAccordion
              expanded={expandedPanel === 'product'}
              title="Product"
              onChange={handlePanelChange('product')}
            >
              <Stack sx={{ gap: 3 }}>
                {withIcons(menuProduct, productIcons).map((item) => (
                  <MegaMenuItem key={item.label} {...item} />
                ))}
                <Divider sx={{ ...dividerSx, mt: -1.5 }} />
                <Stack sx={{ gap: 3 }}>
                  <Typography sx={{ fontSize: '0.875rem', fontWeight: 400 }}>Distributions</Typography>
                  {withIcons(menuDistributions, distributionIcons).map((item) => (
                    <MegaMenuItem key={item.label} {...item} />
                  ))}
                  <Divider sx={{ ...dividerSx, mt: -1.5 }} />
                </Stack>
                <Stack sx={{ gap: 2 }}>
                  <Typography sx={{ fontSize: '0.875rem', fontWeight: 400 }}>Explore</Typography>
                  {menuExplore.map((item) => (
                    <MegaMenuItem key={item.label} {...item} />
                  ))}
                  <Divider sx={{ ...dividerSx, mt: -0.5 }} />
                </Stack>
              </Stack>
            </MenuAccordion>

            {/* Solutions */}
            <MenuAccordion
              expanded={expandedPanel === 'solutions'}
              title="Solutions"
              onChange={handlePanelChange('solutions')}
            >
              <Stack sx={{ gap: 2 }}>
                {menuSolutions.map((item) => (
                  <MegaMenuItem key={item.label} {...item} />
                ))}
                <Divider sx={{ ...dividerSx, mt: -0.5 }} />
              </Stack>
            </MenuAccordion>

            {/* Resources */}
            <MenuAccordion
              expanded={expandedPanel === 'resources'}
              title="Resources"
              onChange={handlePanelChange('resources')}
            >
              <Stack sx={{ gap: 2.5 }}>
                {withIcons(menuResources, resourceIcons).map((item) => (
                  <MegaMenuItem key={item.label} {...item} />
                ))}
                <Divider sx={{ ...dividerSx, mt: -1 }} />
              </Stack>
            </MenuAccordion>

            {/* Pricing */}
            <Link href="/pricing" style={{ textDecoration: 'none', color: 'inherit' }}>
              <Typography sx={{ color: 'text.secondary', fontSize: '1rem', fontWeight: 600 }}>
                Pricing
              </Typography>
            </Link>

            <Divider sx={dividerSx} />

            {/* Discord + Auth */}
            <Stack sx={{ gap: 1.5 }}>
              <DiscordButton sx={{ color: 'text.secondary', width: '100%', justifyContent: 'center' }} />
              <MobileDownloadLink />
              <MobileAuthButtons />
            </Stack>
          </Stack>
        </Stack>
      </ShellDrawer>
    </>
  );
}
