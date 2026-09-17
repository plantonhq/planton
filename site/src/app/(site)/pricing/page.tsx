import { Metadata } from 'next';
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

export const metadata: Metadata = {
  title: 'Pricing | Planton',
  description:
    'Plans for every stage — on Planton.ai or your own infrastructure. A free tier that never bills, one team plan, run-it-yourself free forever with paid licenses, and on Planton.ai an AI assistant on prepaid credits with spend protection.',
};

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
