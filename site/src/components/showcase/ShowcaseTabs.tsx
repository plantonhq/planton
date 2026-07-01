"use client";

import * as React from "react";
import { Monitor, TerminalSquare } from "lucide-react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { AppFrame } from "@/components/showcase/AppFrame";
import { Terminal } from "@/components/showcase/Terminal";
import type { TerminalLineData } from "@/components/showcase/TerminalLine";

export interface ShowcaseTabsProps {
  /**
   * The desktop app view. A real `screenshot` wins; else `media` (rendered
   * content such as the architecture graph); else a labeled placeholder.
   */
  desktop: {
    title?: string;
    screenshot?: { src: string; alt: string };
    media?: React.ReactNode;
    label?: string;
  };
  /** The equivalent terminal view, rendered from structured lines. */
  terminal: { title?: string; lines: TerminalLineData[] };
  defaultTab?: "desktop" | "terminal";
  className?: string;
}

const triggerCls =
  "gap-2 rounded-full px-4 data-[state=active]:bg-background data-[state=active]:text-foreground";

/**
 * The reusable "same thing, two ways" component: one tab shows the Planton
 * desktop app, the other shows the equivalent in the terminal. Used everywhere a
 * product view appears so the desktop/CLI duality is felt consistently.
 */
export function ShowcaseTabs({
  desktop,
  terminal,
  defaultTab = "desktop",
  className,
}: ShowcaseTabsProps) {
  return (
    <Tabs defaultValue={defaultTab} className={className}>
      <div className="mb-4 flex justify-center">
        <TabsList className="rounded-full bg-secondary p-1">
          <TabsTrigger value="desktop" className={triggerCls}>
            <Monitor className="size-4" />
            Desktop
          </TabsTrigger>
          <TabsTrigger value="terminal" className={triggerCls}>
            <TerminalSquare className="size-4" />
            Terminal
          </TabsTrigger>
        </TabsList>
      </div>

      <TabsContent value="desktop">
        <AppFrame title={desktop.title} screenshot={desktop.screenshot} label={desktop.label}>
          {desktop.media}
        </AppFrame>
      </TabsContent>
      <TabsContent value="terminal">
        <Terminal title={terminal.title} lines={terminal.lines} />
      </TabsContent>
    </Tabs>
  );
}

export default ShowcaseTabs;
