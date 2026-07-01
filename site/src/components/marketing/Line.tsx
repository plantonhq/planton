import * as React from "react";
import { cn } from "@/lib/utils";
import { Speaker } from "@/components/marketing/Speaker";

/** The two canonical voices of the story's editorial text. */
export type LineTone = "muted" | "bright";

export interface LineProps {
  /** Optional speaker label (e.g. "You", "Planton"). Omit for narration. */
  speaker?: string;
  /**
   * "muted"  — the reader's voice / the appeal (quiet, secondary).
   * "bright" — Planton's statement or the honest catch (emphasized).
   */
  tone: LineTone;
  className?: string;
  children: React.ReactNode;
}

const TONE: Record<LineTone, string> = {
  muted: "text-lg leading-relaxed text-muted-foreground sm:text-xl",
  bright: "text-2xl font-medium leading-snug text-foreground sm:text-3xl",
};

/**
 * The story's tone atom: an optional speaker label plus one paragraph at a
 * canonical tone. It is the single source of the "You = muted, Planton = bright"
 * typography, so that decision lives in exactly one place.
 *
 * Used across the narrative beats — the You/Planton dialogue (with a speaker) AND
 * speaker-less narration/statements (e.g. the ladder opener and the WhereItFits
 * payoff/bridge). Give it a `speaker` for dialogue; omit it for narration.
 *
 * Deliberately NOT used by: the `Origin` crescendo (a bespoke two-tier
 * statement) and the `Hero`/`FinalCta` display headlines. Those are distinct
 * rhetorical units at their own scale; routing them through `Line` would need
 * extra tones for no gain.
 */
export function Line({ speaker, tone, className, children }: LineProps) {
  return (
    <div className={className}>
      {speaker && <Speaker>{speaker}</Speaker>}
      <p className={cn(TONE[tone])}>{children}</p>
    </div>
  );
}

export default Line;
