import { Metadata } from 'next';
import EnterprisesPage from '@/components/product/solutions/enterprises';

export const metadata: Metadata = {
  title: 'Enterprises | Planton',
  description:
    'Enterprise controls without enterprise friction. Runner security, compliance-ready, multi-cloud governance.',
};

export default function Page() {
  return <EnterprisesPage />;
}
