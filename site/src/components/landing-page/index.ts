/**
 * The active landing page version. Each version is a dated folder holding
 * only what is version-specific: its section composition and its copy. The
 * primitives every version composes live in `@/components/marketing` and are
 * never versioned. Only the active version stays in the tree; earlier ones
 * live in git history.
 */
export * from './v6-2026-09-23';
