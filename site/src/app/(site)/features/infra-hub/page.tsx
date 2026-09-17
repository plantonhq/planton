import { Metadata } from 'next';
import { Box } from '@mui/material';
import { InfraHubHero, InfraHubCapabilities, InfraHubCTA } from '@/components/product/infra-hub';

export const metadata: Metadata = {
  title: 'Infra Hub | Planton',
  description:
    'Deploy any cloud resource — from databases to Kubernetes clusters — in minutes. 700+ resource types, Infra Charts, dependency-aware pipelines, and auditable stack jobs.',
};

export default function InfraHubPage() {
  return (
    <Box>
      <InfraHubHero />
      <InfraHubCapabilities />
      <InfraHubCTA />
    </Box>
  );
}
