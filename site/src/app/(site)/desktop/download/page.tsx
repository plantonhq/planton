import { Box } from '@mui/material';
import { DesktopDownload } from '@/components/desktop/DesktopDownload';
import { DESKTOP_DOWNLOAD_PATH, DESKTOP_PLATFORMS, DOWNLOADS_LATEST } from '@/data/desktop-download';
import { DESKTOP_RELEASE } from '@/data/desktop-release';
import { SITE, sitePage } from '@/data/site-pages';
import { pageMetadata } from '@/lib/page-metadata';

// The download page keeps its own Open Graph poster (captured by the
// harness's `download-og` scene); everything else comes from the registry.
const page = sitePage(DESKTOP_DOWNLOAD_PATH);
export const metadata = pageMetadata(DESKTOP_DOWNLOAD_PATH, {
  openGraph: {
    type: 'website',
    siteName: SITE.name,
    url: `${SITE.url}${DESKTOP_DOWNLOAD_PATH}`,
    title: page.title,
    description: page.description,
    images: [{ url: '/_site/images/og/download.png', width: 1200, height: 630, alt: 'Download Planton Desktop' }],
  },
  twitter: { card: 'summary_large_image' },
});

// Structured data so a search result for "planton desktop download" carries
// the platforms, the price (none), and the direct download URL. Built from the
// same data module as the buttons, so it cannot name a platform the page does
// not offer.
const softwareApplication = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  name: 'Planton Desktop',
  applicationCategory: 'DeveloperApplication',
  operatingSystem: DESKTOP_PLATFORMS.filter((p) => p.available)
    .map((p) => p.name)
    .join(', '),
  ...(DESKTOP_RELEASE.version ? { softwareVersion: DESKTOP_RELEASE.version.replace(/^v/, '') } : {}),
  downloadUrl: `${DOWNLOADS_LATEST}/`,
  offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
  url: `${SITE.url}${DESKTOP_DOWNLOAD_PATH}`,
  description: page.description,
};

export default function DownloadPage() {
  return (
    <Box>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareApplication) }} />
      <DesktopDownload />
    </Box>
  );
}
