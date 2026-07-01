import * as React from "react";
import { DEPLOY_CLOUDS } from "@/site";

/** The "Deploys to" row — the four clouds at deploy parity, from real logos. */
export function DeploysTo() {
  return (
    <div className="mt-10">
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-faint">Deploys to</p>
      <div className="mt-5 flex flex-wrap items-center gap-8">
        {DEPLOY_CLOUDS.map((cloud) => (
          // Brand logos have varying aspect ratios; a plain img keeps a uniform
          // height without next/image's fixed-dimension constraints.
          // eslint-disable-next-line @next/next/no-img-element
          <img
            key={cloud.name}
            src={cloud.logo}
            alt={cloud.name}
            className="h-7 w-auto opacity-70 transition-opacity hover:opacity-100"
          />
        ))}
      </div>
    </div>
  );
}

export default DeploysTo;
