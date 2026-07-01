import * as React from "react";
import { cn } from "@/lib/utils";

interface ArchNode {
  name: string;
  detail: string;
  state: "online" | "provisioning";
}

// A small, honest preview of the read-only architecture graph the desktop shows.
const NODES: ArchNode[] = [
  { name: "Account", detail: "aws · us-east-1", state: "online" },
  { name: "Network", detail: "vpc · 10.0.0.0/16", state: "online" },
  { name: "Service", detail: "ecs · fargate", state: "online" },
  { name: "Database", detail: "rds · provisioning…", state: "provisioning" },
];

function Node({ node }: { node: ArchNode }) {
  return (
    <div className="min-w-[9rem] flex-1 rounded-lg border border-border bg-secondary px-4 py-3">
      <div className="flex items-center gap-2">
        <span
          className={cn(
            "size-2 rounded-full",
            node.state === "online" ? "bg-success" : "border border-muted-foreground",
          )}
        />
        <span className="text-sm font-medium text-foreground">{node.name}</span>
      </div>
      <p className="mt-1 font-mono text-xs text-muted-foreground">{node.detail}</p>
    </div>
  );
}

/**
 * The read-only resource graph the desktop shows before deploy — the node body
 * only, WITHOUT window chrome, so it can be dropped inside an `AppFrame` (which
 * provides the window). Rendered from data, never a screenshot.
 */
export function ArchitectureGraph({ className }: { className?: string }) {
  return (
    <div className={cn("flex flex-col items-stretch gap-3 sm:flex-row sm:items-center", className)}>
      {NODES.map((node, i) => (
        <React.Fragment key={node.name}>
          <Node node={node} />
          {i < NODES.length - 1 && (
            <span className="mx-auto h-4 w-px bg-border sm:h-px sm:w-6" aria-hidden />
          )}
        </React.Fragment>
      ))}
    </div>
  );
}

export default ArchitectureGraph;
