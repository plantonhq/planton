import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";

/** Act 07 — why you can trust it: vetted modules, not a blank file. */
export function TrustModules() {
  return (
    <StoryStation index="07" label="Why you can trust it">
      <Line speaker="Planton" tone="bright">
        And you&rsquo;re never starting from a blank file. Every module is already
        written — and vetted for secure, well-architected, cost-efficient
        infrastructure. Sure, Claude can write you Terraform in minutes. But writing
        it was never the hard part. Trusting it is.
      </Line>
    </StoryStation>
  );
}

export default TrustModules;
