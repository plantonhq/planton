'use client';

import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import { FC } from 'react';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  Card,
  FeatureTitle,
  BodyText,
  CheckIcon,
  SecondaryButton,
} from './shared';
import { PLATFORM_STATS } from '@/data/platform-stats';

/**
 * The no-lock-in section. The boundary sentence is load-bearing and
 * approved: open source owns HOW to deploy; the platform owns workflow
 * and governance. The exit path must always be stated concretely (export
 * YAML, keep deploying with the CLI) — never as a vague "no lock-in" badge.
 */

const auditPoints = [
  'All Terraform and Pulumi deployment modules',
  'Schema definitions and validation rules',
  'The CLI and its deploy engine',
];

export const OpenSource: FC = () => (
  <Section id="open-source">
    <Box className="text-center mb-10">
      <SectionTitle>Open Infrastructure Modules — Not a Black Box</SectionTitle>
      <SectionSubtitle className="mx-auto">
        {PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} catalog kinds on
        GitHub, Apache-2.0. Audit them, fork them, or use them without the
        platform.
      </SectionSubtitle>
    </Box>

    <Box className="flex flex-col md:flex-row gap-6 max-w-5xl mx-auto mb-8">
      <Card className="flex-1">
        <FeatureTitle className="mb-3">Audit Every Line</FeatureTitle>
        <BodyText className="mb-4">
          Every module Planton runs lives in the public repository. Want to
          know exactly what IAM permissions a deployment needs? Read the code.
        </BodyText>
        <Box className="space-y-2.5 mb-5">
          {auditPoints.map((point) => (
            <Box key={point} className="flex items-start gap-2">
              <Box className="mt-0.5 shrink-0">
                <CheckIcon />
              </Box>
              <Typography className="text-sm text-[#b0b0b0]">{point}</Typography>
            </Box>
          ))}
        </Box>
        <Link
          href="https://github.com/plantonhq/planton"
          target="_blank"
          rel="noopener noreferrer"
        >
          <SecondaryButton>View on GitHub</SecondaryButton>
        </Link>
      </Card>

      <Card className="flex-1">
        <FeatureTitle className="mb-3">The Exit Path, Built In</FeatureTitle>
        <BodyText className="mb-4">
          If you ever leave, you keep deploying. Export all your
          configurations as YAML manifests and run them with the open-source
          CLI — same modules, same results, no platform required.
        </BodyText>
        <Box className="rounded-lg bg-[#111] border border-[#2a2a2a] px-4 py-3 font-mono text-xs text-[#b0b0b0]">
          $ planton apply -f manifest.yaml
        </Box>
      </Card>
    </Box>

    <Typography className="text-center text-sm text-[#a0a0a0]">
      Use Planton because it is the best platform —{' '}
      <strong className="text-white">not because switching is too expensive.</strong>
    </Typography>
  </Section>
);
