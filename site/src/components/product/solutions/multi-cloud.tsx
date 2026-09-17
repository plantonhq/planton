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
  TerminalWindow,
} from '@/components/marketing';

interface CapabilitySectionProps {
  number: string;
  title: string;
  description: string;
  details: string[];
  modules: string[];
}

const CapabilityBlock = ({ number, title, description, details, modules }: CapabilitySectionProps) => (
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

const capabilities: CapabilitySectionProps[] = [
  {
    number: '01',
    title: 'One Manifest Format',
    description:
      'KRM YAML manifests with provider-specific specs. Not a generic abstraction that hides capabilities — full access to each cloud\'s features through typed specifications.',
    details: [
      'Same apiVersion/kind/metadata/spec structure for every resource',
      'Provider-specific spec fields — GCP, AWS, and Azure each get their native options',
      'Validated at write time, not at apply time',
    ],
    modules: ['Open Source', 'CLI'],
  },
  {
    number: '02',
    title: 'Centralized Credentials',
    description:
      'One place to manage cloud integrations and credentials across providers. Switch clouds without switching credential workflows or learning new secret management patterns.',
    details: [
      'AWS IAM roles, GCP service accounts, and Azure service principals in one interface',
      'Environment-scoped defaults — dev uses one account, prod uses another',
      'Secrets encrypted at rest, resolved only at execution time',
    ],
    modules: ['Security'],
  },
  {
    number: '03',
    title: 'Same Workflow',
    description:
      'planton apply works the same whether you\'re deploying GCP Cloud SQL, AWS RDS, or Azure SQL Database. One command, one review flow, one audit trail.',
    details: [
      'planton apply / planton destroy — same verbs, every provider',
      'Preview changes before applying, regardless of target cloud',
      'Consistent state management across all providers',
    ],
    modules: ['CLI', 'Infra Hub'],
  },
  {
    number: '04',
    title: 'Provider-Specific Where It Matters',
    description:
      'Planton open source gives you full access to each provider\'s features through typed specifications — not a limited intersection that forces you into the lowest common denominator.',
    details: [
      'GCP-specific fields for Cloud Run, GKE, and Cloud SQL',
      'AWS-specific fields for ECS, EKS, and RDS',
      'No capability hiding — if the provider supports it, you can configure it',
    ],
    modules: ['Open Source'],
  },
  {
    number: '05',
    title: 'Multi-Cloud Pipelines',
    description:
      'Infra Charts can template resources across providers in a single DAG, letting you orchestrate cross-cloud deployments from one definition.',
    details: [
      'Provision GCP networking and AWS compute in a single chart',
      'Dependency ordering across providers',
      'Reusable templates for common multi-cloud patterns',
    ],
    modules: ['Infra Hub', 'Runner'],
  },
];

export const MultiCloud = () => {
  return (
    <>
      {/* Hero */}
      <Section className="pt-24 md:pt-32">
        <Box className="max-w-4xl mx-auto text-center">
          <Badge className="mb-6">Use Case</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Same workflow, every cloud
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Multi-cloud means multi-everything: different CLIs, different credential formats,
            different state management, different modules. Your team context-switches between
            AWS, GCP, and Azure tooling all day.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            One YAML manifest format, one CLI, one console — whether deploying to AWS, GCP, Azure,
            or all three. Planton open source provides provider-specific specs, not lowest-common-denominator
            abstraction.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center mb-12">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Start Free</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <Box className="max-w-3xl mx-auto">
          <TerminalWindow title="planton cli">
            <Box className="space-y-1">
              <Typography className="text-[#a0a0a0] text-sm font-mono">
                <span className="text-[#10b981]">$</span> planton apply -f gcp-cloud-sql.yaml
              </Typography>
              <Typography className="text-[#666] text-sm font-mono">
                ✓ GCP Cloud SQL instance provisioned
              </Typography>
              <Typography className="text-[#a0a0a0] text-sm font-mono mt-3">
                <span className="text-[#10b981]">$</span> planton apply -f aws-rds.yaml
              </Typography>
              <Typography className="text-[#666] text-sm font-mono">
                ✓ AWS RDS instance provisioned
              </Typography>
              <Typography className="text-[#a0a0a0] text-sm font-mono mt-3">
                <span className="text-[#10b981]">$</span> planton apply -f azure-sql.yaml
              </Typography>
              <Typography className="text-[#666] text-sm font-mono">
                ✓ Azure SQL Database provisioned
              </Typography>
            </Box>
          </TerminalWindow>
        </Box>
      </Section>

      {/* Capabilities */}
      <Section>
        <Box className="text-center mb-12">
          <SectionTitle>How Planton unifies multi-cloud</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Five capabilities that eliminate cloud-specific context switching.
          </SectionSubtitle>
        </Box>
        <Box className="space-y-12 max-w-3xl mx-auto">
          {capabilities.map((cap) => (
            <CapabilityBlock key={cap.number} {...cap} />
          ))}
        </Box>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
          <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-4">
            One platform, every cloud
          </Typography>
          <Typography className="text-2xl md:text-3xl font-semibold text-white mb-4">
            Stop context-switching between clouds
          </Typography>
          <BodyText className="!text-base mx-auto max-w-xl mb-8">
            Deploy to AWS, GCP, and Azure with the same manifests, the same CLI,
            and the same credential workflow.
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
