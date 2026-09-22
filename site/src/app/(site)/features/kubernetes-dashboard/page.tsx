import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/kubernetes-dashboard');

export default function Page() {
  return <RetiredRoute from="/features/kubernetes-dashboard" />;
}
