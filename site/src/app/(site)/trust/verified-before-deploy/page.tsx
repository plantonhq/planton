import { TrustPage } from '@/components/trust/TrustPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/trust/verified-before-deploy');

export default function Page() {
  return <TrustPage path="/trust/verified-before-deploy" />;
}
