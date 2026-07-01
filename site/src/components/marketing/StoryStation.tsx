import * as React from "react";
import { Reveal } from "@/components/motion";
import { StationRail } from "@/components/marketing/StationRail";
import { StationBody } from "@/components/marketing/StationBody";

export interface StoryStationProps {
  index: string;
  label: string;
  marker?: "idle" | "active";
  /** Let the body span the full column (for terminals, graphs, logo rows). */
  wide?: boolean;
  children: React.ReactNode;
}

/**
 * One beat of the origin story: a numbered left rail + an editorial body,
 * revealed on scroll. Every numbered station on the page composes this.
 */
export function StoryStation({ index, label, marker, wide, children }: StoryStationProps) {
  return (
    <Reveal className="mx-auto grid max-w-5xl grid-cols-1 gap-6 px-4 py-16 sm:px-6 md:grid-cols-[9rem_1fr] md:gap-10 lg:px-8">
      <StationRail index={index} label={label} marker={marker} />
      <StationBody className={wide ? "max-w-none" : undefined}>{children}</StationBody>
    </Reveal>
  );
}

export default StoryStation;
