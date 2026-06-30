import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { AppRouterCacheProvider } from '@mui/material-nextjs/v15-appRouter';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { theme } from '@/theme/theme';
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://planton.dev"),
  title: {
    default: "Planton — Open‑Source Multi‑Cloud Infrastructure Framework",
    template: "%s | Planton",
  },
  description:
    "Kubernetes‑style manifests for multi‑cloud infrastructure. Define once. Deploy anywhere.",
  keywords: [
    "Planton",
    "multi‑cloud",
    "infrastructure",
    "Pulumi",
    "OpenTofu",
    "Kubernetes",
    "Protobuf",
    "Buf",
  ],
  openGraph: {
    type: "website",
    url: "/",
    title: "Planton — Open‑Source Multi‑Cloud Infrastructure Framework",
    description:
      "Kubernetes‑style manifests for multi‑cloud infrastructure. Define once. Deploy anywhere.",
    siteName: "Planton",
    images: [
      { url: "/icon.png", width: 512, height: 512, alt: "Planton" },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Planton — Open‑Source Multi‑Cloud Infrastructure Framework",
    description:
      "Kubernetes‑style manifests for multi‑cloud infrastructure. Define once. Deploy anywhere.",
    images: ["/icon.png"],
  },
  icons: {
    icon: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen bg-slate-950 text-white`}
      >
        <AppRouterCacheProvider>
          <ThemeProvider theme={theme}>
            <CssBaseline />
            {children}
          </ThemeProvider>
        </AppRouterCacheProvider>
      </body>
    </html>
  );
}
