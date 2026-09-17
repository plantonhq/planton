'use client';

import Link from 'next/link';
import { Box } from '@mui/material';
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

const heroLines: TerminalLine[] = [
  { text: '$ planton apply -f my-postgres.yaml', className: 'text-white' },
  { text: '' },
  { text: '✓ Spec validated against protobuf schema', className: 'text-[#10b981]' },
  { text: '⏳ Provisioning via Pulumi module...', className: 'text-[#f59e0b]' },
  { text: '✓ GcpCloudSql created', className: 'text-[#10b981]' },
  { text: '' },
  { text: '  No SaaS required. Fully portable.', className: 'text-[#666]' },
];

export const OpenSourceHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Open Source</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Your infrastructure definitions shouldn&apos;t be locked inside a
            vendor&apos;s platform
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Proprietary platforms trap your infrastructure definitions. When you want
            to leave, you rewrite everything.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            The open-source foundation that powers every Planton deployment.
            Protobuf-defined APIs, Pulumi and Terraform modules, portable KRM YAML
            manifests.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://github.com/plantonhq/planton" target="_blank">
              <PrimaryButton>Explore Planton on GitHub</PrimaryButton>
            </Link>
            <Link href="https://planton.ai/signup" target="_blank">
              <SecondaryButton>Try Planton Free</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="planton cli"
            lineDelay={350}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Box>
    </Section>
  );
};
