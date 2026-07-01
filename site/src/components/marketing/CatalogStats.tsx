import * as React from "react";
import { Reveal } from "@/components/motion";
import { stats } from "@/site";

/** A quiet proof line — breadth, single-sourced from code (never a typed count). */
export function CatalogStats() {
  const items = [
    `${stats.componentsLabel} components`,
    `${stats.providersLabel} providers`,
    stats.chartsPhrase,
    "backed by Terraform and Pulumi",
  ];
  return (
    <Reveal className="mx-auto max-w-4xl px-4 pb-4 pt-10 sm:px-6 lg:px-8">
      <p className="flex flex-wrap justify-center gap-x-3 gap-y-1 text-center text-sm text-muted-foreground">
        {items.map((item, i) => (
          <React.Fragment key={item}>
            <span>{item}</span>
            {i < items.length - 1 && <span className="text-faint">·</span>}
          </React.Fragment>
        ))}
      </p>
    </Reveal>
  );
}

export default CatalogStats;
