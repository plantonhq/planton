import * as React from "react";
import Image from "next/image";
import { cn } from "@/lib/utils";
import { WindowChrome } from "@/components/showcase/WindowChrome";

export interface AppFrameProps {
  title?: string;
  /** Real screenshot; when omitted, a labeled placeholder is shown. */
  screenshot?: { src: string; alt: string };
  /** Placeholder label used until a real screenshot is captured. */
  label?: string;
  className?: string;
}

/**
 * A desktop app window frame. Renders a real screenshot when provided; until the
 * keynote captures exist it shows a clean, honest placeholder (we never
 * fabricate a screenshot).
 */
export function AppFrame({
  title = "Planton",
  screenshot,
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
      <div className="relative aspect-[16/10] w-full bg-background">
        {screenshot ? (
          <Image
            src={screenshot.src}
            alt={screenshot.alt}
            fill
            sizes="(max-width: 768px) 100vw, 900px"
            className="object-cover object-top"
          />
        ) : (
          <div className="flex h-full flex-col items-center justify-center gap-1.5">
            <span className="font-mono text-xs uppercase tracking-[0.2em] text-muted-foreground">
              {label}
            </span>
            <span className="text-xs text-faint">screenshot placeholder</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default AppFrame;
