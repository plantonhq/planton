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
 * The orientation beat between the hero and the origin story: a calm decision
 * map of the three ways to run a product today, then Planton's slot. It answers
 * "where does this fit, and why choose it?" before the conversation earns it
 * emotionally. Deliberately NOT a StoryStation — a centered map reads distinctly
 * from the numbered dialogue that follows.
 */
export function WhereItFits() {
  return (
    <Reveal className="mx-auto max-w-3xl px-4 py-20 sm:px-6 sm:py-28 lg:px-8">
      <div className="text-center">
        <h2 className="text-balance font-display text-3xl font-bold tracking-tight sm:text-4xl">
          You&rsquo;ve got something to ship. There&rsquo;s more than one good way
          to run it.
        </h2>
        <p className="mx-auto mt-5 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          There are a few good options, and they&rsquo;re all great — it usually
          comes down to preference. Here&rsquo;s the map, and where Planton fits.
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

      <div className="mt-14 space-y-6">
        <Line tone="bright">
          Planton is for the last two — especially solo developers and indie
          hackers who want their own cloud and real control. It gives you the ease
          of a console and the ownership of raw Terraform, with the ceremony of
          neither.
        </Line>
        <Line tone="muted">
          It&rsquo;s a free app that runs on your machine and uses the cloud
          you&rsquo;re already signed into — no service in the middle. Prewritten,
          vetted modules stand up secure, cost-efficient infrastructure — whole
          environments on your own AWS, GCP, Azure, or DigitalOcean, with full
          control and no abstractions.
        </Line>
      </div>

      <Line tone="muted" className="mt-10">
        It wasn&rsquo;t always this easy to have both.
      </Line>
    </Reveal>
  );
}

export default WhereItFits;
