import * as React from "react";
import { cn } from "@/lib/utils";

export interface WindowChromeProps {
  /** Centered title in the title bar (e.g. "planton — zsh"). */
  title?: string;
  className?: string;
}

/**
 * A macOS-style window title bar (three traffic-light dots + centered title).
 * Shared by the terminal and the desktop app frame so both read as real windows.
 */
export function WindowChrome({ title, className }: WindowChromeProps) {
  return (
    <div
      className={cn(
        "relative flex h-9 items-center gap-2 border-b border-border px-4",
        className,
      )}
    >
      <span className="size-3 rounded-full bg-border" />
      <span className="size-3 rounded-full bg-border" />
      <span className="size-3 rounded-full bg-border" />
      {title && (
        <span className="pointer-events-none absolute inset-x-0 text-center font-mono text-xs text-muted-foreground">
          {title}
        </span>
      )}
    </div>
  );
}

export default WindowChrome;
