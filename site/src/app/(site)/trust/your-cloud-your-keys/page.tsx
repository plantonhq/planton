import { TrustPage } from '@/components/trust/TrustPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/trust/your-cloud-your-keys');

export default function Page() {
  return <TrustPage path="/trust/your-cloud-your-keys" />;
}
