/**
 * Shared pieces the content pages (changelog, docs) and the pricing page
 * compose. Marketing-page primitives live in
 * `@/components/marketing`; anything added here must have a consumer, and a
 * file nobody imports is deleted, not kept for later.
 */
export * from './typography';
export * from './CodeBlock';
export { default as MermaidDiagram } from './MermaidDiagram';
export { PageActions } from './PageActions';
