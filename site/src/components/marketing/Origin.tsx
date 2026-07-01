import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Speaker } from "@/components/marketing/Speaker";
import { InlineCode } from "@/components/marketing/InlineCode";

/** Act 04 — the hinge: Planton speaks and frees both gifts for every cloud. */
export function Origin() {
  return (
    <StoryStation index="04" label="The origin">
      <Speaker>Planton</Speaker>
      <p className="text-2xl font-medium leading-snug text-foreground sm:text-3xl">
        We loved all of it too — the apply, the charts. We just couldn&rsquo;t accept
        that the best experience in infrastructure stopped at Kubernetes. So we
        brought it to your whole cloud.
      </p>
      <p className="mt-8 text-xl leading-relaxed text-foreground sm:text-2xl">
        That&rsquo;s Planton. <InlineCode>planton apply -f</InlineCode> for a single
        manifest. <InlineCode>planton chart install</InlineCode> for a whole
        environment. The gestures you already know — for AWS, GCP, Azure, and
        Kubernetes. With the UI Kubernetes never had, and real
        infrastructure-as-code underneath.
      </p>
    </StoryStation>
  );
}

export default Origin;
