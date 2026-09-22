import { ProductPage } from '@/components/product/ProductPage';
import { pageMetadata } from '@/lib/page-metadata';

export const metadata = pageMetadata('/product/coding-agents');

export default function Page() {
  return <ProductPage path="/product/coding-agents" />;
}
