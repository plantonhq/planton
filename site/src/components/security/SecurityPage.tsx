import Link from 'next/link';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { ExperienceLink } from '@/components/marketing/ExperienceLink';
import { DemoLink } from '@/components/marketing/DemoLink';
import { Artifact, Check } from '@/components/marketing/Artifact';
import { SECURITY as S } from '@/data/security';
import styles from './security.module.css';

/**
 * The Security page: each answer beside the real thing behind it (a
 * permission list, a trust policy, a table of who may approve), then the docs
 * that go deeper and the two doors. Every word comes from src/data/security.ts.
 */
export function SecurityPage() {
  return (
    <HomepageAppearance>
      <main>
        <div className={styles.container}>
          <section className={styles.top} aria-labelledby="security-title">
            <p className={styles.eyebrow}>{S.eyebrow}</p>
            <h1 id="security-title">{S.headline}</h1>
            <p className={styles.intro}>{S.lede}</p>
          </section>
          {S.sections.map((section) => (
            <section key={section.id} id={section.id} className={styles.answer} aria-labelledby={`${section.id}-title`}>
              <div className={styles.words}>
                <h2 id={`${section.id}-title`} className={styles.heading}>
                  {section.title}
                </h2>
                {section.paragraphs.map((p) => (
                  <p key={p}>{p}</p>
                ))}
                {section.check && <Check check={section.check} />}
              </div>
              {section.artifact && <Artifact artifact={section.artifact} id={section.id} />}
            </section>
          ))}
          <section className={styles.more} aria-labelledby="read-more-title">
            <div>
              <h2 id="read-more-title" className={styles.heading}>
                {S.more.title}
              </h2>
              <p className={styles.contact}>
                {S.more.contact.text} <a href={`mailto:${S.more.contact.address}`}>{S.more.contact.address}</a>
              </p>
              <div className={styles.actions}>
                <ExperienceLink door="hosted" location="security" primary />
                <DemoLink location="security" />
              </div>
            </div>
            <ul>
              {S.more.docs.map((doc) => (
                <li key={doc.href}>
                  <Link href={doc.href}>{`${doc.label} →`}</Link>
                </li>
              ))}
            </ul>
          </section>
        </div>
      </main>
    </HomepageAppearance>
  );
}
