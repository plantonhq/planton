import { Metadata } from 'next';
import { Box } from '@mui/material';
import { AgentFleetHero, AgentFleetCapabilities, AgentFleetCTA } from '@/components/product/agent-fleet';

export const metadata: Metadata = {
  title: 'Agent Fleet | Planton',
  description:
    'Purpose-built AI agents for DevOps. Browse the marketplace, encode your runbooks as skills, orchestrate sub-agents, and stream every action in real time.',
};

export default function AgentFleetPage() {
  return (
    <Box>
      <AgentFleetHero />
      <AgentFleetCapabilities />
      <AgentFleetCTA />
    </Box>
  );
}
