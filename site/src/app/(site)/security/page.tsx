import { SecurityPage } from '@/components/security/SecurityPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/security');

export default function Page() {
  return <SecurityPage />;
}
