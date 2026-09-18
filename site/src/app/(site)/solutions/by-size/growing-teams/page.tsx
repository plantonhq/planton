import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/solutions/by-size/growing-teams');

export default function Page() {
  return <RetiredRoute from="/solutions/by-size/growing-teams" />;
}
