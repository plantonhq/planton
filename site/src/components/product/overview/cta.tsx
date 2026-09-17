'use client';

import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import { Section, PrimaryButton, SecondaryButton } from '@/components/marketing';

export const ProductCTA = () => {
  return (
    <Section>
      <Box className="text-center max-w-2xl mx-auto py-8">
        <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
          Ready to simplify your infrastructure?
        </Typography>
        <Typography className="text-[#a0a0a0] mb-8">
          Start deploying in minutes. No credit card required. The free tier is free forever — automation is never metered.
        </Typography>
        <Box className="flex flex-col sm:flex-row gap-3 justify-center">
          <Link href="https://planton.ai/signup" target="_blank">
            <PrimaryButton>Start Free</PrimaryButton>
          </Link>
          <Link href="/book-demo">
            <SecondaryButton>Book a Demo</SecondaryButton>
          </Link>
        </Box>
      </Box>
    </Section>
  );
};
