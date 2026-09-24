import { SolutionsIndex } from '@/components/solutions/SolutionsIndex';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/solutions');

export default function Page() {
  return <SolutionsIndex />;
}
