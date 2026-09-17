import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/features/planton-copilot');

export default function Page() {
  return <RetiredRoute from="/features/planton-copilot" />;
}
