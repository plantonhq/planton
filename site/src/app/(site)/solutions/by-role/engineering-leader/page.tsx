import { pageMetadata } from '@/lib/page-metadata';
import EngineeringLeadersPage from '@/components/product/solutions/engineering-leaders';

export const metadata = pageMetadata('/solutions/by-role/engineering-leader');

export default function Page() {
  return <EngineeringLeadersPage />;
}
