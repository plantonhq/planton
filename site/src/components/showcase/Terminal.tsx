import * as React from "react";
import { cn } from "@/lib/utils";
import { WindowChrome } from "@/components/showcase/WindowChrome";
import { TerminalLine, type TerminalLineData } from "@/components/showcase/TerminalLine";

export interface TerminalProps {
  title?: string;
  lines: TerminalLineData[];
  className?: string;
}

/** A read-only terminal window rendered from structured lines (not a screenshot). */
export function Terminal({ title = "planton — zsh", lines, className }: TerminalProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-border bg-card shadow-2xl shadow-black/40",
        className,
      )}
    >
      <WindowChrome title={title} />
      <div className="space-y-1.5 p-5 font-mono text-[13px] leading-relaxed">
        {lines.map((line, i) => (
          <TerminalLine key={i} line={line} />
        ))}
      </div>
    </div>
  );
}

export default Terminal;
export type { TerminalLineData };
