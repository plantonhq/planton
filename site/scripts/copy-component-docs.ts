#!/usr/bin/env node

import * as fs from 'fs';
import * as path from 'path';

/**
 * Build script to copy deployment component documentation from apis/ to site/public/docs/catalog/
 *
 * Scans: apis/dev/planton/provider/{provider}/{component}/v1/catalog-page.md (or docs/README.md)
 * Also:  apis/dev/planton/provider/{provider}/{component}/v1/presets/*.yaml + *.md
 *
 * Outputs:
 *   site/public/docs/catalog/{provider}/{slug}/index.md          (catalog page)
 *   site/public/docs/catalog/{provider}/{slug}/presets/index.md   (preset list page)
 *   site/public/docs/catalog/{provider}/{slug}/presets/{name}.md  (preset detail - raw + page source)
 *   site/public/docs/catalog/{provider}/{slug}/presets/{name}.yaml (raw YAML)
 *
 * Title extraction: Extracts from first `# ` heading in content (hand-written, always correct).
 * Slug generation: Title -> lowercase, spaces to hyphens (e.g., "Route53 DNS Record" -> "route53-dns-record").
 * Generates frontmatter for each component and creates provider index pages.
 */

interface PresetFile {
  slug: string;          // e.g., "01-production-ha"
  rank: string;          // e.g., "01"
  title: string;         // Extracted from MD heading
  excerpt: string;       // First meaningful paragraph from MD
  yamlSourcePath: string;
  mdSourcePath: string;
}

interface ComponentDoc {
  provider: string;
  component: string;   // Original directory name (e.g., "awsroute53dnsrecord")
  slug: string;        // URL-friendly slug (e.g., "route53-dns-record")
  sourcePath: string;
  content: string;
  title: string;       // Sidebar label (e.g., "Route53 DNS Record")
  presets: PresetFile[];
}

interface Stats {
  total: number;
  copied: number;
  skipped: number;
  presetsCopied: number;
  providers: Set<string>;
}

// ---------------------------------------------------------------------------
// Provider display names (directory name -> proper uppercase display name)
// Sourced from CloudResourceProvider.display_name in cloud_resource_provider.proto
// ---------------------------------------------------------------------------

const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  'alicloud': 'ALIBABA CLOUD',
  'atlas': 'ATLAS',
  'auth0': 'AUTH0',
  'aws': 'AWS',
  'azure': 'AZURE',
  'civo': 'CIVO',
  'cloudflare': 'CLOUDFLARE',
  'confluent': 'CONFLUENT',
  'digitalocean': 'DIGITAL OCEAN',
  'gcp': 'GCP',
  'hetznercloud': 'HETZNER CLOUD',
  'kubernetes': 'KUBERNETES',
  'oci': 'OCI',
  'openfga': 'OPENFGA',
  'openstack': 'OPENSTACK',
  'scaleway': 'SCALEWAY',
  'snowflake': 'SNOWFLAKE',
};

// ---------------------------------------------------------------------------
// Provider prefix mapping (for stripping from content headings)
// ---------------------------------------------------------------------------

const PROVIDER_PREFIXES: Record<string, string[]> = {
  'alicloud': ['AliCloud ', 'Alibaba Cloud '],
  'atlas': ['MongoDB Atlas ', 'Atlas '],
  'auth0': ['Auth0 '],
  'aws': ['AWS '],
  'azure': ['Azure '],
  'civo': ['Civo '],
  'cloudflare': ['Cloudflare '],
  'confluent': ['Confluent '],
  'digitalocean': ['DigitalOcean '],
  'gcp': ['GCP '],
  'hetznercloud': ['Hetzner Cloud '],
  'kubernetes': ['Kubernetes '],
  'oci': ['OCI '],
  'openfga': ['OpenFGA '],
  'openstack': ['OpenStack '],
  'scaleway': ['Scaleway '],
  'snowflake': ['Snowflake '],
};

/**
 * Extract the title from the first `# ` heading in content.
 * Returns null if no heading is found.
 */
function extractTitleFromContent(content: string): string | null {
  const match = content.match(/^#\s+(.+)$/m);
  return match ? match[1].trim() : null;
}

/**
 * Strip the provider prefix from a title for use as a sidebar label.
 * e.g., "AWS Route53 DNS Record" -> "Route53 DNS Record"
 */
function stripProviderPrefix(title: string, provider: string): string {
  const prefixes = PROVIDER_PREFIXES[provider.toLowerCase()] || [];
  for (const prefix of prefixes) {
    if (title.startsWith(prefix)) {
      return title.substring(prefix.length);
    }
  }
  return title;
}

/**
 * Generate a URL-friendly slug from a title.
 * e.g., "Route53 DNS Record" -> "route53-dns-record"
 * e.g., "ALB" -> "alb"
 * e.g., "GKE Cluster" -> "gke-cluster"
 */
function generateSlug(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
    .replace(/\s+/g, '-')          // Spaces to hyphens
    .replace(/-+/g, '-')           // Collapse multiple hyphens
    .replace(/^-|-$/g, '');        // Trim leading/trailing hyphens
}

/**
 * Generate human-readable title from component name (fallback for legacy docs without headings).
 * Examples:
 *   awsalb -> ALB
 *   gcpgkecluster -> GKE Cluster
 *   argocdkubernetes -> ArgoCD
 *   clickhousekubernetes -> ClickHouse
 */
function generateTitle(component: string, provider: string): string {
  // Remove provider prefix if component starts with it
  let name = component;
  if (name.toLowerCase().startsWith(provider.toLowerCase())) {
    name = name.substring(provider.length);
  }

  // For kubernetes components, also remove the "kubernetes" suffix
  if (provider.toLowerCase() === 'kubernetes' && name.toLowerCase().endsWith('kubernetes')) {
    name = name.substring(0, name.length - 'kubernetes'.length);
  }

  // Handle special cases for proper capitalization
  const specialCases: Record<string, string> = {
    'argocd': 'ArgoCD',
    'mongodb': 'MongoDB',
    'postgresql': 'PostgreSQL',
    'mysql': 'MySQL',
    'clickhouse': 'ClickHouse',
    'elasticsearch': 'Elasticsearch',
    'opensearch': 'OpenSearch',
    'kafka': 'Kafka',
    'redis': 'Redis',
    'postgres': 'Postgres',
    'gitlab': 'GitLab',
    'jenkins': 'Jenkins',
    'grafana': 'Grafana',
    'prometheus': 'Prometheus',
    'istio': 'Istio',
    'nginx': 'Nginx',
    'harbor': 'Harbor',
    'keycloak': 'Keycloak',
    'solr': 'Solr',
    'neo4j': 'Neo4j',
    'nats': 'NATS',
    'openfga': 'OpenFGA',
    'signoz': 'SigNoz',
    'locust': 'Locust',
    'temporal': 'Temporal',
    'percona': 'Percona',
    'altinity': 'Altinity',
    'certmanager': 'Cert Manager',
    'externaldns': 'External DNS',
    'externalsecrets': 'External Secrets',
    'ingressnginx': 'Ingress Nginx',
    'helmrelease': 'Helm Release',
    'cronjob': 'CronJob',
    'microservice': 'Microservice',
    'httpendpoint': 'HTTP Endpoint',
    'jobrunnner': 'Job Runner',
    'operator': 'Operator',
    'cluster': 'Cluster',
    'nodepool': 'Node Pool',
    'workloadidentitybinding': 'Workload Identity Binding',
    'artifactregistryrepo': 'Artifact Registry Repo',
    'artifactregistry': 'Artifact Registry',
    'containerregistry': 'Container Registry',
    'secretsmanager': 'Secrets Manager',
    'serviceaccount': 'Service Account',
    'subnetwork': 'Subnetwork',
    'routernat': 'Router NAT',
    'natgateway': 'NAT Gateway',
    'cloudfront': 'CloudFront',
    'cloudcdn': 'Cloud CDN',
    'cloudfunction': 'Cloud Function',
    'cloudrun': 'Cloud Run',
    'cloudsql': 'Cloud SQL',
    'gcsbucket': 'GCS Bucket',
    'ecrrepo': 'ECR Repo',
    'clientvpn': 'Client VPN',
    'dnszone': 'DNS Zone',
    'keyvault': 'Key Vault',
    'kmskey': 'KMS Key',
    'securitygroup': 'Security Group',
    'loadbalancer': 'Load Balancer',
    'appplatformservice': 'App Platform Service',
    'databasecluster': 'Database Cluster',
    'computeinstance': 'Compute Instance',
    'ipaddress': 'IP Address',
    'kvnamespace': 'KV Namespace',
    'd1database': 'D1 Database',
    'r2bucket': 'R2 Bucket',
    'zerotrustaccessapplication': 'Zero Trust Access Application',
    'database': 'Database',
    'repo': 'Repo',
    'cert': 'Certificate',
    'managercert': 'Manager Certificate',
    'certificate': 'Certificate',
    'bucket': 'Bucket',
    'volume': 'Volume',
    'firewall': 'Firewall',
    'function': 'Function',
    'worker': 'Worker',
    'droplet': 'Droplet',
    'lambda': 'Lambda',
    'instance': 'Instance',
    'role': 'Role',
    'user': 'User',
    'key': 'Key',
    'zone': 'Zone',
    'keypair': 'Keypair',
    'subnet': 'Subnet',
    'router': 'Router',
    'image': 'Image',
    'project': 'Project',
  };

  // Check if the entire name matches a special case
  const lowerName = name.toLowerCase();
  if (specialCases[lowerName]) {
    return specialCases[lowerName];
  }

  // Handle specific compound components that need special handling
  const compoundSpecialCases: Record<string, string> = {
    'certmanagercert': 'Cert Manager Certificate',
    'perconapostgresqloperator': 'Percona PostgreSQL Operator',
    'perconaservermongodboperator': 'Percona MongoDB Operator',
    'perconaservermysqloperator': 'Percona MySQL Operator',
    'postgresoperator': 'Postgres Operator',
    'solroperator': 'Solr Operator',
    'elasticoperator': 'Elastic Operator',
    'kafkaoperator': 'Kafka Operator',
    'altinityoperator': 'Altinity Operator',
    // OpenStack compound components
    'routerinterface': 'Router Interface',
    'securitygrouprule': 'Security Group Rule',
    'floatingip': 'Floating IP',
    'floatingipassociate': 'Floating IP Associate',
    'networkport': 'Network Port',
    'servergroup': 'Server Group',
    'volumeattach': 'Volume Attach',
    'applicationcredential': 'Application Credential',
    'roleassignment': 'Role Assignment',
    'loadbalancerlistener': 'Load Balancer Listener',
    'loadbalancerpool': 'Load Balancer Pool',
    'loadbalancermember': 'Load Balancer Member',
    'loadbalancermonitor': 'Load Balancer Monitor',
    'dnsrecord': 'DNS Record',
    'containerclustertemplate': 'Container Cluster Template',
    'containercluster': 'Container Cluster',
  };

  if (compoundSpecialCases[lowerName]) {
    return compoundSpecialCases[lowerName];
  }

  // Check for compound names (e.g., "perconapostgresqloperator")
  for (const [key, value] of Object.entries(specialCases)) {
    if (lowerName.includes(key)) {
      name = name.replace(new RegExp(key, 'gi'), value);
    }
  }

  // Handle common acronyms
  const acronyms = ['ALB', 'EKS', 'GKE', 'VPC', 'DNS', 'IAM', 'ACM', 'S3', 'EC2', 'ECS', 'RDS', 'CDN', 'HTTP', 'HTTPS', 'API', 'SDK', 'CLI', 'NAT', 'IP', 'SSL', 'TLS', 'WAF', 'KV', 'D1', 'R2', 'GCS'];

  // Insert spaces before uppercase letters
  let spaced = name.replace(/([A-Z])/g, ' $1').trim();

  // Uppercase known acronyms
  acronyms.forEach(acronym => {
    const regex = new RegExp(`\\b${acronym}\\b`, 'gi');
    spaced = spaced.replace(regex, acronym);
  });

  // Capitalize first letter if not already capitalized
  if (spaced.length > 0 && spaced[0] === spaced[0].toLowerCase()) {
    spaced = spaced.charAt(0).toUpperCase() + spaced.slice(1);
  }

  return spaced;
}

/**
 * Resolve a title for a component doc.
 * Priority: extract from content heading (stripped of provider prefix) > generateTitle() fallback.
 */
function resolveTitle(content: string, component: string, provider: string): string {
  const heading = extractTitleFromContent(content);
  if (heading) {
    return stripProviderPrefix(heading, provider);
  }
  return generateTitle(component, provider);
}

/**
 * Scan a component's v1/presets/ directory for preset YAML/MD pairs.
 * Returns an array of PresetFile objects sorted by rank.
 */
function scanPresets(componentPath: string): PresetFile[] {
  const presetsDir = path.join(componentPath, 'v1', 'presets');
  if (!fs.existsSync(presetsDir)) {
    return [];
  }

  const files = fs.readdirSync(presetsDir);
  const yamlFiles = files.filter(f => f.endsWith('.yaml'));
  const presets: PresetFile[] = [];

  for (const yamlFile of yamlFiles) {
    const baseName = yamlFile.replace(/\.yaml$/, '');
    const mdFile = `${baseName}.md`;

    if (!files.includes(mdFile)) {
      continue; // Skip YAML files without a matching MD
    }

    // Extract rank from filename (e.g., "01" from "01-production-ha")
    const rankMatch = baseName.match(/^(\d+)-/);
    const rank = rankMatch ? rankMatch[1] : '99';

    // Read the MD file to extract title and excerpt
    const mdPath = path.join(presetsDir, mdFile);
    const mdContent = fs.readFileSync(mdPath, 'utf-8');

    const title = extractTitleFromContent(mdContent) || formatPresetSlug(baseName);
    const excerpt = extractExcerptFromPresetMd(mdContent);

    presets.push({
      slug: baseName,
      rank,
      title,
      excerpt,
      yamlSourcePath: path.join(presetsDir, yamlFile),
      mdSourcePath: mdPath,
    });
  }

  // Sort by rank (string comparison works for zero-padded numbers)
  return presets.sort((a, b) => a.rank.localeCompare(b.rank));
}

/**
 * Format a preset slug into a readable title (fallback when MD has no heading).
 * e.g., "01-production-ha" -> "Production HA"
 */
function formatPresetSlug(slug: string): string {
  return slug
    .replace(/^\d+-/, '')           // Remove rank prefix
    .replace(/-/g, ' ')            // Hyphens to spaces
    .replace(/\b\w/g, l => l.toUpperCase()); // Title case
}

/**
 * Extract the first meaningful paragraph from a preset MD file for use as an excerpt.
 * Skips the heading and "When to Use" section, grabs the first paragraph.
 */
function extractExcerptFromPresetMd(content: string): string {
  const lines = content.split('\n');
  let foundFirstParagraph = false;
  const paragraphLines: string[] = [];

  for (const line of lines) {
    // Skip headings
    if (line.match(/^#{1,6}\s/)) {
      if (foundFirstParagraph) break; // Stop at the next heading
      continue;
    }
    // Skip empty lines before first paragraph
    if (!foundFirstParagraph && line.trim() === '') continue;

    if (line.trim() !== '') {
      foundFirstParagraph = true;
      paragraphLines.push(line.trim());
    } else if (foundFirstParagraph) {
      break; // End of first paragraph
    }
  }

  const excerpt = paragraphLines.join(' ');
  if (excerpt.length > 200) {
    const truncated = excerpt.substring(0, 200);
    const lastSpace = truncated.lastIndexOf(' ');
    return (lastSpace > 150 ? truncated.substring(0, lastSpace) : truncated) + '...';
  }
  return excerpt;
}

/**
 * Escape a string for safe embedding in YAML double-quoted values.
 * Replaces literal double quotes and backslashes.
 */
function yamlEscape(value: string): string {
  return value.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
}

/**
 * Generate frontmatter for a component doc
 */
function generateFrontmatter(title: string, component: string, description?: string): string {
  const desc = description || `${title} deployment documentation`;
  return `---
title: "${yamlEscape(title)}"
description: "${yamlEscape(desc)}"
icon: "package"
order: 100
componentName: "${component}"
---`;
}

/**
 * Scan a provider directory for components with docs
 * Handles both flat structures (e.g., aws/awsalb/) and any potential nested subdirectories
 */
function scanProvider(providerPath: string, provider: string): ComponentDoc[] {
  const docs: ComponentDoc[] = [];

  if (!fs.existsSync(providerPath)) {
    return docs;
  }

  const items = fs.readdirSync(providerPath);

  for (const item of items) {
    const componentPath = path.join(providerPath, item);
    const stat = fs.statSync(componentPath);

    if (!stat.isDirectory()) {
      continue;
    }

    // Prefer catalog-page.md (hand-written), fall back to docs/README.md (legacy)
    const catalogPath = path.join(componentPath, 'v1', 'catalog-page.md');
    const legacyPath = path.join(componentPath, 'v1', 'docs', 'README.md');
    const docPath = fs.existsSync(catalogPath) ? catalogPath : legacyPath;

    if (fs.existsSync(docPath)) {
      const content = fs.readFileSync(docPath, 'utf-8');
      const title = resolveTitle(content, item, provider);
      const slug = generateSlug(title);
      const presets = scanPresets(componentPath);

      docs.push({
        provider,
        component: item,
        slug,
        sourcePath: docPath,
        content,
        title,
        presets,
      });
    } else {
      // If no docs at this level, check subdirectories
      // Only scan one level deeper to avoid infinite recursion
      const subitems = fs.readdirSync(componentPath);
      for (const subitem of subitems) {
        const subComponentPath = path.join(componentPath, subitem);
        const subStat = fs.statSync(subComponentPath);

        if (!subStat.isDirectory()) {
          continue;
        }

        const subCatalogPath = path.join(subComponentPath, 'v1', 'catalog-page.md');
        const subLegacyPath = path.join(subComponentPath, 'v1', 'docs', 'README.md');
        const subDocPath = fs.existsSync(subCatalogPath) ? subCatalogPath : subLegacyPath;

        if (fs.existsSync(subDocPath)) {
          const content = fs.readFileSync(subDocPath, 'utf-8');
          const title = resolveTitle(content, subitem, provider);
          const slug = generateSlug(title);
          const presets = scanPresets(subComponentPath);

          docs.push({
            provider,
            component: subitem,
            slug,
            sourcePath: subDocPath,
            content,
            title,
            presets,
          });
        }
      }
    }
  }

  return docs;
}

/**
 * Rewrite internal catalog links so that component directory names are replaced
 * with URL-friendly slugs. Source catalog-page.md files use stable directory names
 * (e.g., /docs/catalog/aws/awsvpc) but the built site uses title-derived slugs
 * (e.g., /docs/catalog/aws/vpc). This function translates one to the other.
 */
function rewriteCatalogLinks(content: string, lookup: Map<string, string>): string {
  return content.replace(
    /\/docs\/catalog\/([a-z0-9-]+)\/([a-z0-9-]+)/g,
    (match, provider, component) => {
      const key = `${provider}/${component}`;
      const slugPath = lookup.get(key);
      return slugPath ? `/docs/catalog/${slugPath}` : match;
    }
  );
}

/**
 * Write component doc to site/public/docs/catalog/{provider}/{slug}/index.md
 * and copy presets if present.
 */
function writeComponentDoc(
  doc: ComponentDoc,
  outputRoot: string
): void {
  // Always use directory-based layout: {provider}/{slug}/index.md
  const componentDir = path.join(outputRoot, doc.provider, doc.slug);
  fs.mkdirSync(componentDir, { recursive: true });

  // Generate output with frontmatter
  const frontmatter = generateFrontmatter(doc.title, doc.component);
  const output = `${frontmatter}\n\n${doc.content}`;

  // Write catalog page as index.md inside the component directory
  const indexPath = path.join(componentDir, 'index.md');
  fs.writeFileSync(indexPath, output, 'utf-8');

  // Write presets if present
  if (doc.presets.length > 0) {
    writePresets(doc, componentDir);
  }
}

/**
 * Write preset files for a component:
 *   {componentDir}/presets/index.md           -- preset list page
 *   {componentDir}/presets/{name}.md          -- preset detail page (also served as raw MD)
 *   {componentDir}/presets/{name}.yaml        -- raw YAML file
 */
function writePresets(doc: ComponentDoc, componentDir: string): void {
  const presetsDir = path.join(componentDir, 'presets');
  fs.mkdirSync(presetsDir, { recursive: true });

  // Copy each preset's YAML and MD files
  for (const preset of doc.presets) {
    // Copy YAML as-is (served as raw file)
    const yamlDest = path.join(presetsDir, `${preset.slug}.yaml`);
    fs.copyFileSync(preset.yamlSourcePath, yamlDest);

    // Read original MD and prepend frontmatter for page rendering
    const originalMd = fs.readFileSync(preset.mdSourcePath, 'utf-8');
    const presetFrontmatter = `---
title: "${yamlEscape(preset.title)}"
description: "${yamlEscape(preset.excerpt)}"
type: "preset"
rank: "${preset.rank}"
presetSlug: "${preset.slug}"
componentSlug: "${doc.slug}"
componentTitle: "${yamlEscape(doc.title)}"
provider: "${doc.provider}"
icon: "package"
order: ${parseInt(preset.rank, 10)}
---`;

    const presetMdOutput = `${presetFrontmatter}\n\n${originalMd}`;
    const mdDest = path.join(presetsDir, `${preset.slug}.md`);
    fs.writeFileSync(mdDest, presetMdOutput, 'utf-8');
  }

  // Generate presets index page
  const presetsListYaml = doc.presets
    .map(p => `  - slug: "${p.slug}"
    rank: "${p.rank}"
    title: "${yamlEscape(p.title)}"
    excerpt: "${yamlEscape(p.excerpt)}"`)
    .join('\n');

  const presetsIndexContent = `---
title: "Presets"
description: "Ready-to-deploy configuration presets for ${doc.title}"
type: "preset-list"
componentSlug: "${doc.slug}"
componentTitle: "${yamlEscape(doc.title)}"
provider: "${doc.provider}"
icon: "package"
order: 200
presets:
${presetsListYaml}
---

# ${doc.title} Presets

Ready-to-deploy configuration presets for ${doc.title}. Each preset is a complete manifest you can copy, customize, and deploy.
`;

  const presetsIndexPath = path.join(presetsDir, 'index.md');
  fs.writeFileSync(presetsIndexPath, presetsIndexContent, 'utf-8');
}

/**
 * Generate provider index page listing all components
 */
function generateProviderIndex(
  provider: string,
  docs: ComponentDoc[],
  outputRoot: string
): void {
  const providerDir = path.join(outputRoot, provider);

  if (!fs.existsSync(providerDir)) {
    fs.mkdirSync(providerDir, { recursive: true });
  }

  const providerTitle = PROVIDER_DISPLAY_NAMES[provider] || provider.toUpperCase();

  // Sort docs alphabetically by title
  const sortedDocs = [...docs].sort((a, b) =>
    a.title.localeCompare(b.title)
  );

  // Generate component list using slug-based URLs
  const componentList = sortedDocs
    .map(doc => `- [${doc.title}](/docs/catalog/${provider}/${doc.slug})`)
    .join('\n');

  const indexContent = `---
title: "${providerTitle}"
description: "Deploy ${providerTitle} resources using Planton"
icon: "cloud"
order: 10
---

# ${providerTitle}

The following ${providerTitle} resources can be deployed using Planton:

${componentList}
`;

  const indexPath = path.join(providerDir, 'index.md');
  fs.writeFileSync(indexPath, indexContent, 'utf-8');
}

/**
 * Get provider icon path
 */
function getProviderIcon(provider: string): string {
  const iconMap: Record<string, string> = {
    'aws': '/images/providers/aws.svg',
    'gcp': '/images/providers/gcp.svg',
    'azure': '/images/providers/azure.svg',
    'cloudflare': '/images/providers/cloudflare.svg',
    'civo': '/images/providers/civo.svg',
    'digitalocean': '/images/providers/digital-ocean.svg',
    'atlas': '/images/providers/mongodb-atlas.svg',
    'confluent': '/images/providers/confluent.svg',
    'kubernetes': '/images/providers/kubernetes.svg',
    'snowflake': '/images/providers/snowflake.svg',
    'openstack': '/images/providers/openstack.svg',
  };
  return iconMap[provider] || '/images/providers/default.svg';
}

/**
 * Get component count for a provider
 */
function getProviderComponentCount(provider: string, allDocs: Map<string, ComponentDoc[]>): number {
  return allDocs.get(provider)?.length || 0;
}

/**
 * Generate main provider index page
 */
function generateMainIndex(_providers: string[], outputRoot: string, _allDocs: Map<string, ComponentDoc[]>): void {
  // The provider grid is now rendered by the CatalogProviderGrid React
  // component at runtime (data-driven from docs-structure.json).  This
  // function only writes the frontmatter and header text so the markdown
  // file exists for the static-generation pipeline.
  const indexContent = `---
title: "Catalog"
description: "Browse deployment components organized by cloud provider"
icon: "package"
order: 50
---

# Catalog

Browse deployment components by cloud provider:
`;

  const indexPath = path.join(outputRoot, 'index.md');
  fs.writeFileSync(indexPath, indexContent, 'utf-8');
}

/**
 * Main function to copy all component docs
 */
async function copyComponentDocs(): Promise<void> {
  console.log('Starting component documentation copy process...\n');

  // Paths
  const scriptDir = __dirname;
  const projectRoot = path.join(scriptDir, '../..');
  const apisRoot = path.join(projectRoot, 'apis/dev/planton/provider');
  const siteDocsRoot = path.join(scriptDir, '../public/docs/catalog');

  // Dynamically discover provider directories to clear (prevents stale files from previous builds)
  const providerDirs = fs.existsSync(siteDocsRoot)
    ? fs.readdirSync(siteDocsRoot).filter(item =>
        fs.statSync(path.join(siteDocsRoot, item)).isDirectory()
      )
    : [];

  // Clear only provider directories (preserve manually created docs like index.md, getting-started.md, etc.)
  for (const provider of providerDirs) {
    const providerPath = path.join(siteDocsRoot, provider);
    if (fs.existsSync(providerPath)) {
      console.log(`  Clearing ${provider} docs`);
      fs.rmSync(providerPath, { recursive: true });
    }
  }

  // Ensure output directory exists
  fs.mkdirSync(siteDocsRoot, { recursive: true });

  // Stats
  const stats: Stats = {
    total: 0,
    copied: 0,
    skipped: 0,
    presetsCopied: 0,
    providers: new Set(),
  };

  // Track docs by provider for index generation
  const docsByProvider: Map<string, ComponentDoc[]> = new Map();

  // Scan all providers
  if (!fs.existsSync(apisRoot)) {
    console.error(`Error: APIs directory not found at ${apisRoot}`);
    process.exit(1);
  }

  const providers = fs.readdirSync(apisRoot).filter(item => {
    const itemPath = path.join(apisRoot, item);
    return fs.statSync(itemPath).isDirectory();
  });

  console.log(`Scanning ${providers.length} providers...\n`);

  // Phase 1: Scan all providers to collect docs and build the component-to-slug lookup.
  // Writing is deferred to Phase 2 so that internal catalog links can be rewritten
  // using the complete lookup (a component in AWS may link to a component in GCP).
  for (const provider of providers) {
    const providerPath = path.join(apisRoot, provider);
    const docs = scanProvider(providerPath, provider);

    if (docs.length > 0) {
      stats.providers.add(provider);
      docsByProvider.set(provider, docs);
      console.log(`${PROVIDER_DISPLAY_NAMES[provider] || provider.toUpperCase()}: Found ${docs.length} components`);
    }
  }

  // Build a global lookup from component directory path to URL slug path.
  // e.g. "aws/awsvpc" -> "aws/vpc", "openstack/openstackinstance" -> "openstack/instance"
  const componentToSlug = new Map<string, string>();
  for (const [provider, docs] of docsByProvider) {
    for (const doc of docs) {
      componentToSlug.set(`${provider}/${doc.component}`, `${provider}/${doc.slug}`);
    }
  }

  console.log(`\nBuilt link lookup for ${componentToSlug.size} components\n`);

  // Phase 2: Write docs with rewritten catalog links, then generate index pages.
  for (const [provider, docs] of docsByProvider) {
    for (const doc of docs) {
      try {
        doc.content = rewriteCatalogLinks(doc.content, componentToSlug);
        writeComponentDoc(doc, siteDocsRoot);
        stats.copied++;
        stats.presetsCopied += doc.presets.length;
        const sourceType = doc.sourcePath.endsWith('catalog-page.md') ? 'catalog-page' : 'legacy';
        const presetInfo = doc.presets.length > 0 ? ` + ${doc.presets.length} presets` : '';
        console.log(`   ${doc.component} -> ${doc.slug}/index.md (${sourceType}${presetInfo})`);
      } catch (error) {
        console.error(`   FAIL ${doc.component}: ${error}`);
        stats.skipped++;
      }
    }

    generateProviderIndex(provider, docs, siteDocsRoot);
    console.log(`   Generated index page\n`);
  }

  // Generate catalog index (now in /docs/catalog/)
  if (stats.providers.size > 0) {
    generateMainIndex(Array.from(stats.providers), siteDocsRoot, docsByProvider);
    console.log(`Generated catalog index\n`);
  }

  // Summary
  console.log('Summary:');
  console.log(`   Providers: ${stats.providers.size}`);
  console.log(`   Components copied: ${stats.copied}`);
  console.log(`   Presets copied: ${stats.presetsCopied}`);
  console.log(`   Components skipped: ${stats.skipped}`);
  console.log(`   Output: ${path.relative(projectRoot, siteDocsRoot)}`);
  console.log('\nComponent documentation copy complete!\n');
}

// Run the script
copyComponentDocs().catch(error => {
  console.error('Error copying component docs:', error);
  process.exit(1);
});
