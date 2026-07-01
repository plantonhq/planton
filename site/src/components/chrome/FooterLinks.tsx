import * as React from "react";
import Link from "next/link";
import { Github } from "lucide-react";
import { DiscordIcon } from "@/components/brand";
import { FOOTER_LINKS } from "@/site";

/** Small leading icon for the two links that carry a brand mark. */
function LinkIcon({ label }: { label: string }) {
  if (label === "GitHub") return <Github className="size-4" />;
  if (label === "Discord") return <DiscordIcon size={16} />;
  return null;
}

/** The footer link row: Docs · Charts · GitHub · Discord · planton.ai */
export function FooterLinks() {
  return (
    <nav className="flex flex-wrap items-center gap-x-6 gap-y-3" aria-label="Footer">
      {FOOTER_LINKS.map((link) => (
        <Link
          key={link.href}
          href={link.href}
          {...(link.external ? { target: "_blank", rel: "noreferrer" } : {})}
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <LinkIcon label={link.label} />
          {link.label}
        </Link>
      ))}
    </nav>
  );
}

export default FooterLinks;
