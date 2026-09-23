import { RetiredRoute } from '@/components/site/RetiredRoute';
import { retiredRouteMetadata } from '@/lib/page-metadata';

export const metadata = retiredRouteMetadata('/hackathon/mobile-vibe-2025');

export default function Page() {
  return <RetiredRoute from="/hackathon/mobile-vibe-2025" />;
}
