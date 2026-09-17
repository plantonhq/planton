import { Metadata } from 'next';
import StartupsPage from '@/components/product/solutions/startups';

export const metadata: Metadata = {
  title: 'Startups | Planton',
  description:
    'Ship production infrastructure without growing your ops team. Free tier, open-source foundation, no lock-in.',
};

export default function Page() {
  return <StartupsPage />;
}
