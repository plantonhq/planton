import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/agents');

export default function Page() {
  return <RetiredRoute from="/agents" />;
}
