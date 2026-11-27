'use client';

import { Artifact, Folded } from '@/components/marketing/Artifact';
import { useCloud } from '@/components/marketing/CloudChoice';
import { DEFAULT_CLOUD } from '@/data/clouds';
import { HOMEPAGE, WATCH_RUN } from '@/data/homepage';
import styles from '@/components/marketing/homepage.module.css';

/**
 * Watch it deploy: the first section that leads with a screen of the product.
 * The heading and the words (left and right on the two columns' line), the
 * run's screen across both, then the record, and the plan, folded, one click
 * away, opening across both. Its run is Google Cloud's, the only cloud with one, so a reader who
 * chose another cloud is told what they're looking at. A client module
 * because the chosen cloud is state; every word comes from homepage.ts.
 */
export function WatchSection() {
  const W = HOMEPAGE.watch;
  const { cloud } = useCloud();
  return (
    <section
      id="watch"
      className={`${styles.section} ${styles.screenFirst}`}
      aria-labelledby="watch-title"
    >
      <h2 id="watch-title" className={styles.heading}>
        {W.title}
      </h2>
      <div className={styles.words}>
        <p className={styles.intro}>{W.intro}</p>
        {cloud !== DEFAULT_CLOUD && <p className={styles.note}>{W.otherCloud}</p>}
      </div>
      <div className={styles.wide}>
        <Artifact artifact={WATCH_RUN.screen} id="watch-run" />
      </div>
      <div className={styles.words}>
        <p>{W.record}</p>
      </div>
      {/* The plan opens across both columns, where its rows fit as the CLI printed them. */}
      <div className={`${styles.words} ${styles.wide}`}>
        <Folded artifact={WATCH_RUN.plan} id="watch-plan" />
      </div>
    </section>
  );
}
