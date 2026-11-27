'use client';
import Link from 'next/link';
import { DOORS, type DoorId } from '@/data/doors';
import { trackExperience, type DemoLocation } from '@/lib/demo-analytics';
import styles from './homepage.module.css';
/**
 * A self-service door on Home, attributed to where it sits. `primary` makes it
 * the section's filled button; Start Free is always the primary door.
 */
export function ExperienceLink({
  door,
  location,
  primary = false,
}: {
  door: DoorId;
  location: DemoLocation;
  primary?: boolean;
}) {
  return (
    <Link
      className={primary ? styles.button : styles.textLink}
      href={DOORS[door].href}
      onClick={() => trackExperience('self_service_click', { location, door })}
    >
      {DOORS[door].label}
      <span aria-hidden="true">↗</span>
    </Link>
  );
}
