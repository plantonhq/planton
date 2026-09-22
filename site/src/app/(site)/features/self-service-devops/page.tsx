import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/self-service-devops');

export default function Page() {
  return <RetiredRoute from="/features/self-service-devops" />;
}
