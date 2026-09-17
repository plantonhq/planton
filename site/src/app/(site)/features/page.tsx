import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import {
  ProductOverviewHero,
  ProductModulesGrid,
  ProductJourney,
  ProductArchitecture,
  ProductCTA,
} from '@/components/product/overview';

export const metadata = pageMetadata('/features');

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
