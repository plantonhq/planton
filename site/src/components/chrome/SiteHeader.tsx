import * as React from "react";
import { HeaderBrand } from "@/components/chrome/HeaderBrand";
import { HeaderNav } from "@/components/chrome/HeaderNav";
import { HeaderActions } from "@/components/chrome/HeaderActions";

export interface SiteHeaderProps {
  /** Optional leading element before the brand (e.g. a docs mobile menu button). */
  leading?: React.ReactNode;
  /** Optional element between nav and actions (e.g. the docs search bar). */
  slot?: React.ReactNode;
}

/**
 * The shared site header: logo left, nav + Download/GitHub right. Reused by the
 * landing page and the docs so brand and navigation are defined exactly once.
 */
export function SiteHeader({ leading, slot }: SiteHeaderProps) {
  return (
    <header className="fixed inset-x-0 top-0 z-50 border-b border-border bg-background/80 backdrop-blur">
      <div className="flex h-16 items-center justify-between gap-4 px-5 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3">
          {leading}
          {/* When a leading control exists (docs mobile menu), the logo yields to
              it on mobile and returns at md — so each viewport shows exactly one
              left control. Without a leading control (landing/download), the logo
              always shows. */}
          <div className={leading ? "hidden md:flex md:items-center" : "flex items-center"}>
            <HeaderBrand />
          </div>
        </div>
        <div className="flex items-center gap-6">
          {slot}
          <HeaderNav />
          <HeaderActions />
        </div>
      </div>
    </header>
  );
}

export default SiteHeader;
