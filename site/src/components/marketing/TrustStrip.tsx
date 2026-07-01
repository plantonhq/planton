import * as React from "react";
import { Check } from "lucide-react";
import { Reveal } from "@/components/motion";

// The four reassurances a skimmer wants before scrolling: free, their cloud,
// no account, open building blocks. Distinct from CatalogStats (which is breadth
// numbers) — this is the "what's the catch" answer, up top.
const POINTS = ["Free forever", "Your own cloud", "No account", "Open-source building blocks"];

/** A compact, checkmarked trust row under the hero — the fast "what's the catch" answer. */
export function TrustStrip() {
  return (
    <Reveal className="mx-auto max-w-4xl px-4 pb-8 pt-2 sm:px-6 lg:px-8">
      <ul className="flex flex-wrap justify-center gap-x-6 gap-y-2">
        {POINTS.map((point) => (
          <li key={point} className="flex items-center gap-2 text-sm text-muted-foreground">
            <Check className="size-4 text-success" aria-hidden />
            {point}
          </li>
        ))}
      </ul>
    </Reveal>
  );
}

export default TrustStrip;
