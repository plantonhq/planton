"use client";

import * as React from "react";
import { Github } from "lucide-react";
import { GITHUB_REPO, GITHUB_URL } from "@/site";

function formatStars(count: number): string {
  return count >= 1000 ? `${(count / 1000).toFixed(1)}k` : String(count);
}

/**
 * A quiet "Star on GitHub" pill with a live count. Best-effort: if the
 * unauthenticated GitHub API is unavailable/rate-limited, it degrades to just
 * the label (never an error).
 */
export function GitHubStars({ className }: { className?: string }) {
  const [stars, setStars] = React.useState<number | null>(null);

  React.useEffect(() => {
    let active = true;
    fetch(`https://api.github.com/repos/${GITHUB_REPO}`)
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => {
        if (active && d) setStars(d.stargazers_count as number);
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);

  return (
    <a
      href={GITHUB_URL}
      target="_blank"
      rel="noreferrer"
      aria-label="Star Planton on GitHub"
      className={`inline-flex items-center gap-2 rounded-full border border-border bg-secondary/60 px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:border-ring hover:text-foreground ${className ?? ""}`}
    >
      <Github className="size-4" />
      <span className="hidden sm:inline">Star</span>
      {stars !== null && (
        <>
          <span className="h-4 w-px bg-border" />
          <span className="font-medium text-foreground">{formatStars(stars)}</span>
        </>
      )}
    </a>
  );
}

export default GitHubStars;
