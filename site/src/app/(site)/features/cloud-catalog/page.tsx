import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/cloud-catalog');

export default function Page() {
  return <RetiredRoute from="/features/cloud-catalog" />;
}
