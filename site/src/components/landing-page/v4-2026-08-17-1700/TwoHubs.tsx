'use client';

import { Box, Typography } from '@mui/material';
import { FC } from 'react';
import {
  Section,
  SectionTitle,
  SectionSubtitle,
  Card,
  FeatureTitle,
  CheckIcon,
} from './shared';
import { POSITIONING } from '@/data/positioning';
import { PLATFORM_STATS } from '@/data/platform-stats';

/**
 * The two halves of the product, each with its ONE scoped analogy — the
 * Level-2 layer of the positioning rule in src/data/positioning.ts. The
 * analogies live here and only here; the hero above never uses them.
 */

const infraHubPoints = [
  'Describe what you need — it composes on a live canvas',
  'Cost and permissions verified before deploy',
  `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} typed kinds across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers`,
  'Publish as an Infra Chart — a template your team reuses',
];

const serviceHubPoints = [
  'Git push to a production URL',
  'No pipeline YAML, no Dockerfile required',
  'Results written back into GitHub checks and deployments',
  'Approval gates, live logs, full audit history',
];

const HubCard: FC<{
  name: string;
  analogy: string;
  points: string[];
}> = ({ name, analogy, points }) => (
  <Card className="flex-1">
    <FeatureTitle className="mb-1">{name}</FeatureTitle>
    <Typography className="text-sm text-[#a0a0a0] font-medium mb-4">
      {analogy}
    </Typography>
    <Box className="space-y-2.5">
      {points.map((point) => (
        <Box key={point} className="flex items-start gap-2">
          <Box className="mt-0.5 shrink-0">
            <CheckIcon />
          </Box>
          <Typography className="text-sm text-[#b0b0b0]">{point}</Typography>
        </Box>
      ))}
    </Box>
  </Card>
);

export const TwoHubs: FC = () => (
  <Section id="two-hubs">
    <Box className="text-center mb-10">
      <SectionTitle>One Platform, Two Halves</SectionTitle>
      <SectionSubtitle className="mx-auto">
        Infra Hub creates the platform. Service Hub ships your applications
        onto it.
      </SectionSubtitle>
    </Box>

    <Box className="flex flex-col md:flex-row gap-6 max-w-5xl mx-auto">
      <HubCard
        name={POSITIONING.infraHub.name}
        analogy={POSITIONING.infraHub.analogy}
        points={infraHubPoints}
      />
      <HubCard
        name={POSITIONING.serviceHub.name}
        analogy={POSITIONING.serviceHub.analogy}
        points={serviceHubPoints}
      />
    </Box>
  </Section>
);
