import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/solutions/by-role/startup-founders');

export default function Page() {
  return <RetiredRoute from="/solutions/by-role/startup-founders" />;
}
