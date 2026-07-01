import * as React from "react";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/motion";

/**
 * A centered narration beat that sits between the movements — the storyteller's
 * voice, not a speaker in the dialogue. Two uses today: the prologue that opens
 * the trade-off conversation and the bridge that closes it and hands into the
 * origin. Deliberately NOT a `StoryStation` (no number, no speaker label) and
 * NOT a `Line` (which carries a speaker); its quiet, centered framing is exactly
 * what separates the narration from the numbered You/Planton exchange.
 */
export function Interlude({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <Reveal className={cn("mx-auto max-w-3xl px-4 py-14 text-center sm:px-6 sm:py-20 lg:px-8", className)}>
      <p className="text-balance text-xl leading-relaxed text-muted-foreground sm:text-2xl">
        {children}
      </p>
    </Reveal>
  );
}

export default Interlude;
