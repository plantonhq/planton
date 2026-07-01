import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Speaker } from "@/components/marketing/Speaker";
import { InlineCode } from "@/components/marketing/InlineCode";

/** The hinge: Planton speaks and makes the easy way and the right way one thing. */
export function Origin() {
  return (
    <StoryStation index="04" label="The origin">
      <Speaker>Planton</Speaker>
      <p className="text-2xl font-medium leading-snug text-foreground sm:text-3xl">
        For years the easy way and the right way were different things. We
        couldn&rsquo;t accept that. So we made the easy thing the right thing — and
        built Planton.
      </p>
      <p className="mt-8 text-xl leading-relaxed text-foreground sm:text-2xl">
        Create like a console: pick a stack, fill a short form. Planton provisions
        it as real Terraform or Pulumi you own — with state and history. And you
        manage your whole cloud the way you manage Kubernetes:{" "}
        <InlineCode>planton apply -f</InlineCode> for a single manifest,{" "}
        <InlineCode>planton chart install</InlineCode> for a whole environment. For
        AWS, GCP, Azure, and Kubernetes — with the UI Kubernetes never had.
      </p>
    </StoryStation>
  );
}

export default Origin;
