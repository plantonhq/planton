import * as React from "react";
import { cn } from "@/lib/utils";

/** A highlighted inline command chip, e.g. `kubectl apply -f` or `planton apply -f`. */
export function InlineCode({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <code
      className={cn(
        "rounded-md bg-secondary px-1.5 py-0.5 font-mono text-[0.9em] text-foreground",
        className,
      )}
    >
      {children}
    </code>
  );
}

export default InlineCode;
