'use client';

import { FC } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import {
  Badge,
  CheckIcon,
} from '@/components/marketing';
import { MarketSelector } from '@/components/market';

/**
 * The pricing page opener: deliberately compact so the plan grid — the
 * page's real content — is visible without scrolling. The headline names
 * the journey; the one-line sub names both homes (our cloud, your
 * cluster) so a visitor who never scrolls still knows self-hosted
 * exists. The money promises render as one row of scannable ticks, not
 * a paragraph, and the market fact lives on the selector itself
 * ("Prices shown for …"), never repeated in prose.
 */

const trustChips = [
  'Free Tier Never Bills',
  'Cancel Anytime',
  '14-Day License Money-Back',
  'Expiry Never Breaks Anything',
];

export const PricingHero: FC = () => {
  return (
    <Stack className="items-center gap-5 pt-14 pb-2 px-4 bg-[#0a0a0a] text-center">
      <Badge>Pricing</Badge>
      <Typography
        variant="h1"
        className="text-3xl md:text-4xl lg:text-5xl font-semibold text-white leading-tight tracking-tight max-w-[900px]"
      >
        Plans for Every Stage of Your Journey
      </Typography>
      <Typography className="text-sm md:text-base text-[#a0a0a0] max-w-[640px]">
        The same platform — on Planton.ai or your own infrastructure.
      </Typography>
      <Box className="flex flex-wrap justify-center gap-x-5 gap-y-2">
        {trustChips.map((chip) => (
          <Box key={chip} className="flex items-center gap-1.5">
            <CheckIcon />
            <Typography className="text-xs text-[#b0b0b0]">{chip}</Typography>
          </Box>
        ))}
      </Box>
      <MarketSelector className="mt-1" />
    </Stack>
  );
};
