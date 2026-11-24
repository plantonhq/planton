/**
 * The numbers the marketing pages state, counted from the open-source
 * repository and carried nowhere else in prose. A page that says "850+"
 * reads it from here; the story records how each number was counted so the
 * next reconciliation repeats the method instead of guessing.
 *
 * Counted 2026-11-24 against plantonhq/planton:
 *   catalog kinds   = folders under catalog/<provider>/ carrying a spec.proto, internal aa_* folders excluded (852:
 *                     aws 205, gcp 186, azure 185, kubernetes 143, cloudflare 66, digitalocean 31, auth0 17, stripe 16, openfga 3)
 *   providers         = the provider folders under catalog/ (aws, azure, gcp, kubernetes, cloudflare, digitalocean, auth0, stripe, openfga)
 *   control profiles  = kinds carrying a controls.yaml (851)
 *   controls          = the control catalog's entries (17, in 6 categories)
 *
 * Planton ships no Infra Charts of its own: every chart belongs to an organization, so no chart count is kept.
 *
 * Rounded figures ("850+") are what a page prints; the exact counts are the
 * record. A change to the catalog changes this file first.
 */
export const PLATFORM_STATS = {
  /** Printed. Catalog kinds in the catalog, rounded down to the fifty. */
  DEPLOYMENT_MODULE_COUNT: '850+',
  CLOUD_PROVIDER_COUNT: '9',
  CONTROL_COUNT: '17',
  CONTROL_CATEGORY_COUNT: '6',
  /** The year the first customer went to production on Planton. */
  IN_PRODUCTION_SINCE: '2023',
} as const;

/**
 * The providers the catalog covers, in the order the folders sit under
 * catalog/, with the brand mark the site holds for each (under
 * public/_site/images/providers/; a provider without one renders its name).
 * Brand marks keep their own colors: identification, not decoration.
 */
export const CLOUD_PROVIDERS = [
  { name: 'AWS', logo: 'aws.svg' },
  { name: 'Azure', logo: 'azure.svg' },
  { name: 'GCP', logo: 'gcp.svg' },
  { name: 'Kubernetes', logo: 'kubernetes.svg' },
  { name: 'Cloudflare', logo: 'cloudflare.svg' },
  { name: 'DigitalOcean', logo: 'digital-ocean.svg' },
  { name: 'Auth0' },
  { name: 'Stripe' },
  { name: 'OpenFGA' },
] as const;

/** Exact counts behind the printed figures, for the record and for llms.txt. */
export const PLATFORM_COUNTS = {
  catalogKinds: 852,
  providers: 9,
  controlProfiles: 851,
  controls: 17,
  controlCategories: 6,
  countedOn: '2026-11-24',
} as const;
