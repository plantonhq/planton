import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";
import { ArchitecturePreview } from "@/components/marketing/ArchitecturePreview";

/** Act 06 — the desktop answer: the dashboard Kubernetes never had. */
export function DesktopAnswer() {
  return (
    <StoryStation index="06" label="The desktop answer" wide>
      <div className="max-w-xl space-y-6">
        <Line speaker="You" tone="muted">
          …and when I&rsquo;d rather click?
        </Line>
        <Line speaker="Planton" tone="bright">
          Open the app — the dashboard Kubernetes never had. See the architecture
          before you deploy, then watch each piece light up as it comes online.
        </Line>
      </div>
      <div className="mt-8 max-w-3xl">
        <ArchitecturePreview />
      </div>
    </StoryStation>
  );
}

export default DesktopAnswer;
