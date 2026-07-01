import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";
import { DeploysTo } from "@/components/marketing/DeploysTo";

/** Act 08 — the payoff: real IaC the whole way down, on your own cloud. */
export function Payoff() {
  return (
    <StoryStation index="08" label="The payoff" marker="active" wide>
      <span className="inline-flex items-center gap-2 rounded-full border border-border bg-secondary px-3 py-1 font-mono text-xs text-muted-foreground">
        <span className="size-2 rounded-full bg-success" />
        all systems running
      </span>
      <Line speaker="Planton" tone="bright" className="mt-6 max-w-xl">
        Either way, it&rsquo;s real infrastructure-as-code the whole way down —
        stored, versioned, every change a diff. On your own cloud. Nothing to lock
        you in.
      </Line>
      <DeploysTo />
    </StoryStation>
  );
}

export default Payoff;
