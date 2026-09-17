import { Metadata } from 'next';
import EngineeringLeadersPage from '@/components/product/solutions/engineering-leaders';

export const metadata: Metadata = {
  title: 'For Engineering Leaders | Planton',
  description:
    'Visibility without micromanagement. Audit trails, team autonomy with guardrails, AI operational intelligence.',
};

export default function Page() {
  return <EngineeringLeadersPage />;
}
