'use client';

import { FC } from 'react';
import Link from 'next/link';
import { Box, Stack, Typography } from '@mui/material';
import {
  Badge,
  PrimaryButton,
  ShieldIcon,
} from '@/components/marketing';
import { useMarket } from '@/components/market';

/**
 * Enterprise presence directly under the plan grid — a full-width band
 * rather than a cramped fifth column. The anchor is PUBLISHED deliberately:
 * a printed number filters unqualified conversations and lets a buyer
 * qualify themselves before any call. Identity capabilities that are still
 * shipping are never claimed as live (the chips name the package;
 * /pricing/enterprise words the Coming Soon plainly).
 */

const chips = ['Up to 250 Seats', 'SSO/SCIM', 'Air-Gap', 'Compliance Reporting', '24×7 SLA'];

export const EnterpriseBand: FC = () => {
  const { market } = useMarket();
  const anchor = market.enterprise[0].perYear;

  return (
    <Box className="w-full px-4 md:px-8 py-6 bg-[#0a0a0a]">
      <Box className="max-w-7xl mx-auto rounded-xl bg-[#151515] border border-[#2a2a2a] hover:border-[#3a3a3a] transition-all duration-300 p-5 md:p-6">
        <Stack
          direction={{ xs: 'column', md: 'row' }}
          className="items-start md:items-center justify-between gap-4"
        >
          <Box className="flex items-start md:items-center gap-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white flex-shrink-0">
              <ShieldIcon />
            </Box>
            <Box>
              <Box className="flex items-baseline gap-2 flex-wrap">
                <Typography className="text-base font-semibold text-white">
                  Enterprise
                </Typography>
                <Typography className="text-sm text-[#a0a0a0]">
                  starts at {anchor}/year — published rate card
                </Typography>
              </Box>
              <Box className="flex flex-wrap gap-1.5 mt-2">
                {chips.map((chip) => (
                  <Badge key={chip} className="!px-2 !py-0.5 !text-[11px]">
                    {chip}
                  </Badge>
                ))}
              </Box>
            </Box>
          </Box>
          <Link href="/pricing/enterprise" className="flex-shrink-0">
            <PrimaryButton>See Enterprise Plans</PrimaryButton>
          </Link>
        </Stack>
      </Box>
    </Box>
  );
};
