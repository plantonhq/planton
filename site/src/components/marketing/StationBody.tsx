import * as React from "react";
import { cn } from "@/lib/utils";

/** The editorial column of a story station. */
export function StationBody({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <div className={cn("max-w-xl", className)}>{children}</div>;
}

export default StationBody;
