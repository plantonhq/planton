import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { ThemeProvider } from '@/providers/theme';
import { GoogleAnalytics } from '@next/third-parties/google';
import { HANDOFF_CAPTURE_SCRIPT } from '@/lib/console-handoff';
import { SITE, sitePage } from '@/data/site-pages';
import { POSITIONING } from '@/data/positioning';

const inter = Inter({
  weight: ['300', '400', '500', '600', '700'],
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
});

// The defaults every page inherits until it sets its own (registered pages do,
// through pageMetadata). The words are the home page's, from the route registry.
const home = sitePage('/');
const siteTitle = `${SITE.name}: ${home.title}`;

export const metadata: Metadata = {
  metadataBase: new URL(SITE.url),
  applicationName: SITE.name,
  icons: { icon: '/favicon.ico' },
  title: siteTitle,
  description: home.description,
  openGraph: {
    siteName: SITE.name,
    type: 'website',
    url: SITE.url,
    title: siteTitle,
    description: home.description,
    images: [{ url: SITE.defaultOgImage, width: 1200, height: 630 }],
  },
};

/**
 * Structured data for the organization and the product, read by search
 * engines and by agents. Facts come from the registry and the positioning
 * vocabulary; nothing here is typed twice.
 */
const structuredData = [
  {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: 'Planton Cloud, Inc.',
    alternateName: SITE.name,
    url: SITE.url,
    logo: `${SITE.url}/_site/images/og/desktop.png`,
    sameAs: ['https://github.com/plantonhq', 'https://discord.gg/pwcSapdQAp'],
  },
  {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: SITE.name,
    url: SITE.url,
    description: POSITIONING.umbrella.sentence,
    applicationCategory: 'DeveloperApplication',
    operatingSystem: 'Web, macOS, Windows, Linux',
    offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
    publisher: { '@type': 'Organization', name: 'Planton Cloud, Inc.', url: SITE.url },
  },
];

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-screen">
      {process.env.NODE_ENV === 'production' && <GoogleAnalytics gaId="G-VWZNWQPEJ0" />}

      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: 'try{if(window.location.hostname==="planton.ai"&&window.location.pathname==="/"&&!window.location.search&&/planton_logged_in=/.test(document.cookie))window.location.replace("/dashboard")}catch(e){}',
          }}
        />
        {/* A console's email handoff (`#email=`) is captured into session
            storage and stripped from the address bar here, synchronously,
            before the analytics script below can record a page URL. */}
        <script dangerouslySetInnerHTML={{ __html: HANDOFF_CAPTURE_SCRIPT }} />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Material+Symbols+Rounded:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&display=optional"
        />
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }} />
      </head>
      {/* The body declares the canvas and the primary text color from the palette so nothing
          inherits a theme default by accident; every surface below reads the same two tokens. */}
      <body className={`${inter.variable} antialiased h-screen bg-cover bg-center bg-canvas text-fg`}>
        <ThemeProvider>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
