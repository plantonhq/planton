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
  Lock as SecretsIcon,
  VpnKey as MultiBackendIcon,
  Shield as RunnerTrustIcon,
  Fingerprint as IdentityIcon,
  VerifiedUser as AuditIcon,
  Key as ConnectionIcon,
  Security as ZeroTrustIcon,
} from '@mui/icons-material';
import {
  MetricsStrip,
  CodeTabs,
  BentoGrid,
  BentoItem,
  AnimatedTerminal,
  ScrollReveal,
  StaggerContainer,
  StaggerItem,
} from '@/components/product/shared';
import type { TerminalLine, CodeTab, MetricItem } from '@/components/product/shared';

const metrics: MetricItem[] = [
  { value: '5+', label: 'Secret Backends' },
  { value: '0', label: 'Plaintext in Production' },
  { value: '100%', label: 'Changes Audited' },
  { value: 'JIT', label: 'Secret Resolution' },
];

const secretRefTabs: CodeTab[] = [
  {
    label: 'GCP Secret Manager',
    code: `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: production-db
spec:
  databasePassword:
    secretRef:
      name: db-password
      backend: gcp-secret-manager`,
  },
  {
    label: 'AWS Secrets Manager',
    code: `apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsInstance
metadata:
  name: production-db
spec:
  masterPassword:
    secretRef:
      name: db-password
      backend: aws-secrets-manager`,
  },
  {
    label: 'HashiCorp Vault',
    code: `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: production-db
spec:
  databasePassword:
    secretRef:
      name: db-password
      backend: hashicorp-vault
      path: secret/data/production`,
  },
];

const auditLines: TerminalLine[] = [
  { text: '$ planton cloud-resource versions my-postgres', className: 'text-white' },
  { text: '' },
  { text: 'VERSION   AUTHOR              MESSAGE                    DATE', className: 'text-[#666]' },
  { text: 'v7        alice@company.com   Enable HA for prod DB      2026-03-24', className: 'text-[#b0b0b0]' },
  { text: 'v6        bob@company.com     Increase storage to 500GB  2026-03-20', className: 'text-[#b0b0b0]' },
  { text: 'v5        ci-bot              Rotate database password   2026-03-15', className: 'text-[#b0b0b0]' },
  { text: 'v4        alice@company.com   Add read replica           2026-03-10', className: 'text-[#b0b0b0]' },
  { text: '' },
  { text: '$ planton cloud-resource diff my-postgres v6 v7', className: 'text-white' },
  { text: '  - highAvailability: false', className: 'text-[#ef4444]' },
  { text: '  + highAvailability: true', className: 'text-[#10b981]' },
];

const SecretsShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <SecretsIcon />
            </Box>
            <FeatureTitle>Secrets Management</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Store secrets encrypted at rest. Reference them by name, not by value &mdash; no plaintext
            in transit or at rest in production. Secrets are resolved just-in-time within the Runner&apos;s
            security boundary, never transiting the control plane.
          </BodyText>
          <Box className="space-y-3">
            {[
              'SecretRef pattern — reference secrets by name in resource specs, never inline values',
              'Encrypted storage with provider-native encryption at rest',
              'Execution-time resolution — secrets are injected only when the stack job runs',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <CodeTabs tabs={secretRefTabs} title="SecretRef Pattern" />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const RunnerTrustShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row-reverse gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <RunnerTrustIcon />
            </Box>
            <FeatureTitle>Runner Trust Model</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Runner executes IaC and operations in YOUR cloud. Credentials are resolved via your cloud
            provider&apos;s native IAM. The Planton control plane never sees them.
          </BodyText>
          <Box className="space-y-3">
            {[
              'Just-in-time credential resolution via native cloud IAM — no long-lived secrets',
              'Runner runs in your VPC with your security policies and network controls',
              'Encrypted tunnel between Runner and control plane with verified identity on both sides',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <Box className="rounded-xl border border-[#2a2a2a] bg-[#0d0d0d] p-6 space-y-4">
            {[
              { label: 'Runner (your VPC)', items: ['IRSA (AWS)', 'Workload Identity (GCP)', 'Managed Identity (Azure)'], accent: false },
              { label: 'Encrypted Tunnel', items: ['Verified identity', 'Outbound only'], accent: true },
              { label: 'Planton Control Plane', items: ['Never sees cloud credentials', 'Orchestration only'], accent: false },
            ].map((layer) => (
              <Box key={layer.label} className={`rounded-lg border p-4 ${layer.accent ? 'border-white/20 bg-white/5' : 'border-[#2a2a2a] bg-[#151515]'}`}>
                <Typography className="text-xs font-semibold uppercase tracking-wider text-[#666] mb-2">{layer.label}</Typography>
                <Box className="flex flex-wrap gap-2">
                  {layer.items.map((item) => (
                    <Box key={item} className="px-2.5 py-1 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#b0b0b0]">
                      {item}
                    </Box>
                  ))}
                </Box>
              </Box>
            ))}
          </Box>
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const FeatureBentoSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Security at every layer of the stack</SectionTitle>
        <SectionSubtitle className="mx-auto">
          From secrets storage to zero-trust networking &mdash; every security control is native, not bolted on.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <MultiBackendIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Multi-Backend Secrets</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Bring your own secrets backend. Or use Planton&apos;s managed backend to get started in seconds.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['AWS SM', 'GCP SM', 'Azure KV', 'Vault', 'K8s Secrets'].map((backend) => (
                <Box key={backend} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {backend}
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
                    <IdentityIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Identity &amp; Access</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Human users and machine identities share one unified identity model. Fine-grained, relationship-driven access control.
                </BodyText>
              </Box>
              <Box className="flex-1 grid grid-cols-2 gap-2">
                {['IdentityAccount', 'Service Accounts', 'API Keys (SHA-256)', 'Org-Scoped RBAC', 'Env-Level AuthZ', 'Relationship-Driven'].map((item) => (
                  <Box key={item} className="px-3 py-2 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#888] text-center">
                    {item}
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
                <ConnectionIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Connection Security</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              All connections use typed SecretRef fields. OAuth tokens rotate automatically. GitHub App installations with zero user-managed secrets.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['SecretRef', 'OAuth Rotation', 'GitHub App', 'Scoped RBAC'].map((item) => (
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
                <ZeroTrustIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Zero-Trust Architecture</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Encrypted tunnels with cryptographic identity for every Runner. No implicit trust between any component.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['Crypto Identity', 'JWT Auth', 'Per-Service AuthZ', 'No Implicit Trust'].map((item) => (
                <Box key={item} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {item}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const AuditTrailsDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Every change has a story</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Version history with Git-like commit messages. Color-coded diffs. Searchable audit log across all resources.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={auditLines}
        title="planton cloud-resource"
        lineDelay={300}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<AuditIcon />}
            title="Version History"
            description="Git-like commit messages for every resource modification. Know who changed what and why."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<SecretsIcon />}
            title="Color-Coded Diffs"
            description="Compare any two versions of a resource with clear add/remove/update highlighting."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<IdentityIcon />}
            title="Searchable Audit Log"
            description="Search across all resources, environments, and identities. Stack job logs preserved for every execution."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

export const SecurityCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <SecretsShowcase />
    <RunnerTrustShowcase />
    <FeatureBentoSection />
    <AuditTrailsDeepDive />
    {/* SCREENSHOT OPPORTUNITY: Security — Audit Trail & Version Diff
       Show: The cloud resource version history page with:
       - A list of 4-5 versions showing author, commit message, and timestamp
       - The diff view comparing two versions with red/green highlighting (e.g., highAvailability toggled)
       Value: Visually proves the "every change is tracked" claim. Diff views are compelling to engineering leaders.
       Suggested: 16:9 aspect ratio, dark theme. Show a meaningful change like enabling HA or rotating credentials.
       Placement: Add an <img> tag wrapped in a <Section> component between AuditTrailsDeepDive and CTA.
    */}
    {/* SCREENSHOT OPPORTUNITY: Security — Connection Management
       Show: The connections list page with 3-4 different connection types (AWS, GCP, GitHub) showing:
       - Connection status (active/inactive), provider icon, last-used timestamp
       - The SecretRef pattern in action — no plaintext values visible anywhere
       Value: Shows the breadth of integrations and the security-first design of connection management.
       Suggested: 16:9 aspect ratio, dark theme.
       Placement: Could be added alongside the Connection Security bento card in FeatureBentoSection.
    */}
  </>
);
