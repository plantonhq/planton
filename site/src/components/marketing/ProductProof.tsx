'use client';
import Image from 'next/image';
import { useId, useState, type MouseEvent } from 'react';
import { ArchitectureViewer } from './ArchitectureViewer';
import { PRODUCT_PROOF as P } from '@/data/homepage-experience';
import { trackExperience } from '@/lib/demo-analytics';
import styles from './experience.module.css';
import { experienceTheme } from './HeroExperience';

/** An actual captured interface, with explanations outside the image. No fabricated live UI. */
export function ProductProof() {
  const [selected, setSelected] = useState(0),
    id = useId();
  const [viewerOpen, setViewerOpen] = useState(false);
  const openViewer = (event: MouseEvent<HTMLAnchorElement>) => {
    event.preventDefault();
    event.currentTarget.focus();
    setViewerOpen(true);
  };
  const note = P.notes[selected];
  return (
    <div className={styles.product} style={experienceTheme}>
      <figure className={styles.productFrame}>
        <div className={styles.frameHeader}>
          <strong>{P.frame}</strong>
          <span>{P.capture}</span>
        </div>
        <a
          href={P.image}
          onClick={openViewer}
          aria-haspopup="dialog"
          aria-label="Open the full architecture screenshot"
        >
          <Image
            src={P.image}
            alt={P.alt}
            width={2676}
            height={1708}
            loading="lazy"
            unoptimized
            sizes="(max-width:760px) 100vw, 1280px"
          />
        </a>
        <figcaption>
          <p>
            <strong className={styles.captionLead}>{P.captionLead}</strong>
            {P.caption}
          </p>
          <a
            href={P.image}
            onClick={openViewer}
            aria-haspopup="dialog"
            className={styles.fullImageLink}
          >
            Open Full-Screen Diagram <span aria-hidden="true">⤢</span>
          </a>
        </figcaption>
      </figure>
      <div className={styles.annotations} aria-label="Architecture view explanations">
        {P.notes.map((item, i) => (
          <button
            type="button"
            key={item.title}
            aria-pressed={selected === i}
            aria-controls={`${id}-note`}
            onClick={() => {
              setSelected(i);
              trackExperience('product_example_select', { selection: item.focus });
            }}
          >
            <span>0{i + 1}</span>
            {item.title}
          </button>
        ))}
      </div>
      <div id={`${id}-note`} className={styles.productNote} aria-live="polite">
        <p>{note.text}</p>
        <span>{note.detail}</span>
      </div>
      {viewerOpen && <ArchitectureViewer onClose={() => setViewerOpen(false)} />}
      <noscript>
        <ul>
          {P.notes.map((item) => (
            <li key={item.title}>
              <strong>{item.title}</strong>
              <p>{item.text}</p>
            </li>
          ))}
        </ul>
      </noscript>
    </div>
  );
}
