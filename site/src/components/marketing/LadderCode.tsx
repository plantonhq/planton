import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";

/** Act 02 — the code: You gains review, Planton names the ceremony it cost. */
export function LadderCode() {
  return (
    <StoryStation index="02" label="The code">
      <div className="space-y-6">
        <Line speaker="You" tone="muted">
          So I moved it all to Terraform. Now every change is a diff I can review.
        </Line>
        <Line speaker="Planton" tone="bright">
          And now a one-line change is a pull request, a plan, and a prayer. It
          works — it&rsquo;s just not how you want to spend your day.
        </Line>
      </div>
    </StoryStation>
  );
}

export default LadderCode;
