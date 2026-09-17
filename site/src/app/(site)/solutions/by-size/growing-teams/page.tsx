import { Metadata } from 'next';
import GrowingTeamsPage from '@/components/product/solutions/growing-teams';

export const metadata: Metadata = {
  title: 'Growing Teams | Planton',
  description:
    'Scale your infrastructure without scaling your ops team. Self-service, standards enforcement, and team visibility.',
};

export default function Page() {
  return <GrowingTeamsPage />;
}
