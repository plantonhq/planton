'use client';

import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  BodyText,
  FeatureTitle,
  PrimaryButton,
  SecondaryButton,
  Badge,
  ArrowRightIcon,
} from '@/components/marketing';
import { MetricsStrip, ScrollReveal } from '@/components/product/shared';
import type { MetricItem } from '@/components/product/shared';
import { PLATFORM_STATS } from '@/data/platform-stats';

const metrics: MetricItem[] = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Resource Types' },
  { value: '17', label: 'Cloud Providers' },
  { value: 'Self-Service', label: 'Infrastructure' },
  { value: 'Full Audit', label: 'Trail' },
];

interface CapabilityBlockProps {
  number: string;
  title: string;
  description: string;
  details: string[];
  modules: string[];
}

const CapabilityBlock = ({ number, title, description, details, modules }: CapabilityBlockProps) => (
  <Box className="flex gap-5 items-start">
    <Box className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center flex-shrink-0 mt-0.5">
      <Typography className="text-sm font-semibold text-white">{number}</Typography>
    </Box>
    <Box className="flex-1">
      <FeatureTitle className="mb-2">{title}</FeatureTitle>
      <BodyText className="mb-3 !text-base">{description}</BodyText>
      <Box className="space-y-1.5 mb-3">
        {details.map((d, i) => (
          <Box key={i} className="flex gap-2 items-start">
            <Box className="w-1.5 h-1.5 rounded-full bg-[#3a3a3a] mt-2 flex-shrink-0" />
            <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{d}</Typography>
          </Box>
        ))}
      </Box>
      <Box className="flex flex-wrap gap-1.5">
        {modules.map((mod) => (
          <Box
            key={mod}
            className="px-2.5 py-1 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#a0a0a0]"
          >
            {mod}
          </Box>
        ))}
      </Box>
    </Box>
  </Box>
);

const capabilities: CapabilityBlockProps[] = [
  {
    number: '01',
    title: 'Self-Service at Scale',
    description:
      'Infra Hub creation wizard and presets let developers provision databases, caches, and clusters without filing tickets or waiting on the ops team.',
    details: [
      `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} resource types with production-ready presets`,
      'Creation wizard with live provider data',
      'Developers pick from approved templates — no drift',
    ],
    modules: ['Infra Hub', 'Open Source'],
  },
  {
    number: '02',
    title: 'Standards Without Bottlenecks',
    description:
      'Infra Charts encode your organization\'s best practices into reusable, versioned blueprints. Teams follow the standard path without waiting on a review board.',
    details: [
      'Parameterized templates for common infrastructure patterns',
      'Community charts or custom org-specific charts',
      'Environment promotion: dev → staging → production',
    ],
    modules: ['Infra Hub'],
  },
  {
    number: '03',
    title: 'Team Visibility',
    description:
      'Every infrastructure change is tracked with who, what, when, and the Git commit message. Leadership gets visibility without micromanagement.',
    details: [
      'Stack Jobs create full audit trail for every change',
      'Version history with diffs for every resource',
      'Centralized credential governance with RBAC',
    ],
    modules: ['Security'],
  },
  {
    number: '04',
    title: 'Multi-Environment Management',
    description:
      'Dev, staging, production — each with consistent configuration. Promote changes through environments with confidence and manage config per-environment.',
    details: [
      'Environment-scoped credential defaults',
      'Same Infra Chart, different values per environment',
      'Agent Fleet monitors health across all environments',
    ],
    modules: ['Infra Hub', 'Agent Fleet'],
  },
];

export default function GrowingTeamsPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>Growing Teams</Badge>
          <SectionTitle>Scale your infrastructure without scaling your ops team</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You&apos;ve grown from 10 to 50 engineers. Infrastructure requests are increasing faster
            than your ops team can handle. Self-service is no longer optional — it&apos;s survival.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Give developers self-service through Infra Hub presets and Service Hub pipelines. Enforce
            standards with Infra Charts and centralized credential management. Keep leadership visibility with audit trails.
          </SectionSubtitle>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-4 mt-4">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>
                Start Free
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Section>

      {/* Metrics */}
      <MetricsStrip metrics={metrics} />

      {/* Capabilities */}
      <Section>
        <SectionTitle className="text-center mb-4">
          Give your team autonomy without losing control
        </SectionTitle>
        <SectionSubtitle className="text-center mx-auto mb-12">
          Four capabilities that turn a growing team into a self-sufficient one.
        </SectionSubtitle>
        <Box className="space-y-12 max-w-3xl mx-auto">
          {capabilities.map((cap, i) => (
            <ScrollReveal key={cap.number} delay={i * 0.1}>
              <CapabilityBlock {...cap} />
            </ScrollReveal>
          ))}
        </Box>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to unblock your team?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. Self-service infrastructure for every developer on your team.
          </SectionSubtitle>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-4 mt-2">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>
                Start Free
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Section>
    </Box>
  );
}
