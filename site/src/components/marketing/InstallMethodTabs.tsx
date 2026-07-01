"use client";

import * as React from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { CommandLine } from "@/components/marketing/CommandLine";
import { CLI_INSTALL_METHODS } from "@/site";

/** Tabbed CLI install methods, driven entirely by site config. */
export function InstallMethodTabs() {
  const methods = CLI_INSTALL_METHODS;
  return (
    <Tabs defaultValue={methods[0]?.id} className="w-full">
      <TabsList className="mb-4 bg-secondary">
        {methods.map((m) => (
          <TabsTrigger key={m.id} value={m.id} className="data-[state=active]:bg-background">
            {m.label}
          </TabsTrigger>
        ))}
      </TabsList>
      {methods.map((m) => (
        <TabsContent key={m.id} value={m.id}>
          <CommandLine command={m.command} />
          {m.note && <p className="mt-2 text-xs text-muted-foreground">{m.note}</p>}
        </TabsContent>
      ))}
    </Tabs>
  );
}

export default InstallMethodTabs;
