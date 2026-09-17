import { Metadata } from 'next';
import { Box } from '@mui/material';
import { RunnerHero, RunnerCapabilities, RunnerCTA } from '@/components/product/runner';

export const metadata: Metadata = {
  title: 'Runner | Planton',
  description:
    'Self-hosted execution agent that runs in your cloud. Planton orchestrates, Runner executes — your credentials never leave your account.',
};

export default function RunnerPage() {
  return (
    <Box>
      <RunnerHero />
      <RunnerCapabilities />
      <RunnerCTA />
    </Box>
  );
}
