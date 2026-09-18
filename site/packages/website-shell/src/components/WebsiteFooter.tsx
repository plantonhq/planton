'use client';

import { Fragment } from 'react';
import Link from 'next/link';
import { Box, Divider, Stack, Typography } from '@mui/material';
import { WebsiteLogo } from './WebsiteLogo';
import { footerGroups, footerTermsLinks, type FooterGroup } from '../data/navigation';
import { tokens } from '../theme/tokens';

function FooterLinkGroup({ title, items }: FooterGroup) {
  return (
    <Stack sx={{ gap: 1.5 }}>
      <Typography sx={{ fontSize: '0.75rem', fontWeight: 400, color: tokens.text.faint }}>{title}</Typography>
      {items.map((item) => (
        <Link
          href={item.href}
          key={item.href}
          style={{ fontSize: '0.75rem', fontWeight: 400, textDecoration: 'none', color: 'inherit' }}
        >
          {item.label}
        </Link>
      ))}
    </Stack>
  );
}

/**
 * Full-width website footer with link groups, terms links, and copyright.
 * Renders a responsive grid: 3-column on mobile, 5-column on desktop.
 */
export function WebsiteFooter() {
  return (
    <Stack
      component="footer"
      sx={{
        px: { md: 12 },
        py: { md: 5 },
        gap: { md: 3 },
        bgcolor: tokens.surface.canvas,
      }}
    >
      {/* Main content row */}
      <Stack
        sx={{
          px: { xs: 2.5, md: 0 },
          py: { xs: 3, md: 0 },
          flexDirection: { md: 'row' },
          gap: 4,
          justifyContent: 'space-between',
        }}
      >
        {/* Logo column */}
        <Stack sx={{ gap: 1.5 }}>
          <WebsiteLogo />
        </Stack>

        {/* Mobile grid: 3 columns */}
        <Box
          sx={{
            display: { xs: 'grid', md: 'none' },
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: 1.5,
          }}
        >
          <Stack sx={{ gap: 1.5 }}>
            {footerGroups
              .filter((g) => ['product', 'explore'].includes(g.id))
              .map((group, index) => (
                <Stack sx={{ gap: 1.5 }} key={group.id}>
                  {index !== 0 && <Divider sx={{ borderColor: tokens.edge.default }} />}
                  <FooterLinkGroup {...group} />
                </Stack>
              ))}
          </Stack>
          <Stack sx={{ gap: 1.5 }}>
            {footerGroups
              .filter((g) => ['open_source', 'get_started'].includes(g.id))
              .map((group, index) => (
                <Stack sx={{ gap: 1.5 }} key={group.id}>
                  {index !== 0 && <Divider sx={{ borderColor: tokens.edge.default }} />}
                  <FooterLinkGroup {...group} />
                </Stack>
              ))}
          </Stack>
          <Stack sx={{ gap: 1.5 }}>
            {footerGroups
              .filter((g) => g.id === 'resources')
              .map((group) => (
                <FooterLinkGroup key={group.id} {...group} />
              ))}
          </Stack>
        </Box>

        {/* Desktop grid: 5 columns */}
        <Box
          sx={{
            display: { xs: 'none', md: 'grid' },
            gridTemplateColumns: 'repeat(5, 1fr)',
            columnGap: 10,
            rowGap: 3,
          }}
        >
          {footerGroups.map((group) => (
            <FooterLinkGroup key={group.id} {...group} />
          ))}
        </Box>
      </Stack>

      <Divider sx={{ borderColor: tokens.edge.default }} />

      {/* Bottom bar: terms + copyright */}
      <Stack
        sx={{
          px: { xs: 2.5, md: 0 },
          py: { xs: 3, md: 0 },
          flexDirection: { md: 'row' },
          justifyContent: { xs: 'center', md: 'space-between' },
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <Stack
          direction="row"
          sx={{
            flexWrap: 'wrap',
            alignItems: 'center',
            justifyContent: 'center',
            columnGap: 2.5,
            rowGap: 0.5,
          }}
        >
          {footerTermsLinks.map((term, index) => (
            <Fragment key={term.href}>
              {index > 0 && (
                <Box
                  sx={{
                    display: { xs: 'none', sm: 'block' },
                    width: 4,
                    aspectRatio: '1',
                    borderRadius: '50%',
                    bgcolor: tokens.text.primary,
                  }}
                />
              )}
              <Link
                href={term.href}
                style={{ fontSize: '0.75rem', fontWeight: 400, whiteSpace: 'nowrap', textDecoration: 'none', color: 'inherit' }}
              >
                {term.label}
              </Link>
            </Fragment>
          ))}
        </Stack>
        <Typography sx={{ fontSize: '0.75rem', fontWeight: 400 }}>
          {`\u00A9${new Date().getFullYear()} Planton Cloud Inc. All Rights Reserved.`}
        </Typography>
      </Stack>
    </Stack>
  );
}
