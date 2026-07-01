import type { Metadata } from "next";
import Link from "next/link";
import { Download } from "lucide-react";
import { SiteHeader, SiteFooter } from "@/components/chrome";
import { Button } from "@/components/ui/button";
import { OPERATING_SYSTEMS, RELEASES_URL } from "@/site";

export const metadata: Metadata = {
  title: "Download",
  description:
    "Download Planton — a free desktop app for your cloud infrastructure. Free forever, no account required.",
};

/** The /download page — the target of every "Download Planton" CTA. */
export default function DownloadPage() {
  return (
    <>
      <SiteHeader />
      <main className="mx-auto max-w-3xl px-4 pb-28 pt-36 sm:px-6 lg:px-8">
        <h1 className="font-display text-4xl font-bold tracking-tight sm:text-5xl">
          Download Planton
        </h1>
        <p className="mt-4 text-lg text-muted-foreground">
          The Desktop App is free, forever — including for commercial use. No
          account, no sign-up.
        </p>

        <section className="mt-12">
          <h2 className="text-sm font-medium uppercase tracking-[0.16em] text-faint">
            Desktop App
          </h2>
          <div className="mt-5 grid gap-3 sm:grid-cols-3">
            {OPERATING_SYSTEMS.map((os) => (
              <Button
                key={os}
                asChild
                variant="secondary"
                className="h-auto justify-start gap-2 rounded-lg border border-border py-3"
              >
                <Link href={RELEASES_URL} target="_blank" rel="noreferrer">
                  <Download className="size-4" />
                  <span>Download for {os}</span>
                </Link>
              </Button>
            ))}
          </div>
        </section>

      </main>
      <SiteFooter />
    </>
  );
}
