import * as React from "react";
import { Section } from "@/components/marketing/Section";
import { Speaker } from "@/components/marketing/Speaker";
import { Line } from "@/components/marketing/Line";
import { Reveal } from "@/components/motion";
import { DeploysTo } from "@/components/marketing/DeploysTo";

/**
 * The trust + payoff band that closes the feature showcase: why the modules are
 * safe to trust, that it's real IaC on your own cloud with nothing to lock you
 * in, and the clouds it deploys to. Statement voice (no "we") — the reassurance
 * is about what you get, not about us.
 */
export function Payoff() {
  return (
    <Section>
      <Reveal>
        <span className="inline-flex items-center gap-2 rounded-full border border-border bg-secondary px-3 py-1 font-mono text-xs text-muted-foreground">
          <span className="size-2 rounded-full bg-success" />
          all systems running
        </span>
        <div className="mt-6 max-w-2xl space-y-6">
          <Speaker>Why you can trust it</Speaker>
          <Line tone="bright">
            You&rsquo;re never starting from a blank file. Every module is already
            written — and vetted for secure, well-architected, cost-efficient
            infrastructure. Claude can write you Terraform in minutes; writing it
            was never the hard part. Trusting it is.
          </Line>
          <Line tone="muted">
            Whichever way you deploy, it&rsquo;s real infrastructure-as-code the
            whole way down — stored, versioned, every change a diff. On your own
            cloud. Nothing to lock you in.
          </Line>
        </div>
        <DeploysTo />
      </Reveal>
    </Section>
  );
}

export default Payoff;
