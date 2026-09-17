import { Metadata } from 'next';
import { Box } from '@mui/material';
import {
  ProductOverviewHero,
  ProductModulesGrid,
  ProductJourney,
  ProductArchitecture,
  ProductCTA,
} from '@/components/product/overview';

export const metadata: Metadata = {
  title: 'Product | Planton',
  description:
    'The Self-Service Cloud Platform — AI-designed infrastructure and Git-to-production deployments, in your own cloud account. Deploy to any cloud with an open source foundation and zero vendor lock-in.',
};

export default function ProductOverviewPage() {
  return (
    <Box>
      <ProductOverviewHero />
      <ProductModulesGrid />
      <ProductJourney />
      <ProductArchitecture />
      <ProductCTA />
    </Box>
  );
}
