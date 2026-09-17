import { Metadata } from 'next';
import { Box } from '@mui/material';
import {
  EnterpriseHero,
  EnterpriseHowItWorks,
  EnterpriseRateCards,
} from '@/components/enterprise';
import { PricingCta } from '@/components/pricing';
import { MarketProvider } from '@/components/market';

export const metadata: Metadata = {
  title: 'Enterprise | Planton',
  description:
    'Enterprise at Planton: a published rate card in your market, enterprise identity, air-gap, compliance reporting, and real SLAs. Self-serve below 25 seats — no sales call required.',
};

export default function EnterprisePage() {
  return (
    <Box>
      {/* One shared market fact for the whole page: the hero's selector,
          the rate cards, and the buying steps flip together — without the
          provider each component would hold its own fallback market state
          and a switch would flip only the hero. */}
      <MarketProvider>
        <EnterpriseHero />
        <EnterpriseRateCards />
        <EnterpriseHowItWorks />
        <PricingCta />
      </MarketProvider>
    </Box>
  );
}
