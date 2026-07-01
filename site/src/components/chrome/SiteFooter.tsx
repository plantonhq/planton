import * as React from "react";
import { FooterBrand } from "@/components/chrome/FooterBrand";
import { FooterLinks } from "@/components/chrome/FooterLinks";

/** The shared site footer. */
export function SiteFooter() {
  const year = new Date().getFullYear();
  return (
    <footer className="border-t border-border">
      <div className="px-5 py-14 sm:px-6 lg:px-8">
        <div className="flex flex-col justify-between gap-10 md:flex-row md:items-start">
          <FooterBrand />
          <FooterLinks />
        </div>
        <p className="mt-12 text-xs text-muted-foreground">
          © {year} Planton · Open source under Apache-2.0
        </p>
      </div>
    </footer>
  );
}

export default SiteFooter;
