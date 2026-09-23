import { ProductIndex } from '@/components/product/ProductIndex';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/product');

export default function Page() {
  return <ProductIndex />;
}
