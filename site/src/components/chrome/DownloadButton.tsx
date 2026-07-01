import * as React from "react";
import Link from "next/link";
import { Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { DOWNLOAD_HREF } from "@/site";

export interface DownloadButtonProps {
  size?: "sm" | "default" | "lg";
  className?: string;
  label?: string;
}

/**
 * The primary "Download Planton" CTA. Targets the single `DOWNLOAD_HREF` from
 * site config, so every instance (header, hero, final CTA) stays in lockstep.
 */
export function DownloadButton({
  size = "default",
  className,
  label = "Download Planton",
}: DownloadButtonProps) {
  return (
    <Button asChild size={size} className={cn("rounded-full font-medium", className)}>
      <Link href={DOWNLOAD_HREF}>
        <Download />
        {label}
      </Link>
    </Button>
  );
}

export default DownloadButton;
