import type { Metadata } from "next";
import { Inter, Manrope, JetBrains_Mono } from "next/font/google";
import { AppRouterCacheProvider } from "@mui/material-nextjs/v15-appRouter";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import { theme } from "@/theme/theme";
import "./globals.css";

// Body / UI text.
const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

// Display headings.
const manrope = Manrope({
  variable: "--font-manrope",
  subsets: ["latin"],
  display: "swap",
});

// Terminals, code, commands.
const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  display: "swap",
});

const TITLE = "Planton — A free Desktop App and CLI for your cloud infrastructure";
const DESCRIPTION =
  "Deploy real infrastructure to your own cloud — without writing Terraform. Planton is a free Desktop App and CLI you download and open: pick a stack and fill a short form, or planton apply -f a manifest, and watch it deploy with clean, auditable infrastructure-as-code underneath.";

export const metadata: Metadata = {
  metadataBase: new URL("https://planton.dev"),
  title: {
    default: TITLE,
    template: "%s | Planton",
  },
  description: DESCRIPTION,
  keywords: [
    "Planton",
    "cloud infrastructure",
    "infrastructure as code",
    "Desktop App",
    "CLI",
    "Terraform",
    "Pulumi",
    "AWS",
    "GCP",
    "Azure",
    "Kubernetes",
  ],
  openGraph: {
    type: "website",
    url: "/",
    title: TITLE,
    description: DESCRIPTION,
    siteName: "Planton",
    images: [{ url: "/og.png", width: 1200, height: 630, alt: "Planton" }],
  },
  twitter: {
    card: "summary_large_image",
    title: TITLE,
    description: DESCRIPTION,
    images: ["/og.png"],
  },
  icons: {
    icon: [
      { url: "/icon.svg", type: "image/svg+xml" },
      { url: "/favicon.ico", sizes: "any" },
    ],
    apple: "/icon.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body
        className={`${inter.variable} ${manrope.variable} ${jetbrainsMono.variable} min-h-screen bg-background text-foreground`}
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
