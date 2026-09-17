'use client';

import { Box, Typography } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  FeatureTitle,
  BodyText,
} from '@/components/marketing';
import {
  Api as OpenSourceApiIcon,
  FileCopy as ManifestIcon,
  Widgets as ChartsIcon,
  Construction as ForgeIcon,
  AddCircleOutline as PlantonAddsIcon,
  Build as InitIcon,
  Code as SpecIcon,
  Extension as ModuleIcon,
  VerifiedUser as ValidateIcon,
  MergeType as PrIcon,
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

const metrics: MetricItem[] = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Resource Types' },
  { value: '2', label: 'Provisioners' },
  { value: '0', label: 'Vendor Lock-in' },
  { value: '100%', label: 'Open Source' },
];

const openSourceTabs: CodeTab[] = [
  {
    label: 'Protobuf Spec',
    code: `message GcpCloudSqlSpec {
  string gcp_project_id = 1;
  string database_version = 2;
  string tier = 3;
  string region = 4;
  bool high_availability = 5;
}`,
  },
  {
    label: 'YAML Manifest',
    code: `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: my-postgres
  labels:
    env: production
    team: backend
spec:
  gcpProjectId: my-project
  databaseVersion: POSTGRES_15
  tier: db-custom-4-16384
  region: us-central1
  highAvailability: true`,
  },
  {
    label: 'CLI Output',
    code: `$ planton apply -f my-postgres.yaml
✓ Spec validated against protobuf schema
⏳ Provisioning via Pulumi module...
✓ GcpCloudSql created

# Export from Planton console
# Version in Git
# Apply with the planton CLI`,
  },
];

const forgeSteps: FlowStep[] = [
  { icon: <SpecIcon fontSize="small" />, label: 'Define Spec', sublabel: 'Protobuf schema' },
  { icon: <ModuleIcon fontSize="small" />, label: 'Implement', sublabel: 'Pulumi + Terraform' },
  { icon: <ValidateIcon fontSize="small" />, label: 'Validate', sublabel: '20 automated checks' },
  { icon: <InitIcon fontSize="small" />, label: 'Generate', sublabel: 'Docs, presets, CLI' },
  { icon: <PrIcon fontSize="small" />, label: 'Submit PR', sublabel: 'Community review' },
];

const forgeLines: TerminalLine[] = [
  { text: '# Component structure for a new deployment component', className: 'text-[#555]' },
  { text: '' },
  { text: 'apis/dev/planton/provider/gcp/gcpcloudsql/v1/', className: 'text-[#b0b0b0]' },
  { text: '  spec.proto            # Typed protobuf spec', className: 'text-[#888]' },
  { text: '  api.proto             # KRM envelope', className: 'text-[#888]' },
  { text: '  stack_input.proto     # IaC module inputs', className: 'text-[#888]' },
  { text: '  stack_outputs.proto   # Deployment outputs', className: 'text-[#888]' },
  { text: '  docs/README.md        # Auto-generated docs', className: 'text-[#888]' },
  { text: '  presets/              # Ranked starter configs', className: 'text-[#888]' },
  { text: '  pulumi/main.go        # Pulumi IaC module', className: 'text-[#888]' },
  { text: '  terraform/main.tf     # Terraform IaC module', className: 'text-[#888]' },
];

const OpenSourceShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="flex flex-col lg:flex-row gap-8 lg:gap-12 items-start">
        <Box className="flex-1 lg:max-w-md">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <OpenSourceApiIcon />
            </Box>
            <FeatureTitle>Planton open source</FeatureTitle>
          </Box>
          <BodyText className="!text-base mb-6">
            Protobuf-defined APIs for {PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} cloud resource types. Pulumi and Terraform modules.
            Validate and deploy from the CLI without any SaaS dependency.
          </BodyText>
          <Box className="space-y-3">
            {[
              'Every resource kind has a typed protobuf spec — not freeform YAML',
              'Pulumi modules in Go, TypeScript, Python and Terraform/OpenTofu modules',
              'CEL validation rules catch configuration errors before deployment',
              'Use standalone or as the foundation for Planton\'s managed platform',
            ].map((detail, i) => (
              <Box key={i} className="flex gap-2.5 items-start">
                <Box className="w-1.5 h-1.5 rounded-full bg-white/30 mt-2 flex-shrink-0" />
                <Typography className="text-sm text-[#b0b0b0] leading-relaxed">{detail}</Typography>
              </Box>
            ))}
          </Box>
        </Box>
        <Box className="flex-1 w-full">
          <CodeTabs tabs={openSourceTabs} title="Planton open source" />
        </Box>
      </Box>
    </ScrollReveal>
  </Section>
);

const ForgeShowcase = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <Box className="flex items-center justify-center gap-3 mb-4">
          <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
            <ForgeIcon />
          </Box>
          <SectionTitle className="!text-xl md:!text-2xl">Contribute with Forge</SectionTitle>
        </Box>
        <SectionSubtitle className="mx-auto">
          Add new cloud resource kinds through Forge. Protobuf spec, IaC module, validation &mdash; structured contribution path.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <FlowSteps steps={forgeSteps} className="mb-10" />
    </ScrollReveal>

    <ScrollReveal delay={0.25}>
      <Box className="flex flex-col lg:flex-row gap-6 items-start">
        <Box className="flex-1 w-full">
          <AnimatedTerminal lines={forgeLines} title="component structure" lineDelay={300} />
        </Box>
        <Box className="flex-1 space-y-4">
          {[
            { title: 'Structured Workflow', desc: 'The Forge process guides contributors through 20 automated validation steps — from proto definition to production-ready component.' },
            { title: 'Define in Protobuf', desc: 'Typed specs mean every field is documented, validated, and generates language bindings automatically.' },
            { title: 'Dual IaC Modules', desc: 'Every component ships with both Pulumi (Go) and Terraform (HCL) modules for maximum flexibility.' },
            { title: 'Community Review', desc: 'Submit a pull request. Automated validation pipeline verifies compatibility before merge.' },
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
        <SectionTitle>Open foundation, no lock-in</SectionTitle>
        <SectionSubtitle className="mx-auto">
          The open-source core that powers Planton. Use it standalone, extend it, or let Planton manage it for you.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <ManifestIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Portable Manifests</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Every resource is a standard KRM YAML manifest. Export, version, apply anywhere.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['apiVersion', 'kind', 'metadata', 'spec'].map((field) => (
                <Box key={field} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {field}
                </Box>
              ))}
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <ChartsIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Infra Charts</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Community-maintained infrastructure templates. Browse, fork, customize. Open repository on GitHub.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['vpc-stack', 'k8s-cluster', 'database-ha', 'cdn-setup'].map((chart) => (
                <Box key={chart} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                  {chart}
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
                    <PlantonAddsIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">What Planton Adds</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  The open-source core handles definitions and execution. Planton adds the platform layer for teams that need more.
                </BodyText>
              </Box>
              <Box className="flex-1 grid grid-cols-2 gap-2">
                {[
                  'Multi-tenancy', 'RBAC & Audit', 'DAG Pipelines', 'Web Console',
                  'CI/CD Integration', 'Runner Fleet', 'Agent Fleet', 'Connections',
                ].map((feature) => (
                  <Box key={feature} className="px-3 py-2 rounded-md bg-white/5 border border-[#2a2a2a] text-xs text-[#888] text-center">
                    {feature}
                  </Box>
                ))}
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

export const OpenSourceCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <OpenSourceShowcase />
    <ForgeShowcase />
    <FeatureBentoSection />
    {/* SCREENSHOT OPPORTUNITY: Open Source — Planton GitHub Repository
       Show: The plantonhq/planton GitHub repository page showing:
       - The directory structure with provider folders (aws, gcp, azure, kubernetes, etc.)
       - Star count, contributor count, and recent commit activity
       - The README with the project description visible
       Value: GitHub is the universal trust signal for open source. Showing an active, well-organized repo builds credibility.
       Suggested: 16:9 aspect ratio, light GitHub theme. Capture the repo root with the directory listing visible.
       Placement: Add an <img> tag wrapped in a <Section> component between FeatureBentoSection and CTA.
    */}
    {/* SCREENSHOT OPPORTUNITY: Open Source — Deployment Component Catalog
       Show: The console's deployment component catalog page showing:
       - A grid of component cards with provider icons, names, and descriptions
       - Filter/search functionality active
       - One component expanded showing its presets tab with ranked configurations
       Value: Proves the "700+ resource types" claim visually with a real, browsable catalog.
       Suggested: 16:9 aspect ratio, dark theme.
       Placement: Could be embedded in the FeatureBentoSection near the "Portable Manifests" or "Infra Charts" cards.
    */}
  </>
);
