/**
 * Alias route: /invest/steps -> /invest/process
 *
 * This is a shorter, memorable URL for the "Investment Process" page.
 * Both URLs serve the same content.
 */
import { Metadata } from 'next';
import { ProcessPage } from '@/components/invest/process/ProcessPage';

export const metadata: Metadata = {
  title: 'Investment Process | Planton',
  description:
    'How to invest in Planton. Step-by-step guide using YC SAFE and Carta for professional, transparent investment processing.',
  openGraph: {
    title: 'Investment Process | Planton',
    description:
      'Step-by-step guide to investing in Planton using YC SAFE and Carta. Professional, transparent, industry-standard process.',
    type: 'website',
  },
};

export default function StepsRoute() {
  return <ProcessPage />;
}
