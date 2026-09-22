/**
 * This landing version's primitives now live in their stable, unversioned
 * home, `@/components/marketing`. This file exists only so the v3 sections
 * beside it keep rendering unchanged while the folder is kept for rollback;
 * nothing outside a versioned landing folder may import from inside one.
 */
export * from '@/components/marketing';
