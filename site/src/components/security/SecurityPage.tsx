import Link from 'next/link';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { ExperienceLink } from '@/components/marketing/ExperienceLink';
import { DemoLink } from '@/components/marketing/DemoLink';
import { SECURITY as S, type SecurityArtifact, type SecurityCheck } from '@/data/security';
import styles from './security.module.css';

/** An outside page opens in a new tab; a page of this site stays in this one. */
function linkProps(href: string) {
  return href.startsWith('http') ? { target: '_blank', rel: 'noreferrer' } : {};
}

function Artifact({ artifact, id }: { artifact: SecurityArtifact; id: string }) {
  const labelId = `${id}-artifact`;
  return (
    <figure aria-labelledby={labelId}>
      <figcaption id={labelId} className={styles.label}>
        {artifact.label}
      </figcaption>
      {artifact.kind === 'code' ? (
        <div className={styles.code}>
          <pre>{artifact.lines.join('\n')}</pre>
        </div>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              {artifact.columns.map((c) => (
                <th key={c} scope="col">
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {artifact.rows.map((row) => (
              <tr key={row[0]}>
                {row.map((cell, i) => (
                  <td key={i}>{cell}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </figure>
  );
}

function Check({ check }: { check: SecurityCheck }) {
  return (
    <p className={styles.check}>
      {check.href ? (
        <Link href={check.href} {...linkProps(check.href)}>
          {check.text}
        </Link>
      ) : (
        check.text
      )}
      {check.command && <code>{check.command}</code>}
    </p>
  );
}

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
