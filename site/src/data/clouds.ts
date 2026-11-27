/**
 * The clouds Home lets a visitor choose between, in the order the picker
 * shows them. A cloud is listed only when Home has a real artifact for it (a
 * real run on that cloud's demo company). Google Cloud is the default: it is
 * the cloud whose whole walk, from a new account to a running app, is proven
 * (the story, chapter 1).
 *
 * The choice is page-wide: every section below the picker reads it. A section
 * with no real artifact for the chosen cloud shows the default cloud's and
 * says so, rather than leaving the reader with nothing.
 */
export type CloudId = 'aws' | 'gcp' | 'azure' | 'cloudflare';

export const CLOUDS: readonly { id: CloudId; name: string }[] = [
  { id: 'aws', name: 'AWS' },
  { id: 'gcp', name: 'Google Cloud' },
  { id: 'azure', name: 'Azure' },
  { id: 'cloudflare', name: 'Cloudflare' },
];

export const DEFAULT_CLOUD: CloudId = 'gcp';
