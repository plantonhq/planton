/**
 * Alias route: /invest/if -> /invest/if-you-are
 *
 * This is a shorter, memorable URL for the "What We're Looking For" page.
 * Both URLs serve the same content.
 */
import { Metadata } from 'next';
import { IfYouArePage } from '@/components/invest/explainer/pages';

export const metadata: Metadata = {
  title: "What We're Looking For in an Investor | Planton",
  description:
    'What Planton looks for in an investor. Our ideal partner profile, check sizes, expectations, and how to move forward.',
  openGraph: {
    title: "What We're Looking For in an Investor | Planton",
    description:
      'Our ideal investor profile, check sizes ($15K-$350K), expectations, and how to partner with Planton.',
    type: 'website',
  },
};

export default function IfRoute() {
  return <IfYouArePage />;
}
