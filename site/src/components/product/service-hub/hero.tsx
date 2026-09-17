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

const heroLines: TerminalLine[] = [
  { text: '$ git push origin main', className: 'text-white' },
  { text: '' },
  { text: '▶ Webhook received — payments-api (main)', className: 'text-[#b0b0b0]' },
  { text: '⏳ Building with Buildpacks...', className: 'text-[#f59e0b]' },
  { text: '✓ Image built: payments-api:a1b2c3d', className: 'text-[#10b981]' },
  { text: '⏳ Deploying to production (Kubernetes)...', className: 'text-[#f59e0b]' },
  { text: '✓ Deployed. Healthy. 3/3 pods ready.', className: 'text-[#10b981]' },
];

const buildModes = ['Buildpacks', 'Dockerfile', 'Pre-built Images'];
const deployTargets = ['Kubernetes', 'ECS', 'Cloud Run', 'Cloudflare Workers'];

export const ServiceHubHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Service Hub</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Ship code from Git to production
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Deploying a microservice shouldn&apos;t require a PhD in Kubernetes. But today,
            shipping a backend service means wrestling with Dockerfiles, Helm charts, CI pipelines,
            ingress configs, and secret management.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Service Hub: connect your Git repo, push code, and Planton builds, deploys, and
            manages your service across any environment. Vercel for Backend, In Your Own Cloud.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Deploy Your First Service</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="git push → production"
            lineDelay={350}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>

        <Box className="flex flex-wrap justify-center gap-6 mt-8">
          <Box className="flex items-center gap-2">
            <Typography component="span" className="text-[10px] text-[#555] uppercase font-semibold tracking-wider">Build</Typography>
            {buildModes.map((m) => (
              <Typography key={m} component="span" className="text-xs text-[#888] px-2.5 py-1 rounded-md border border-[#1a1a1a] bg-[#111]">
                {m}
              </Typography>
            ))}
          </Box>
          <Box className="flex items-center gap-2">
            <Typography component="span" className="text-[10px] text-[#555] uppercase font-semibold tracking-wider">Deploy</Typography>
            {deployTargets.map((t) => (
              <Typography key={t} component="span" className="text-xs text-[#888] px-2.5 py-1 rounded-md border border-[#1a1a1a] bg-[#111]">
                {t}
              </Typography>
            ))}
          </Box>
        </Box>
      </Box>
    </Section>
  );
};
