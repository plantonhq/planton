/**
 * Alias route: /invest/why -> /invest/and-you-get
 *
 * This is a shorter, memorable URL for the "What You Get When You Invest" page.
 * Both URLs serve the same content.
 */
import { Metadata } from 'next';
import { AndYouGetPage } from '@/components/invest/explainer/pages';

export const metadata: Metadata = {
  title: 'What You Get When You Invest | Planton',
  description:
    'What you get when you invest in Planton. SAFE terms, valuation cap, ownership calculator, and transparent risk disclosure.',
  openGraph: {
    title: 'What You Get When You Invest | Planton',
    description:
      'SAFE terms, $7M valuation cap, ownership calculator, and transparent risk disclosure for Planton investors.',
    type: 'website',
  },
};

export default function WhyRoute() {
  return <AndYouGetPage />;
}
