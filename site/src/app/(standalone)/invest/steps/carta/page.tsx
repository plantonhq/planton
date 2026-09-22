import { Metadata } from 'next';
import { CartaWalkthroughPage } from '@/components/invest/carta-walkthrough/CartaWalkthroughPage';

export const metadata: Metadata = {
  title: 'Carta Walkthrough | Planton Investment',
  description:
    'See the actual SAFE creation process on Carta. Step-by-step screenshots showing cap table, terms configuration, Mercury bank integration, and investor portal.',
  openGraph: {
    title: 'Carta Walkthrough | Planton Investment',
    description:
      'See the actual SAFE creation process on Carta. Step-by-step screenshots showing how your investment is professionally managed.',
    type: 'website',
  },
};

export default function CartaWalkthroughRoute() {
  return <CartaWalkthroughPage />;
}
