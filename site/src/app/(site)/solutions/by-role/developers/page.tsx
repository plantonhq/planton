import { Metadata } from 'next';
import DevelopersPage from '@/components/product/solutions/developers';

export const metadata: Metadata = {
  title: 'For Developers | Planton',
  description:
    'Deploy your code, not your weekend. Git-to-deploy, self-service infrastructure, without deep Kubernetes expertise.',
};

export default function Page() {
  return <DevelopersPage />;
}
