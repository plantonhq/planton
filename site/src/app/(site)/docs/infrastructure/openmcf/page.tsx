import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/docs/infrastructure/openmcf');

export default function Page() {
  return <RetiredRoute from="/docs/infrastructure/openmcf" />;
}
