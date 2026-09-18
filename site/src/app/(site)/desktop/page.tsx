import { Box } from '@mui/material';
import { LandingHero, LandingCapabilities, LandingCTA } from '@/components/product/desktop';
import { DESKTOP_LANDING_PATH } from '@/data/desktop-download';
import { SITE, sitePage } from '@/data/site-pages';
import { pageMetadata } from '@/lib/page-metadata';

// The desktop landing keeps its own Open Graph poster (captured by the
// harness's `desktop-og` scene); everything else comes from the registry.
const page = sitePage(DESKTOP_LANDING_PATH);
export const metadata = pageMetadata(DESKTOP_LANDING_PATH, {
  openGraph: {
    type: 'website',
    siteName: SITE.name,
    url: `${SITE.url}${DESKTOP_LANDING_PATH}`,
    title: 'Planton Desktop: the whole platform on your laptop, free',
    description: page.description,
    images: [{ url: '/_site/images/og/desktop.png', width: 1200, height: 630, alt: 'Planton Desktop' }],
  },
  twitter: { card: 'summary_large_image' },
});

export default function DesktopPage() {
  return (
    <Box>
      <LandingHero />
      <LandingCapabilities />
      <LandingCTA />
    </Box>
  );
}
