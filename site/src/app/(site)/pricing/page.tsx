import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import {
  EnterpriseBand,
  Faqs,
  PlanGrid,
  PricingCta,
  PricingHero,
  RunnersNote,
  ValueMatrix,
} from '@/components/pricing';
import { MarketProvider } from '@/components/market';

export const metadata = pageMetadata('/pricing');

export default function PricingPage() {
  return (
    <Box>
      <MarketProvider>
        <PricingHero />
        <PlanGrid />
        <RunnersNote />
        <EnterpriseBand />
        <ValueMatrix />
        <Faqs />
        <PricingCta />
      </MarketProvider>
    </Box>
  );
}
