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

export const InfraHubCTA = () => {
  return (
    <>
      <RelatedModules modules={['runner', 'security', 'open-source']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              Built on Planton open source
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Deploy your first resource in minutes
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Every Infra Hub deployment is powered by Planton&apos;s open-source modules.
              Your infrastructure definitions are portable &mdash; no lock-in, ever.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Start Free</PrimaryButton>
              </Link>
              <Link href="/features/open-source">
                <SecondaryButton>Explore open source</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
