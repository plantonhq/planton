import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/demo');

export default function Page() {
  return <RetiredRoute from="/demo" />;
}
