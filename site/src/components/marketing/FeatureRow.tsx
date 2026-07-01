import * as React from "react";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/motion";
import { Speaker } from "@/components/marketing/Speaker";

export interface FeatureRowProps {
  /** Small tracked eyebrow above the title (e.g. "The desktop app"). */
  eyebrow?: string;
  title: React.ReactNode;
  body: React.ReactNode;
  /** A real, rendered product visual (the architecture graph, a terminal). */
  media: React.ReactNode;
  /** Put the media on the left; alternate down the page so the eye zig-zags. */
  reverse?: boolean;
  className?: string;
}

/**
 * One feature of the product, in the "your win" voice of the feature showcase
 * (movement 3): an editorial copy column beside a product visual. This is the
 * deliberate design-system switch away from the numbered `StoryStation` used by
 * the conversation — rows alternate sides via `reverse`. Copy leads with the
 * benefit; `media` must be a real rendered showcase piece (graph, terminal),
 * never a fabricated screenshot.
 */
export function FeatureRow({ eyebrow, title, body, media, reverse, className }: FeatureRowProps) {
  return (
    <Reveal className={cn("mx-auto max-w-6xl px-4 py-14 sm:px-6 sm:py-16 lg:px-8", className)}>
      <div className="grid items-center gap-10 md:grid-cols-2 md:gap-16">
        <div className={cn("max-w-xl", reverse ? "md:order-2" : "md:order-1")}>
          {eyebrow && <Speaker>{eyebrow}</Speaker>}
          <h2 className="text-balance font-display text-2xl font-semibold tracking-tight sm:text-3xl">
            {title}
          </h2>
          <div className="mt-4 space-y-4 text-lg leading-relaxed text-muted-foreground">{body}</div>
        </div>
        <div className={cn("min-w-0", reverse ? "md:order-1" : "md:order-2")}>{media}</div>
      </div>
    </Reveal>
  );
}

export default FeatureRow;
