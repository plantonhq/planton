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
  CodeIcon,
  ShieldIcon,
} from '@/components/marketing';
import {
  FlowSteps,
  AnimatedTerminal,
  ScrollReveal,
} from '@/components/product/shared';
import type { FlowStep, TerminalLine } from '@/components/product/shared';
import {
  Code as PushIcon,
  Build as BuildIcon,
  CloudDone as DeployIcon,
  Terminal as LogsIcon,
  BugReport as DebugIcon,
  Settings as ConfigIcon,
  AutoAwesome as SimpleIcon,
} from '@mui/icons-material';

const workflowSteps: FlowStep[] = [
  { icon: <PushIcon fontSize="small" />, label: 'Push Code', sublabel: 'Git commit triggers pipeline' },
  { icon: <BuildIcon fontSize="small" />, label: 'Auto Build', sublabel: 'Container image via Buildpacks' },
  { icon: <DeployIcon fontSize="small" />, label: 'Deploy', sublabel: 'Rollout to Kubernetes or ECS' },
  { icon: <LogsIcon fontSize="small" />, label: 'Access Logs', sublabel: 'Pod shell, logs, port-forward' },
];

const deployTerminalLines: TerminalLine[] = [
  { text: '$ git push origin main', className: 'text-white' },
  { text: '' },
  { text: '▶ Pipeline triggered for acme-backend', className: 'text-[#b0b0b0]' },
  { text: '  Commit: "Add user profile endpoint"', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Building container image...', className: 'text-[#f59e0b]' },
  { text: '  Detected: Go 1.22 via Cloud Native Buildpacks', className: 'text-[#666]' },
  { text: '✓ Image pushed to registry in 48s.', className: 'text-[#10b981]' },
  { text: '' },
  { text: '⏳ Deploying to staging...', className: 'text-[#f59e0b]' },
  { text: '✓ Rollout complete. 3/3 pods ready.', className: 'text-[#10b981]' },
  { text: '' },
  { text: '  Endpoint → https://staging.acme-backend.example.com', className: 'text-[#666]' },
  { text: '  Logs     → planton service logs acme-backend', className: 'text-[#666]' },
];

const valueProps = [
  {
    title: 'Push Code, Get a Deployment',
    description:
      'Service Hub connects to your Git repo and handles container builds, rollouts, and rollbacks. Your job ends at git push.',
    icon: <CodeIcon />,
  },
  {
    title: 'Provision Infrastructure Without Tickets',
    description:
      'Need a Postgres database? Infra Hub presets let you spin one up in minutes. No Jira ticket, no waiting on the ops team.',
    icon: <RocketIcon />,
  },
  {
    title: 'Access Logs and Containers',
    description:
      'Pod access gives you logs, shell, and port-forwarding without kubeconfig files. Debug production issues from the Planton console.',
    icon: <DebugIcon />,
  },
  {
    title: 'Local Dev Config',
    description:
      'CLI dot-env export pulls your environment variables into a local .env file. Develop against real config without copy-pasting from dashboards.',
    icon: <ConfigIcon />,
  },
  {
    title: 'Environment Variables & Secrets',
    description:
      'VariablesGroup and SecretsGroup let you manage config and secrets in one place. Update a secret, deploy picks it up automatically.',
    icon: <ShieldIcon />,
  },
  {
    title: 'No Kubernetes Expertise Required',
    description:
      'You don\'t need to learn Helm, Kustomize, or kubectl. Planton abstracts the complexity so you can focus on your application code.',
    icon: <SimpleIcon />,
  },
];

export default function DevelopersPage() {
  return (
    <Box>
      {/* Hero */}
      <Section className="pt-20 md:pt-28">
        <Stack className="items-center text-center gap-5 max-w-3xl mx-auto">
          <Badge>For Developers</Badge>
          <SectionTitle>Deploy your code, not your weekend</SectionTitle>
          <BodyText className="max-w-2xl text-center">
            You just want to ship code. But deploying a backend service means learning Kubernetes,
            writing Helm charts, configuring CI pipelines, and waiting for the ops team to provision
            a database. You didn&apos;t sign up for this.
          </BodyText>
          <SectionSubtitle className="max-w-2xl text-center">
            Service Hub turns your Git push into a deployment. Infra Hub lets you provision a database
            in minutes. CLI gives you logs and env vars without kubeconfig files.
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

      {/* Workflow */}
      <Section>
        <ScrollReveal>
          <Box className="text-center mb-8">
            <SectionTitle>Your workflow with Planton</SectionTitle>
            <SectionSubtitle className="mx-auto">
              From git push to production logs &mdash; four steps, zero ops tickets.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <FlowSteps steps={workflowSteps} />
        </ScrollReveal>
      </Section>

      {/* Value Propositions */}
      <Section>
        <SectionTitle className="text-center mb-8">
          Focus on your code, not the plumbing
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
              Push a commit. Service Hub builds, deploys, and gives you the endpoint.
            </SectionSubtitle>
          </Box>
        </ScrollReveal>
        <ScrollReveal delay={0.15}>
          <AnimatedTerminal
            lines={deployTerminalLines}
            title="developer workflow"
            lineDelay={300}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Section>

      {/* Bottom CTA */}
      <Section>
        <Stack className="items-center text-center gap-5">
          <SectionTitle>Ready to just ship code?</SectionTitle>
          <SectionSubtitle className="text-center max-w-xl">
            Free tier. No credit card. Deploy your first service in under 10 minutes.
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
