import * as React from "react";
import { Reveal } from "@/components/motion";
import { DownloadButton } from "@/components/chrome";
import { OPERATING_SYSTEMS } from "@/site";

/** The closing call to action. */
export function FinalCta() {
  return (
    <Reveal className="mx-auto max-w-2xl px-4 py-28 text-center sm:px-6 lg:px-8">
      <h2 className="text-balance font-display text-3xl font-bold tracking-tight sm:text-5xl">
        Stop choosing between easy and accountable.
      </h2>
      <p className="mx-auto mt-5 max-w-xl text-muted-foreground">
        Download Planton, open it, and deploy to your own cloud in minutes.
      </p>
      <div className="mt-8 flex flex-col items-center gap-3">
        <DownloadButton size="lg" />
        <p className="text-xs uppercase tracking-[0.14em] text-faint">
          Runs on {OPERATING_SYSTEMS.join(" · ")}
        </p>
      </div>
    </Reveal>
  );
}

export default FinalCta;
