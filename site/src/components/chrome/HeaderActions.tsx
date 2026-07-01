import * as React from "react";
import { GitHubStars } from "@/components/chrome/GitHubStars";
import { DownloadButton } from "@/components/chrome/DownloadButton";

/** Right-aligned header actions: GitHub stars + the Download CTA. */
export function HeaderActions() {
  return (
    <div className="flex items-center gap-3">
      <GitHubStars />
      <DownloadButton size="sm" />
    </div>
  );
}

export default HeaderActions;
