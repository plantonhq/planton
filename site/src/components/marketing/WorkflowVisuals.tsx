import { HOMEPAGE as H } from '@/data/homepage';
import styles from './homepage.module.css';

/** Shared node icon. Geometry remains editable; text is live HTML at every size. */
export function ResourceMark({ kind = 0 }: { kind?: number }) {
  return <svg viewBox="0 0 40 40" fill="none" stroke="currentColor" strokeWidth="1" aria-hidden="true" className={styles.resourceMark}>
    {kind === 2 ? <><ellipse cx="20" cy="11" rx="12" ry="5"/><path d="M8 11v18c0 7 24 7 24 0V11M8 20c0 7 24 7 24 0"/></> : kind === 1 ? <><path d="m20 5 14 8v15l-14 8-14-8V13Z M6 13l14 8 14-8M20 21v15"/><path d="m13 9 14 8"/></> : <><rect x="14" y="14" width="12" height="12" rx="2"/><path d="M20 4v10m0 12v10M4 20h10m12 0h10M8 8l7 7m10 10 7 7M8 32l7-7m10-10 7-7"/></>}
  </svg>;
}

/** A template and two environments, intentionally not a product screenshot. */
export function InfrastructureVisual() {
  const v=H.visuals;
  return <figure className={styles.visual} aria-label={v.configuration}>
    <div className={styles.visualTop}><span>{v.template}</span><span aria-hidden="true">↗</span></div>
    <div className={styles.templateName}><ResourceMark/>{v.templateFile}</div>
    <div className={styles.resourceRow}>{v.resources.map((r,i)=><div key={r}><ResourceMark kind={i}/><span>{r}</span></div>)}</div>
    <div className={styles.fork} aria-hidden="true"/>
    <div className={styles.environmentPair}>{v.environments.map((e,i)=><div key={e}><span className={styles.environmentIndex}>0{i+1}</span><strong>{e}</strong><div className={styles.miniNodes} aria-hidden="true"><i/><b/><i/><b/><i/></div></div>)}</div>
    <figcaption className={styles.visualCaption}>{v.configuration}<span>{H.illustration}</span></figcaption>
  </figure>;
}

/** A responsive timeline; mobile changes direction rather than shrinking text. */
export function DeliveryVisual() {
  const v=H.visuals;
  return <figure className={`${styles.visual} ${styles.deliveryVisual}`} aria-label={v.deliveryTitle}>
    <div className={styles.visualTop}><span>{v.service}</span><span>⑂ {v.branch}</span></div>
    <ol className={styles.deliverySteps}>{v.stages.map((stage,i)=><li key={stage}><span className={styles.stepNode} aria-hidden="true">{i===3 ? 'Ⅱ' : i===4 ? '↗' : '✓'}</span><div><strong>{stage}</strong><span>{v.stageDetails[i]}</span></div></li>)}</ol>
    <figcaption className={styles.visualCaption}>{v.deliveryNote}<span>{H.illustration}</span></figcaption>
  </figure>;
}

/** Coverage is part of the illustration; examples never pretend to be live evidence. */
export function ReviewVisual() {
  const v=H.visuals;
  return <figure className={`${styles.visual} ${styles.reviewVisual}`}>
    <div className={styles.visualTop}><span>{v.reviewTitle}</span><span aria-hidden="true">↗</span></div>
    <h3>{v.reviewEnvironment}</h3>
    <dl className={styles.evidence}>{v.reviewRows.map(([label,value])=><div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
    <div className={styles.approval}><span aria-hidden="true">Ⅱ</span>{v.approval}</div>
    <div className={styles.record}><span className={styles.mono}>{v.recordTitle}</span>{v.recordFields.map(f=><span key={f}>↳ {f}</span>)}</div>
    <figcaption className={styles.visualCaption}>{v.evidenceNote}<span>{H.illustration}</span></figcaption>
  </figure>;
}

export function OwnershipVisual() {
  const v=H.visuals;
  return <figure className={`${styles.visual} ${styles.ownership}`}>
    <div className={styles.visualTop}>{v.ownershipTitle}</div>
    <div className={styles.postures}>{v.postures.map(p=><span key={p}>{p}</span>)}</div>
    <div className={styles.downlink} aria-hidden="true">↓</div>
    <div className={styles.cloudBoundary}><span className={styles.mono}>{v.boundary}</span><div>{v.workloads.map((w,i)=><span key={w}><ResourceMark kind={i}/>{w}</span>)}</div></div>
    <figcaption className={styles.visualCaption}>{v.ownershipNote}<span>{H.illustration}</span></figcaption>
  </figure>;
}
