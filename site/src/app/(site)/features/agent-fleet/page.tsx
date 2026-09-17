import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { AgentFleetHero, AgentFleetCapabilities, AgentFleetCTA } from '@/components/product/agent-fleet';

export const metadata = pageMetadata('/features/agent-fleet');

export default function AgentFleetPage() {
  return (
    <Box>
      <AgentFleetHero />
      <AgentFleetCapabilities />
      <AgentFleetCTA />
    </Box>
  );
}
