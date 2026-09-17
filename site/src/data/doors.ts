/**
 * The door vocabulary: every call to action the marketing pages offer, each
 * with its one label and its one destination. A page names its doors by key
 * (`hosted`, `demo`, ...) and the Doors primitive renders the pair; no
 * section writes its own button, so the same door reads the same way on
 * every page and a destination changes in one place.
 *
 * Labels are Title Case (they are chrome). The hosted free tier leads on the
 * user's pages because it works on every device a link is opened on,
 * including the phone a thread link is often read on; the free desktop app is
 * the second door and the one the fine print explains. Trust pages lead with
 * a demo because their reader signs rather than installs.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { DESKTOP_DOWNLOAD_PATH } from './desktop-download.ts';
import { EVALUATION_URL } from './pricing.ts';

export type DoorId = 'hosted' | 'desktop' | 'demo' | 'pricing' | 'selfHostedDocs' | 'evaluation' | 'codingAgentsDocs' | 'cliDocs';

export interface Door {
  /** Title Case; the button's text. */
  label: string;
  /** A site path, a console path the link gate allows, or an absolute URL. */
  href: string;
}

export const DOORS: Record<DoorId, Door> = {
  hosted: { label: 'Start Free', href: '/signup' },
  desktop: { label: 'Download Planton Desktop', href: DESKTOP_DOWNLOAD_PATH },
  demo: { label: 'Book a Demo', href: '/book-demo' },
  pricing: { label: 'Pricing', href: '/pricing' },
  selfHostedDocs: { label: 'Read the Self-Hosting Docs', href: '/docs/self-hosting' },
  evaluation: { label: 'Start an Evaluation', href: EVALUATION_URL },
  codingAgentsDocs: { label: 'Set Up Your Coding Agent', href: '/docs/coding-agents' },
  cliDocs: { label: 'Install the CLI', href: '/docs/cli' },
};

/** A pair of doors: the one a page leads with and the one beside it. */
export interface DoorPair {
  primary: DoorId;
  secondary: DoorId;
}

/** The user's pair (landing, Product pages): start free, or download the desktop app. */
export const START_DOORS: DoorPair = { primary: 'hosted', secondary: 'desktop' };

/** The buyer's pair (Trust pages): book a demo, or read the pricing. */
export const BUYER_DOORS: DoorPair = { primary: 'demo', secondary: 'pricing' };
