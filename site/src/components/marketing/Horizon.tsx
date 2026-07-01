import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";

/** Act 09 — on the horizon (honestly labeled future work). */
export function Horizon() {
  return (
    <StoryStation index="09" label="On the horizon">
      <p className="text-lg leading-relaxed text-muted-foreground sm:text-xl">
        Soon: point Planton at a cloud you already use, and it brings what&rsquo;s
        already there under management. Describe what you want, and watch it
        assemble.
      </p>
    </StoryStation>
  );
}

export default Horizon;
