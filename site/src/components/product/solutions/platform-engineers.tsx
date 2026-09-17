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
  FeatureCard,
  Card,
  Badge,
  ArrowRightIcon,
  ShieldIcon,
  CodeIcon,
  CpuIcon,
  GitBranchIcon,
  RocketIcon,
} from '@/components/marketing';
import { CodeTabs, ScrollReveal } from '@/components/product/shared';
import type { CodeTab } from '@/components/product/shared';

const goldenPathTabs: CodeTab[] = [
  {
    label: 'Chart.yaml',
    code: `apiVersion: infra-hub.planton.ai/v1
kind: InfraChart
metadata:
  name: Production Backend Stack
spec:
  selector:
    kind: platform
  description: >
    VPC, ALB, ECS Fargate, Aurora PostgreSQL,
    ElastiCache Redis, and SQS — production-ready.`,
  },
  {
    label: 'values.yaml',
    code: `params:
  - name: environment
    description: Target environment
    value: production

  - name: database_enabled
    description: Create Aurora PostgreSQL cluster
    type: bool
    value: true

  - name: cache_enabled
    description: Create ElastiCache Redis
    type: bool
    value: true`,
  },
  {
    label: 'templates/compute.yaml',
    code: `apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsCluster
metadata:
  name: "{{ values.environment }}-ecs"
  group: compute
spec:
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT`,
  },
];

const defineItems = [
  'Infra Charts with approved patterns',
  'Presets for each resource type',
  'Custom modules via private Git repos',
  'RBAC rules and credential scopes',
  'Environment-scoped connection defaults',
];

const selfServeItems = [
  'Pick a preset, fill in basics, deploy',
  'Provision databases and caches in minutes',
  'Push code, get production deployment',
  'Access logs and shells without kubeconfig',
  'Fetch env vars for local development',
];

const valueProps = [
  {
    title: 'Define Golden Paths',
    description:
      'Infra Charts and presets let you encode your best practices into reusable templates. Developers follow the path you designed — no drift.',
    icon: <CodeIcon />,
  },
  {
    title: 'Credential Governance',
    description:
      'Environment defaults and RBAC mean cloud credentials are provisioned automatically with the right scope. No more manual access management.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Managed CI/CD',
    description:
      'Service Hub runs managed pipelines so developers deploy without asking you. You define the pipeline; they push code.',
    icon: <RocketIcon />,
  },
  {
    title: 'Security Boundary',
    description:
      'Runner executes in your cloud. Credentials stay inside your VPC. Your security posture doesn\'t depend on a third-party SaaS.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Agent Fleet for Operational Intelligence',
    description:
      'AI agents surface anomalies, suggest optimizations, and handle routine operational tasks so your team can focus on platform work.',
    icon: <CpuIcon />,
  },
  {
    title: 'Open Source Modules',
    description:
      'Planton\'s modules are open source. Inspect every line, customize any module, contribute back. No black boxes in your infrastructure.',
    icon: <GitBranchIcon />,
  },
];

export default function PlatformEngineersPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>For Platform Engineers</Badge>
          <SectionTitle>Build golden paths, not bottleneck queues</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You&apos;re the bottleneck. Every infrastructure request goes through you. Every
            deployment issue gets escalated to you. You spend more time on tickets than on building
            the platform your org needs.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Infra Hub presets and Infra Charts let you define golden paths that developers follow.
            Centralized credential management and RBAC govern access without manual ticket queues. Service Hub handles CI/CD
            so developers don&apos;t need your help to deploy.
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

      {/* Define vs Self-Serve */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-10">
            <SectionTitle>You define the guardrails. Developers self-serve.</SectionTitle>
            <SectionSubtitle className="mx-auto">
              Platform engineers set standards and govern credentials. Developers provision and deploy within those boundaries.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <Box className="flex flex-col md:flex-row gap-6 max-w-4xl mx-auto">
            <Card hover={false} className="flex-1">
              <FeatureTitle className="mb-4">You Define</FeatureTitle>
              <Box className="space-y-2.5">
                {defineItems.map((item) => (
                  <Box key={item} className="flex gap-2.5 items-start">
                    <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                    <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{item}</Typography>
                  </Box>
                ))}
              </Box>
            </Card>
            <Card hover={false} className="flex-1">
              <FeatureTitle className="mb-4">Developers Self-Serve</FeatureTitle>
              <Box className="space-y-2.5">
                {selfServeItems.map((item) => (
                  <Box key={item} className="flex gap-2.5 items-start">
                    <Box className="w-1.5 h-1.5 rounded-full bg-[#10b981]/50 mt-2 flex-shrink-0" />
                    <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{item}</Typography>
                  </Box>
                ))}
              </Box>
            </Card>
          </Box>
        </ScrollReveal>
      </Section>

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          Automate the toil, own the platform
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

      {/* Golden Path Code Example */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-8">
            <SectionTitle>Encode your standards as Infra Charts</SectionTitle>
            <SectionSubtitle className="mx-auto">
              Parameterized templates that package your approved infrastructure patterns.
              Developers deploy with one click &mdash; your standards are built in.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <Box className="flex flex-col lg:flex-row gap-6 items-start">
            <Box className="flex-1 w-full">
              <CodeTabs tabs={goldenPathTabs} title="Infra Chart — Production Backend" />
            </Box>
            <Box className="flex-1 space-y-4">
              {[
                {
                  title: 'Bring Your Own Modules',
                  desc: 'Connect your private Git repo with custom Terraform modules. Only contract: the variables file — Planton doesn\'t inspect the module contents.',
                },
                {
                  title: 'Dependency-Aware DAG',
                  desc: 'VPC first, then subnets, then the database. Resources deploy in the right sequence automatically.',
                },
                {
                  title: 'Environment Promotion',
                  desc: 'Same chart, different values. Promote from dev to staging to production with confidence.',
                },
                {
                  title: 'Community + Org Charts',
                  desc: 'Start from community charts in the public infra-charts repo, or create your own for org-specific patterns.',
                },
              ].map((item) => (
                <Box key={item.title} className="p-4 rounded-lg border border-[#2a2a2a] bg-[#111]">
                  <Typography className="text-sm font-semibold text-white mb-1">{item.title}</Typography>
                  <Typography className="text-xs text-[#888] leading-relaxed">{item.desc}</Typography>
                </Box>
              ))}
            </Box>
          </Box>
        </ScrollReveal>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to stop being the bottleneck?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. Build your internal developer platform in days, not quarters.
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
