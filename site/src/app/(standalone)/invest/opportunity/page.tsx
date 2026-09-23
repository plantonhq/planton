import { Metadata } from 'next';
import { OpportunityPage } from '@/components/invest/opportunity/OpportunityPage';

export const metadata: Metadata = {
  title: 'The Opportunity | Planton Investment',
  description:
    'Where Planton sits in the DevOps and platform engineering market. Competitor analysis, market context, and our positioning.',
  openGraph: {
    title: 'The Opportunity | Planton Investment',
    description:
      'Where Planton sits in the DevOps and platform engineering market. Competitor analysis, market context, and our positioning.',
    type: 'website',
  },
};

export default function OpportunityRoute() {
  return <OpportunityPage />;
}
