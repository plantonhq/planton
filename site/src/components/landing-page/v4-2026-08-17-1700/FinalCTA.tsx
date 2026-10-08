'use client';

import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import { FC } from 'react';
import {
  Section,
  SectionTitle,
  PrimaryButton,
  SecondaryButton,
  ArrowRightIcon,
} from './shared';
import { FREE_TIER_SEATS } from '@/data/pricing';
import { SITE } from '@/data/site-pages';

/**
 * The close. One promise, two doors, no new claims — everything above
 * this section already made the argument.
 */

export const FinalCTA: FC = () => (
  <Section id="get-started">
    <Box className="text-center py-8">
      <SectionTitle className="mb-3">Start Deploying Today</SectionTitle>
      <Typography className="text-sm md:text-base text-[#a0a0a0] max-w-xl mx-auto mb-8">
        Free for up to {FREE_TIER_SEATS} people, no credit card required.
        Deploy your first infrastructure in under an hour.
      </Typography>

      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        className="gap-3 items-center justify-center"
      >
        <Link href={`${SITE.url}/login`}>
          <PrimaryButton className="text-sm px-8 py-3">
            Start Free
            <ArrowRightIcon />
          </PrimaryButton>
        </Link>
        <Link href="/book-demo">
          <SecondaryButton className="text-sm px-8 py-3">
            Book a Demo
          </SecondaryButton>
        </Link>
      </Stack>
    </Box>
  </Section>
);
