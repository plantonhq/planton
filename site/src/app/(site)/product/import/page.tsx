import { ProductPage } from '@/components/product/ProductPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/product/import');

export default function Page() {
  return <ProductPage path="/product/import" />;
}
