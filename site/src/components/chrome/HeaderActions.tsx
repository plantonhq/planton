import * as React from "react";
import { GitHubStars } from "@/components/chrome/GitHubStars";
import { DownloadButton } from "@/components/chrome/DownloadButton";
import { DiscordIcon } from "@/components/brand";
import { DISCORD_URL } from "@/site";

/** Right-aligned header actions: Discord + GitHub stars + the Download CTA. */
export function HeaderActions() {
  return (
    <div className="flex items-center gap-3">
      <a
        href={DISCORD_URL}
        target="_blank"
        rel="noreferrer"
        aria-label="Join the Planton Discord"
        className="inline-flex size-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-secondary/60 hover:text-foreground"
      >
        <DiscordIcon size={18} />
      </a>
      <GitHubStars />
      <DownloadButton size="sm" />
    </div>
  );
}

export default HeaderActions;
