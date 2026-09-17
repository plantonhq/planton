import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/cli');

export default function Page() {
  return <RetiredRoute from="/cli" />;
}
