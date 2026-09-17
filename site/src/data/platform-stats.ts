/**
 * The numbers the marketing pages state, counted from the open-source
 * repository and carried nowhere else in prose. A page that says "700+"
 * reads it from here; the story records how each number was counted so the
 * next reconciliation repeats the method instead of guessing.
 *
 * Counted 2026-09-17 against plantonhq/planton:
 *   component kinds   = folders under catalog/<provider>/ carrying a spec.proto (719)
 *   providers         = the provider folders under catalog/ (aws, azure, gcp, kubernetes, cloudflare, digitalocean, auth0, openfga)
 *   Infra Charts      = Chart.yaml files under charts/ (18; the earlier "50+" was wrong)
 *   control profiles  = kinds carrying a controls.yaml (718)
 *   controls          = the control catalog's entries (17, in 6 categories)
 *   crosswalks        = framework files (HIPAA Security Rule, SOC 2 TSC, FedRAMP Moderate, CIS AWS Foundations)
 *
 * Rounded figures ("700+") are what a page prints; the exact counts are the
 * record. A change to the catalog changes this file first.
 */
export const PLATFORM_STATS = {
  /** Printed. Component kinds in the catalog, rounded down to the hundred. */
  DEPLOYMENT_MODULE_COUNT: '700+',
  CLOUD_PROVIDER_COUNT: '8',
  /** Printed exactly; the number is small enough to be honest about. */
  INFRA_CHART_COUNT: '18',
  CONTROL_COUNT: '17',
  CONTROL_CATEGORY_COUNT: '6',
  FRAMEWORK_CROSSWALK_COUNT: '4',
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
  { name: 'OpenFGA' },
] as const;

/** Exact counts behind the printed figures, for the record and for llms.txt. */
export const PLATFORM_COUNTS = {
  componentKinds: 719,
  providers: 8,
  infraCharts: 18,
  controlProfiles: 718,
  controls: 17,
  controlCategories: 6,
  frameworkCrosswalks: 4,
  countedOn: '2026-09-17',
} as const;
