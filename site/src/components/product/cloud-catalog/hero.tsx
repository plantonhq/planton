'use client';

import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  PrimaryButton,
  SecondaryButton,
  Badge,
  BodyText,
} from '@/components/marketing';
import { AnimatedTerminal, ScrollReveal } from '@/components/product/shared';
import type { TerminalLine } from '@/components/product/shared';
import { PLATFORM_STATS } from '@/data/platform-stats';

const heroLines: TerminalLine[] = [
  { text: '# Found in Cloud Catalog: AWS EKS Cluster', className: 'text-[#555]' },
  { text: '$ planton apply -f eks-cluster.yaml', className: 'text-white' },
  { text: '' },
  { text: '▶ Applying AwsEksCluster/production-cluster', className: 'text-[#b0b0b0]' },
  { text: '  Provider:     Amazon Web Services', className: 'text-[#666]' },
  { text: '  Environment:  production', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Previewing changes...', className: 'text-[#f59e0b]' },
  { text: '  + aws:eks:Cluster       (create)', className: 'text-[#666]' },
  { text: '  + aws:eks:NodeGroup     (create)', className: 'text-[#666]' },
  { text: '  + aws:iam:Role          (create)', className: 'text-[#666]' },
  { text: '✓ Preview complete. 12 resources to create.', className: 'text-[#10b981]' },
  { text: '' },
  { text: '⏳ Provisioning...', className: 'text-[#f59e0b]' },
  { text: '✓ AwsEksCluster provisioned in 8m 12s.', className: 'text-[#10b981]' },
];

const moduleTypes = ['Lego Blocks', 'Infra Charts'];
const topProviders = ['AWS', 'GCP', 'Azure', 'Kubernetes', 'DigitalOcean', 'Cloudflare', 'Auth0'];

export const CloudCatalogHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Cloud Catalog</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Find it. Deploy it. Done.
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Finding the right infrastructure module shouldn&apos;t require tribal knowledge.
            Teams waste hours searching docs, comparing Terraform modules, and copy-pasting
            configs across providers.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Cloud Catalog: Browse {PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} pre-built deployment
            modules across {PLATFORM_STATS.CLOUD_PROVIDER_COUNT} cloud providers. Filter by
            provider, preview configurations, and deploy to your cloud in minutes.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="/cloud-catalog">
              <PrimaryButton>Explore Catalog</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="catalog → deploy"
            lineDelay={300}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>

        <Box className="flex flex-wrap justify-center gap-6 mt-8">
          <Box className="flex items-center gap-2">
            <Typography component="span" className="text-[10px] text-[#555] uppercase font-semibold tracking-wider">Modules</Typography>
            {moduleTypes.map((m) => (
              <Typography key={m} component="span" className="text-xs text-[#888] px-2.5 py-1 rounded-md border border-[#1a1a1a] bg-[#111]">
                {m}
              </Typography>
            ))}
          </Box>
          <Box className="flex items-center gap-2">
            <Typography component="span" className="text-[10px] text-[#555] uppercase font-semibold tracking-wider">Providers</Typography>
            {topProviders.map((t) => (
              <Typography key={t} component="span" className="text-xs text-[#888] px-2.5 py-1 rounded-md border border-[#1a1a1a] bg-[#111]">
                {t}
              </Typography>
            ))}
          </Box>
        </Box>
      </Box>
    </Section>
  );
};
