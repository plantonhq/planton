import { pageMetadata } from '@/lib/page-metadata';
import EnterprisesPage from '@/components/product/solutions/enterprises';

export const metadata = pageMetadata('/solutions/by-size/enterprises');

export default function Page() {
  return <EnterprisesPage />;
}
