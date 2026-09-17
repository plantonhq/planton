import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { InfraHubHero, InfraHubCapabilities, InfraHubCTA } from '@/components/product/infra-hub';

export const metadata = pageMetadata('/features/infra-hub');

export default function InfraHubPage() {
  return (
    <Box>
      <InfraHubHero />
      <InfraHubCapabilities />
      <InfraHubCTA />
    </Box>
  );
}
