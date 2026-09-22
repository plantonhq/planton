import { Box, Typography } from '@mui/material';
import type { FC } from 'react';
import { TerminalWindow } from '@/components/marketing';

/**
 * A still frame of the moment the story is about: one sentence asked for,
 * the components it became, and the three verdicts stamped before anything
 * exists. Every figure here is an illustration and says so ("est."); the
 * real ones come from the verified catalog at deploy time. It does not
 * animate: the proof is the resting frame, and a page that depends on a
 * clock cannot be compared between two builds or read by a person who asked
 * for reduced motion.
 */

const PROMPT = 'A production environment on AWS: private VPC, ECS services behind a load balancer, RDS PostgreSQL encrypted with our own KMS key.';

const COMPOSED = [
  { name: 'AWS VPC', cost: 'No charge' },
  { name: 'NAT Gateway', cost: '~$33/mo est.' },
  { name: 'Application Load Balancer', cost: '~$16/mo est.' },
  { name: 'ECS Service', cost: 'Usage-based' },
  { name: 'RDS PostgreSQL (Multi-AZ)', cost: '~$122/mo est.' },
  { name: 'KMS Key', cost: '$1/mo est.' },
];

const VERDICTS = [
  { label: 'verified cost', value: '~$172/mo est. \u00b7 5 of 6 priced' },
  { label: 'permissions', value: 'least-privilege policy \u00b7 14 actions' },
  { label: 'controls', value: 'encryption at rest \u00b7 encryption in transit \u00b7 no public exposure' },
];

export const ProofMoment: FC = () => (
  <TerminalWindow title="Planton Desktop" className="text-left">
    <Box className="rounded-lg border border-edge bg-panel px-4 py-3 mb-4">
      <Typography className="text-xs text-fg-faint mb-1">What do you want to build?</Typography>
      <Typography className="text-sm text-fg font-mono">{PROMPT}</Typography>
    </Box>

    <Box className="grid grid-cols-2 sm:grid-cols-3 gap-2 mb-4">
      {COMPOSED.map((component) => (
        <Box key={component.name} className="rounded-lg border border-edge bg-card px-3 py-2">
          <Typography className="text-xs text-fg font-medium leading-tight">{component.name}</Typography>
          <Typography className="text-[11px] text-fg-muted mt-0.5">{component.cost}</Typography>
        </Box>
      ))}
    </Box>

    <Box className="grid grid-cols-1 sm:grid-cols-[12rem_1fr] gap-x-4 gap-y-1.5 border-t border-edge pt-4">
      {VERDICTS.map((verdict) => (
        <Box key={verdict.label} className="contents">
          <Typography className="text-xs text-fg-muted font-mono">{verdict.label}</Typography>
          <Typography className="text-xs text-fg font-mono">{verdict.value}</Typography>
        </Box>
      ))}
    </Box>
    <Typography className="text-xs text-fg-muted mt-4">
      An illustration of the shape. Figures marked est. are examples; a real deploy carries its own.
    </Typography>
  </TerminalWindow>
);
