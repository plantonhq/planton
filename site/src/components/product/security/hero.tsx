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
  { text: '▶ Resolving secrets for GcpCloudSql/production-db', className: 'text-[#b0b0b0]' },
  { text: '' },
  { text: '  db-password  → gcp-secret-manager  ✓', className: 'text-[#10b981]' },
  { text: '  api-key      → aws-secrets-manager  ✓', className: 'text-[#10b981]' },
  { text: '  tls-cert     → hashicorp-vault      ✓', className: 'text-[#10b981]' },
  { text: '' },
  { text: '⏳ Injecting at execution boundary...', className: 'text-[#f59e0b]' },
  { text: '✓ 3 secrets resolved. Zero plaintext.', className: 'text-[#10b981]' },
];

export const SecurityHero = () => {
  return (
    <Section className="pt-24 md:pt-32">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <Badge className="mb-6">Security</Badge>
          <SectionTitle className="!text-3xl md:!text-5xl !leading-tight mb-4">
            Security isn&apos;t a bolt-on. It&apos;s built into every layer.
          </SectionTitle>
          <SectionSubtitle className="mx-auto !max-w-3xl mb-4">
            You need SaaS convenience but can&apos;t hand over your cloud credentials.
            Your team is growing but secret sprawl is growing faster. Auditors want a
            trail for every infrastructure change.
          </SectionSubtitle>
          <BodyText className="!text-base md:!text-lg mx-auto max-w-3xl mb-8">
            Secrets management, identity and access control, and full audit trails are
            built into every layer of Planton &mdash; from how credentials are stored to
            how infrastructure changes are executed.
          </BodyText>
          <Box className="flex flex-col sm:flex-row gap-3 justify-center items-center">
            <Link href="https://planton.ai/signup" target="_blank">
              <PrimaryButton>Start Free</PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton>Book a Demo</SecondaryButton>
            </Link>
          </Box>
        </Box>

        <ScrollReveal delay={0.2}>
          <AnimatedTerminal
            lines={heroLines}
            title="secret resolution"
            lineDelay={350}
            className="max-w-3xl mx-auto"
          />
        </ScrollReveal>
      </Box>
    </Section>
  );
};
