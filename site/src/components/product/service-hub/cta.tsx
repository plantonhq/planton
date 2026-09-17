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

export const ServiceHubCTA = () => {
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
              Git to Production
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Start deploying
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Connect your repo, push your code, and let Planton handle the rest &mdash;
              builds, environments, ingress, and runtime access out of the box.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Deploy Your First Service</PrimaryButton>
              </Link>
              <Link href="/docs/ci-cd">
                <SecondaryButton>Read the Docs</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
