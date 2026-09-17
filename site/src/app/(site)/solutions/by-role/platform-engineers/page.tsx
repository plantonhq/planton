import { Metadata } from 'next';
import PlatformEngineersPage from '@/components/product/solutions/platform-engineers';

export const metadata: Metadata = {
  title: 'For Platform Engineers | Planton',
  description:
    'Build golden paths, not bottleneck queues. Define standards, govern credentials, and let developers self-serve.',
};

export default function Page() {
  return <PlatformEngineersPage />;
}
