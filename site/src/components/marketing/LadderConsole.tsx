import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";

/** Act 01 — the opener, then the console: You loves it, Planton names the cost. */
export function LadderConsole() {
  return (
    <StoryStation index="01" label="The console">
      <Line tone="bright">
        Every way of running cloud infrastructure has asked you to give something up.
      </Line>
      <div className="mt-10 space-y-6">
        <Line speaker="You" tone="muted">
          The cloud console is easy — I just fill a form and go.
        </Line>
        <Line speaker="Planton" tone="bright">
          But nothing&rsquo;s written down. No history. No review. No source of truth.
        </Line>
      </div>
    </StoryStation>
  );
}

export default LadderConsole;
