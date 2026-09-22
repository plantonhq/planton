import type { Metadata } from 'next';
import { BookDemoPage } from '@/components/book-demo/BookDemoPage';

export const metadata: Metadata = {
  title: 'Book a Demo — Planton',
  description:
    'See how teams deploy production infrastructure and ship services across AWS, GCP, and Azure in minutes — without a dedicated DevOps team. Book a walkthrough with the Planton team.',
  openGraph: {
    title: 'Book a Demo — Planton',
    description:
      'See how teams deploy production infrastructure and ship services across AWS, GCP, and Azure in minutes — without a dedicated DevOps team.',
    url: 'https://planton.ai/book-demo',
  },
};

export default function BookDemoRoute() {
  return <BookDemoPage />;
}
