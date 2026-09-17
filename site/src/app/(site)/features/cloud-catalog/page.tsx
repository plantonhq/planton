import { Metadata } from 'next';
import { Box } from '@mui/material';
import { CloudCatalogHero, CloudCatalogCapabilities, CloudCatalogCTA } from '@/components/product/cloud-catalog';

export const metadata: Metadata = {
  title: 'Cloud Catalog | Planton',
  description:
    'Browse 700+ pre-built deployment modules across 8 cloud providers. Filter by provider, preview YAML configurations, and deploy to your cloud in minutes.',
};

export default function CloudCatalogPage() {
  return (
    <Box>
      <CloudCatalogHero />
      <CloudCatalogCapabilities />
      <CloudCatalogCTA />
    </Box>
  );
}
