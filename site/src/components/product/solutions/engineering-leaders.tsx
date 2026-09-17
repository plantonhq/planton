'use client';

import { Box, Stack } from '@mui/material';
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
  Badge,
  ArrowRightIcon,
  ShieldIcon,
  CpuIcon,
  CodeIcon,
} from '@/components/marketing';
import {
  MetricsStrip,
  BentoGrid,
  BentoItem,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { MetricItem } from '@/components/product/shared';
import {
  PlaylistAddCheck as AuditIcon,
  Security as AccessIcon,
  Bookmark as StandardsIcon,
  Psychology as IntelIcon,
} from '@mui/icons-material';
import { PLATFORM_STATS } from '@/data/platform-stats';

const metrics: MetricItem[] = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Resource Types' },
  { value: '17', label: 'Cloud Providers' },
  { value: 'Full Audit Trail', label: '' },
  { value: 'Zero Credential Exposure', label: '' },
];

const valueProps = [
  {
    title: 'Full Audit Trail',
    description:
      'Every infrastructure change is tracked with who, what, when, and the Git commit message. Compliance teams get the evidence they need.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Team Autonomy with Guardrails',
    description:
      'Presets and RBAC let developers self-serve within boundaries you define. Autonomy for them, governance for you.',
    icon: <CodeIcon />,
  },
  {
    title: 'Cost Visibility',
    description:
      'ROI calculator and usage tracking show infrastructure spend by team, project, and environment. Make data-driven resourcing decisions.',
    icon: <CpuIcon />,
  },
  {
    title: 'Compliance',
    description:
      'The Security module provides encrypted secrets, IAM governance, and audit-ready reports. Pass your next SOC 2 audit without scrambling.',
    icon: <ShieldIcon />,
  },
  {
    title: 'AI Operational Intelligence',
    description:
      'Agent Fleet monitors infrastructure health, surfaces anomalies, and provides operational insights so you can lead proactively.',
    icon: <CpuIcon />,
  },
  {
    title: 'Standardization Without Bottlenecks',
    description:
      'Infra Charts encode your architecture standards into reusable templates. Teams follow the standard path without waiting on a review board.',
    icon: <CodeIcon />,
  },
];

export default function EngineeringLeadersPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>For Engineering Leaders</Badge>
          <SectionTitle>Visibility without micromanagement</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You need to know what&apos;s deployed, who changed it, and whether the team is following
            standards. But you also need your engineers to be autonomous. Enforcing controls
            shouldn&apos;t mean creating bottlenecks.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Planton gives your team autonomy with guardrails. Audit trails show every change. RBAC
            ensures the right access. Infra Hub presets enforce standards without blocking developers.
            Agent Fleet handles operational intelligence.
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

      {/* Metrics Strip */}
      <MetricsStrip metrics={metrics} />

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          Lead with clarity, not control
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

      {/* Governance Bento Grid */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-10">
            <SectionTitle>Governance that scales with your team</SectionTitle>
            <SectionSubtitle className="mx-auto">
              Audit trails, access control, standards enforcement, and operational intelligence &mdash; built into the platform.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>

        <StaggerContainer stagger={0.1}>
          <BentoGrid>
            <StaggerItem>
              <BentoItem span="wide">
                <Box className="flex flex-col md:flex-row gap-5">
                  <Box className="flex-1">
                    <Box className="flex items-center gap-3 mb-3">
                      <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                        <AuditIcon fontSize="small" />
                      </Box>
                      <FeatureTitle className="!text-base">Audit Trail</FeatureTitle>
                    </Box>
                    <BodyText className="!text-sm">
                      Every infrastructure change creates a Stack Job &mdash; tracked with who
                      triggered it, what changed, when it happened, and the commit message.
                      Compliance teams get evidence without scrambling.
                    </BodyText>
                  </Box>
                  <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                    <pre className="whitespace-pre-wrap">
{`▶ Stack Job: sjb-7f2a1c
  Operation:  update
  Resource:   prod-postgres
  Triggered:  alice@company.com
  Commit:     "Enable HA for prod DB"

✓ 1 resource updated in 2m 34s.`}
                    </pre>
                  </Box>
                </Box>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <AccessIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Access Control</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  RBAC, connection-level credential governance, and environment-scoped defaults.
                  The right people get the right access to the right resources.
                </BodyText>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <StandardsIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Standards Enforcement</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  Infra Charts and presets encode your architectural standards into reusable
                  templates. Teams follow the standard path without waiting on a review board.
                </BodyText>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <IntelIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Operational Intelligence</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  Agent Fleet monitors infrastructure health, surfaces anomalies, and provides
                  actionable insights so you can lead proactively.
                </BodyText>
              </BentoItem>
            </StaggerItem>
          </BentoGrid>
        </StaggerContainer>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to lead with clarity?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. Give your team the platform they need — and the visibility you need.
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
