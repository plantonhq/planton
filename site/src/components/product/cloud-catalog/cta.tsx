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

export const CloudCatalogCTA = () => {
  return (
    <>
      <RelatedModules modules={['infra-hub', 'runner', 'open-source']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              Open infrastructure catalog
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Start exploring
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Browse the full catalog without signing up. When you find the right module,
              sign in and deploy it to your cloud in minutes.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="/cloud-catalog">
                <PrimaryButton>Explore Catalog</PrimaryButton>
              </Link>
              <Link href="/docs/infrastructure">
                <SecondaryButton>Read the Docs</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
