import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Speaker } from "@/components/marketing/Speaker";
import { InlineCode } from "@/components/marketing/InlineCode";
import { Terminal } from "@/components/showcase";
import type { TerminalLineData } from "@/components/showcase";

// A whole environment via `planton chart install` — the `helm install` parallel.
const CHART_INSTALL: TerminalLineData[] = [
  { kind: "command", text: "planton chart install aws-ecs --name api --env dev --values values.yaml" },
  { kind: "output", text: "→ installing chart · aws-ecs" },
  { kind: "resource", name: "network.vpc", status: "created · 4.2s", state: "done" },
  { kind: "resource", name: "subnet.private", status: "created · 2.1s", state: "done" },
  { kind: "resource", name: "ecr.repository", status: "created · 3.8s", state: "done" },
  { kind: "resource", name: "ecs.service", status: "creating…", state: "running" },
];

/** Act 05 — the CLI answer: a whole environment coming up live in the terminal. */
export function CliAnswer() {
  return (
    <StoryStation index="05" label="The CLI answer" wide>
      <Speaker>The CLI</Speaker>
      <p className="max-w-xl text-xl leading-relaxed text-foreground sm:text-2xl">
        Run it from your terminal and watch the whole environment come up, live —
        every resource, as it happens. Not fire-and-forget like{" "}
        <InlineCode>kubectl apply</InlineCode>.
      </p>
      <div className="mt-8 max-w-2xl">
        <Terminal lines={CHART_INSTALL} />
      </div>
    </StoryStation>
  );
}

export default CliAnswer;
