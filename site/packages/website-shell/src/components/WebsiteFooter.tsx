'use client';

import { Fragment } from 'react';
import Link from 'next/link';
import { Box, Divider, Stack, Typography } from '@mui/material';
import { WebsiteLogo } from './WebsiteLogo';
import { footerGroups, footerTermsLinks, type FooterGroup } from '../data/navigation';
import { scopedTokens as tokens } from '../theme/tokens';

function FooterLinkGroup({ title, items }: FooterGroup) {
  return (
    <Stack sx={{ gap: 1.5 }}>
      <Typography sx={{ fontSize: '0.75rem', fontWeight: 400, color: tokens.text.faint }}>
        {title}
      </Typography>
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
 * Full-width website footer with link groups, terms links, and copyright,
 * the same on every page. Renders a responsive grid: 2-column on mobile,
 * one column per group on desktop.
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
        '& a': { fontSize: '0.875rem !important', lineHeight: 1.6 },
        borderTop: `1px solid ${tokens.edge.default}`,
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

        {/* Mobile grid: 2 columns */}
        <Box
          sx={{
            display: { xs: 'grid', md: 'none' },
            gridTemplateColumns: 'repeat(2, 1fr)',
            columnGap: 1.5,
            rowGap: 3,
          }}
        >
          {footerGroups.map((group) => (
            <FooterLinkGroup key={group.id} {...group} />
          ))}
        </Box>

        {/* Desktop grid: one column per group */}
        <Box
          sx={{
            display: { xs: 'none', md: 'grid' },
            // The groups fill the row after the logo, so their columns spread
            // evenly at every width instead of packing against the right edge.
            flex: 1,
            gridTemplateColumns: `repeat(${footerGroups.length}, 1fr)`,
            columnGap: 10,
            rowGap: 3,
          }}
        >
          {footerGroups.map((group) => (
            <FooterLinkGroup key={group.id} {...group} />
          ))}
        </Box>
      </Stack>

      <Divider style={{ borderBottom: `1px solid ${tokens.edge.default}` }} />

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
                style={{
                  fontSize: '0.75rem',
                  fontWeight: 400,
                  whiteSpace: 'nowrap',
                  textDecoration: 'none',
                  color: 'inherit',
                }}
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
