'use client';

import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import {
  Section,
  PrimaryButton,
  SecondaryButton,
  Card,
  BodyText,
  Divider,
} from '@/components/marketing';
import { ScrollReveal, RelatedModules } from '@/components/product/shared';

export const CliCTA = () => {
  return (
    <>
      <RelatedModules modules={['infra-hub', 'service-hub', 'open-source']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              One CLI, every operation
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Install in seconds, deploy in minutes
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              brew install plantonhq/tap/planton &mdash; then apply your first manifest.
              Same commands work locally, in CI, everywhere.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Get Started</PrimaryButton>
              </Link>
              <Link href="/book-demo">
                <SecondaryButton>Book a Demo</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
