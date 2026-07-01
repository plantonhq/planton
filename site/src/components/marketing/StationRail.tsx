import * as React from "react";
import { cn } from "@/lib/utils";

export interface StationRailProps {
  /** Zero-padded station number, e.g. "01". */
  index: string;
  /** Uppercase station label, e.g. "THE CONSOLE". */
  label: string;
  /** "active" shows a filled status dot (the payoff "all systems running" beat). */
  marker?: "idle" | "active";
}

/** The left rail of a story station: a marker, the number, and the label. */
export function StationRail({ index, label, marker = "idle" }: StationRailProps) {
  return (
    <div className="flex items-center gap-3 md:flex-col md:items-start md:gap-4">
      <span
        className={cn(
          "size-2.5 rounded-full",
          marker === "active" ? "bg-success" : "border border-border",
        )}
        aria-hidden
      />
      <div className="flex items-center gap-3 md:flex-col md:items-start md:gap-2">
        <span className="font-mono text-xs text-faint">{index}</span>
        <span className="text-xs font-medium uppercase tracking-[0.18em] text-faint">
          {label}
        </span>
      </div>
    </div>
  );
}

export default StationRail;
