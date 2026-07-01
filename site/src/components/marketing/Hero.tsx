import * as React from "react";
import { DownloadButton } from "@/components/chrome";
import { ShowcaseTabs, ArchitectureGraph } from "@/components/showcase";
import type { TerminalLineData } from "@/components/showcase";
import { OPERATING_SYSTEMS } from "@/site";

// A single-manifest apply — the terminal shown as a companion to the app.
const HERO_TERMINAL: TerminalLineData[] = [
  { kind: "command", text: "planton apply -f ecs-service.yaml" },
  { kind: "output", text: "→ applying aws-ecs-service · api" },
  { kind: "resource", name: "ecs.service", status: "created · 6.4s", state: "done" },
];

/** The hero: desktop-first positioning, the primary CTA, and a rendered product preview. */
export function Hero() {
  return (
    <section className="px-4 pt-32 sm:px-6 sm:pt-40 lg:px-8">
      <div className="mx-auto max-w-3xl text-center">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-faint">
          A free desktop app for your cloud infrastructure
        </p>
        <h1 className="mt-6 text-balance font-display text-4xl font-bold leading-[1.05] tracking-tight sm:text-6xl">
          Deploy real infrastructure to your own cloud — without writing Terraform.
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground">
          Planton is a free desktop app you download and open. It finds the cloud
          you&rsquo;re already signed into — pick a stack, fill a short form, and
          watch it deploy, with clean, auditable infrastructure-as-code running
          underneath. No account. No connections. No ceremony.
        </p>
        <div className="mt-9 flex flex-col items-center gap-3">
          <DownloadButton size="lg" />
          <p className="text-xs tracking-[0.08em] text-faint">
            Runs on {OPERATING_SYSTEMS.join(" · ")}
          </p>
        </div>
      </div>

      <div className="mx-auto mt-16 max-w-4xl sm:mt-20">
        <ShowcaseTabs
          desktop={{ title: "Planton — Architecture", media: <ArchitectureGraph /> }}
          terminal={{ title: "planton — zsh", lines: HERO_TERMINAL }}
        />
      </div>
    </section>
  );
}

export default Hero;
