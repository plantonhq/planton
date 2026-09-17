import { pageMetadata } from '@/lib/page-metadata';
import StartupsPage from '@/components/product/solutions/startups';

export const metadata = pageMetadata('/solutions/by-size/startups');

export default function Page() {
  return <StartupsPage />;
}
