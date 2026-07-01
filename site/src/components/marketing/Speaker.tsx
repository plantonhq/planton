import * as React from "react";
import { cn } from "@/lib/utils";

/** A small uppercase, tracked speaker/eyebrow label (e.g. "You — the console"). */
export function Speaker({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <p className={cn("mb-3 text-xs font-medium uppercase tracking-[0.18em] text-faint", className)}>
      {children}
    </p>
  );
}

export default Speaker;
