"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import { useClipboard } from "@/lib/useClipboard";

export interface CommandLineProps {
  command: string;
  /** Show a leading "$" prompt. */
  prompt?: boolean;
  className?: string;
}

/** A single copy-pasteable shell command with a copy button. */
export function CommandLine({ command, prompt = true, className }: CommandLineProps) {
  const { copied, copy } = useClipboard();
  return (
    <div
      className={cn(
        "flex items-center justify-between gap-4 rounded-lg border border-border bg-card px-4 py-3",
        className,
      )}
    >
      <code className="overflow-x-auto whitespace-nowrap font-mono text-sm text-foreground">
        {prompt && <span className="mr-2 select-none text-muted-foreground">$</span>}
        {command}
      </code>
      <button
        type="button"
        onClick={() => copy(command)}
        aria-label={copied ? "Copied" : "Copy command"}
        className="shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      >
        {copied ? <Check className="size-4 text-success" /> : <Copy className="size-4" />}
      </button>
    </div>
  );
}

export default CommandLine;
