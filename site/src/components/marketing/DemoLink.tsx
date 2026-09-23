'use client';

import Link from 'next/link';
import { DOORS } from '@/data/doors';
import { trackDemo, type DemoLocation } from '@/lib/demo-analytics';
import styles from './homepage.module.css';

/** The existing demo door with explicit, non-identifying attribution. */
export function DemoLink({ location }: { location: DemoLocation }) {
  return <Link href={DOORS.demo.href} className={styles.button} onClick={() => trackDemo('demo_cta_click', location)}>
    {DOORS.demo.label}<span aria-hidden="true">↗</span>
  </Link>;
}
