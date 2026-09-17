'use client';

import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  BodyText,
  FeatureTitle,
  PrimaryButton,
  SecondaryButton,
  Card,
  Badge,
  Grid,
  CheckIcon,
  XIcon,
} from '@/components/marketing';
import { FlowSteps, ScrollReveal } from '@/components/product/shared';
import type { FlowStep } from '@/components/product/shared';
import { Cloud, Route, Person, Visibility } from '@mui/icons-material';

interface CapabilityProps {
  title: string;
  description: string;
  modules: string[];
}

const Capability = ({ title, description, modules }: CapabilityProps) => (
  <Card className="h-full">
    <FeatureTitle className="mb-2">{title}</FeatureTitle>
    <BodyText className="mb-4">{description}</BodyText>
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
  </Card>
);

const capabilities: CapabilityProps[] = [
  {
    title: 'Self-Service Infrastructure',
    description:
      'Infra Hub presets and the creation wizard let developers provision databases, caches, queues, and clusters without filing tickets or waiting for ops.',
    modules: ['Infra Hub', 'Open Source'],
  },
  {
    title: 'Golden Paths',
    description:
      'Infra Charts and Service templates encode your organization\'s best practices into reusable, versioned blueprints that every team inherits.',
    modules: ['Infra Hub', 'Service Hub'],
  },
  {
    title: 'Managed CI/CD',
    description:
      'Service Hub handles build and deploy with managed pipelines — no YAML in every repo, no Jenkins to maintain, no GitHub Actions sprawl.',
    modules: ['Service Hub', 'Runner'],
  },
  {
    title: 'Access Control',
    description:
      'Centralized credential management for cloud integrations, with RBAC, secrets encryption, and audit trails — without manual permission spreadsheets.',
    modules: ['Security'],
  },
  {
    title: 'AI Assistance',
    description:
      'Agent Fleet provides AI agents that understand your infrastructure context — diagnose failures, suggest configurations, and execute operations.',
    modules: ['Agent Fleet'],
  },
  {
    title: 'Developer Experience',
    description:
      'Console UI, CLI, and direct pod access — developers choose the interface that fits their workflow. One platform, multiple entry points.',
    modules: ['CLI', 'Console'],
  },
];

interface ComparisonRow {
  feature: string;
  backstage: boolean | 'partial';
  planton: boolean | 'partial';
}

const comparisonRows: ComparisonRow[] = [
  { feature: 'Service catalog', backstage: true, planton: true },
  { feature: 'Infrastructure provisioning', backstage: 'partial', planton: true },
  { feature: 'Built-in CI/CD', backstage: false, planton: true },
  { feature: 'Credential management', backstage: false, planton: true },
  { feature: 'Multi-cloud support', backstage: 'partial', planton: true },
  { feature: 'Works out of the box', backstage: false, planton: true },
  { feature: 'No plugin maintenance', backstage: false, planton: true },
  { feature: 'AI-powered operations', backstage: false, planton: true },
];

const StatusCell = ({ value }: { value: boolean | 'partial' }) => {
  if (value === true) return <CheckIcon />;
  if (value === false) return <XIcon />;
  return <Typography className="text-xs text-[#f59e0b]">Partial</Typography>;
};

const flowSteps: FlowStep[] = [
  { icon: <Cloud fontSize="small" />, label: 'Connect Cloud', sublabel: 'Provider credentials' },
  { icon: <Route fontSize="small" />, label: 'Define Golden Paths', sublabel: 'Infra Charts & presets' },
  { icon: <Person fontSize="small" />, label: 'Developers Self-Serve', sublabel: 'Provision & deploy' },
  { icon: <Visibility fontSize="small" />, label: 'Track Everything', sublabel: 'Audit trail & RBAC' },
];

export const InternalDeveloperPlatform = () => {
  return (
    <>
      {/* Hero */}
      <Section className="pt-24 md:pt-32">
        <Box className="max-w-4xl mx-auto text-center">
          <Badge className="mb-6">Use Case</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Build an IDP without building an IDP
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Platform teams spend months building internal developer platforms from scratch with
            Backstage, Terraform modules, and custom tooling. Developers still wait for ops.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Planton gives you an IDP out of the box. Self-service infrastructure, managed CI/CD,
            controlled access, and AI assistance — integrated, ready to use, and maintained.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Start Free</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>
      </Section>

      {/* Flow Steps */}
      <Section>
        <ScrollReveal>
          <FlowSteps steps={flowSteps} />
        </ScrollReveal>
      </Section>

      {/* Capabilities */}
      <Section>
        <Box className="text-center mb-12">
          <SectionTitle>Everything an IDP needs, nothing to build</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Six integrated capabilities that replace months of custom platform engineering.
          </SectionSubtitle>
        </Box>
        <Grid cols={3}>
          {capabilities.map((cap) => (
            <Capability key={cap.title} {...cap} />
          ))}
        </Grid>
      </Section>

      {/* Comparison */}
      <Section>
        <Box className="text-center mb-12">
          <SectionTitle>Backstage + custom tooling vs. Planton</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Backstage is a framework — you still need to build the platform. Planton is the platform.
          </SectionSubtitle>
        </Box>
        <Card hover={false} className="overflow-hidden !p-0 max-w-2xl mx-auto">
          <Box className="grid grid-cols-3 gap-0">
            {/* Header */}
            <Box className="p-4 border-b border-[#2a2a2a] bg-[#1a1a1a]">
              <Typography className="text-sm font-semibold text-[#666]">Feature</Typography>
            </Box>
            <Box className="p-4 border-b border-l border-[#2a2a2a] bg-[#1a1a1a] text-center">
              <Typography className="text-sm font-semibold text-[#a0a0a0]">Backstage + DIY</Typography>
            </Box>
            <Box className="p-4 border-b border-l border-[#2a2a2a] bg-[#1a1a1a] text-center">
              <Typography className="text-sm font-semibold text-white">Planton</Typography>
            </Box>
            {/* Rows */}
            {comparisonRows.map((row) => (
              <Box key={row.feature} className="contents">
                <Box className="p-4 border-b border-[#2a2a2a]">
                  <Typography className="text-sm text-[#b0b0b0]">{row.feature}</Typography>
                </Box>
                <Box className="p-4 border-b border-l border-[#2a2a2a] flex justify-center">
                  <StatusCell value={row.backstage} />
                </Box>
                <Box className="p-4 border-b border-l border-[#2a2a2a] flex justify-center">
                  <StatusCell value={row.planton} />
                </Box>
              </Box>
            ))}
          </Box>
        </Card>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
          <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
            Ready to ship
          </Typography>
          <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
            Your IDP, ready in minutes
          </Typography>
          <BodyText className="!text-base mx-auto max-w-xl mb-8">
            Stop building platform tooling. Start shipping the infrastructure your developers
            actually need.
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
      </Section>
    </>
  );
};
