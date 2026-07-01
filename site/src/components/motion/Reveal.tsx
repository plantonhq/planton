"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { useReveal } from "@/components/motion/useReveal";

export interface RevealProps {
  children: React.ReactNode;
  /** Stagger start via a small delay (ms). */
  delayMs?: number;
  className?: string;
}

/**
 * Fades and lifts its children into view on first scroll. The scroll-as-
 * deployment motif — each beat "comes online" as you reach it.
 */
export function Reveal({ children, delayMs = 0, className }: RevealProps) {
  const { ref, visible } = useReveal<HTMLDivElement>();
  return (
    <div
      ref={ref}
      style={{ transitionDelay: visible ? `${delayMs}ms` : "0ms" }}
      className={cn(
        "transition-all duration-700 ease-out motion-reduce:transition-none",
        visible ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0",
        className,
      )}
    >
      {children}
    </div>
  );
}

export default Reveal;
