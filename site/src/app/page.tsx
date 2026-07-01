import { SiteHeader, SiteFooter } from "@/components/chrome";
import {
  Hero,
  CatalogStats,
  TrustStrip,
  Interlude,
  Conversation,
  Origin,
  Features,
  Payoff,
  WhereItFits,
  Horizon,
  Catch,
  FinalCta,
} from "@/components/marketing";

/**
 * planton.app — the front door, in three movements:
 *   1. Skimmer top     — hero + proof + the "what's the catch" trust strip, so a
 *      visitor gets it and can download without scrolling.
 *   2. The conversation — a prologue, the numbered agree-first trade-off dialogue,
 *      a bridge, then the origin synthesis.
 *   3. Feature showcase — the product shown as alternating feature rows in the
 *      "your win" voice, trust/deploys, positioning, and the closing CTA.
 */
export default function HomePage() {
  return (
    <>
      <SiteHeader />
      <main>
        {/* Movement 1 — for the skimmer */}
        <Hero />
        <CatalogStats />
        <TrustStrip />

        {/* Movement 2 — the conversation */}
        <Interlude>
          Every way of running cloud infrastructure has asked you to give something up.
        </Interlude>
        <Conversation />
        <Interlude>
          Console, Terraform, Kubernetes — each one brilliant, each one asking you
          to pick, and live without the rest. Picking was the problem.
        </Interlude>
        <Origin />

        {/* Movement 3 — the feature showcase */}
        <Features />
        <Payoff />
        <WhereItFits />
        <Horizon />
        <Catch />
        <FinalCta />
      </main>
      <SiteFooter />
    </>
  );
}
