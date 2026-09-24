'use client';

import Image from 'next/image';
import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react';
import { PRODUCT_PROOF as P } from '@/data/homepage-experience';
import styles from './experience.module.css';

// Keep Tab within the viewer, including browsers that otherwise let it reach browser chrome.
function containTab(event: KeyboardEvent<HTMLDialogElement>) {
  if (event.key !== 'Tab') return;
  const stops = event.currentTarget.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [tabindex="0"]'
  );
  const first = stops[0],
    last = stops[stops.length - 1];
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last?.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first?.focus();
  }
}

/** Native dialog provides modal focus containment and an inert page underneath.
 * Keep zoom inside a scrollable viewport; never crop the source image itself. */
export function ArchitectureViewer({ onClose }: { onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const closeButton = useRef<HTMLButtonElement>(null);
  const viewport = useRef<HTMLDivElement>(null);
  const [fit, setFit] = useState({ width: 0, height: 0 });
  const [zoom, setZoom] = useState(1);
  const titleId = useId();

  useEffect(() => {
    const element = dialog.current;
    if (!element) return;
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    element.showModal();
    closeButton.current?.focus();
    return () => {
      element.close();
      document.body.style.overflow = previousOverflow;
      opener?.focus({ preventScroll: true });
    };
  }, []);

  useEffect(() => {
    const element = viewport.current;
    if (!element) return;
    const measure = () => {
      const scale = Math.min((element.clientWidth - 32) / 2676, (element.clientHeight - 32) / 1708);
      setFit({ width: 2676 * scale, height: 1708 * scale });
    };
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    measure();
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const element = viewport.current;
    if (element)
      element.scrollTo({
        left: (element.scrollWidth - element.clientWidth) / 2,
        top: (element.scrollHeight - element.clientHeight) / 2,
      });
  }, [zoom, fit]);

  return (
    <dialog
      ref={dialog}
      className={styles.imageDialog}
      aria-labelledby={titleId}
      onClose={onClose}
      onKeyDown={containTab}
    >
      <div className={styles.viewerToolbar}>
        <div>
          <h2 id={titleId}>Planton on GKE</h2>
          <p>{P.capture}</p>
        </div>
        <div className={styles.viewerActions}>
          <button
            type="button"
            aria-label="Zoom out"
            disabled={zoom === 1}
            onClick={() => setZoom((value) => value - 1)}
          >
            −
          </button>
          <button type="button" onClick={() => setZoom(1)} aria-label="Fit diagram to screen">
            Fit
          </button>
          <button
            type="button"
            aria-label="Zoom in"
            disabled={zoom === 4}
            onClick={() => setZoom((value) => value + 1)}
          >
            +
          </button>
          <button
            aria-label="Close full-screen diagram"
            ref={closeButton}
            type="button"
            onClick={onClose}
            className={styles.viewerClose}
          >
            Close <span aria-hidden="true">×</span>
          </button>
        </div>
      </div>
      <div
        ref={viewport}
        className={styles.viewerViewport}
        tabIndex={0}
        aria-label="Architecture diagram. Use arrow keys to scroll when zoomed."
      >
        <div
          className={styles.viewerImage}
          style={{
            width: Math.max(0, fit.width * zoom + 32),
            height: Math.max(0, fit.height * zoom + 32),
          }}
        >
          <Image src={P.image} alt={P.alt} width={2676} height={1708} unoptimized loading="eager" />
        </div>
      </div>
      <p className={styles.viewerHint} aria-live="polite">
        {zoom === 1
          ? 'Complete diagram · zoom in to inspect resources'
          : `${zoom}× zoom · scroll to explore`}
      </p>
    </dialog>
  );
}
