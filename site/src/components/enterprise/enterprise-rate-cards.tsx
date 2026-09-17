'use client';

import { FC } from 'react';
import Link from 'next/link';
import { Box, Grid2, Stack, Typography } from '@mui/material';
import {
  CheckIcon,
  FeatureTitle,
  PrimaryButton,
} from '@/components/marketing';
import { useMarket } from '@/components/market';
import { MARKETS } from '@/data/pricing';

/**
 * Published rate anchors in the VISITOR'S market (founder direction
 * 2026-08-20): the headline price is always the market the visitor is in —
 * an Indian buyer reads ₹ first, never as a subtitle under a dollar figure.
 * India additionally gets a subtle USD line for comparability (the USD
 * anchor is the price for every market outside India); non-India visitors
 * see no INR anywhere, matching the sitewide India gate. A printed number
 * remains the trust move, and India prices are set India numbers, never an
 * FX conversion.
 */

const tierBullets: Record<string, string[]> = {
  'Enterprise Standard': [
    'Enterprise SSO (SAML/OIDC) and SCIM directory sync — Coming Soon',
    'Supported air-gap posture',
    'Compliance reporting',
    'Named business-hours support',
  ],
  'Enterprise Plus': [
    'Everything in Enterprise Standard',
    '24×7 support with SLA and a dedicated channel',
    'Procurement and security-review assistance',
    'Contracting on your paper',
  ],
};

export const EnterpriseRateCards: FC = () => {
  const { market, marketId } = useMarket();
  return (
    <Stack className="items-center px-4 md:px-8 py-8 gap-4 bg-[#0a0a0a]">
      <Grid2 container spacing={4} className="w-full max-w-5xl items-stretch">
        {market.enterprise.map((tier, tierIndex) => {
          const usTier = MARKETS.us.enterprise[tierIndex];
          return (
            <Grid2 size={{ xs: 12, md: 6 }} key={tier.name} className="flex">
              <Box className="w-full rounded-xl border border-[#2a2a2a] bg-[#151515] hover:border-[#3a3a3a] transition-all duration-300 p-6 md:p-8">
                <Stack className="gap-5 h-full justify-between">
                  <Stack className="gap-4">
                    <FeatureTitle>{tier.name}</FeatureTitle>
                    <Box>
                      <Box className="flex items-baseline gap-1.5">
                        <Typography className="text-3xl font-bold text-white">
                          {tier.perYear}
                        </Typography>
                        <Typography className="text-sm text-[#a0a0a0]">/year</Typography>
                      </Box>
                      {marketId === 'in' && (
                        <Typography className="text-sm text-[#a0a0a0] mt-1">
                          {usTier.perYear}/year — US and other markets
                        </Typography>
                      )}
                    </Box>
                    <Typography className="text-sm font-medium text-[#c0c0c0]">
                      Up to {tier.seatCeiling} seats
                    </Typography>
                    <Stack className="gap-2">
                      {tierBullets[tier.name].map((bullet) => (
                        <Box key={bullet} className="flex items-start gap-2">
                          <Box className="mt-1 flex-shrink-0">
                            <CheckIcon />
                          </Box>
                          <Typography className="text-sm text-[#c0c0c0] leading-relaxed">
                            {bullet}
                          </Typography>
                        </Box>
                      ))}
                    </Stack>
                  </Stack>
                  <Link href="/book-demo" className="self-start">
                    <PrimaryButton>Talk to Us</PrimaryButton>
                  </Link>
                </Stack>
              </Box>
            </Grid2>
          );
        })}
      </Grid2>
      <Typography className="text-sm text-[#8a8a8a] max-w-[760px] text-center">
        Beyond 250 seats, contracts are quoted individually — the published
        packages are where the price sheet stops, not where the product does.
      </Typography>
      <Typography className="text-sm text-[#8a8a8a] max-w-[760px] text-center">
        Features marked Coming Soon are decided packaging that is still
        shipping — they become live as they ship.
      </Typography>
    </Stack>
  );
};
