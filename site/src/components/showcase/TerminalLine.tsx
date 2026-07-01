import * as React from "react";
import { cn } from "@/lib/utils";

/** A resource line's provisioning status, driving its glyph and color. */
export type ResourceStatus = "done" | "running" | "pending";

export type TerminalLineData =
  | { kind: "command"; text: string }
  | { kind: "comment"; text: string }
  | { kind: "output"; text: string }
  | {
      kind: "resource";
      /** e.g. "network.vpc" */
      name: string;
      /** e.g. "created · 4.2s" or "creating…" */
      status: string;
      state: ResourceStatus;
    };

const GLYPH: Record<ResourceStatus, string> = {
  done: "✓",
  running: "◇",
  pending: "·",
};

/** Renders one line of terminal output from data — no ANSI, just tokens. */
export function TerminalLine({ line }: { line: TerminalLineData }) {
  switch (line.kind) {
    case "command":
      return (
        <div className="flex gap-2">
          <span className="select-none text-muted-foreground">$</span>
          <span className="text-foreground">{line.text}</span>
        </div>
      );
    case "comment":
    case "output":
      return <div className="text-muted-foreground">{line.text}</div>;
    case "resource":
      return (
        <div className="flex items-center justify-between gap-4">
          <span className="flex items-center gap-2">
            <span
              className={cn(
                "w-3 select-none",
                line.state === "done" ? "text-success" : "text-muted-foreground",
              )}
            >
              {GLYPH[line.state]}
            </span>
            <span className="text-foreground">{line.name}</span>
          </span>
          <span className="text-muted-foreground">{line.status}</span>
        </div>
      );
  }
}

export default TerminalLine;
