import { HOMEPAGE as H } from '@/data/homepage';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { HeroExperience } from '@/components/marketing/HeroExperience';
import { SetupPromptExample } from '@/components/marketing/SetupPromptExample';
import { DemoLink } from '@/components/marketing/DemoLink';
import { ExperienceLink } from '@/components/marketing/ExperienceLink';
import styles from '@/components/marketing/homepage.module.css';

/**
 * Home, told as one story and built one or two sections at a time: the stand
 * and the team it serves (the hero), a brand-new cloud account to a running
 * app (the first app), and the start (the close).
 */
export function Homepage() {
  const F = H.firstApp;
  return (
    <HomepageAppearance>
      <main>
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
          <section id="first-app" className={styles.firstApp} aria-labelledby="first-app-title">
            <div className={styles.firstAppHeader}>
              <h2 id="first-app-title" className={styles.heading}>
                {F.title}
              </h2>
              <p className={styles.intro}>{F.intro}</p>
            </div>
            <figure className={styles.promptFigure}>
              <figcaption id="setup-prompt-label" className={styles.promptLabel}>
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
      </main>
    </HomepageAppearance>
  );
}
