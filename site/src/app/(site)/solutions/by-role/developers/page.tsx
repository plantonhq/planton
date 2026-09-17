import { pageMetadata } from '@/lib/page-metadata';
import DevelopersPage from '@/components/product/solutions/developers';

export const metadata = pageMetadata('/solutions/by-role/developers');

export default function Page() {
  return <DevelopersPage />;
}
