'use client';

import Link from 'next/link';
import { Box } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  PrimaryButton,
  SecondaryButton,
  Badge,
  BodyText,
} from '@/components/marketing';
import { AnimatedTerminal, ScrollReveal } from '@/components/product/shared';
import type { TerminalLine } from '@/components/product/shared';

const heroLines: TerminalLine[] = [
  { text: '# Register and install Runner', className: 'text-[#555]' },
  { text: '$ planton runner install --channel my-channel', className: 'text-white' },
  { text: '' },
  { text: '✓ Runner binary downloaded (v1.4.2)', className: 'text-[#10b981]' },
  { text: '✓ Secure identity provisioned', className: 'text-[#10b981]' },
  { text: '✓ API key provisioned and stored', className: 'text-[#10b981]' },
  { text: '✓ Connected to control plane (outbound only)', className: 'text-[#10b981]' },
  { text: '' },
  { text: 'Runner is live. No inbound ports opened.', className: 'text-[#b0b0b0]' },
];

export const RunnerHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Runner</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Execute in your cloud, orchestrate from ours
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            SaaS platforms want your cloud credentials. On-prem tools require you to manage everything.
            You shouldn&apos;t have to choose between convenience and security.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Planton Runner: A single binary that runs in your cloud. Planton orchestrates. Runner
            executes. Your credentials never leave your account.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Install Runner</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="planton runner install"
            lineDelay={300}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Box>
    </Section>
  );
};
