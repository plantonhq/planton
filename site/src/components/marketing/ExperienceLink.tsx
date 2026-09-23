'use client';
import Link from 'next/link';
import { DOORS, type DoorId } from '@/data/doors';
import { trackExperience, type DemoLocation } from '@/lib/demo-analytics';
import styles from './homepage.module.css';
export function ExperienceLink({
  door,
  location,
}: {
  door: DoorId;
  location: DemoLocation | 'agents';
}) {
  return (
    <Link
      className={styles.textLink}
      href={DOORS[door].href}
      onClick={() => trackExperience('self_service_click', { location, door })}
    >
      {DOORS[door].label}
      <span aria-hidden="true">↗</span>
    </Link>
  );
}
