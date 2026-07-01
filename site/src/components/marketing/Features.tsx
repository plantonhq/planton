import * as React from "react";
import { FeatureRow } from "@/components/marketing/FeatureRow";
import { InlineCode } from "@/components/marketing/InlineCode";
import { AppFrame, ArchitectureGraph, Terminal } from "@/components/showcase";
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

/**
 * The feature showcase (movement 3): the product shown as alternating rows in
 * the "your win" voice. The desktop app leads (it's the product); the CLI is
 * shown as its companion, never a standalone tool. Media is always a real
 * rendered showcase piece — never a fabricated screenshot.
 */
export function Features() {
  return (
    <>
      <FeatureRow
        eyebrow="The desktop app"
        title="See it before you ship it."
        body={
          <>
            <p>
              Open the app — the dashboard Kubernetes never had. See your whole
              architecture before you deploy, then watch each piece light up as it
              comes online.
            </p>
            <p>No blind applies, no guessing what a change will touch.</p>
          </>
        }
        media={
          <AppFrame title="Planton — Architecture">
            <ArchitectureGraph />
          </AppFrame>
        }
      />

      <FeatureRow
        reverse
        eyebrow="The CLI, a companion"
        title="Prefer the terminal? It&rsquo;s right there."
        body={
          <>
            <p>
              The same deploys, from your shell. The CLI drives the very engine the
              app brings online — managed state, ready-made charts, and history all
              included — so there&rsquo;s nothing extra to wire up.
            </p>
            <p>
              Run <InlineCode>planton chart install</InlineCode> and watch a whole
              environment come up live, every resource as it happens.
            </p>
          </>
        }
        media={<Terminal lines={CHART_INSTALL} />}
      />
    </>
  );
}

export default Features;
