'use client';

import { useId, useState } from 'react';
import { HOMEPAGE as H } from '@/data/homepage';
import styles from './homepage.module.css';

/** Editable vector architecture: one graph repeated through three steps. */
export function ArchitecturePlanes() {
  const [selected, setSelected] = useState(0);
  const id = useId().replaceAll(':', '');
  const nodes = [[-125, 0], [-45, -52], [112, 3], [36, 44], [-20, 3]];
  const links = [[0, 1], [1, 2], [2, 3], [3, 4], [4, 0], [1, 4]];
  return <figure className={styles.architecture}>
    <div className={styles.artboard}>
      <svg viewBox="0 0 620 620" role="img" aria-labelledby={`${id}-title ${id}-desc`}>
        <title id={`${id}-title`}>{H.visuals.heroTitle}</title><desc id={`${id}-desc`}>{H.visuals.heroDescription}</desc>
        <defs>
          <linearGradient id={`${id}-plane`} x1="0" y1="0" x2="1" y2="1"><stop stopColor="var(--hp-ink)" stopOpacity=".29"/><stop offset="1" stopColor="var(--hp-ink)" stopOpacity=".035"/></linearGradient>
          <linearGradient id={`${id}-edge`}><stop stopColor="var(--hp-ink)" stopOpacity=".55"/><stop offset="1" stopColor="var(--hp-ink)" stopOpacity=".14"/></linearGradient>
        </defs>
        {[2, 1, 0].map((layer) => <g key={layer} className={styles.plane} opacity={selected === layer ? 1 : .53} transform={`translate(265 ${143 + layer * 162})`}>
          <path d="M-191 0 0-109 191 0 0 109Z" transform="translate(0 4)" fill="var(--hp-panel)" stroke="var(--hp-line)"/>
          <path d="M-191 0 0-109 191 0 0 109Z" fill="var(--hp-canvas)"/>
          <path d="M-191 0 0-109 191 0 0 109Z" fill={`url(#${id}-plane)`} stroke={`url(#${id}-edge)`}/>
          {links.map(([a,b], i) => <path key={i} d={`M${nodes[a].join(' ')}L${nodes[b].join(' ')}`} fill="none" stroke="var(--hp-ink)" strokeOpacity=".45" strokeDasharray={layer === 1 ? '3 4' : undefined}/>)}
          {nodes.map(([x,y], i) => <g key={i}><path d={`M${x-11} ${y}l11 -6 11 6 -11 6Z`} fill="var(--hp-panel)" stroke="var(--hp-ink)" strokeOpacity=".7"/><path d={`M${x-11} ${y}v5l11 6 11-6v-5M${x} ${y+6}v5`} fill="none" stroke="var(--hp-ink)" strokeOpacity=".4"/></g>)}
        </g>)}
        <path d="M265 143V467" stroke="var(--hp-ink)" strokeWidth="1.5"/>
        {[0,1,2].map(i => <g key={i}>
          <path d={`M265 ${143+i*162}H477`} stroke="var(--hp-ink)" strokeOpacity=".3"/>
          <ellipse cx="265" cy={143+i*162} rx={selected===i ? 10 : 6} ry={selected===i ? 6 : 4} fill="var(--hp-ink)"/>
          <circle cx="477" cy={143+i*162} r="2" fill="var(--hp-ink)"/>
        </g>)}
      </svg>
      {H.layers.map((layer,i) => <button key={layer.name} type="button" className={styles.layerButton} style={{ top: `${(143+i*162)/620*100}%` }} aria-pressed={selected === i} aria-controls={`${id}-detail`} onClick={() => setSelected(i)}><span>0{i+1}</span>{layer.name}</button>)}
    </div>
    <figcaption className={styles.artCaption}><span>FIG. 01 / {H.illustration}</span><span>DESIGN → REVIEW → DEPLOY</span></figcaption>
    <div id={`${id}-detail`} className={styles.layerDetail} aria-live="polite" aria-atomic="true"><span className={styles.mono}>0{selected+1}</span><div><strong>{H.layers[selected].title}</strong><p>{H.layers[selected].text}</p></div></div>
  </figure>;
}
