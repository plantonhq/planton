import { pageMetadata } from '@/lib/page-metadata';
import { BookDemoPage } from '@/components/book-demo/BookDemoPage';

export const metadata = pageMetadata('/book-demo');
export default function BookDemoRoute() { return <BookDemoPage />; }
