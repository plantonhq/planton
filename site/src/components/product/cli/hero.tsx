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

const heroLines: TerminalLine[] = [
  { text: '$ brew install plantonhq/tap/planton', className: 'text-white' },
  { text: '✓ planton installed', className: 'text-[#10b981]' },
  { text: '' },
  { text: '$ planton apply -f postgres.yaml', className: 'text-white' },
  { text: '▶ Applying GcpCloudSql/my-postgres', className: 'text-[#b0b0b0]' },
  { text: '  Organization: acme-corp', className: 'text-[#666]' },
  { text: '  Environment:  production', className: 'text-[#666]' },
  { text: '' },
  { text: '⏳ Creating stack job...', className: 'text-[#f59e0b]' },
  { text: '✓ Stack job sjb-d4f21a created', className: 'text-[#10b981]' },
  { text: '⏳ Provisioning resources...', className: 'text-[#f59e0b]' },
  { text: '✓ GcpCloudSql created in 3m 12s', className: 'text-[#10b981]' },
];

const installMethods = [
  { label: 'Homebrew', command: 'brew install plantonhq/tap/planton' },
];

export const CliHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">CLI</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Everything Planton does, from your terminal
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            Context-switching between terminal, console, and CI. Different tools for
            different operations. Copy-pasting YAML between systems.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            kubectl-inspired commands, manifest-driven workflows, and real-time
            infrastructure operations &mdash; all from one CLI.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Get Started</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="planton cli"
            lineDelay={280}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>

        <Box className="flex flex-wrap justify-center gap-3 mt-8">
          {installMethods.map((m) => (
            <Box key={m.label} className="flex items-center gap-2 px-3 py-1.5 rounded-md border border-[#1a1a1a] bg-[#111]">
              <Typography component="span" className="text-[10px] text-[#666] uppercase font-semibold tracking-wider">
                {m.label}
              </Typography>
              <Typography component="span" className="text-xs text-[#888] font-mono">
                {m.command}
              </Typography>
            </Box>
          ))}
        </Box>
      </Box>
    </Section>
  );
};
