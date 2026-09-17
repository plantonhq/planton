import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { CloudCatalogHero, CloudCatalogCapabilities, CloudCatalogCTA } from '@/components/product/cloud-catalog';

export const metadata = pageMetadata('/features/cloud-catalog');

export default function CloudCatalogPage() {
  return (
    <Box>
      <CloudCatalogHero />
      <CloudCatalogCapabilities />
      <CloudCatalogCTA />
    </Box>
  );
}
