import { ComparePage } from '@/components/compare/ComparePage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/compare');

export default function Page() {
  return <ComparePage />;
}
