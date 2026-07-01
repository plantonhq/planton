import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";
import { InlineCode } from "@/components/marketing/InlineCode";

/** Act 03 — the peak: You loves Kubernetes' two gifts, Planton names the one catch. */
export function LadderKubernetes() {
  return (
    <StoryStation index="03" label="Kubernetes">
      <div className="space-y-6">
        <Line speaker="You" tone="muted">
          Then Kubernetes showed me something better. Write a manifest,{" "}
          <InlineCode>kubectl apply -f</InlineCode>, done — and it&rsquo;s saved to
          disk. Helm charts go further: a whole stack in one file. It&rsquo;s the
          best infrastructure has ever felt.
        </Line>
        <Line speaker="Planton" tone="bright">
          But all of it only works for Kubernetes. No UI. And it was never really
          managing your cloud — no real state, no history.
        </Line>
      </div>
    </StoryStation>
  );
}

export default LadderKubernetes;
