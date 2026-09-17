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

export const OpenSourceCTA = () => {
  return (
    <>
      <RelatedModules modules={['infra-hub', 'cli', 'runner']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              No lock-in, ever
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Your infrastructure, your definitions, your choice
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Start with Planton open source for portable, open-source infrastructure definitions.
              Add Planton when you need the managed platform.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://github.com/plantonhq/planton" target="_blank">
                <PrimaryButton>Explore Planton on GitHub</PrimaryButton>
              </Link>
              <Link href="https://planton.ai/signup" target="_blank">
                <SecondaryButton>Try Planton Free</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
