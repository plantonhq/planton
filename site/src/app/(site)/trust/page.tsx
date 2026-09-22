import { TrustIndex } from '@/components/trust/TrustIndex';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/trust');

export default function Page() {
  return <TrustIndex />;
}
