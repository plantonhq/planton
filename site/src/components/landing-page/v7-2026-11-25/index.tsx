import { HOMEPAGE as H, SELF_HOSTED_WALK } from '@/data/homepage';
import { Artifact } from '@/components/marketing/Artifact';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { HeroExperience } from '@/components/marketing/HeroExperience';
import { SetupPromptExample } from '@/components/marketing/SetupPromptExample';
import { CloudChoiceProvider } from '@/components/marketing/CloudChoice';
import { InfrastructureSection } from './InfrastructureSection';
import { DemoLink } from '@/components/marketing/DemoLink';
import { ExperienceLink } from '@/components/marketing/ExperienceLink';
import styles from '@/components/marketing/homepage.module.css';

/**
 * Home, told as one story and built one or two sections at a time: the stand
 * and the team it serves (the hero), a brand-new cloud account to a running
 * app (the first app), infrastructure asked of a coding agent (what it writes
 * and what already deploys it), the same Planton on the visitor's own
 * Kubernetes cluster (run it yourself), and the start (the close). Later
 * sections go in above "run it yourself", which stays last before the close.
 */
export function Homepage() {
  const F = H.firstApp;
  const S = H.selfHosted;
  return (
    <HomepageAppearance>
      <main>
        <CloudChoiceProvider>
          <div className={styles.container}>
            <section className={styles.hero} aria-labelledby="homepage-title">
              <div className={styles.heroCopy}>
                <h1 id="homepage-title">{H.headline}</h1>
                <p className={styles.intro}>{H.intro}</p>
                <div className={styles.actions}>
                  <ExperienceLink door="hosted" location="hero" primary />
                  <DemoLink location="hero" />
                </div>
                <p className={styles.heroCaption}>{H.caption}</p>
              </div>
              <HeroExperience />
            </section>
            <section id="first-app" className={styles.section} aria-labelledby="first-app-title">
              <div className={styles.sectionHeader}>
                <h2 id="first-app-title" className={styles.heading}>
                  {F.title}
                </h2>
                <p className={styles.intro}>{F.intro}</p>
              </div>
              <figure className={styles.figure}>
                <figcaption id="setup-prompt-label" className={styles.label}>
                  {F.promptLabel}
                </figcaption>
                <SetupPromptExample labelledBy="setup-prompt-label" />
              </figure>
              <div className={styles.who}>
                {[F.you, F.agent].map((side) => (
                  <div key={side.title}>
                    <h3>{side.title}</h3>
                    <ul>
                      {side.steps.map((step) => (
                        <li key={step}>{step}</li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </section>
            <InfrastructureSection />
            <section
              id="run-it-yourself"
              className={`${styles.section} ${styles.split}`}
              aria-labelledby="run-it-yourself-title"
            >
              <div className={`${styles.words} ${styles.lead}`}>
                <h2 id="run-it-yourself-title" className={styles.heading}>
                  {S.title}
                </h2>
                <p className={styles.intro}>{S.intro}</p>
                <p>{S.setup}</p>
                <p>{S.same}</p>
              </div>
              <div className={styles.artifacts}>
                <Artifact artifact={SELF_HOSTED_WALK.install} id="run-it-yourself-install" />
                <Artifact artifact={SELF_HOSTED_WALK.status} id="run-it-yourself-status" />
              </div>
              <div className={`${styles.words} ${styles.proof}`}>
                <p>{S.seats}</p>
                <p className={styles.note}>{S.requirements}</p>
                <ExperienceLink door="selfHostedDocs" location="self-hosted" />
              </div>
            </section>
            <section className={styles.close} aria-labelledby="close-title">
              <h2 id="close-title" className={styles.heading}>
                {H.close.title}
              </h2>
              <p>{H.close.text}</p>
              <div className={styles.closeActions}>
                <ExperienceLink door="hosted" location="close" primary />
                <DemoLink location="close" />
              </div>
              <small>{H.close.note}</small>
            </section>
          </div>
        </CloudChoiceProvider>
      </main>
    </HomepageAppearance>
  );
}
