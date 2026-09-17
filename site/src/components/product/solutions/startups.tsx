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
  RocketIcon,
  CodeIcon,
  ShieldIcon,
} from '@/components/marketing';
import {
  BentoGrid,
  BentoItem,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import {
  Storage,
  Code,
  Shield,
  Terminal,
  AttachMoney as PricingIcon,
  Groups as TeamIcon,
} from '@mui/icons-material';
import { PLATFORM_STATS } from '@/data/platform-stats';

const valueProps = [
  {
    title: 'Speed to Production',
    description:
      'Infra Hub presets give you production-ready Kubernetes, databases, and caches in minutes. Pick a preset, fill in the basics, deploy.',
    icon: <RocketIcon />,
  },
  {
    title: 'Git-to-Deploy',
    description:
      'Service Hub connects your Git repo and handles the rest — container builds, rollouts, rollbacks. Push code, get a deployment.',
    icon: <CodeIcon />,
  },
  {
    title: 'Free Tier + Transparent Pricing',
    description:
      'Start free — no credit card, no metered automation. Transparent per-seat pricing when you are ready to scale.',
    icon: <PricingIcon />,
  },
  {
    title: 'No Lock-In',
    description:
      'Planton open source means every infrastructure manifest is portable. If Planton disappears tomorrow, your infrastructure still works.',
    icon: <ShieldIcon />,
  },
  {
    title: 'CLI-First Workflow',
    description:
      'Manage infrastructure from your terminal. Create resources, fetch env vars, tail logs — no browser tab required.',
    icon: <Terminal />,
  },
  {
    title: 'Self-Serve From Day One',
    description:
      'Presets encode best practices so your engineers can provision production infrastructure without Kubernetes or Terraform expertise.',
    icon: <TeamIcon />,
  },
];

export default function StartupsPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>Startups</Badge>
          <SectionTitle>Ship production infrastructure without growing your ops team</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You&apos;re a startup with 2–10 engineers. You need production infrastructure but
            cannot justify a dedicated ops team yet. You&apos;re choosing between learning Terraform
            yourself, using a PaaS that limits you, or outsourcing to expensive contractors.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Planton gives you self-service infrastructure from day one. Free tier to start, scale as
            you grow. Open-source foundation means no lock-in.
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

      {/* Startup Capabilities Bento Grid */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-10">
            <SectionTitle>Everything a startup needs, nothing it doesn&apos;t</SectionTitle>
            <SectionSubtitle className="mx-auto">
              Production-grade infrastructure, CI/CD, security, and CLI &mdash; without the complexity.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>

        <StaggerContainer stagger={0.1}>
          <BentoGrid>
            <StaggerItem>
              <BentoItem span="wide">
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <Storage fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Infrastructure on Demand</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  {PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} resource types across {PLATFORM_STATS.CLOUD_PROVIDER_COUNT} cloud providers. Pick a preset, fill in the basics,
                  deploy. Kubernetes clusters, databases, caches, DNS &mdash; all from one interface.
                </BodyText>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <Code fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Git-to-Deploy CI/CD</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  Service Hub connects your Git repo and handles the rest &mdash; container builds,
                  rollouts, rollbacks. No pipeline YAML to maintain.
                </BodyText>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <Shield fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Security Built In</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  Encrypted secrets, no plaintext credential storage, full audit trails.
                  Enterprise-grade security from day one.
                </BodyText>
              </BentoItem>
            </StaggerItem>

            <StaggerItem>
              <BentoItem>
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <Terminal fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">CLI-First</FeatureTitle>
                </Box>
                <BodyText className="!text-sm">
                  Everything from your terminal: create resources, fetch env vars, tail logs,
                  manage deployments. No browser tab required.
                </BodyText>
              </BentoItem>
            </StaggerItem>
          </BentoGrid>
        </StaggerContainer>
      </Section>

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          Everything you need to go from zero to production
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

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to ship?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. No credit card. Deploy your first infrastructure in under an hour.
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
