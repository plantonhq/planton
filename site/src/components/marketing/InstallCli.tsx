import * as React from "react";
import { Reveal } from "@/components/motion";
import { InstallMethodTabs } from "@/components/marketing/InstallMethodTabs";

/** The "prefer the terminal?" section — install the open-source CLI directly. */
export function InstallCli() {
  return (
    <Reveal className="mx-auto max-w-2xl px-4 py-16 sm:px-6 lg:px-8">
      <h2 className="font-display text-2xl font-semibold tracking-tight sm:text-3xl">
        Prefer the terminal? Install the CLI.
      </h2>
      <p className="mt-3 text-muted-foreground">
        The <code className="font-mono text-foreground">planton</code> CLI is open
        source (Apache-2.0). Install it and{" "}
        <code className="font-mono text-foreground">planton apply -f</code> from
        anywhere.
      </p>
      <div className="mt-6">
        <InstallMethodTabs />
      </div>
    </Reveal>
  );
}

export default InstallCli;
