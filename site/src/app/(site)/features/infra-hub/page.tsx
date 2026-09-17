import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/infra-hub');

export default function Page() {
  return <RetiredRoute from="/features/infra-hub" />;
}
