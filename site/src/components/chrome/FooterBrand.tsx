import * as React from "react";
import { PlantonMark, Wordmark } from "@/components/brand";

/** Footer brand lockup with a short, honest licensing line. */
export function FooterBrand() {
  return (
    <div className="max-w-sm">
      <div className="flex items-center gap-2.5 text-foreground">
        <PlantonMark size={22} />
        <Wordmark />
      </div>
      <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
        A free Desktop App and CLI for your cloud infrastructure. The building
        blocks, stacks, and CLI are open source under Apache-2.0. No lock-in.
      </p>
    </div>
  );
}

export default FooterBrand;
