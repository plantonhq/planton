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
  CloudIcon,
  CpuIcon,
  CodeIcon,
  GitBranchIcon,
} from '@/components/marketing';
import { ScrollReveal } from '@/components/product/shared';

const valueProps = [
  {
    title: 'Runner Security Boundary',
    description:
      'Runner executes in your cloud. Credentials never leave your VPC. Your security team gets the isolation they require.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Compliance-Ready',
    description:
      'Audit trails, RBAC, and encrypted secrets give your compliance team the evidence they need for SOC 2, HIPAA, and ISO audits.',
    icon: <CodeIcon />,
  },
  {
    title: 'Multi-Cloud Governance',
    description:
      'Centralized credential management across AWS, GCP, and Azure. One governance layer across every cloud provider.',
    icon: <CloudIcon />,
  },
  {
    title: 'Change Management',
    description:
      'Pipeline approvals, DAG execution, and Git-backed state mean every infrastructure change is reviewable, reversible, and traceable.',
    icon: <GitBranchIcon />,
  },
  {
    title: 'Self-Hosted Option',
    description:
      'Planton Operator runs the entire platform inside your infrastructure. Air-gapped environments, on-prem data centers — fully supported.',
    icon: <CpuIcon />,
  },
  {
    title: 'Enterprise Support & SLAs',
    description:
      'Dedicated account management, priority support, and customizable SLAs to keep your operations running at enterprise scale.',
    icon: <CpuIcon />,
  },
];

interface SecurityLevelProps {
  level: number;
  title: string;
  description: string;
  module: string;
}

const securityLevels: SecurityLevelProps[] = [
  {
    level: 1,
    title: 'Encrypted Secret Backends',
    description:
      'Keep secrets in your own store — AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, HashiCorp Vault, or OpenBAO — as native values under readable names your own tools can consume. Resolved just-in-time during execution.',
    module: 'Security',
  },
  {
    level: 2,
    title: 'Runner in Your Cloud',
    description:
      'Deploy the Planton Runner in your VPC. Credentials never leave your cloud boundary. Outbound-only connection — no inbound firewall rules required.',
    module: 'Runner',
  },
  {
    level: 3,
    title: 'Full Self-Hosted Control Plane',
    description:
      'Deploy the entire Planton platform on your Kubernetes cluster. Single kubectl apply manifest. All data stays within your infrastructure. Air-gapped and on-prem supported.',
    module: 'Planton Operator',
  },
];

interface ComplianceRow {
  standard: string;
  mapping: string;
}

const complianceRows: ComplianceRow[] = [
  {
    standard: 'SOC 2',
    mapping: 'Full audit trail via Stack Jobs — every change tracked with actor, timestamp, and resource context',
  },
  {
    standard: 'HIPAA',
    mapping: 'Zero credential exposure via Runner — credentials never leave customer infrastructure',
  },
  {
    standard: 'PCI DSS',
    mapping: 'No plaintext credential storage — Connection specs store only secret references, never values',
  },
  {
    standard: 'Data Residency',
    mapping: 'Execution in your cloud region — Runner executes IaC using local credentials in your VPC',
  },
];

export default function EnterprisesPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>Enterprises</Badge>
          <SectionTitle>Enterprise controls without enterprise friction</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            Enterprise infrastructure means compliance requirements, security audits, multi-cloud
            mandates, and change management processes. Most platforms either don&apos;t meet security
            requirements or add so much friction that developers work around them.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Runner executes in your cloud — credentials never leave your VPC. Centralized credentials
            and IAM govern who can use them. Infra Hub Pipelines enforce change management. Security provides the
            audit trail compliance teams need.
          </SectionSubtitle>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-4 mt-4">
            <Link href="/book-demo">
              <PrimaryButton>
                Book a Demo
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="https://planton.ai/signup" target="_blank">
              <SecondaryButton>Start Free</SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Section>

      {/* 3-Level Security Posture */}
      <Section>
        <Box className="text-center mb-12">
          <SectionTitle>Three levels of security posture</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Choose the level that fits your requirements — from encrypted secrets to full self-hosting.
          </SectionSubtitle>
        </Box>
        <Box className="max-w-2xl mx-auto space-y-4">
          {securityLevels.map((sl, i) => (
            <ScrollReveal key={sl.level} delay={i * 0.12}>
              <Card hover={false}>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center flex-shrink-0">
                    <Typography className="text-sm font-semibold text-white">{sl.level}</Typography>
                  </Box>
                  <FeatureTitle>{sl.title}</FeatureTitle>
                </Box>
                <BodyText className="mb-3">{sl.description}</BodyText>
                <Box className="flex flex-wrap gap-1.5">
                  <Box className="px-2.5 py-1 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#a0a0a0]">
                    {sl.module}
                  </Box>
                </Box>
              </Card>
            </ScrollReveal>
          ))}
        </Box>
      </Section>

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          Security and compliance built into every layer
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

      {/* Compliance Mapping */}
      <Section>
        <Box className="text-center mb-10">
          <SectionTitle>Compliance mapping</SectionTitle>
          <SectionSubtitle className="mx-auto">
            How Planton capabilities map to common compliance requirements.
          </SectionSubtitle>
        </Box>
        <ScrollReveal>
          <Card hover={false} className="overflow-hidden !p-0 max-w-3xl mx-auto">
            <Box className="grid grid-cols-[140px_1fr] md:grid-cols-[180px_1fr]">
              <Box className="p-4 border-b border-[#2a2a2a] bg-[#1a1a1a]">
                <Typography className="text-sm font-semibold text-[#666]">Standard</Typography>
              </Box>
              <Box className="p-4 border-b border-l border-[#2a2a2a] bg-[#1a1a1a]">
                <Typography className="text-sm font-semibold text-[#666]">Planton Capability</Typography>
              </Box>
              {complianceRows.map((row) => (
                <Box key={row.standard} className="contents">
                  <Box className="p-4 border-b border-[#2a2a2a] flex items-center">
                    <Typography className="text-sm font-semibold text-white">{row.standard}</Typography>
                  </Box>
                  <Box className="p-4 border-b border-l border-[#2a2a2a] flex items-center">
                    <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{row.mapping}</Typography>
                  </Box>
                </Box>
              ))}
            </Box>
          </Card>
        </ScrollReveal>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to modernize your infrastructure platform?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Talk to our enterprise team. Custom SLAs, self-hosted deployment, and dedicated support.
          </SectionSubtitle>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-4 mt-2">
            <Link href="/book-demo">
              <PrimaryButton>
                Book a Demo
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="https://planton.ai/signup" target="_blank">
              <SecondaryButton>Start Free</SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Section>
    </Box>
  );
}
