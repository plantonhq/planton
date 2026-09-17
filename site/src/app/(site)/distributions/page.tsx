import { DistributionsIndex } from '@/components/distributions/DistributionsIndex';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/distributions');

export default function Page() {
  return <DistributionsIndex />;
}
