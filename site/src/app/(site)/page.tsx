import { pageMetadata } from '@/lib/page-metadata';
import { Homepage } from '@/components/landing-page';
import { HOMEPAGE } from '@/data/homepage';

export const metadata = pageMetadata('/', {
  openGraph: { type: 'website', url: 'https://planton.ai', siteName: 'Planton', title: HOMEPAGE.title, description: HOMEPAGE.description, images: [{ url: 'https://planton.ai/_site/images/og/homepage.png', width: 1200, height: 630 }] },
  twitter: { card: 'summary_large_image', title: HOMEPAGE.title, description: HOMEPAGE.description, images: ['https://planton.ai/_site/images/og/homepage.png'] },
});

export default Homepage;
