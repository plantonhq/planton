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
  Hub as ArchitectureIcon,
  CloudSync as CloudOpsIcon,
  BuildCircle as IaCIcon,
  Security as SecurityModelIcon,
  RocketLaunch as DeploymentIcon,
  Cable as TunnelIcon,
  Language as BrowserIcon,
  Api as ApiIcon,
  Shield as GatewayIcon,
  Lock as EncryptedIcon,
  Dns as RunnerNodeIcon,
  Terminal as KubectlIcon,
} from '@mui/icons-material';
import {
  MetricsStrip,
  FlowSteps,
  BentoGrid,
  BentoItem,
  AnimatedTerminal,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { TerminalLine, MetricItem, FlowStep } from '@/components/product/shared';

const metrics: MetricItem[] = [
  { value: '1', label: 'Binary' },
  { value: '0', label: 'Inbound Ports' },
  { value: '< 5 ms', label: 'Tunnel Overhead' },
  { value: '3', label: 'Deploy Modes' },
];

const iacLines: TerminalLine[] = [
  { text: '$ planton stack-job watch sjb-runner-7f3a', className: 'text-white' },
  { text: '' },
  { text: '▶ Stack Job: sjb-runner-7f3a', className: 'text-[#b0b0b0]' },
  { text: '  Runner: runner-prod-us-east-1', className: 'text-[#666]' },
  { text: '  Operation: update', className: 'text-[#666]' },
  { text: '  Resource: prod-rds (AwsRdsInstance)', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Resolving IAM credentials via IRSA...', className: 'text-[#f59e0b]' },
  { text: '✓ Assumed role: arn:aws:iam::123456789012:role/planton-runner', className: 'text-[#10b981]' },
  { text: '⏳ Previewing changes...', className: 'text-[#f59e0b]' },
  { text: '  ~ aws:rds:Instance  (update)', className: 'text-[#666]' },
  { text: '    + instanceClass: "db.r6g.xlarge"', className: 'text-[#666]' },
  { text: '' },
  { text: '✓ Preview complete. 1 resource to update.', className: 'text-[#10b981]' },
  { text: '⏳ Applying changes...', className: 'text-[#f59e0b]' },
  { text: '✓ Update complete in 3m 12s.', className: 'text-[#10b981]' },
  { text: '✓ Progress streamed in real time.', className: 'text-[#10b981]' },
];

const tunnelSteps: FlowStep[] = [
  { icon: <BrowserIcon fontSize="small" />, label: 'Browser / CLI', sublabel: 'User request' },
  { icon: <ApiIcon fontSize="small" />, label: 'Planton API', sublabel: 'Route to Runner' },
  { icon: <GatewayIcon fontSize="small" />, label: 'Gateway', sublabel: 'Identity verified' },
  { icon: <EncryptedIcon fontSize="small" />, label: 'Encrypted Tunnel', sublabel: 'Outbound only' },
  { icon: <RunnerNodeIcon fontSize="small" />, label: 'Runner', sublabel: 'Your cloud' },
  { icon: <KubectlIcon fontSize="small" />, label: 'kubectl / Cloud API', sublabel: 'Executed locally' },
];

const ArchitectureShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>How Runner works</SectionTitle>
        <SectionSubtitle className="mx-auto">
          A single binary that bridges your cloud and Planton&apos;s control plane &mdash;
          without exposing your credentials or opening inbound ports.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <Box className="max-w-3xl mx-auto mb-10">
        <Box className="rounded-xl border border-[#2a2a2a] bg-[#0d0d0d] p-6 md:p-8 space-y-4">
          <Box className="rounded-lg border border-white/20 bg-white/5 p-4">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-3">
              Planton Control Plane (SaaS)
            </Typography>
            <Box className="flex flex-wrap gap-2">
              {['API Server', 'Workflow Engine', 'Console UI', 'Orchestration'].map((item) => (
                <Box key={item} className="px-3 py-1.5 rounded-md bg-[#151515] border border-[#2a2a2a] text-xs text-[#b0b0b0]">
                  {item}
                </Box>
              ))}
            </Box>
          </Box>

          <Box className="flex justify-center">
            <Box className="flex flex-col items-center">
              <Box className="w-px h-4 bg-[#3a3a3a]" />
              <Box className="px-3 py-1 rounded-full border border-[#3a3a3a] bg-[#151515] text-[10px] text-[#666]">
                Encrypted Tunnel &middot; Outbound Only
              </Box>
              <Box className="w-px h-4 bg-[#3a3a3a]" />
            </Box>
          </Box>

          <Box className="rounded-lg border border-[#2a2a2a] bg-[#151515] p-4">
            <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-3">
              Runner (Your Cloud)
            </Typography>
            <Box className="flex flex-wrap gap-2">
              {['Cloud Ops', 'IaC Executor', 'Provider APIs', 'Your Resources'].map((item) => (
                <Box key={item} className="px-3 py-1.5 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#b0b0b0]">
                  {item}
                </Box>
              ))}
            </Box>
          </Box>
        </Box>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.25}>
      <Box className="max-w-3xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-4">
        <Box className="p-4 rounded-lg border border-[#2a2a2a] bg-[#111]">
          <Box className="flex items-center gap-2 mb-2">
            <CloudOpsIcon fontSize="small" className="text-white" />
            <Typography className="text-sm font-semibold text-white">CloudOps Mode</Typography>
          </Box>
          <Typography className="text-xs text-[#888] leading-relaxed">
            Real-time cloud operations proxied through Runner. kubectl, cloud APIs, cluster inspection &mdash; all with IAM-scoped access.
          </Typography>
        </Box>
        <Box className="p-4 rounded-lg border border-[#2a2a2a] bg-[#111]">
          <Box className="flex items-center gap-2 mb-2">
            <IaCIcon fontSize="small" className="text-white" />
            <Typography className="text-sm font-semibold text-white">IaC Execution Mode</Typography>
          </Box>
          <Typography className="text-xs text-[#888] leading-relaxed">
            Stack jobs execute on Runner using Pulumi or Terraform. Your cloud provider&apos;s native IAM authenticates Runner to your resources.
          </Typography>
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const FeatureBentoSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Built for security-conscious teams</SectionTitle>
        <SectionSubtitle className="mx-auto">
          From cryptographic identity to deployment options &mdash; Runner gives you SaaS convenience with self-hosted security.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <CloudOpsIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">CloudOps</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Real-time cloud operations through the Planton console. kubectl, cloud provider APIs, cluster inspection.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['AWS', 'GCP', 'Azure', 'Kubernetes'].map((provider) => (
                <Box key={provider} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {provider}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem span="wide">
            <Box className="flex flex-col md:flex-row gap-5">
              <Box className="flex-1">
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <IaCIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">IaC Execution</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Stack jobs execute on Runner using Pulumi or Terraform. Native IAM authenticates Runner to your resources. No long-lived credentials.
                </BodyText>
              </Box>
              <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                <pre className="whitespace-pre-wrap">{`⏳ Resolving IAM via IRSA...
✓ Assumed role: planton-runner
⏳ Previewing changes...
✓ 1 resource to update.
✓ Update complete in 3m 12s.`}</pre>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <SecurityModelIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Security Model</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Cryptographic identity for every Runner. SHA-256 hashed API keys. Anti-impersonation validation.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['Crypto ID', 'SHA-256', 'Least Privilege', 'Anti-Spoof'].map((item) => (
                <Box key={item} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {item}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <DeploymentIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Deployment Options</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Kubernetes DaemonSet, standalone binary, or Docker. Install in minutes.
            </BodyText>
            <Box className="flex items-center gap-3">
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">K8s</Typography>
              </Box>
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">Binary</Typography>
              </Box>
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">Docker</Typography>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const IaCDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Watch IaC execution in real time</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Stack jobs stream live from Runner to your console. Preview, apply, and track every change with full audit trail.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={iacLines}
        title="planton stack-job watch"
        lineDelay={300}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<IaCIcon />}
            title="Resumable Execution"
            description="Jobs pick up where they left off. Automatic retries with reliable execution guarantees."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<SecurityModelIcon />}
            title="JIT Secrets"
            description="Secrets fetched at execution time, never stored on disk. Resolved via your cloud provider's IAM."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<ArchitectureIcon />}
            title="Preflight Checks"
            description="Credentials, state backend, and provider connectivity validated before execution begins."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

const SecureTunnelSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <Box className="flex items-center justify-center gap-3 mb-4">
          <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
            <TunnelIcon />
          </Box>
          <SectionTitle className="!text-xl md:!text-2xl">Secure Tunnel</SectionTitle>
        </Box>
        <SectionSubtitle className="mx-auto">
          Outbound-only connection from Runner to control plane. ~1&ndash;5ms overhead. Automatic reconnection. Built-in monitoring.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <FlowSteps steps={tunnelSteps} />
    </ScrollReveal>
  </Section>
);

export const RunnerCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <ArchitectureShowcase />
    <FeatureBentoSection />
    <IaCDeepDive />
    <SecureTunnelSection />
    {/* SCREENSHOT OPPORTUNITY: Runner Enrollment
       Show: The Planton console's Runners page showing a freshly enrolled Runner with:
       - The runner token that admitted it (attribution on the detail page)
       - Runner status showing "Connected" with a green indicator
       - The tunnel connection details (channel ID, last heartbeat)
       Value: Proves the "start with a token, it enrolls itself" claim and shows the polished console experience.
       Suggested: 16:9 aspect ratio, dark theme, show the connected state.
       Placement: Add an <img> tag wrapped in a <Section> component between SecureTunnelSection and CTA.
    */}
  </>
);
