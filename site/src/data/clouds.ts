/**
 * The clouds Home lets a visitor choose between, in the order the picker
 * shows them. A cloud is listed only when Home has a real artifact for it (a
 * real run on that cloud's demo company); Cloudflare joins when its demo
 * company's run exists. Google Cloud is the default: it is the cloud whose
 * whole walk, from a new account to a running app, is proven (the story,
 * chapter 1).
 */
export type CloudId = 'aws' | 'gcp' | 'azure';

export const CLOUDS: readonly { id: CloudId; name: string }[] = [
  { id: 'aws', name: 'AWS' },
  { id: 'gcp', name: 'Google Cloud' },
  { id: 'azure', name: 'Azure' },
];

export const DEFAULT_CLOUD: CloudId = 'gcp';
