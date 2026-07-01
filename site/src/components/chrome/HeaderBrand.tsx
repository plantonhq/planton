import * as React from "react";
import Link from "next/link";
import { PlantonMark } from "@/components/brand";

/** The clickable logo mark that returns home. Mark only — no wordmark. */
export function HeaderBrand() {
  return (
    <Link href="/" className="flex items-center text-foreground" aria-label="Planton home">
      <PlantonMark size={26} />
    </Link>
  );
}

export default HeaderBrand;
