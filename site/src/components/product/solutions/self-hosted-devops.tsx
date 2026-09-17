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
} from '@/components/marketing';

interface CapabilityCardProps {
  title: string;
  description: string;
  details: string[];
  modules: string[];
}

const CapabilityCard = ({ title, description, details, modules }: CapabilityCardProps) => (
  <Card className="h-full">
    <FeatureTitle className="mb-2">{title}</FeatureTitle>
    <BodyText className="mb-3">{description}</BodyText>
    <Box className="space-y-1.5 mb-4">
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
  </Card>
);

const capabilities: CapabilityCardProps[] = [
  {
    title: 'Split Architecture',
    description:
      'Control plane (SaaS) handles orchestration, UI, and workflow management. Runner (your cloud) handles execution with your credentials.',
    details: [
      'SaaS control plane manages state, scheduling, and user interface',
      'Runner executes IaC, builds, and deployments inside your VPC',
      'Clean boundary — orchestration logic and sensitive execution never share a runtime',
    ],
    modules: ['Runner'],
  },
  {
    title: 'Zero Credential Exposure',
    description:
      'Runner uses IRSA, Workload Identity, or Managed Identity. No long-lived credentials cross the boundary between your cloud and the control plane.',
    details: [
      'Cloud-native identity federation — no static keys to rotate',
      'Credentials resolved at execution time inside your cloud boundary',
      'Control plane never sees, stores, or proxies your cloud credentials',
      'Secret backend options: platform Vault, bring-your-own Vault or OpenBAO, or your cloud\'s own secret manager — values stored native in your store',
    ],
    modules: ['Runner', 'Security'],
  },
  {
    title: 'Encrypted Tunnel',
    description:
      'Outbound-only connection from Runner to the control plane. No inbound firewall rules required. Cryptographic identity verification on every connection.',
    details: [
      'Runner initiates all connections — no open ports in your network',
      'Encrypted end-to-end with verified identity on both sides',
      'Automatic credential rotation and renewal',
    ],
    modules: ['Runner'],
  },
  {
    title: 'Compliance',
    description:
      'Full audit trail, RBAC, and secrets encrypted at rest. Designed for organizations with strict security, data residency, and regulatory requirements.',
    details: [
      'Every operation logged with actor, timestamp, and resource context',
      'Fine-grained, relationship-based access control',
      'Data residency — execution happens in your cloud region of choice',
    ],
    modules: ['Security'],
  },
  {
    title: 'Planton Operator',
    description:
      'For teams that need full self-hosting, the Planton Operator runs the entire platform on your Kubernetes cluster — no external dependencies.',
    details: [
      'Single Helm chart installs the complete Planton control plane',
      'All data stays within your cluster and network boundary',
      'Same API, same CLI, same console — just fully self-hosted',
    ],
    modules: ['Planton Operator'],
  },
];

const architectureLayers = [
  {
    label: 'Level 1 — Encrypted Secret Backends',
    items: ['Platform-provided Vault', 'Bring Your Own Vault / OpenBAO', 'AWS / GCP / Azure Secret Managers', 'Native values in your store'],
    border: 'border-[#3a3a3a]',
  },
  {
    label: 'Level 2 — Runner Execution Boundary',
    items: ['Outbound-only Connection', 'Cloud-native IAM (IRSA / Workload Identity)', 'Credentials Resolved Locally', 'No Inbound Firewall Rules'],
    border: 'border-[#2a2a2a]',
  },
  {
    label: 'Level 3 — Full Self-Hosted Control Plane',
    items: ['Single kubectl apply Manifest', 'Kubernetes Operator', 'Air-gapped Support', 'Complete Data Sovereignty'],
    border: 'border-[#3a3a3a]',
  },
];

export const SelfHostedDevOps = () => {
  return (
    <>
      {/* Hero */}
      <Section className="pt-24 md:pt-32">
        <Box className="max-w-4xl mx-auto text-center">
          <Badge className="mb-6">Use Case</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Enterprise security with SaaS convenience
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            You need the ease of a managed platform but your security requirements prohibit sending
            cloud credentials to a third party. Fully self-hosted solutions require managing the
            entire control plane yourself.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Planton splits orchestration (SaaS) from execution (your cloud). Runner executes IaC
            and operations in your VPC with your cloud provider&apos;s native IAM. The control plane
            never touches your credentials.
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

      {/* Architecture Diagram */}
      <Section>
        <Box className="text-center mb-10">
          <SectionTitle>How the split architecture works</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Three levels of security — choose the posture that fits your requirements.
          </SectionSubtitle>
        </Box>
        <Box className="max-w-2xl mx-auto space-y-4">
          {architectureLayers.map((layer, i) => (
            <Card key={layer.label} hover={false} className={`${layer.border}`}>
              <Box className="flex items-center gap-3 mb-3">
                <Box className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center flex-shrink-0">
                  <Typography className="text-sm font-semibold text-white">{i + 1}</Typography>
                </Box>
                <FeatureTitle>{layer.label}</FeatureTitle>
              </Box>
              <Box className="flex flex-wrap gap-2">
                {layer.items.map((item) => (
                  <Box
                    key={item}
                    className="px-3 py-1.5 rounded-md bg-white/5 border border-[#2a2a2a] text-sm text-[#b0b0b0]"
                  >
                    {item}
                  </Box>
                ))}
              </Box>
            </Card>
          ))}
        </Box>
      </Section>

      {/* Capabilities */}
      <Section>
        <Box className="text-center mb-12">
          <SectionTitle>Security without compromise</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Every layer designed for teams with strict security and compliance requirements.
          </SectionSubtitle>
        </Box>
        <Grid cols={3}>
          {capabilities.slice(0, 3).map((cap) => (
            <CapabilityCard key={cap.title} {...cap} />
          ))}
        </Grid>
        <Box className="mt-6">
          <Grid cols={2}>
            {capabilities.slice(3).map((cap) => (
              <CapabilityCard key={cap.title} {...cap} />
            ))}
          </Grid>
        </Box>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
          <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
            Your cloud, your rules
          </Typography>
          <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
            Keep credentials in your boundary
          </Typography>
          <BodyText className="!text-base mx-auto max-w-xl mb-8">
            Get the convenience of a managed platform with the security posture your organization
            requires. Start with the SaaS control plane or go fully self-hosted.
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
