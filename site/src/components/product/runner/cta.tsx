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

export const RunnerCTA = () => {
  return (
    <>
      <RelatedModules modules={['infra-hub', 'security', 'cli']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              Self-hosted execution
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Install Runner in minutes
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              One binary. Outbound-only. Your credentials never leave your cloud.
              Register in the console and run the install command.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Install Runner</PrimaryButton>
              </Link>
              <Link href="/docs/runner">
                <SecondaryButton>Read the Docs</SecondaryButton>
              </Link>
            </Box>
          </Card>
        </ScrollReveal>
      </Section>
    </>
  );
};
