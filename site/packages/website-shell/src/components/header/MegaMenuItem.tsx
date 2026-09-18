'use client';

import Link from 'next/link';
import { Box, Stack, Typography } from '@mui/material';
import type { MenuItem } from '../../data/navigation';
import { tokens } from '../../theme/tokens';

/** The description under a label reads at the secondary role at rest and rises to primary under the pointer; it is the line that says what the page is, never fine print. */
export function MegaMenuItem({ label, subLabel, icon, href, onClick, alignWithMarks = false }: MenuItem & { onClick?: () => void; alignWithMarks?: boolean }) {
  return (
    <Link href={href ?? ''} onClick={onClick} style={{ width: '100%', textDecoration: 'none', color: 'inherit' }}>
      <Stack
        direction="row"
        sx={{
          gap: 2,
          alignItems: { md: 'center' },
          borderRadius: 2,
          px: 1,
          py: 0.75,
          mx: -1,
          transition: 'background-color 150ms ease',
          '&:hover': { bgcolor: 'rgba(255,255,255,0.1)' },
        }}
      >
        {icon ?? (alignWithMarks ? <Box sx={{ width: { xs: 16, md: 24 }, flexShrink: 0 }} aria-hidden /> : null)}
        <Stack sx={{ justifyContent: 'flex-start' }}>
          <Typography sx={{ color: tokens.text.primary, fontWeight: subLabel ? 600 : 400, fontSize: '0.875rem' }}>
            {label}
          </Typography>
          {subLabel && (
            <Typography
              sx={{
                color: tokens.text.secondary,
                fontSize: '0.875rem',
                fontWeight: 400,
                transition: 'color 150ms ease',
                '.MuiStack-root:hover &': { color: tokens.text.primary },
              }}
            >
              {subLabel}
            </Typography>
          )}
        </Stack>
      </Stack>
    </Link>
  );
}
