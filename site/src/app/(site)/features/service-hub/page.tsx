import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { ServiceHubHero, ServiceHubCapabilities, ServiceHubCTA } from '@/components/product/service-hub';

export const metadata = pageMetadata('/features/service-hub');

export default function ServiceHubPage() {
  return (
    <Box>
      <ServiceHubHero />
      <ServiceHubCapabilities />
      <ServiceHubCTA />
    </Box>
  );
}
