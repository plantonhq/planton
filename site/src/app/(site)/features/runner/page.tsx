import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { RunnerHero, RunnerCapabilities, RunnerCTA } from '@/components/product/runner';

export const metadata = pageMetadata('/features/runner');

export default function RunnerPage() {
  return (
    <Box>
      <RunnerHero />
      <RunnerCapabilities />
      <RunnerCTA />
    </Box>
  );
}
