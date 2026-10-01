import Image from 'next/image';
import { HOMEPAGE as H } from '@/data/homepage';
import { PRODUCT_PROOF as P, CONTROL_COPY as C } from '@/data/homepage-experience';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { HeroExperience } from '@/components/marketing/HeroExperience';
import { OverviewVideo } from '@/components/marketing/OverviewVideo';
import { ProductProof } from '@/components/marketing/ProductProof';
import { ControlOwnership } from '@/components/marketing/ControlOwnership';
import { DemoLink } from '@/components/marketing/DemoLink';
import { ExperienceLink } from '@/components/marketing/ExperienceLink';
import {
  ArchitectureStories,
  DeliveryStory,
  AgentStory,
} from '@/components/marketing/DeferredWorkflows';
import styles from '@/components/marketing/homepage.module.css';
import workflowStyles from '@/components/marketing/workflows/workflows.module.css';

/** Engineering leaders: understand the platform, see the proof, then choose a next step. */
export function Homepage() {
  return (
    <HomepageAppearance>
      <main>
        <div className={styles.container}>
          <section className={styles.hero} aria-labelledby="homepage-title">
            <div className={styles.heroCopy}>
              <p className={styles.eyebrow}>{H.eyebrow}</p>
              <h1 id="homepage-title">
                {H.headline.map((line) => (
                  <span key={line}>{line}</span>
                ))}
              </h1>
              <p className={styles.intro}>{H.intro}</p>
              <div className={styles.actions}>
                <DemoLink location="hero" />
                <a className={styles.textLink} href="#how-it-works">
                  {H.seeHow}
                  <span aria-hidden="true">↓</span>
                </a>
              </div>
              <p className={styles.heroCaption}>{H.caption}</p>
            </div>
            <HeroExperience />
          </section>
          <OverviewVideo />
          <section id="how-it-works" className={styles.overview} aria-labelledby="overview-title">
            <p className={styles.eyebrow}>{H.overview.eyebrow}</p>
            <div className={styles.overviewHeader}>
              <h2 id="overview-title" className={styles.heading}>
                {H.overview.title}
              </h2>
              <p>{H.overview.intro}</p>
            </div>
            <ol className={styles.steps}>
              {H.overview.steps.map((s, i) => (
                <li key={s.title}>
                  <div className={styles.stepHeader}>
                    <span className={styles.stepNumber}>0{i + 1}</span>
                  </div>
                  <h3>{s.title}</h3>
                  <p>{s.text}</p>
                </li>
              ))}
            </ol>
          </section>
          <section
            id="infrastructure"
            className={workflowStyles.section}
            aria-labelledby="infrastructure-title"
          >
            <p className={styles.eyebrow}>{H.infrastructure.eyebrow}</p>
            <div className={workflowStyles.heading}>
              <h2 id="infrastructure-title" className={styles.heading}>
                {H.infrastructure.title}
              </h2>
              <div className={workflowStyles.copy}>
                <p>{H.infrastructure.intro}</p>
              </div>
            </div>
            <ArchitectureStories />
          </section>
          <section
            id="living-architecture"
            className={styles.overview}
            aria-labelledby="architecture-title"
          >
            <p className={styles.eyebrow}>{P.eyebrow}</p>
            <div className={styles.overviewHeader}>
              <h2 id="architecture-title" className={styles.heading}>
                {P.title}
              </h2>
              <p>{P.intro}</p>
            </div>
            <ProductProof />
          </section>
          <section
            id="delivery"
            className={workflowStyles.section}
            aria-labelledby="delivery-title"
          >
            <p className={styles.eyebrow}>{H.delivery.eyebrow}</p>
            <div className={workflowStyles.heading}>
              <h2 id="delivery-title" className={styles.heading}>
                {H.delivery.title}
              </h2>
              <div className={workflowStyles.copy}>
                <p>{H.delivery.intro}</p>
              </div>
            </div>
            <DeliveryStory />
          </section>
          <section id="coding-agents" className={styles.agents} aria-labelledby="agents-title">
            <p className={styles.eyebrow}>{H.agents.eyebrow}</p>
            <div className={styles.overviewHeader}>
              <h2 id="agents-title" className={styles.heading}>
                {H.agents.title}
              </h2>
              <p>{H.agents.intro}</p>
            </div>
            <AgentStory />
            <div className={styles.agentSetup}>
              <p>{H.agents.setup}</p>
              <ExperienceLink door="codingAgentsDocs" location="agents" />
            </div>
          </section>
          <section id="controls" className={styles.overview} aria-labelledby="controls-title">
            <p className={styles.eyebrow}>{C.eyebrow}</p>
            <div className={styles.overviewHeader}>
              <h2 id="controls-title" className={styles.heading}>
                {C.title}
              </h2>
              <p>{C.intro}</p>
            </div>
            <ControlOwnership />
            <p className={styles.adoptionNote}>{C.adoption}</p>
            <div className={styles.actions}>
              <DemoLink location="controls" />
              <ExperienceLink door="hosted" location="controls" />
            </div>
          </section>
          <section className={styles.proof} aria-labelledby="proof-title">
            <p className={styles.eyebrow}>{H.proof.eyebrow}</p>
            <h2 className={styles.heading} id="proof-title">
              {H.proof.title}
            </h2>
            <div className={styles.quotes}>
              {H.proof.quotes.map((q) => (
                <blockquote key={q.name}>
                  <p>{q.quote}</p>
                  <footer>
                    {q.avatar && <Image className={styles.quotePortrait} src={q.avatar} alt="" width={64} height={64} />}
                    <div className={styles.quoteAttribution}>
                      <strong>{q.name}</strong>
                      <span>
                        {q.role} · {q.company}
                      </span>
                    </div>
                  </footer>
                </blockquote>
              ))}
            </div>
          </section>
          <section className={styles.faq} aria-labelledby="faq-title">
            <div>
              <p className={styles.eyebrow}>{H.faq.eyebrow}</p>
              <h2 id="faq-title" className={styles.heading}>
                {H.faq.title}
              </h2>
            </div>
            <div className={styles.questions}>
              {H.faq.questions.map((q) => (
                <details key={q.question}>
                  <summary>{q.question}</summary>
                  <p>{q.answer}</p>
                </details>
              ))}
            </div>
          </section>
          <section className={styles.close} aria-labelledby="close-title">
            <p className={styles.eyebrow}>{H.close.eyebrow}</p>
            <h2 id="close-title" className={styles.heading}>
              {H.close.title}
            </h2>
            <p>{H.close.text}</p>
            <div className={styles.closeActions}>
              <DemoLink location="close" />
              <ExperienceLink door="hosted" location="close" />
            </div>
            <small>{H.close.note}</small>
          </section>
        </div>
      </main>
    </HomepageAppearance>
  );
}
