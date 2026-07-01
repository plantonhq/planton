import * as React from "react";
import Link from "next/link";
import { HEADER_NAV } from "@/site";

/** Text navigation links in the header (Docs, Charts). */
export function HeaderNav() {
  return (
    <nav className="hidden items-center gap-6 sm:flex" aria-label="Primary">
      {HEADER_NAV.map((link) => (
        <Link
          key={link.href}
          href={link.href}
          {...(link.external ? { target: "_blank", rel: "noreferrer" } : {})}
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          {link.label}
        </Link>
      ))}
    </nav>
  );
}

export default HeaderNav;
