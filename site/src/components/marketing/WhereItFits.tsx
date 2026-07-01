import * as React from "react";
import { Reveal } from "@/components/motion";
import { Line } from "@/components/marketing/Line";

// The three legitimate ways to run a product today. Kept parallel-by-construction
// (one array, one row template) so the tiers can never drift apart. Naming rule:
// the fully-managed tier names representative platforms as a category illustration;
// the "own cloud" tiers are evoked, never pointed at a direct peer.
const WAYS: { label: string; body: string }[] = [
  {
    label: "Fully managed platforms",
    body:
      "Heroku, Render, Fly, Railway. Push your code and they run it — the fastest way to get live when you\u2019d rather not think about infrastructure at all.",
  },
  {
    label: "Your own cloud, managed for you",
    body:
      "A hosted service connects to your cloud account and gives you a console for standing up environments. Your cloud, with someone else\u2019s control plane in front of it — an account, a plan, and a connection to wire up first.",
  },
  {
    label: "Your own cloud, by hand",
    body:
      "Click through your cloud provider\u2019s console, or write the Terraform yourself (or have an AI write it) and ship it from your laptop or CI. Total control — and it\u2019s all yours to get right, secure, and keep running.",
  },
];

/**
 * The orientation map (movement 3): a calm comparison of the ways to run what
 * you ship, then Planton's slot. It answers "where does this fit among my
 * options?" for the evaluator who has already seen the product. Deliberately NOT
 * a StoryStation — a centered map reads distinctly from the numbered dialogue.
 * The full synthesis lives in `Origin`, so this stays a placement, not a re-pitch.
 */
export function WhereItFits() {
  return (
    <Reveal className="mx-auto max-w-3xl px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="text-center">
        <h2 className="text-balance font-display text-3xl font-bold tracking-tight sm:text-4xl">
          Where Planton fits
        </h2>
        <p className="mx-auto mt-5 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          A few good ways to run what you ship — all of them fine, mostly down to
          preference. Here&rsquo;s the map.
        </p>
      </div>

      <div className="mt-14 border-t border-border">
        {WAYS.map((way) => (
          <div
            key={way.label}
            className="grid gap-2 border-b border-border py-6 sm:grid-cols-[16rem_1fr] sm:gap-8"
          >
            <h3 className="font-medium text-foreground">{way.label}</h3>
            <p className="text-muted-foreground">{way.body}</p>
          </div>
        ))}
      </div>

      <Line tone="muted" className="mt-12">
        Planton is for the last two — your own cloud and real control, without the
        ceremony. Especially for solo developers and indie hackers.
      </Line>
    </Reveal>
  );
}

export default WhereItFits;
