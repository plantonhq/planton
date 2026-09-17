import { DistributionPage } from '@/components/distributions/DistributionPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/distributions/hosted');

export default function Page() {
  return <DistributionPage path="/distributions/hosted" />;
}
