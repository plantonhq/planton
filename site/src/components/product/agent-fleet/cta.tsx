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

export const AgentFleetCTA = () => {
  return (
    <>
      <RelatedModules modules={['service-hub', 'cli', 'security']} />
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <ScrollReveal>
          <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
              Stop firefighting
            </Typography>
            <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
              Let agents handle the toil
            </Typography>
            <BodyText className="!text-base mx-auto max-w-xl mb-8">
              Deploy your first agent in minutes. Purpose-built for your infrastructure,
              with real access to your Planton resources.
            </BodyText>
            <Box className="flex flex-col sm:flex-row gap-3 justify-center">
              <Link href="https://planton.ai/signup" target="_blank">
                <PrimaryButton>Explore Agent Fleet</PrimaryButton>
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
