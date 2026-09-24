/**
 * Landing page v5 (2026-09-17): the page is the story.
 *
 * What changed from v4, and why:
 * - Every sentence about what Planton is or does reads from src/data/story.ts
 *   (the thirteen chapters), src/data/positioning.ts (the vocabulary law),
 *   src/data/platform-stats.ts, src/data/pricing.ts, and
 *   src/data/testimonials.ts. No section carries a claim of its own; a
 *   section is the composition of one chapter and the labels that frame it.
 * - The order is the story's order: the wall, what Planton is, verified
 *   before it exists, your rules hold, the record, services from Git, bring
 *   what you have, runs where you decide, who it is for, proof, and a close
 *   that folds how-it-compares and start into one beat.
 * - The hero's animated typing loop is gone. It depended on wall-clock timing,
 *   ignored prefers-reduced-motion, and could not be compared between two
 *   builds. Its replacement is a still frame of the proof moment: one prompt,
 *   the composed components, the cost fact, the policy, the posture.
 * - No shared.tsx of its own: sections import the marketing primitives from
 *   `@/components/marketing` and use the palette's role classes only.
 * - Sections are server components; the page ships as HTML first.
 *
 * Each section's intent and its prohibitions are stated at the top of its
 * file. To roll back, point src/components/landing-page/index.ts at v4.
 */
export { Hero } from './Hero';
export { TheWall } from './TheWall';
export { WhatPlantonIs } from './WhatPlantonIs';
export { VerifiedBeforeItExists } from './VerifiedBeforeItExists';
export { YourRulesHold } from './YourRulesHold';
export { TheRecord } from './TheRecord';
export { ServicesShipFromGit } from './ServicesShipFromGit';
export { BringWhatYouHave } from './BringWhatYouHave';
export { RunsWhereYouDecide } from './RunsWhereYouDecide';
export { WhoItIsFor } from './WhoItIsFor';
export { Proof } from './Proof';
export { Close } from './Close';

export { Homepage } from './Homepage';
