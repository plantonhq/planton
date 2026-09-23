import { HOMEPAGE as H } from '@/data/homepage';
import { HomepageAppearance } from '@/components/marketing/HomepageAppearance';
import { ArchitecturePlanes } from '@/components/marketing/ArchitecturePlanes';
import { DemoLink } from '@/components/marketing/DemoLink';
import { ReviewVisual, OwnershipVisual } from '@/components/marketing/WorkflowVisuals';
import styles from '@/components/marketing/homepage.module.css';
import { ArchitectureTabs } from '@/components/marketing/workflows/ArchitectureTabs';
import { WorkflowExplainer } from '@/components/marketing/workflows/WorkflowExplainer';
import workflowStyles from '@/components/marketing/workflows/workflows.module.css';

/** Engineering-leader homepage. Copy and evidence live in the homepage record. */
export function Homepage() {
  return <HomepageAppearance><main>
    <div className={styles.container}>
      <section className={styles.hero} aria-labelledby="homepage-title">
        <div className={styles.heroCopy}>
          <p className={styles.eyebrow}>{H.eyebrow}</p>
          <h1 id="homepage-title">{H.headline.map(line=><span key={line}>{line}</span>)}</h1>
          <p className={styles.intro}>{H.intro}</p>
          <div className={styles.actions}><DemoLink location="hero"/><a className={styles.textLink} href="#how-it-works">{H.seeHow}<span aria-hidden="true">↓</span></a></div>
          <p className={styles.heroCaption}>{H.caption}</p>
        </div>
        <ArchitecturePlanes/>
      </section>
      <section id="how-it-works" className={styles.overview} aria-labelledby="overview-title">
        <p className={styles.eyebrow}>{H.overview.eyebrow}</p>
        <div className={styles.overviewHeader}><h2 id="overview-title" className={styles.heading}>{H.overview.title}</h2><p>{H.overview.intro}</p></div>
        <ol className={styles.steps}>{H.overview.steps.map((s,i)=><li key={s.title}><div className={styles.stepHeader}><span className={styles.stepNumber}>0{i+1}</span></div><h3>{s.title}</h3><p>{s.text}</p><span className={styles.stepLabel}>{s.label}</span></li>)}</ol>
      </section>
      <section id="coding-agents" className={styles.agents} aria-labelledby="agents-title">
        <p className={styles.eyebrow}>{H.agents.eyebrow}</p>
        <div className={styles.overviewHeader}><h2 id="agents-title" className={styles.heading}>{H.agents.title}</h2><p>{H.agents.intro}</p></div>
        <WorkflowExplainer storyId="agents" />
        <div className={styles.agentSetup}><p>{H.agents.setup}</p><a className={styles.textLink} href={H.agents.href}>{H.agents.link}<span aria-hidden="true">↗</span></a></div>
      </section>
      <section id="infrastructure" className={workflowStyles.section} aria-labelledby="infrastructure-title">
        <p className={styles.eyebrow}>{H.infrastructure.eyebrow}</p>
        <div className={workflowStyles.heading}>
          <h2 id="infrastructure-title" className={styles.heading}>{H.infrastructure.title}</h2>
          <div className={workflowStyles.copy}><p>{H.infrastructure.intro}</p><p>{H.infrastructure.note}</p></div>
        </div>
        <ArchitectureTabs />
      </section>
      <section id="delivery" className={workflowStyles.section} aria-labelledby="delivery-title">
        <p className={styles.eyebrow}>{H.delivery.eyebrow}</p>
        <div className={workflowStyles.heading}>
          <h2 id="delivery-title" className={styles.heading}>{H.delivery.title}</h2>
          <div className={workflowStyles.copy}><p>{H.delivery.intro}</p><p>{H.delivery.note}</p></div>
        </div>
        <WorkflowExplainer storyId="delivery" />
      </section>
      <section className={styles.feature} aria-labelledby="controls-title">
        <div className={styles.featureCopy}><p className={styles.eyebrow}>{H.controls.eyebrow}</p><h2 id="controls-title" className={styles.heading}>{H.controls.title}</h2><p>{H.controls.intro}</p><ul className={styles.controlPoints}>{H.controls.points.map(p=><li key={p.title}><h3>{p.title}</h3><p>{p.text}</p></li>)}</ul><p>{H.controls.note}</p><div className={styles.actions}><DemoLink location="controls"/></div></div>
        <ReviewVisual/>
      </section>
      <section className={styles.proof} aria-labelledby="proof-title"><p className={styles.eyebrow}>{H.proof.eyebrow}</p><h2 id="proof-title" className={styles.heading}>{H.proof.title}</h2><div className={styles.quotes}>{H.proof.quotes.map(q=><blockquote key={q.name}><p>{q.quote}</p><footer><strong>{q.name}</strong><span>{q.role} · {q.company}</span></footer></blockquote>)}</div></section>
      <section className={styles.feature} aria-labelledby="adoption-title"><div className={styles.featureCopy}><p className={styles.eyebrow}>{H.adoption.eyebrow}</p><h2 id="adoption-title" className={styles.heading}>{H.adoption.title}</h2><p>{H.adoption.intro}</p>{H.adoption.paragraphs.map(p=><p key={p}>{p}</p>)}</div><OwnershipVisual/></section>
      <section className={styles.faq} aria-labelledby="faq-title"><div><p className={styles.eyebrow}>{H.faq.eyebrow}</p><h2 id="faq-title" className={styles.heading}>{H.faq.title}</h2></div><div className={styles.questions}>{H.faq.questions.map(q=><details key={q.question}><summary>{q.question}</summary><p>{q.answer}</p></details>)}</div></section>
      <section className={styles.close} aria-labelledby="close-title"><p className={styles.eyebrow}>{H.close.eyebrow}</p><h2 id="close-title" className={styles.heading}>{H.close.title}</h2><p>{H.close.text}</p><DemoLink location="close"/><small>{H.close.note}</small></section>
    </div>
  </main></HomepageAppearance>;
}
