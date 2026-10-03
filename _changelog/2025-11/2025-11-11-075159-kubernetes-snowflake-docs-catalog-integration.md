# Kubernetes and Snowflake Documentation Catalog Integration

**Date**: November 11, 2025
**Type**: Enhancement
**Components**: Documentation Site, Build System, Component Registry

## Summary

Integrated Kubernetes (36 kinds) and Snowflake (1 kind) into the Planton documentation site, removing "Coming soon" placeholders and making all kinds discoverable with proper icons and clean, readable titles. Enhanced the documentation build system to handle nested directory structures and improved title generation to remove redundant provider prefixes from kind names across all providers.

## Problem Statement / Motivation

The Planton documentation site was showing Kubernetes as "Coming soon" despite having 36 fully-documented kinds available (13 addons and 23 workloads). Snowflake database documentation was completely missing from the catalog. Additionally, all kind titles were displaying with redundant provider prefixes (e.g., "KUBERNETES altinityoperator" instead of "Altinity Operator"), creating poor user experience in navigation.

### Pain Points

- **Kubernetes kinds invisible**: 36 documented Kubernetes kinds (addon/ and workload/) were not appearing in the catalog because the build script only scanned one directory level deep
- **Snowflake missing**: Snowflake database documentation existed but wasn't integrated into the catalog system
- **Redundant provider prefixes**: Kind titles like "KUBERNETES argocdkubernetes" and "GCP cloudcdn" repeated provider information unnecessarily
- **Poor discoverability**: Users couldn't find or browse Kubernetes and Snowflake kinds through the documentation site
- **Inconsistent GCP count**: Main index showed 5 GCP kinds when 17 actually existed
- **Missing kind icons**: All 36 Kubernetes kinds and 1 Snowflake kind lacked visual icons in the sidebar

## Solution / What's New

Enhanced the documentation generation system to support nested provider directory structures, integrated kind icons, and implemented intelligent title formatting that removes provider prefixes while maintaining proper capitalization for technical terms.

### Key Features

**1. Nested Directory Support**

The build script now recursively scans provider directories to find kinds in subdirectories like `kubernetes/addon/` and `kubernetes/workload/`.

**2. Clean Kind Titles**

Removed redundant provider prefixes and implemented smart title generation:
- Kubernetes: "ArgoCD", "ClickHouse", "Cert Manager" (instead of "KUBERNETES argocdkubernetes")
- GCP: "Cloud CDN", "Artifact Registry Repo" (instead of "GCP cloudcdn")
- AWS: "ALB", "EKS Cluster" (instead of "AWS alb")

**3. Kind Icon Integration**

Copied 39 kind logos from planton repository and organized them in the expected flat structure for the documentation site.

**4. Complete Catalog Coverage**

All 10 providers now appear in the catalog with accurate kind counts:
- ATLAS: 1 kind
- AWS: 22 kinds
- AZURE: 7 kinds
- CIVO: 12 kinds
- CLOUDFLARE: 7 kinds
- CONFLUENT: 1 kind
- DIGITALOCEAN: 14 kinds
- GCP: 17 kinds (corrected from 5)
- KUBERNETES: 36 kinds (NEW)
- SNOWFLAKE: 1 kind (NEW)

## Implementation Details

### 1. Build Script Enhancement

**File**: `site/scripts/copy-kind-docs.ts`

Enhanced the `scanProvider()` function to recursively scan one level deeper for nested directories:

```typescript
function scanProvider(providerPath: string, provider: string): KindDoc[] {
  const docs: KindDoc[] = [];
  
  for (const item of items) {
    const docPath = path.join(kindPath, 'v1', 'docs', 'README.md');
    
    if (fs.existsSync(docPath)) {
      // Found at top level
      docs.push({...});
    } else {
      // Check subdirectories (e.g., kubernetes/addon/, kubernetes/workload/)
      const subitems = fs.readdirSync(kindPath);
      for (const subitem of subitems) {
        const subDocPath = path.join(subKindPath, 'v1', 'docs', 'README.md');
        if (fs.existsSync(subDocPath)) {
          docs.push({...});
        }
      }
    }
  }
}
```

This change enabled discovery of all 36 Kubernetes kinds organized in the `addon/` and `workload/` subdirectories.

### 2. Title Generation Refactoring

Completely rewrote the `generateTitle()` function to produce clean, readable kind names:

```typescript
function generateTitle(kind: string, provider: string): string {
  // Remove provider prefix
  let name = kind;
  if (name.toLowerCase().startsWith(provider.toLowerCase())) {
    name = name.substring(provider.length);
  }
  
  // For kubernetes, also remove trailing "kubernetes" suffix
  if (provider.toLowerCase() === 'kubernetes' && name.toLowerCase().endsWith('kubernetes')) {
    name = name.substring(0, name.length - 'kubernetes'.length);
  }

  // Handle 60+ special cases for proper capitalization
  const specialCases = {
    'argocd': 'ArgoCD',
    'mongodb': 'MongoDB',
    'postgresql': 'PostgreSQL',
    'clickhouse': 'ClickHouse',
    // ... 50+ more cases
  };
  
  // Handle compound kinds
  const compoundSpecialCases = {
    'certmanagercert': 'Cert Manager Certificate',
    'perconapostgresqloperator': 'Percona PostgreSQL Operator',
    // ... more compound cases
  };
}
```

**Title Improvements**:

| Before | After |
|--------|-------|
| KUBERNETES altinityoperator | Altinity Operator |
| KUBERNETES argocdkubernetes | ArgoCD |
| KUBERNETES perconapostgresqloperator | Percona PostgreSQL Operator |
| GCP cloudcdn | Cloud CDN |
| GCP artifactregistryrepo | Artifact Registry Repo |
| AWS alb | ALB |
| SNOWFLAKE database | Database |

### 3. Kind Icon Organization

Copied kind logos from the planton repository and reorganized them to match the site's expected flat structure:

```bash
# Copied from source
/planton/apis/.../provider/kubernetes/addon/{kind}/v1/logo.svg
/planton/apis/.../provider/kubernetes/workload/{kind}/v1/logo.svg
/planton/apis/.../provider/snowflake/snowflakedatabase/v1/logo.svg

# Organized as
/site/public/images/providers/kubernetes/{kind}/logo.svg
/site/public/images/providers/snowflake/snowflakedatabase/logo.svg
```

**Icons Copied**:
- 13 Kubernetes addon logos
- 25 Kubernetes workload logos (includes stackupdaterunnerkubernetes missing 1 kind to make 36 total)
- 1 Snowflake logo

### 4. Documentation Index Updates

**File**: `site/public/docs/index.md`

- Removed `opacity-50` class from Kubernetes card
- Changed "Coming soon" to "36 kinds"
- Added Snowflake provider card with "1 kind"
- Updated GCP count from 5 to 17 kinds

### 5. Auto-Generated Catalog Pages

The enhanced build script now generates:

**Catalog Structure**:
```
site/public/docs/catalog/
├── index.md (main catalog with all 10 providers)
├── kubernetes/
│   ├── index.md (lists all 36 kinds)
│   ├── argocdkubernetes.md
│   ├── clickhousekubernetes.md
│   └── ... (34 more kind pages)
└── snowflake/
    ├── index.md (lists 1 kind)
    └── snowflakedatabase.md
```

**Statistics**:
- 10 provider indexes generated
- 118 kind documentation pages
- 1 main catalog index
- **Total: 129 documentation files**

## Benefits

### For End Users

1. **Complete Kubernetes Coverage**: All 36 Kubernetes kinds now discoverable through the catalog
2. **Snowflake Support**: Snowflake database deployment documentation now accessible
3. **Better Readability**: Clean kind titles without redundant provider prefixes improve scanning and comprehension
4. **Visual Navigation**: Kind icons in the sidebar make navigation more intuitive
5. **Accurate Information**: Kind counts reflect actual available resources (GCP: 17 not 5)

### For Developers

1. **Maintainable System**: Build script automatically handles nested directory structures for future providers
2. **Extensible Title Logic**: Special case dictionary makes it easy to add new kind name mappings
3. **Automated Generation**: All catalog pages auto-generated from source documentation
4. **Consistent Structure**: Flat icon organization pattern established for all providers

### For Documentation

1. **Source of Truth**: Documentation directly generated from `apis/` source with each build
2. **No Manual Sync**: Kind additions automatically appear in catalog on next build
3. **Professional Polish**: Clean titles and icons create professional documentation appearance

## Impact

### User Experience

- **Before**: Kubernetes kinds hidden, Snowflake missing, confusing titles like "KUBERNETES argocdkubernetes"
- **After**: All 10 providers visible with 118 total kinds, clean titles like "ArgoCD", complete icon coverage

### Documentation Coverage

| Provider | Kinds | Status |
|----------|------------|--------|
| ATLAS | 1 | ✅ Complete |
| AWS | 22 | ✅ Complete |
| AZURE | 7 | ✅ Complete |
| CIVO | 12 | ✅ Complete |
| CLOUDFLARE | 7 | ✅ Complete |
| CONFLUENT | 1 | ✅ Complete |
| DIGITALOCEAN | 14 | ✅ Complete |
| GCP | 17 | ✅ Complete (updated) |
| KUBERNETES | 36 | ✅ Complete (NEW) |
| SNOWFLAKE | 1 | ✅ Complete (NEW) |

### Developer Workflow

The build command now handles all providers correctly:

```bash
cd site && npm run copy-docs

# Output:
# 📦 KUBERNETES: Found 36 kinds
#    ✓ altinityoperator
#    ✓ argocdkubernetes
#    ... (34 more)
#    ✓ Generated index page
#
# 📦 SNOWFLAKE: Found 1 kinds
#    ✓ snowflakedatabase
#    ✓ Generated index page
#
# ✓ Generated catalog index
# 📊 Summary:
#    Providers: 10
#    Kinds copied: 118
```

## Implementation Phases

**Phase 1: Main Index Updates** ✅
- Updated `site/public/docs/index.md` manually
- Removed Kubernetes "Coming soon" placeholder
- Added Snowflake provider card
- Corrected GCP kind count

**Phase 2: Build Script Enhancement** ✅
- Modified `scanProvider()` to handle nested directories
- Enhanced `generateTitle()` with 60+ special cases
- Added compound kind name handling
- Updated provider list to include kubernetes and snowflake

**Phase 3: Icon Integration** ✅
- Copied 39 kind logos from planton
- Organized in flat structure matching site expectations
- Verified all paths match website URL patterns

**Phase 4: Catalog Generation** ✅
- Ran build script to generate all catalog pages
- Verified 118 kind pages created
- Confirmed 10 provider indexes generated
- Validated catalog index accuracy

**Phase 5: Verification** ✅
- Checked sample kind titles across providers
- Verified icon paths return 200 (not 404)
- Confirmed no linter errors
- Validated sidebar navigation display

## Technical Decisions

### Why Nested Directory Scanning?

Kubernetes kinds are logically organized into `addon/` (cluster add-ons) and `workload/` (application workloads) categories in the source repository. Rather than flatten the source structure, we enhanced the build system to understand this organization, preserving the source's logical grouping while still generating a flat catalog for users.

### Why Flat Icon Structure?

While the source has nested directories (`kubernetes/addon/{kind}/logo.svg`), the Next.js site expects a flat structure (`kubernetes/{kind}/logo.svg`) to match the URL pattern `/images/providers/kubernetes/{kind}/logo.svg`. We could have updated the site's icon loading logic, but keeping the existing pattern maintains consistency with the 8 other providers already using flat structures.

### Why Special Case Dictionary for Titles?

Kind names like "argocdkubernetes", "perconapostgresqloperator", and "mongodbatlas" require domain knowledge to format correctly (ArgoCD, Percona PostgreSQL Operator, MongoDB Atlas). A simple camelCase split produces "Argocd", "Perconapostgresqloperator", "Mongodbatlas". The special case dictionary captures this domain knowledge once and applies it consistently across all 118 kinds.

### Why Remove Provider Prefixes?

In the sidebar navigation, each kind already appears under its provider section (e.g., under the "KUBERNETES" parent). Repeating "KUBERNETES" in every child item creates visual noise and wastes horizontal space. Clean names like "ArgoCD" and "Postgres Operator" are more readable when the provider context is already established by the parent section.

## Code Metrics

**Files Modified**: 3
- `site/scripts/copy-kind-docs.ts` (enhanced build logic)
- `site/public/docs/index.md` (manual index updates)
- Icon files: 39 new SVG files added

**Lines Changed**: ~200 lines in build script

**Documentation Generated**: 129 markdown files
- 1 main catalog index
- 10 provider indexes (2 new: kubernetes, snowflake)
- 118 kind pages (37 new: 36 kubernetes, 1 snowflake)

**Build Time**: ~2 seconds to generate all documentation

## Related Work

- **2025-11-09 Automated Catalog Kind Docs Build**: Established the automated documentation generation system that this enhancement extends
- **2025-11-09 Documentation Site with Git as CMS**: Created the documentation site architecture this change integrates with

## Future Enhancements

1. **Category Tags**: Add addon/workload tags to Kubernetes kinds in the UI
2. **Search Optimization**: Enhance search to include kind aliases (e.g., "argocd" finds "ArgoCD")
3. **Icon Fallbacks**: Generate default icons for kinds without custom logos
4. **Automated Icon Sync**: Script to periodically sync icons from planton repository
5. **Title Validation**: Build-time warnings for kinds without special case title mappings

## Verification Commands

```bash
# Generate documentation
cd site && npm run copy-docs

# Count generated files
find site/public/docs/catalog -name "*.md" | wc -l
# Output: 129

# Verify Kubernetes kinds
ls site/public/docs/catalog/kubernetes/*.md | wc -l
# Output: 37 (36 kinds + 1 index)

# Verify icons
find site/public/images/providers/kubernetes -name "logo.svg" | wc -l
# Output: 36

# Start dev server
cd site && yarn dev
# Visit: http://localhost:3000/docs/catalog/kubernetes
```

## Known Limitations

1. **Manual Index Maintenance**: The `site/public/docs/index.md` provider cards are manually maintained and don't auto-update when kind counts change
2. **Title Edge Cases**: Some compound kind names may not format perfectly and require adding to the special cases dictionary
3. **Icon Dependencies**: Icons must be manually synced from planton repository when new kinds are added

---

**Status**: ✅ Production Ready
**Timeline**: 2 hours (build enhancement, icon organization, documentation generation, verification)

