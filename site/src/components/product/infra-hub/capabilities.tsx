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
  Storage as CloudResourcesIcon,
  Dashboard as ChartsIcon,
  PlaylistAddCheck as StackJobsIcon,
  Bookmark as PresetsIcon,
  SwapHoriz as MultiProvisionerIcon,
  Description as TemplateIcon,
  DataObject as ValuesIcon,
  Layers as RenderIcon,
  Timeline as DagIcon,
  CheckCircle as DeployedIcon,
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
import { PLATFORM_STATS } from '@/data/platform-stats';

// ---------------------------------------------------------------------------
// DATA
// ---------------------------------------------------------------------------

const metrics: MetricItem[] = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Resource Types' },
  { value: '17', label: 'Cloud Providers' },
  { value: '< 3 min', label: 'Average Deploy' },
  { value: '0', label: 'Vendor Lock-in' },
];

const cloudResourceTabs: CodeTab[] = [
  {
    label: 'GCP',
    code: `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: my-postgres
spec:
  gcpProjectId: my-project
  databaseVersion: POSTGRES_15
  tier: db-f1-micro
  region: us-central1
  highAvailability: true`,
  },
  {
    label: 'AWS',
    code: `apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsInstance
metadata:
  name: my-postgres
spec:
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.micro
  region: us-east-1
  multiAz: true
  allocatedStorageGb: 20`,
  },
  {
    label: 'Azure',
    code: `apiVersion: azure.planton.dev/v1alpha1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: my-postgres
spec:
  resourceGroup: my-rg
  skuName: Standard_B1ms
  version: "15"
  location: eastus
  highAvailability:
    mode: ZoneRedundant`,
  },
];

const infraChartYaml: CodeTab[] = [
  {
    label: 'Chart.yaml',
    code: `apiVersion: infra-hub.planton.ai/v1
kind: InfraChart
metadata:
  name: AWS Microservices Backend
spec:
  selector:
    kind: platform
  description: VPC, ALB, ECS Fargate, Aurora PostgreSQL, ElastiCache Redis, and SQS.`,
  },
  {
    label: 'values.yaml',
    code: `params:
  - name: availability_zone_1
    description: First AZ for the subnet pair
    value: us-east-1a

  - name: service_name
    description: Prefix for ECS cluster and ALB
    value: my-service

  - name: databaseEnabled
    description: Create Aurora PostgreSQL cluster
    type: bool
    value: true`,
  },
  {
    label: 'templates/compute.yaml',
    code: `apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsCluster
metadata:
  name: "{{ values.service_name }}-ecs-cluster"
  group: compute
spec:
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT`,
  },
];

const chartFlowSteps: FlowStep[] = [
  { icon: <TemplateIcon fontSize="small" />, label: 'Chart Template', sublabel: 'Parameterized YAML' },
  { icon: <ValuesIcon fontSize="small" />, label: 'Values', sublabel: 'Per-environment config' },
  { icon: <RenderIcon fontSize="small" />, label: 'Render', sublabel: 'Resolved resources' },
  { icon: <DagIcon fontSize="small" />, label: 'DAG Pipeline', sublabel: 'Dependency ordering' },
  { icon: <DeployedIcon fontSize="small" />, label: 'Deployed', sublabel: 'Running infrastructure' },
];

const stackJobLines: TerminalLine[] = [
  { text: '$ planton stack-job watch sjb-abc123', className: 'text-white' },
  { text: '' },
  { text: '▶ Stack Job: sjb-abc123', className: 'text-[#b0b0b0]' },
  { text: '  Operation: update', className: 'text-[#666]' },
  { text: '  Resource: my-postgres (GcpCloudSql)', className: 'text-[#666]' },
  { text: '  Triggered by: alice@company.com', className: 'text-[#666]' },
  { text: '  Commit: "Enable HA for production DB"', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Previewing changes...', className: 'text-[#f59e0b]' },
  { text: '  ~ gcp:sql:DatabaseInstance  (update)', className: 'text-[#666]' },
  { text: '    + settings.availabilityType: "REGIONAL"', className: 'text-[#666]' },
  { text: '' },
  { text: '✓ Preview complete. 1 resource to update.', className: 'text-[#10b981]' },
  { text: '⏳ Applying changes...', className: 'text-[#f59e0b]' },
  { text: '✓ Update complete in 2m 34s.', className: 'text-[#10b981]' },
];

const presetTags = [
  'production-ha', 'dev-minimal', 'staging-standard',
  'gpu-ml-training', 'edge-low-latency', 'cost-optimized',
];

// ---------------------------------------------------------------------------
// SECTIONS
// ---------------------------------------------------------------------------

const CloudResourcesShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <CloudResourcesIcon />
            </Box>
            <FeatureTitle>Cloud Resources</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            {PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} resource types across every major provider. Each one is a typed,
            validated YAML manifest &mdash; not a bash script and a prayer.
          </BodyText>
          <Box className="space-y-3">
            {[
              'Typed specs per provider with CEL validation — catch errors before deployment',
              'Creation wizard with live provider data and preset configurations',
              'Every change creates an auditable stack job with full diff views',
              'Version history with commit messages for every resource modification',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <CodeTabs tabs={cloudResourceTabs} title="Resource Manifest" />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const InfraChartsShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <Box className="flex items-center justify-center gap-3 mb-4">
          <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
            <ChartsIcon />
          </Box>
          <SectionTitle className="!text-xl md:!text-2xl">From template to deployment</SectionTitle>
        </Box>
        <SectionSubtitle className="mx-auto">
          Infra Charts template your infrastructure patterns. Parameterize once, deploy across every environment.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <FlowSteps steps={chartFlowSteps} className="mb-10" />
    </ScrollReveal>

    <ScrollReveal delay={0.25}>
      <Box className="flex flex-col lg:flex-row gap-6 items-start">
        <Box className="flex-1 w-full">
          <CodeTabs tabs={infraChartYaml} title="Infra Chart" />
        </Box>
        <Box className="flex-1 space-y-4">
          {[
            {
              title: 'Jinjava Templating',
              desc: 'Use {{ values.x }} placeholders and {% if %} conditionals. valueFrom references wire cross-resource dependencies.',
            },
            {
              title: 'Dependency-Aware DAG',
              desc: 'VPC first, then subnets, then the database. Resources deploy in the right sequence automatically.',
            },
            {
              title: 'Community Charts',
              desc: 'Public charts in the infra-charts repository. Fork, customize, or contribute your own.',
            },
            {
              title: 'Immutable Projects',
              desc: 'Render once into an Infra Project for versioned, reproducible deployments.',
            },
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
        <SectionTitle>Built for teams that ship fast</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Every tool you need to manage infrastructure at scale &mdash; from execution tracking to multi-provisioner support.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        {/* Stack Jobs — wide card with mini terminal */}
        <StaggerItem>
          <BentoItem span="wide">
            <Box className="flex flex-col md:flex-row gap-5">
              <Box className="flex-1">
                <Box className="flex items-center gap-3 mb-3">
                  <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                    <StackJobsIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Stack Jobs</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Every infrastructure change is a tracked, auditable execution. Preview before
                  you apply. Full diffs. Commit messages. Version history.
                </BodyText>
                <Box className="flex flex-wrap gap-2">
                  {['Preview', 'Apply', 'Destroy', 'Refresh'].map((op) => (
                    <Box
                      key={op}
                      className="px-2.5 py-1 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#888]"
                    >
                      {op}
                    </Box>
                  ))}
                </Box>
              </Box>
              <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                <pre className="whitespace-pre-wrap">
{`⏳ Previewing changes...
  ~ gcp:sql:DatabaseInstance (update)
    + availabilityType: "REGIONAL"

✓ Preview complete. 1 resource.
✓ Update complete in 2m 34s.`}
                </pre>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        {/* Presets — tag cloud */}
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <PresetsIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Presets</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Start from battle-tested configurations, not blank YAML files.
              Community and org presets for every resource type.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {presetTags.map((tag) => (
                <Box
                  key={tag}
                  className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono"
                >
                  {tag}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        {/* Multi-Provisioner — visual toggle */}
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <MultiProvisionerIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Multi-Provisioner</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Choose Pulumi or Terraform. Same workflow, same API.
              Bring your own modules or use Planton&apos;s.
            </BodyText>
            <Box className="flex items-center gap-3">
              <Box className="flex-1 p-3 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white mb-0.5">Pulumi</Typography>
                <Typography className="text-[10px] text-[#666]">Go, TS, Python</Typography>
              </Box>
              <Typography className="text-[#555] text-xs font-medium">or</Typography>
              <Box className="flex-1 p-3 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white mb-0.5">Terraform</Typography>
                <Typography className="text-[10px] text-[#666]">HCL, OpenTofu</Typography>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const StackJobsDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>Every change is tracked. Every change is auditable.</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Watch infrastructure changes in real time. Preview before you apply. Full diffs. Commit messages.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={stackJobLines}
        title="planton stack-job watch"
        lineDelay={350}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            }
            title="Preview Before Apply"
            description="See exactly what will change before any resource is modified. Full resource diffs with add, update, and delete operations."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            }
            title="Per-Resource Locking"
            description="Concurrent modifications are prevented at the resource level. No conflicting deployments, no state corruption."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            }
            title="Preflight Checks"
            description="Credentials, state backend access, and provider connectivity are validated before any execution begins."
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

{/* SCREENSHOT OPPORTUNITY: InfraHub Dashboard - Resource List View
   Show: The cloud resources list with 5-6 resources of different providers (GCP, AWS, Azure),
   each showing status badges (Running, Pending), last-deployed timestamps, and the provider icon.
   Also show the left sidebar with InfraHub navigation (Resources, Charts, Projects, Pipelines).
   Value: Proves the multi-cloud resource management claim visually. Shows the console is real.
   Suggested: 16:9 aspect ratio, dark theme, with realistic resource names like "prod-postgres", "staging-vpc".
   Placement: Add an <img> tag wrapped in a <Section> component between StackJobsDeepDive and CTA.
*/}

// ---------------------------------------------------------------------------
// MAIN EXPORT
// ---------------------------------------------------------------------------

export const InfraHubCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <CloudResourcesShowcase />
    <InfraChartsShowcase />
    <FeatureBentoSection />
    <StackJobsDeepDive />
  </>
);
