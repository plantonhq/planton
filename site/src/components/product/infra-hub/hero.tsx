'use client';

import Link from 'next/link';
import { Box, Typography } from '@mui/material';
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
import { PLATFORM_STATS } from '@/data/platform-stats';

const providers = [
  'AWS', 'GCP', 'Azure', 'Kubernetes',
  'DigitalOcean', 'Cloudflare', 'Auth0', 'OpenFGA',
];

const heroTerminalLines: TerminalLine[] = [
  { text: '$ planton chart install api aws-ecs-environment -f values.yaml', className: 'text-white' },
  { text: '' },
  { text: '  Resolved from platform chart catalog', className: 'text-[#b0b0b0]' },
  { text: '' },
  { text: '✓ VPC and networking provisioned (2m 15s)', className: 'text-[#a0a0a0]' },
  { text: '✓ Load balancer configured (1m 30s)', className: 'text-[#a0a0a0]' },
  { text: '✓ ECS cluster created (3m 45s)', className: 'text-[#a0a0a0]' },
  { text: '✓ Container service deployed (1m 18s)', className: 'text-[#a0a0a0]' },
  { text: '✓ DNS and TLS certificates ready (42s)', className: 'text-[#a0a0a0]' },
  { text: '' },
  { text: '⚡ Environment ready in 9 minutes', className: 'text-amber-400' },
  { text: '' },
  { text: '→ https://api.acmecorp.io', className: 'text-white' },
];

export const InfraHubHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Infra Hub</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Infrastructure deployment without the wait
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Your developers wait days for infrastructure. Your ops team is the bottleneck.
            Every cloud has different tools, different workflows, different nightmares.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Infra Hub &mdash; Cursor for Cloud Infrastructure. Describe what you need, or use the
            wizard, and deploy any cloud resource &mdash; from a Postgres database to an entire
            Kubernetes cluster &mdash; in minutes. One workflow. Any provider.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Deploy Your First Resource</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroTerminalLines}
            title="planton cli"
            lineDelay={300}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>

        <Box className="flex flex-wrap justify-center gap-2 mt-8">
          {providers.map((provider) => (
            <Typography
              key={provider}
              component="span"
              className="text-xs text-[#666] px-2.5 py-1 rounded-md border border-[#1a1a1a] bg-[#111]"
            >
              {provider}
            </Typography>
          ))}
          <Typography
            component="span"
            className="text-xs text-[#555] px-2.5 py-1"
          >
            {PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers via Planton open source
          </Typography>
        </Box>
      </Box>
    </Section>
  );
};
