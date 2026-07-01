import * as React from "react";
import { StoryStation } from "@/components/marketing/StoryStation";
import { Line } from "@/components/marketing/Line";
import { InlineCode } from "@/components/marketing/InlineCode";

interface TradeOff {
  /** Zero-padded station number. */
  index: string;
  /** Uppercase station label. */
  label: string;
  /** The reader's appeal (first person "I"), quiet. */
  you: React.ReactNode;
  /**
   * Planton's reply: affirm the tool first, then name the ONE honest catch.
   * Agree-first, never "you're wrong", and never the word "we" here — the praise
   * is of the tool itself, so the focus stays on the reader's experience.
   */
  planton: React.ReactNode;
}

/**
 * The trade-off ladder as one source of truth. Each rung is a genuine follow-up
 * to the last (console -> code -> Kubernetes), so it reads as a continuous
 * conversation rather than three disconnected beats. Kept as data + one renderer
 * (mirroring `WhereItFits`' "one array, can't drift" rule) so the three rungs
 * can never style-drift apart.
 */
const TRADE_OFFS: TradeOff[] = [
  {
    index: "01",
    label: "The console",
    you: "The cloud console is easy — I just fill a form and go.",
    planton: (
      <>
        Right? Consoles are the best — nothing beats filling in a form and
        watching it appear. The one catch: nothing&rsquo;s written down. No
        history, no diff, nothing to reproduce it from.
      </>
    ),
  },
  {
    index: "02",
    label: "The code",
    you: "Exactly why I moved to Terraform — now every change is a diff I can review.",
    planton: (
      <>
        And Terraform is the one thing that gives you that — a source of truth you
        can trust. The catch: a one-line change becomes a pull request, a plan,
        and a prayer. You lose the console&rsquo;s just-fill-a-form ease.
      </>
    ),
  },
  {
    index: "03",
    label: "Kubernetes",
    you: (
      <>
        The closest it ever felt right was Kubernetes — write a manifest,{" "}
        <InlineCode>kubectl apply -f</InlineCode>, done, saved to disk. Helm even
        does a whole stack in one file.
      </>
    ),
    planton: (
      <>
        Kubernetes nailed the gesture — the best infrastructure has ever felt. The
        catch: it stops at Kubernetes. No UI, and no real cloud state or history
        behind it.
      </>
    ),
  },
];

/** The numbered trade-off conversation (movement 2). */
export function Conversation() {
  return (
    <>
      {TRADE_OFFS.map((rung) => (
        <StoryStation key={rung.index} index={rung.index} label={rung.label}>
          <div className="space-y-6">
            <Line speaker="You" tone="muted">
              {rung.you}
            </Line>
            <Line speaker="Planton" tone="bright">
              {rung.planton}
            </Line>
          </div>
        </StoryStation>
      ))}
    </>
  );
}

export default Conversation;
