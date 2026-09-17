'use client';

import { Box, Typography } from '@mui/material';
import {
  Hub as InfraHubIcon,
  RocketLaunch as ServiceHubIcon,
  Shield as SecurityIcon,
  Psychology as AgentFleetIcon,
  LinkOutlined as ConnectIcon,
} from '@mui/icons-material';
import { Section, SectionSubtitle } from '@/components/marketing';
import { FlowSteps } from '@/components/product/shared';
import type { FlowStep } from '@/components/product/shared';

const journeySteps: FlowStep[] = [
  {
    icon: <ConnectIcon sx={{ fontSize: 20 }} />,
    label: 'Connect Cloud',
    sublabel: 'AWS, GCP, or Azure',
  },
  {
    icon: <InfraHubIcon sx={{ fontSize: 20 }} />,
    label: 'Deploy Infra',
    sublabel: 'Infra Hub',
  },
  {
    icon: <ServiceHubIcon sx={{ fontSize: 20 }} />,
    label: 'Ship Code',
    sublabel: 'Service Hub',
  },
  {
    icon: <SecurityIcon sx={{ fontSize: 20 }} />,
    label: 'Secure',
    sublabel: 'Security',
  },
  {
    icon: <AgentFleetIcon sx={{ fontSize: 20 }} />,
    label: 'Automate',
    sublabel: 'Agent Fleet',
  },
];

export const ProductJourney = () => {
  return (
    <Section>
      <Box className="text-center mb-8">
        <Typography className="text-lg font-semibold text-white mb-2">
          A typical journey
        </Typography>
        <SectionSubtitle className="mx-auto !text-sm">
          From cloud credentials to production workloads in a single workflow.
        </SectionSubtitle>
      </Box>
      <FlowSteps steps={journeySteps} />
    </Section>
  );
};
