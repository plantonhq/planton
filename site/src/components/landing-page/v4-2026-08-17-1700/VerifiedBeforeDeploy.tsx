'use client';

import { Box, Typography } from '@mui/material';
import { FC } from 'react';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  Grid,
  FeatureCard,
} from './shared';

/**
 * The differentiators that did not exist when the last landing page was
 * written: verified cost, permissions, and compliance data on the catalog,
 * surfaced BEFORE deployment. Every claim here is factual and must stay
 * free of dollar-savings math — this section never quotes a customer
 * saving or a salary comparison.
 */

const facts = [
  {
    title: 'The Bill, Before Deployment',
    description:
      'An itemized cloud bill for the architecture you composed — each line citing the provider price document it was priced from, and the date it was verified.',
  },
  {
    title: 'Least-Privilege Permissions',
    description:
      'A downloadable IAM policy scoped to exactly the architecture on screen — the answer to the first question every security review asks.',
  },
  {
    title: 'Compliance Posture',
    description:
      'Per-kind controls mapped to HIPAA, SOC 2, FedRAMP, and CIS requirements, with evidence for what is enforced. Stated plainly, never as a certification.',
  },
];

export const VerifiedBeforeDeploy: FC = () => (
  <Section id="verified-before-deploy">
    <Box className="text-center mb-10">
      <SectionTitle>Verified Data, Not Model Memory</SectionTitle>
      <SectionSubtitle className="mx-auto">
        Ask a coding agent what your architecture costs and it guesses. The
        Planton catalog carries verified answers — available before anything
        touches your cloud.
      </SectionSubtitle>
    </Box>

    <Grid cols={3} className="max-w-5xl mx-auto">
      {facts.map((fact) => (
        <FeatureCard
          key={fact.title}
          title={fact.title}
          description={fact.description}
        />
      ))}
    </Grid>

    <Typography className="text-center text-xs text-[#555] mt-8">
      Coverage grows with the catalog — components without verified data say
      so, instead of guessing.
    </Typography>
  </Section>
);
