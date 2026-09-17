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
  Description as ManifestIcon,
  Terminal as StackJobIcon,
  Cloud as ConnectionIcon,
  Sailing as KubernetesIcon,
  VpnKey as EnvConfigIcon,
  SystemUpdateAlt as UpdateIcon,
} from '@mui/icons-material';
import {
  MetricsStrip,
  AnimatedTerminal,
  BentoGrid,
  BentoItem,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { TerminalLine, MetricItem } from '@/components/product/shared';

const metrics: MetricItem[] = [
  { value: '1', label: 'CLI for Everything' },
  { value: '< 1 min', label: 'Install Time' },
  { value: '3', label: 'Platforms' },
  { value: 'Real-time', label: 'Streaming' },
];

const applyLines: TerminalLine[] = [
  { text: '$ planton apply -f postgres.yaml', className: 'text-white' },
  { text: '' },
  { text: '▶ Applying GcpCloudSql/my-postgres', className: 'text-[#b0b0b0]' },
  { text: '  Organization: acme-corp', className: 'text-[#666]' },
  { text: '  Environment:  production', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Creating stack job...', className: 'text-[#f59e0b]' },
  { text: '✓ Stack job sjb-d4f21a created', className: 'text-[#10b981]' },
  { text: '⏳ Provisioning resources...', className: 'text-[#f59e0b]' },
  { text: '✓ GcpCloudSql created in 3m 12s', className: 'text-[#10b981]' },
  { text: '' },
  { text: 'Connection string written to:', className: 'text-[#666]' },
  { text: '  planton secret get my-postgres-conn-url', className: 'text-[#666]' },
];

const kubectlLines: TerminalLine[] = [
  { text: '$ planton kubectl --cluster prod-us-east-1 \\', className: 'text-white' },
  { text: '    get pods -n backend', className: 'text-white' },
  { text: '' },
  { text: 'NAME                      READY   STATUS', className: 'text-[#b0b0b0]' },
  { text: 'api-server-7d4f8b-x9k2m   1/1    Running', className: 'text-[#10b981]' },
  { text: 'worker-5c8a9f-p3n7j        1/1    Running', className: 'text-[#10b981]' },
  { text: 'scheduler-6b2e1d-m8w4r     1/1    Running', className: 'text-[#10b981]' },
  { text: '' },
  { text: '# No kubeconfig files. No VPN.', className: 'text-[#555]' },
  { text: '# IAM-scoped, auditable access.', className: 'text-[#555]' },
];

const dotenvLines: TerminalLine[] = [
  { text: '$ planton service dot-env \\', className: 'text-white' },
  { text: '    --service api-server --env dev', className: 'text-white' },
  { text: '' },
  { text: 'DATABASE_URL=postgres://...', className: 'text-[#b0b0b0]' },
  { text: 'REDIS_URL=redis://...', className: 'text-[#b0b0b0]' },
  { text: 'API_KEY=sk-...', className: 'text-[#b0b0b0]' },
  { text: 'STRIPE_SECRET=whsec-...', className: 'text-[#b0b0b0]' },
  { text: '' },
  { text: '✓ 12 variables written to .env', className: 'text-[#10b981]' },
  { text: '  4 secrets resolved from vault', className: 'text-[#666]' },
];

const ManifestDrivenShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <ManifestIcon />
            </Box>
            <FeatureTitle>Manifest-Driven</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Deploy any cloud resource or service from YAML. Same manifests work in CI, in console, from terminal.
          </BodyText>
          <Box className="space-y-3">
            {[
              '`planton apply -f manifest.yaml` — one command for any resource type',
              'Validated against typed protobuf specs before submission',
              'Diff preview before apply — see exactly what will change',
              'Works identically in local terminal, CI pipeline, and console',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <AnimatedTerminal lines={applyLines} title="planton apply" lineDelay={300} />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const KubernetesAccessShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row-reverse gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <KubernetesIcon />
            </Box>
            <FeatureTitle>Kubernetes Access</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Access any connected cluster without kubeconfig files. IAM-scoped, auditable, zero local configuration.
          </BodyText>
          <Box className="space-y-3">
            {[
              '`planton kubectl` — proxy through Planton with org-level RBAC',
              'No kubeconfig files, no VPN tunnels, no local credential management',
              'Cluster auto-discovery from your connected providers',
              'Full audit trail of every kubectl command executed',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <AnimatedTerminal lines={kubectlLines} title="planton kubectl" lineDelay={300} />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const FeatureBentoSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Your infrastructure, one command away</SectionTitle>
        <SectionSubtitle className="mx-auto">
          From deploying resources to streaming logs to managing secrets &mdash; the Planton CLI covers your entire workflow.
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
                    <StackJobIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Stack Job Operations</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Stream IaC execution logs in real-time. Preview, apply, destroy from command line.
                </BodyText>
                <Box className="flex flex-wrap gap-2">
                  {['watch', 'list', 'cancel', 'retry'].map((cmd) => (
                    <Box key={cmd} className="px-2.5 py-1 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#888] font-mono">
                      planton stack-job {cmd}
                    </Box>
                  ))}
                </Box>
              </Box>
              <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                <pre className="whitespace-pre-wrap">{`⏳ Previewing changes...
  ~ gcp:sql:DatabaseInstance (update)
    + availabilityType: "REGIONAL"

✓ Preview complete. 1 resource.
✓ Update complete in 2m 34s.`}</pre>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <ConnectionIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Connection Management</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Create, list, and manage cloud connections. Guided setup for every supported provider.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['create', 'list', 'inspect', 'rotate'].map((cmd) => (
                <Box key={cmd} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {cmd}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <UpdateIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Built-in Upgrade</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Run <code className="text-[#b0b0b0]">planton upgrade</code> to check for updates and install the latest version in seconds.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['upgrade', 'upgrade --check', 'upgrade --force'].map((cmd) => (
                <Box key={cmd} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  planton {cmd}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const EnvConfigDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Local dev, production secrets</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Pull environment variables for local development. Secrets resolve straight from your own
          vault &mdash; never stored in Planton&apos;s database.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={dotenvLines}
        title="planton service dot-env"
        lineDelay={350}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<EnvConfigIcon />}
            title="Secrets Resolved Securely"
            description="Secrets are fetched at runtime from your vault, never stored in plaintext on disk."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<ConnectionIcon />}
            title="Environment-Specific"
            description="Separate variable sets for dev, staging, and production. Switch with a single flag."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<ManifestIcon />}
            title="Vault Integration"
            description="Works with AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault, and more."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

{/* SCREENSHOT OPPORTUNITY: CLI in Action - Terminal Session
   Show: A real terminal screenshot (e.g. iTerm2 or macOS Terminal) showing a sequence of CLI commands:
   - `planton auth login` with success output
   - `planton apply -f postgres.yaml` with real-time streaming output
   - The output should show the actual CLI formatting, colors, and progress indicators
   Value: Shows the CLI is a real, polished tool — not just a marketing concept.
   Suggested: 16:10 aspect ratio, dark terminal theme, realistic command output.
   Placement: Add an <img> tag wrapped in a <Section> component between EnvConfigDeepDive and CTA.
*/}

export const CliCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <ManifestDrivenShowcase />
    <KubernetesAccessShowcase />
    <FeatureBentoSection />
    <EnvConfigDeepDive />
  </>
);
