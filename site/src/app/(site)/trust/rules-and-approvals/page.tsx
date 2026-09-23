import { TrustPage } from '@/components/trust/TrustPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/trust/rules-and-approvals');

export default function Page() {
  return <TrustPage path="/trust/rules-and-approvals" />;
}
