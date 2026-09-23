'use client';

import { useCallback, useRef, useState } from 'react';
import { BookDemoForm } from './BookDemoForm';
import { BookDemoScheduler } from './BookDemoScheduler';
import type { DemoFormData } from './types';
import { DEMO_COPY as C } from '@/data/homepage';
import styles from './book-demo.module.css';

/** The same form → calendar journey, with distinct submitted and booked states. */
export function BookDemoPage() {
  const [formData, setFormData] = useState<DemoFormData | null>(null);
  const [booked, setBooked] = useState(false);
  const heading = useRef<HTMLHeadingElement>(null);
  const handleSuccess = useCallback((data: DemoFormData) => {
    setFormData(data);
    requestAnimationFrame(() => { heading.current?.focus(); window.scrollTo({ top: 0, behavior: 'instant' }); });
  }, []);
  const handleBooked = useCallback(() => setBooked(true), []);
  return <main className={styles.page}>
    <div className={styles.layout}>
      <section className={styles.copy}>
        <p className={styles.eyebrow}>{C.eyebrow}</p>
        <h1 ref={heading} tabIndex={-1}>{formData ? C.schedulerTitle : C.title}</h1>
        <p className={styles.intro}>{formData ? C.schedulerText : C.intro}</p>
        <div className={styles.agenda}><h2>{C.agendaTitle}</h2><ol>{C.agenda.map((item,i)=><li key={item}><span>0{i+1}</span>{item}</li>)}</ol><p>{C.note}</p></div>
      </section>
      <section className={styles.formColumn} aria-label={formData ? C.steps[1] : C.steps[0]}>
        <ol className={styles.progress}>{C.steps.map((step,i)=><li key={step} aria-current={(formData ? i===1 : i===0) ? 'step' : undefined}><span>0{i+1}</span>{step}</li>)}</ol>
        {formData ? <BookDemoScheduler formData={formData} onBooked={handleBooked}/> : <BookDemoForm onSuccess={handleSuccess}/>}
        {booked && <div className={styles.confirmation} role="status"><h2>{C.confirmation}</h2><p>{C.confirmationText}</p></div>}
      </section>
    </div>
  </main>;
}
