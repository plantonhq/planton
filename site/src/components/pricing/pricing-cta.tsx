'use client';

import { FC } from 'react';
import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import {
  BodyText,
  Card,
  PrimaryButton,
  SecondaryButton,
} from '@/components/marketing';

/** The closing CTA: start free, or talk to a human if that helps. */
export const PricingCta: FC = () => {
  return (
    <Box className="w-full px-4 md:px-8 pb-16 bg-[#0a0a0a]">
      <Box className="max-w-3xl mx-auto">
        <Card hover={false} className="!p-8 md:!p-10 text-center">
          <Typography className="text-2xl md:text-3xl font-semibold text-white mb-3">
            Ready to Try Planton?
          </Typography>
          <BodyText className="!text-base mx-auto max-w-xl mb-7">
            Start on the free tier — no credit card required. Or run it
            yourself free forever with the community edition.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center">
            <Link href="https://planton.ai" target="_blank">
              <PrimaryButton>Start Free</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Card>
      </Box>
    </Box>
  );
};
