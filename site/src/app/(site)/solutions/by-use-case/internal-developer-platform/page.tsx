import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/solutions/by-use-case/internal-developer-platform');

export default function Page() {
  return <RetiredRoute from="/solutions/by-use-case/internal-developer-platform" />;
}
