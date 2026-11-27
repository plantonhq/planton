'use client';

import { Artifact, Check } from '@/components/marketing/Artifact';
import { CloudPicker, useCloud } from '@/components/marketing/CloudChoice';
import { AGENT_RUNS, HOMEPAGE } from '@/data/homepage';
import styles from '@/components/marketing/homepage.module.css';

/**
 * Infrastructure asked of a coding agent, on the visitor's cloud: the picker,
 * the claim and the ask (left), what the agent wrote and the check it ran
 * (right), then the code that deploys it and how to prove it (left, under the
 * ask). The picker here sets the cloud for every section below it. A client module
 * because the chosen cloud is state; every word comes from homepage.ts.
 */
export function InfrastructureSection() {
  const I = HOMEPAGE.infrastructure;
  const { cloud } = useCloud();
  const run = AGENT_RUNS[cloud];
  return (
    <section
      id="infrastructure"
      className={`${styles.section} ${styles.split}`}
      aria-labelledby="infrastructure-title"
    >
      <div className={`${styles.words} ${styles.lead}`}>
        <CloudPicker label={I.pickerLabel} />
        <h2 id="infrastructure-title" className={styles.heading}>
          {I.title}
        </h2>
        <p className={styles.intro}>{I.intro}</p>
        <p className={styles.intro}>{I.next}</p>
        <figure className={styles.ask}>
          <figcaption>{run.askLabel}</figcaption>
          <blockquote>{run.ask}</blockquote>
        </figure>
      </div>
      <div className={styles.artifacts}>
        <Artifact artifact={run.manifest} id="infrastructure-manifest" />
        <Artifact artifact={run.validation} id="infrastructure-validation" />
      </div>
      <div className={`${styles.words} ${styles.proof}`}>
        <Artifact artifact={run.proven} id="infrastructure-proven" />
        {run.checks.map((check) => (
          <Check key={check.text} check={check} />
        ))}
        <p className={styles.note}>{run.reach}</p>
      </div>
    </section>
  );
}
