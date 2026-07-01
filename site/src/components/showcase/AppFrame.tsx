import * as React from "react";
import Image from "next/image";
import { cn } from "@/lib/utils";
import { WindowChrome } from "@/components/showcase/WindowChrome";

export interface AppFrameProps {
  title?: string;
  /** Real screenshot; takes precedence when provided. */
  screenshot?: { src: string; alt: string };
  /**
   * Rendered content to show inside the window (e.g. the architecture graph) —
   * an honest, crisp stand-in until real screenshots exist. Used when no
   * `screenshot` is given.
   */
  children?: React.ReactNode;
  /** Placeholder label used when neither a screenshot nor children are given. */
  label?: string;
  className?: string;
}

/**
 * A desktop app window frame. Priority: a real `screenshot` if provided, else
 * rendered `children` (honest rendered content, e.g. the architecture graph),
 * else a clean labeled placeholder. We never fabricate a screenshot.
 */
export function AppFrame({
  title = "Planton",
  screenshot,
  children,
  label = "Planton Desktop",
  className,
}: AppFrameProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-border bg-card shadow-2xl shadow-black/40",
        className,
      )}
    >
      <WindowChrome title={title} />
      {screenshot ? (
        <div className="relative aspect-[16/10] w-full bg-background">
          <Image
            src={screenshot.src}
            alt={screenshot.alt}
            fill
            sizes="(max-width: 768px) 100vw, 900px"
            className="object-cover object-top"
          />
        </div>
      ) : children ? (
        <div className="bg-background p-6">{children}</div>
      ) : (
        <div className="flex aspect-[16/10] w-full flex-col items-center justify-center gap-1.5 bg-background">
          <span className="font-mono text-xs uppercase tracking-[0.2em] text-muted-foreground">
            {label}
          </span>
          <span className="text-xs text-faint">screenshot placeholder</span>
        </div>
      )}
    </div>
  );
}

export default AppFrame;
