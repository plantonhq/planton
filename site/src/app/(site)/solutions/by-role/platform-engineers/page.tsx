import { pageMetadata } from '@/lib/page-metadata';
import PlatformEngineersPage from '@/components/product/solutions/platform-engineers';

export const metadata = pageMetadata('/solutions/by-role/platform-engineers');

export default function Page() {
  return <PlatformEngineersPage />;
}
