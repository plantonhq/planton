import * as React from "react";
import { cn } from "@/lib/utils";

export interface SectionProps {
  id?: string;
  className?: string;
  children: React.ReactNode;
}

/** A consistent vertical rhythm + max-width wrapper for marketing sections. */
export function Section({ id, className, children }: SectionProps) {
  return (
    <section id={id} className={cn("mx-auto max-w-5xl px-4 py-20 sm:px-6 sm:py-28 lg:px-8", className)}>
      {children}
    </section>
  );
}

export default Section;
