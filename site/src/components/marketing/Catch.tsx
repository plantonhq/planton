import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";

/** Act 10 — the catch: free forever for solo devs, a dry reassurance. */
export function Catch() {
  return (
    <StoryStation index="10" label="The catch">
      <div className="space-y-6">
        <Line speaker="You" tone="muted">
          Okay — so what&rsquo;s the catch?
        </Line>
        <Line speaker="Planton" tone="bright">
          Honestly? We&rsquo;d rather you ship than reach for your credit card.
          Planton is free forever for solo developers and indie hackers — commercial
          use included. We make our money from teams and companies, not from you.
        </Line>
      </div>
    </StoryStation>
  );
}

export default Catch;
