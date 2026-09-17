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

export const SecurityCTA = () => {
  return (
    <>
      <RelatedModules modules={['runner', 'open-source', 'infra-hub']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              Built-in Security
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              See how Planton secures your infrastructure
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Secrets management, identity and access, audit trails, zero-trust networking &mdash;
              all native to the platform.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Start Free</PrimaryButton>
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
