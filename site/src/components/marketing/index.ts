/**
 * The marketing primitive library: the one set of building blocks every
 * marketing page composes. Colors come from the palette's role classes
 * (bg-canvas, bg-card, text-fg-secondary, border-edge); nothing here types a
 * hex, and nothing here carries copy. Pages import from this barrel and
 * never from a versioned landing folder.
 */
export * from './section';
export * from './typography';
export * from './buttons';
export * from './card';
export * from './badge';
export * from './grid';
export * from './icons';
export * from './divider';
export * from './quote';
export * from './metric';
export * from './terminal-window';
export * from './comparison-cell';
export * from './step';
export * from './record-window';
// Page sections: the pieces every story page composes (a hero, proof beside a
// record, doors to sibling pages, the doors themselves, a command to paste).
export * from './page-hero';
export * from './proof-list';
export * from './page-card';
export * from './doors';
export * from './command-block';
export * from './page-artifact';
export * from './provider-strip';
