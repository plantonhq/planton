import * as React from "react";
import { Reveal } from "@/components/motion";
import { DownloadButton } from "@/components/chrome";
import { OPERATING_SYSTEMS } from "@/site";

/** The closing call to action — the "make the easy thing the right thing" thesis, paid off. */
export function FinalCta() {
  return (
    <Reveal className="mx-auto max-w-2xl px-4 py-28 text-center sm:px-6 lg:px-8">
      <p className="font-display text-lg italic text-muted-foreground sm:text-xl">
        &ldquo;Make the easy thing the right thing.&rdquo;
      </p>
      <h2 className="mt-4 text-balance font-display text-3xl font-bold tracking-tight sm:text-5xl">
        So we built Planton.
      </h2>
      <p className="mx-auto mt-5 max-w-xl text-muted-foreground">
        The ease of a console, the rigor of code — no trade-off. Download it and
        deploy to your own cloud in minutes.
      </p>
      <div className="mt-8 flex flex-col items-center gap-3">
        <DownloadButton size="lg" />
        <p className="text-xs tracking-[0.08em] text-faint">
          Runs on {OPERATING_SYSTEMS.join(" · ")}
        </p>
      </div>
    </Reveal>
  );
}

export default FinalCta;
