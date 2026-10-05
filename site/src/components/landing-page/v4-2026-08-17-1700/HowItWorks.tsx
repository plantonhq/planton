'use client';

import { Box } from '@mui/material';
import { FC } from 'react';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  Step,
  CodeIcon,
  ShieldIcon,
  RocketIcon,
  CloudIcon,
  GitBranchIcon,
} from './shared';

/**
 * The self-service loop, stated as five plain steps. This is the demo
 * sequence and the product thesis in one row: one person composes and
 * verifies; from the publish step on, every deployment is a form fill.
 */

const steps = [
  {
    title: 'Describe',
    description: 'Say what you need in one sentence. The AI composes it from a typed catalog of 700+ kinds.',
    icon: <CodeIcon />,
  },
  {
    title: 'Verify',
    description: 'See the itemized Cloud Bill, the least-privilege IAM policy, and the compliance posture — before anything is created.',
    icon: <ShieldIcon />,
  },
  {
    title: 'Deploy',
    description: 'Tested Terraform and Pulumi modules run in your own cloud account, with live execution progress.',
    icon: <RocketIcon />,
  },
  {
    title: 'Publish',
    description: 'Save it as an Infra Chart — a template your organization owns and governs.',
    icon: <CloudIcon />,
  },
  {
    title: 'Self-Serve',
    description: 'Every developer deploys the same architecture through a form. No Terraform, no tickets.',
    icon: <GitBranchIcon />,
  },
];

export const HowItWorks: FC = () => (
  <Section id="how-it-works">
    <Box className="text-center mb-10">
      <SectionTitle>Compose Once. Publish. Developers Self-Serve.</SectionTitle>
      <SectionSubtitle className="mx-auto">
        A platform engineer builds and verifies the architecture. The whole
        team deploys it from a template.
      </SectionSubtitle>
    </Box>

    <Box className="flex flex-col lg:flex-row gap-8 lg:gap-4">
      {steps.map((step, index) => (
        <Step
          key={step.title}
          number={index + 1}
          title={step.title}
          description={step.description}
          icon={step.icon}
          isLast={index === steps.length - 1}
        />
      ))}
    </Box>
  </Section>
);
