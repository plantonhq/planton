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
  Widgets as LegoBlocksIcon,
  AccountTree as InfraChartIcon,
  Tune as PresetsIcon,
  Search as SearchIcon,
  CloudQueue as MultiCloudIcon,
  RocketLaunch as DeployIcon,
  GitHub as OpenSourceIcon,
  Code as YamlIcon,
  FilterAlt as FilterIcon,
  AutoAwesome as SmartIcon,
  Speed as SpeedIcon,
  Storefront as CatalogIcon,
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
import { PLATFORM_STATS } from '@/data/platform-stats';

const metrics: MetricItem[] = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Deployment Modules' },
  { value: PLATFORM_STATS.CLOUD_PROVIDER_COUNT, label: 'Cloud Providers' },
  { value: PLATFORM_STATS.INFRA_CHART_COUNT, label: 'Infra Charts' },
  { value: '< 5 min', label: 'First Deploy' },
];

const deployFlow: FlowStep[] = [
  { icon: <CatalogIcon fontSize="small" />, label: 'Browse', sublabel: 'Cloud Catalog' },
  { icon: <SearchIcon fontSize="small" />, label: 'Find', sublabel: 'Filter by provider' },
  { icon: <PresetsIcon fontSize="small" />, label: 'Configure', sublabel: 'Pick a preset' },
  { icon: <DeployIcon fontSize="small" />, label: 'Deploy', sublabel: 'One click' },
];

const yamlLines: TerminalLine[] = [
  { text: '# AWS EKS Cluster — production preset', className: 'text-[#555]' },
  { text: 'apiVersion: infrahub.planton.ai/v1', className: 'text-[#b0b0b0]' },
  { text: 'kind: AwsEksCluster', className: 'text-white' },
  { text: 'metadata:', className: 'text-[#b0b0b0]' },
  { text: '  name: production-cluster', className: 'text-[#b0b0b0]' },
  { text: 'spec:', className: 'text-[#b0b0b0]' },
  { text: '  region: us-east-1', className: 'text-[#666]' },
  { text: '  kubernetesVersion: "1.30"', className: 'text-[#666]' },
  { text: '  nodeGroups:', className: 'text-[#666]' },
  { text: '    - name: default', className: 'text-[#666]' },
  { text: '      instanceType: t3.large', className: 'text-[#666]' },
  { text: '      desiredSize: 3', className: 'text-[#666]' },
  { text: '' },
  { text: '$ planton apply -f eks-cluster.yaml', className: 'text-white' },
  { text: '✓ AwsEksCluster provisioned in 8m 12s.', className: 'text-[#10b981]' },
];

const TwoModuleTypes = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <SectionTitle>Two ways to deploy</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Lego Blocks for individual resources. Infra Charts for composed stacks.
          Both browsable, both deployable, both open source.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <Box className="max-w-3xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-6">
        <Box className="p-6 rounded-xl border border-[#2a2a2a] bg-[#0d0d0d]">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <LegoBlocksIcon />
            </Box>
            <FeatureTitle className="!text-lg">Lego Blocks</FeatureTitle>
          </Box>
          <BodyText className="!text-sm mb-4">
            Individual cloud resources &mdash; a database, a Kubernetes cluster, a load balancer.
            Each one is a typed, protobuf-defined API backed by an open-source Pulumi or Terraform module.
          </BodyText>
          <Box className="flex flex-wrap gap-1.5">
            {['EKS Cluster', 'RDS Instance', 'S3 Bucket', 'Cloud Run', 'GKE Cluster'].map((item) => (
              <Box key={item} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                {item}
              </Box>
            ))}
          </Box>
        </Box>

        <Box className="p-6 rounded-xl border border-[#2a2a2a] bg-[#0d0d0d]">
          <Box className="flex items-center gap-3 mb-4">
            <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
              <InfraChartIcon />
            </Box>
            <FeatureTitle className="!text-lg">Infra Charts</FeatureTitle>
          </Box>
          <BodyText className="!text-sm mb-4">
            Composed infrastructure stacks that wire multiple resources together with dependency-aware
            ordering. Deploy an entire environment with one manifest.
          </BodyText>
          <Box className="flex flex-wrap gap-1.5">
            {['Production Stack', 'Microservice Infra', 'Data Pipeline', 'Networking', 'Observability'].map((item) => (
              <Box key={item} className="px-2.5 py-1 rounded-full bg-white/5 border border-[#2a2a2a] text-[11px] text-[#777] font-mono">
                {item}
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
        <SectionTitle>Built for discovery</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Everything you need to find the right module, understand what it does, and deploy it to your cloud.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <StaggerContainer stagger={0.1}>
      <BentoGrid>
        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <FilterIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Filter by Provider</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Browse modules by cloud provider. AWS, GCP, Azure, Kubernetes, and more &mdash; each with provider-specific icons and grouping.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['AWS', 'GCP', 'Azure', 'Kubernetes', 'Cloudflare'].map((provider) => (
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
                    <PresetsIcon fontSize="small" />
                  </Box>
                  <FeatureTitle className="!text-base">Ready-to-Deploy Presets</FeatureTitle>
                </Box>
                <BodyText className="!text-sm mb-3">
                  Every deployment module comes with YAML presets for common configurations. Production, staging, minimal &mdash; pick one, customize, and deploy.
                </BodyText>
              </Box>
              <Box className="flex-1 rounded-lg bg-[#0d0d0d] border border-[#222] p-3 font-mono text-[11px] text-[#888] leading-relaxed min-w-0">
                <pre className="whitespace-pre-wrap">{`kind: AwsEksCluster
# Production preset
spec:
  region: us-east-1
  nodeGroups:
    - instanceType: t3.large
      desiredSize: 3`}</pre>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>

        <StaggerItem>
          <BentoItem>
            <Box className="flex items-center gap-3 mb-3">
              <Box className="w-9 h-9 rounded-lg bg-white/10 flex items-center justify-center text-white">
                <OpenSourceIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">Open Source</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Every module is backed by an open-source Planton module. Audit the code, fork it, or use it standalone with the Planton CLI.
            </BodyText>
            <Box className="flex flex-wrap gap-1.5">
              {['GitHub', 'Pulumi', 'Terraform', 'Portable'].map((item) => (
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
                <SmartIcon fontSize="small" />
              </Box>
              <FeatureTitle className="!text-base">No Sign-up Required</FeatureTitle>
            </Box>
            <BodyText className="!text-sm mb-4">
              Browse the entire catalog without an account. Search, filter, preview configurations, and read documentation. Sign in when you&apos;re ready to deploy.
            </BodyText>
            <Box className="flex items-center gap-3">
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">Browse</Typography>
              </Box>
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">Preview</Typography>
              </Box>
              <Box className="flex-1 p-2.5 rounded-lg bg-white/5 border border-[#2a2a2a] text-center">
                <Typography className="text-xs font-semibold text-white">Deploy</Typography>
              </Box>
            </Box>
          </BentoItem>
        </StaggerItem>
      </BentoGrid>
    </StaggerContainer>
  </Section>
);

const YamlDeepDive = () => (
  <Section className="!bg-[#0e0e0e]">
    <ScrollReveal>
      <Box className="text-center mb-8">
        <SectionTitle>YAML-native from catalog to cluster</SectionTitle>
        <SectionSubtitle className="mx-auto">
          Every module in the catalog is a typed YAML manifest. Browse, preview the spec,
          customize the preset, and deploy &mdash; no proprietary abstractions.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <AnimatedTerminal
        lines={yamlLines}
        title="eks-cluster.yaml"
        lineDelay={250}
        className="max-w-3xl mx-auto mb-10"
      />
    </ScrollReveal>

    <StaggerContainer stagger={0.1} className="max-w-3xl mx-auto">
      <Grid cols={3} gap="md">
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<YamlIcon />}
            title="Typed Manifests"
            description="Every field validated by protobuf schemas. No runtime surprises."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<SpeedIcon />}
            title="Instant Preview"
            description="See exactly what will be created before you deploy. Full resource DAG."
          />
        </StaggerItem>
        <StaggerItem className="h-full">
          <FeatureCard
            className="h-full"
            icon={<MultiCloudIcon />}
            title="Any Provider"
            description={`${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers. Same workflow. Same manifest structure. Zero context switching.`}
          />
        </StaggerItem>
      </Grid>
    </StaggerContainer>
  </Section>
);

const DeployFlowSection = () => (
  <Section>
    <ScrollReveal>
      <Box className="text-center mb-10">
        <Box className="flex items-center justify-center gap-3 mb-4">
          <Box className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center text-white">
            <DeployIcon />
          </Box>
          <SectionTitle className="!text-xl md:!text-2xl">Catalog to cluster in four steps</SectionTitle>
        </Box>
        <SectionSubtitle className="mx-auto">
          No Terraform expertise required. No YAML from scratch. Browse, pick, deploy.
        </SectionSubtitle>
      </Box>
    </ScrollReveal>

    <ScrollReveal delay={0.15}>
      <FlowSteps steps={deployFlow} />
    </ScrollReveal>
  </Section>
);

export const CloudCatalogCapabilities = () => (
  <>
    <MetricsStrip metrics={metrics} />
    <TwoModuleTypes />
    <FeatureBentoSection />
    <YamlDeepDive />
    <DeployFlowSection />
  </>
);
