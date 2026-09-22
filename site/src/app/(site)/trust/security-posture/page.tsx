import { TrustPage } from '@/components/trust/TrustPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/trust/security-posture');

export default function Page() {
  return <TrustPage path="/trust/security-posture" />;
}
