import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/iac-workflows');

export default function Page() {
  return <RetiredRoute from="/features/iac-workflows" />;
}
