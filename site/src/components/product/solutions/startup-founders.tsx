'use client';

import { Box, Stack } from '@mui/material';
import Link from 'next/link';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  BodyText,
  PrimaryButton,
  SecondaryButton,
  FeatureCard,
  Badge,
  ArrowRightIcon,
  RocketIcon,
  ShieldIcon,
  CodeIcon,
} from '@/components/marketing';
import {
  FlowSteps,
  AnimatedTerminal,
  ScrollReveal,
} from '@/components/product/shared';
import type { FlowStep, TerminalLine } from '@/components/product/shared';
import {
  PersonAdd,
  Cloud,
  Tune,
  Storage,
  RocketLaunch,
} from '@mui/icons-material';

const onboardingSteps: FlowStep[] = [
  { icon: <PersonAdd fontSize="small" />, label: 'Sign Up', sublabel: 'Free, no credit card' },
  { icon: <Cloud fontSize="small" />, label: 'Connect Cloud', sublabel: 'AWS, GCP, or Azure' },
  { icon: <Tune fontSize="small" />, label: 'Pick Preset', sublabel: 'Production-ready config' },
  { icon: <Storage fontSize="small" />, label: 'Deploy Infra', sublabel: 'Database, cache, cluster' },
  { icon: <RocketLaunch fontSize="small" />, label: 'Ship Code', sublabel: 'Git push → production' },
];

const deployTerminalLines: TerminalLine[] = [
  { text: '$ planton apply -f postgres.yaml', className: 'text-white' },
  { text: '' },
  { text: '▶ Applying GcpCloudSql/my-postgres', className: 'text-[#b0b0b0]' },
  { text: '  Organization: my-startup', className: 'text-[#666]' },
  { text: '  Environment:  production', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Provisioning resources...', className: 'text-[#f59e0b]' },
  { text: '✓ PostgreSQL database created in 2m 18s.', className: 'text-[#10b981]' },
  { text: '' },
  { text: '  Connection string → planton secret get my-postgres-conn-url', className: 'text-[#666]' },
];

const valueProps = [
  {
    title: 'Production in an Afternoon',
    description:
      'Infra Hub presets and Service Hub get your backend running in hours, not weeks. Kubernetes, databases, CI/CD — all configured out of the box.',
    icon: <RocketIcon />,
  },
  {
    title: 'Free Tier',
    description:
      'Free forever for small teams, with unlimited automation runs. Get your MVP deployed and serving customers before you spend a dollar on the platform.',
    icon: <RocketIcon />,
  },
  {
    title: 'No Lock-In',
    description:
      'Planton open source produces portable manifests. Your infrastructure runs on standard Kubernetes and Terraform. Leave any time — your infra comes with you.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Self-Serve From Day One',
    description:
      'Your engineers can provision production infrastructure without Kubernetes or Terraform expertise. Save that headcount for product engineers.',
    icon: <CodeIcon />,
  },
  {
    title: 'Transparent Pricing',
    description:
      'Scale from free to transparent per-seat pricing. No surprise bills. No per-resource charges that spike when you get your first traffic burst.',
    icon: <RocketIcon />,
  },
  {
    title: 'Enterprise-Ready When You Get There',
    description:
      'Start with the free tier. When you raise your Series A and need RBAC, audit trails, and control-posture evidence — Planton already has them.',
    icon: <ShieldIcon />,
  },
];

export default function StartupFoundersPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>For Startup Founders</Badge>
          <SectionTitle>Your CTO&apos;s first infrastructure decision</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You need production infrastructure but every option seems wrong. PaaS platforms lock you
            in. IaC tools need an ops engineer. Cloud consoles are a maze. You want to focus on your
            product, not your infrastructure.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Planton gets you to production in an afternoon. Free tier. Open-source foundation means
            you can always leave. Scale pricing as your company grows.
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

      {/* Onboarding Flow */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-8">
            <SectionTitle>From sign-up to production in five steps</SectionTitle>
            <SectionSubtitle className="mx-auto">
              Production infrastructure without deep cloud expertise on the team.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <FlowSteps steps={onboardingSteps} />
        </ScrollReveal>
      </Section>

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          From zero to production, without the overhead
        </SectionTitle>
        <Box className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {valueProps.map((vp) => (
            <FeatureCard
              key={vp.title}
              icon={vp.icon}
              title={vp.title}
              description={vp.description}
            />
          ))}
        </Box>
      </Section>

      {/* Terminal Demo */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-8">
            <SectionTitle>See it in action</SectionTitle>
            <SectionSubtitle className="mx-auto">
              One command to provision a production database. Connection string ready in minutes.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <AnimatedTerminal
            lines={deployTerminalLines}
            title="planton apply"
            lineDelay={350}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to launch?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. No credit card. Get your product to production this afternoon.
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
