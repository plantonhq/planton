import { pageMetadata } from '@/lib/page-metadata';
import GrowingTeamsPage from '@/components/product/solutions/growing-teams';

export const metadata = pageMetadata('/solutions/by-size/growing-teams');

export default function Page() {
  return <GrowingTeamsPage />;
}
