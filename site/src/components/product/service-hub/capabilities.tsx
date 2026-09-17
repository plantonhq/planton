'use client';

import { Box, Typography } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  FeatureTitle,
  BodyText,
  FeatureCard,
  Grid,
} from '@/components/marketing';
import {
  RocketLaunch as GitToProductionIcon,
  SwapVert as PromotionIcon,
  CloudQueue as DeployAnywhereIcon,
  Language as IngressIcon,
  Terminal as PodAccessIcon,
  Code as DevIcon,
  Science as StagingIcon,
  CheckCircle as ProdIcon,
} from '@mui/icons-material';
import {
  MetricsStrip,
  CodeTabs,
  FlowSteps,
  BentoGrid,
  BentoItem,
  AnimatedTerminal,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { TerminalLine, CodeTab, MetricItem, FlowStep } from '@/components/product/shared';

const metrics: MetricItem[] = [
  { value: '3', label: 'Build Modes' },
  { value: '4+', label: 'Deploy Targets' },
  { value: 'Git Push', label: 'to Production' },
  { value: 'Multi-env', label: 'Promotion' },
];

const gitToProdTabs: CodeTab[] = [
  {
    label: 'Buildpacks',
    code: `apiVersion: service-hub.planton.ai/v1
kind: Service
metadata:
  name: payments-api
  org: acme-corp
spec:
  gitRepo:
    cloneUrl: https://github.com/acme/payments.git
    defaultBranch: main
    gitRepoProvider: github
  packageType: container_image
  pipelineConfiguration:
    imageBuildMethod: buildpacks
    pipelineBranches:
      - main`,
  },
  {
    label: 'Dockerfile',
    code: `apiVersion: service-hub.planton.ai/v1
kind: Service
metadata:
  name: payments-api
  org: acme-corp
spec:
  gitRepo:
    cloneUrl: https://github.com/acme/payments.git
    defaultBranch: main
    gitRepoProvider: github
  packageType: container_image
  pipelineConfiguration:
    imageBuildMethod: dockerfile
    dockerfilePath: ./Dockerfile
    pipelineBranches:
      - main
      - staging`,
  },
];

const envPromotionSteps: FlowStep[] = [
  { icon: <DevIcon fontSize="small" />, label: 'Dev', sublabel: 'Push to main' },
  { icon: <StagingIcon fontSize="small" />, label: 'Staging', sublabel: 'Auto-promote' },
  { icon: <ProdIcon fontSize="small" />, label: 'Production', sublabel: 'Approval gate' },
];

const envPromotionTabs: CodeTab[] = [
  {
    label: 'Pipeline Output',
    code: `▶ Pipeline: spl-a7c3f1 (payments-api)
  Trigger: push to main (abc1234)

  ✓ Build     — 2m 18s (Buildpacks)
  ✓ Deploy    — dev (auto)
  ✓ Deploy    — staging (auto)
  ⏳ Approval — production
    Waiting for: platform-team`,
  },
  {
    label: 'Kustomize Overlay',
    code: `# _kustomize/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
patches:
  - path: resource-limits.yaml
configMapGenerator:
  - name: app-config
    literals:
      - LOG_LEVEL=warn
      - FEATURE_FLAGS=payments-v2`,
  },
];

const podAccessLines: TerminalLine[] = [
  { text: '$ planton service logs payments-api --env production', className: 'text-white' },
  { text: '' },
  { text: '▶ Streaming logs: payments-api (production)', className: 'text-[#b0b0b0]' },
  { text: '  Pod: payments-api-7d4f8b-x2k9p', className: 'text-[#666]' },
  { text: '  Container: app', className: 'text-[#666]' },
  { text: '' },
  { text: '[2026-03-25T10:15:32Z] INFO  Server started on :8080', className: 'text-[#b0b0b0]' },
  { text: '[2026-03-25T10:15:33Z] INFO  Connected to database', className: 'text-[#b0b0b0]' },
  { text: '[2026-03-25T10:15:34Z] INFO  Processing payment txn-4f2a...', className: 'text-[#b0b0b0]' },
  { text: '[2026-03-25T10:15:34Z] INFO  Payment completed in 142ms', className: 'text-[#10b981]' },
];

const GitToProductionShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <GitToProductionIcon />
            </Box>
            <FeatureTitle>Git to Production</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Connect your GitHub or GitLab repo. Planton detects your build method, builds with managed pipelines, and deploys to your target.
          </BodyText>
          <Box className="space-y-3">
            {[
              'Three build modes — Buildpacks for zero-config, Dockerfile for control, pre-built images for CI flexibility',
              'Webhook-triggered pipelines that fire on push, tag, or PR events',
              'Sparse checkout for monorepos — build only what changed',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <CodeTabs tabs={gitToProdTabs} title="Service Manifest" />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const EnvironmentPromotionShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <Box className="flex items-center justify-center gap-3 mb-4">
          <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
            <PromotionIcon />
          </Box>
          <SectionTitle className="!text-xl md:!text-2xl">Multi-Environment Promotion</SectionTitle>
        </Box>
        <SectionSubtitle className="mx-auto">
          Define your environments. Push to a branch, trigger the right pipeline. Promote across environments with approval gates.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <FlowSteps steps={envPromotionSteps} className="mb-10" />
    </ScrollReveal>

    <ScrollReveal delay={0.25}>
      <Box className="flex flex-col lg:flex-row gap-6 items-start">
        <Box className="flex-1 w-full">
          <CodeTabs tabs={envPromotionTabs} title="Environment Config" />
        </Box>
        <Box className="flex-1 space-y-4">
          {[
            { title: 'Tag & PR Flows', desc: 'Deploy on merge to main or on version tag. Configurable triggers per environment.' },
            { title: 'Sequential Stages', desc: 'Promotion rules ensure code flows dev → staging → production in the right order.' },
            { title: 'Approval Gates', desc: 'Automated or manual approval with configurable rollback policies.' },
            { title: 'Kustomize Overlays', desc: 'Environment-specific config in your repo. Variables, secrets, resource limits — all version-controlled.' },
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
);

const FeatureBentoSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Everything you need to ship and run services at scale</SectionTitle>
        <SectionSubtitle className="mx-auto">
          From a single API to a fleet of microservices &mdash; Service Hub handles the build, deploy, and operate lifecycle.
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
                    <DeployAnywhereIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Deploy Anywhere</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Kubernetes, ECS, Cloud Run, Cloudflare Workers &mdash; same service definition, different targets. Switch providers without rewriting your pipeline.
                </BodyText>
              </Box>
              <Box className="flex-1 grid grid-cols-2 gap-2">
                {['Kubernetes', 'ECS', 'Cloud Run', 'Workers'].map((target) => (
                  <Box key={target} className="px-3 py-2.5 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#888] text-center">
                    {target}
                  </Box>
                ))}
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <IngressIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Ingress Made Simple</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Domain, TLS, routing &mdash; Planton handles the Istio and cert-manager plumbing. Point your domain, push your code.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['DNS', 'TLS', 'Routing', 'Canary'].map((cap) => (
                <Box key={cap} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {cap}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <PodAccessIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Pod Access</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Stream logs, exec into containers, inspect resources &mdash; all from the Planton console, with IAM-scoped access.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['logs', 'exec', 'describe', 'events'].map((cmd) => (
                <Box key={cmd} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {cmd}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const PodAccessDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Debug in production without kubectl config</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Stream logs, exec into containers, inspect resources &mdash; all from the Planton console or CLI, with IAM-scoped access.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={podAccessLines}
        title="planton service logs"
        lineDelay={350}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<PodAccessIcon />}
            title="Real-time Logs"
            description="Stream logs from any pod and container. Filter by severity, search by keyword."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<GitToProductionIcon />}
            title="Container Exec"
            description="Exec into running containers for debugging — scoped by IAM role, not cluster access."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<DeployAnywhereIcon />}
            title="Resource Browser"
            description="Browse pods, services, and config maps with RBAC-controlled visibility."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

{/* SCREENSHOT OPPORTUNITY: ServiceHub Dashboard - Service Overview
   Show: The service detail page for a service like "payments-api" with:
   - Pipeline history showing recent builds (green checkmarks for success, timestamps)
   - Environment tabs (dev, staging, production) with deployment status per env
   - The service settings showing git repo connection, build method, and deploy target
   Value: Demonstrates the "Vercel for Backend, In Your Own Cloud" Service Hub line by showing a polished, real deployment dashboard.
   Suggested: 16:9 aspect ratio, dark theme, show at least one successful and one in-progress pipeline.
   Placement: Add an <img> tag wrapped in a <Section> component between PodAccessDeepDive and CTA.
*/}

export const ServiceHubCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <GitToProductionShowcase />
    <EnvironmentPromotionShowcase />
    <FeatureBentoSection />
    <PodAccessDeepDive />
  </>
);
