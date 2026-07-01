import * as React from "react";
import { cn } from "@/lib/utils";

export interface WordmarkProps {
  className?: string;
}

/** The "Planton" wordmark, set in the display face. Pair with `PlantonMark`. */
export function Wordmark({ className }: WordmarkProps) {
  return (
    <span
      className={cn(
        "font-display text-lg font-semibold tracking-tight text-foreground",
        className,
      )}
    >
      Planton
    </span>
  );
}

export default Wordmark;
