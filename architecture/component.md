# Component: Definition and Ideal State

## What is a Component?

A **component** in Planton is a self-contained, production-ready package that enables declarative deployment of a specific infrastructure resource or application workload to a cloud provider or Kubernetes cluster.

### Technical Definition

A component consists of:

1. **API Definition (Protobuf)** - A strongly-typed, language-neutral schema that defines:
   - The configuration interface (`spec.proto`)
   - The deployment inputs (`input.proto`)
   - The deployment outputs (`outputs.proto`)
   - Field-level validation rules

2. **Infrastructure-as-Code Modules** - Executable deployment logic in both:
   - Pulumi (Go-based, using real programming language)
   - Terraform/OpenTofu (HCL-based, declarative)

3. **Documentation** - Multi-layered documentation serving different audiences:
   - User-facing documentation (`README.md` and `catalog.md`, the catalog page)
   - Authored operational judgment (`GUIDE.md`, where the component carries non-obvious judgment)
   - Ready-to-deploy presets and a validated example manifest (`e2e/manifest.yaml`)

4. **Verified Fact-Sheets** - Machine-checked data every surface answers from:
   - Cost facts (`cost.yaml`: billing model, dimensions, price provenance) with generated per-preset estimates priced from the pinned price books
   - Technical control posture with evidence (`controls.yaml`, citing the central control catalog and framework crosswalks in `catalog/_compliance/`)
   - Least-privilege provisioning permissions (`iac/permissions.yaml`), validated against the providers' own published action/scope inventories

### Role in Planton

Components are the **atomic units of deployment** in Planton. They serve as:

- **The Menu Items** - In the restaurant analogy from the main README, components are the individual dishes available for order
- **Reusable Building Blocks** - Platform engineers compose multiple components to build complete application stacks
- **Provider-Specific Implementations** - Each component targets a specific provider (AWS, GCP, Azure, Kubernetes, etc.) with provider-specific configuration
- **The Bridge** - Between high-level declarative manifests and low-level cloud provider APIs

### Relationship to Kubernetes Resource Model (KRM)

Planton adopts the Kubernetes Resource Model philosophy but extends it beyond Kubernetes:

**Structural Consistency:**
```yaml
apiVersion: <provider>.planton.dev/<version>
kind: <ComponentType>
metadata:
  name: <resource-name>
  org: <organization>
  env: <environment>
spec:
  # Provider-specific configuration
status:
  # System-managed outputs (read-only)
```

**Key Differences from Kubernetes:**
- **Protocol Buffers vs Go Structs** - Planton uses protobuf for language neutrality and multi-language SDK generation
- **Provider-Specific vs Abstracted** - Each cloud provider has its own components (no artificial abstraction layer)
- **Dual IaC Support** - Both Pulumi and Terraform implementations (Kubernetes only uses Go-based controllers)
- **Documentation-First** - Research-driven design grounded in the provider's authoritative API

### Examples of Components

**Cloud Provider Resources:**
- `AwsRdsInstance` - PostgreSQL/MySQL on AWS RDS
- `GcpCloudSql` - PostgreSQL/MySQL on Google Cloud SQL
- `AzureAksCluster` - Managed Kubernetes on Azure
- `GcpCertManagerCert` - SSL/TLS certificates on GCP

**Kubernetes Workloads:**
- `KubernetesPostgres` - PostgreSQL deployed to any Kubernetes cluster
- `KubernetesKafka` - Apache Kafka deployed to any Kubernetes cluster
- `KubernetesDeployment` - Containerized workload deployment

**Identity and Authorization Platforms:**
- `Auth0Client` - Auth0 application client
- `OpenFgaStore` - OpenFGA authorization store

---

## What Does "Complete" Mean?

Completeness of a component is **contextual and principle-driven**, not a simple checklist of every possible feature.

### Philosophy of Completeness

**The Provider Parity Standard:**

A complete component models **100% of the configurable argument surface** of the provider resources it maps to, at the pinned provider version. Partial coverage is not a design choice available to any component; where something is not covered, that is an explicitly recorded decision with a reason, never a silent omission. This means:

- **Schema as the Parity Target** - The provider's canonical API at the pinned version defines the parity target: read it fully, model it fully. For cloud providers with an official Terraform provider, that cloned provider source at the pinned version is the reference. For Kubernetes, there is no single canonical provider catalog: the reference is the pinned upstream artifact for that component (Helm chart values/templates, CRD schemas, or core Kubernetes API types), read locally — never designed from web memory. Kubernetes admitted components model complete typed coverage of that upstream surface (with typed escape hatches only as safety valves on top of a fully modeled spec), never a deliberately trimmed subset. References are read for facts (field sets, defaults, constraints, ForceNew behavior) — never as a source of code text: the Terraform provider repos are MPL-2.0 and HashiCorp's BUSL repos are not open source, so catalog modules are always original Apache-2.0 authorship informed by those facts (consuming the Apache-2.0 Pulumi SDKs as dependencies is, of course, normal usage).
- **Unconditional Depth** - Every configurable argument of the mapped provider resources remains representable in the spec. Restructuring is encouraged — idiomatic proto shapes, renames that read better — as long as nothing the provider can configure becomes inexpressible. A kind that cannot express a provider argument is incomplete, regardless of how rarely that argument is used.
- **Research-Driven** - Coverage is grounded in the provider schema at the pinned version, not guesswork or web memory.
- **Opinionated Defaults** - Provide sensible defaults so full coverage never costs usability: advanced fields default well and stay out of the way until a user needs them.

**Example:** For `GcpCertManagerCert`, the most-reached fields are:
- `gcp_project_id` - Where to deploy
- `primary_domain_name` - What domain to secure
- `cloud_dns_zone_id` - Where to create validation records
- `certificate_type` - MANAGED vs LOAD_BALANCER

Advanced fields like certificate scope, location, and labels are equally modeled, with sensible defaults so the full provider surface is reachable without burdening the common case.

### Contextual vs Absolute Completeness

**Contextual Completeness** means a component is complete when:

1. **Coverage Is Accounted For** - Every coverage decision is recorded where the parity tooling reads it: compositions (provider resources deliberately folded into this kind) and exclusions (deprecated or superseded surface only), each with its reason, along with the pinned provider version parity is declared against. The durable rationale -- why the kind is shaped the way it is -- lives in the component's `GUIDE.md`; research working notes are never committed

2. **Proto Schema Achieves Parity** - The `spec.proto` models the full configurable argument surface of the mapped provider resources at the pinned provider version -- no deliberate field exclusions, however rarely a field is used

3. **Both IaC Modules Implement the Schema** - Every field defined in `spec.proto` is actually used in both Pulumi and Terraform modules (no unused fields, no missing implementations)

4. **Presets Validate the API** - The preset YAML files contain working, deployable configurations that demonstrate the API's capabilities and validate against the current schema

5. **Documentation Explains Decisions** - Users understand the coverage accounting and the rationale for every composition and recorded exclusion, reducing support burden

**The recorded-exclusion rule** (omission is a decision, and every decision is recorded):

- Deprecated or superseded provider surface may be excluded -- always *with a recorded reason*, never silently; these are the only permitted exclusion categories
- Beta-only provider capability enters the GA parity baseline only through the admission list -- `pkg/providerparity/admissions/<channel>.yaml`, one entry per resource per consuming kind, each with a recorded reason and the path where its promotion to GA is tracked -- implemented with the idiomatic per-resource beta-provider declaration inside the module (`provider = google-beta` on exactly the admitted resource blocks, the channel pinned to the baseline's line). The list is a gate, not a note: the module census records every such attachment, the accounting fails an attachment with no entry and an entry the GA provider has since made unnecessary, and `hack/guards/ensure_beta_admissions.sh` holds the two in step on every module edit. Wholesale beta parity is rejected
- Supporting every possible deployment method remains out of scope; a component models its mapped provider resources, not every way to deploy the technology
- Field count is never the goal for its own sake -- every covered field must be tested, parity-verified, and deploy-validated

### Quality At Scale

A component reaches parity by covering the provider's full configurable surface **with** the same rigor -- every field researched, documented, validated, and exercised in both engines. Breadth never excuses a hastily-added, undocumented, or untested field: quality is the constant, and coverage is raised to the parity target on top of it. Parity without that rigor is not parity; it is untested surface area.

Breadth is never a reason to lower the bar on any single unit of the catalog -- not across providers (the number of clouds supported never justifies a weaker experience for one cloud), and not within a provider (the number of components in a cloud's catalog never justifies a shallow experience for one component). Build and evolve every component as if the entire platform existed solely to give the best possible self-service experience for that one resource. Design and implementation decisions are never driven by implementation complexity, and never by how another component or another provider happened to solve a similar problem; the only test that decides a design is the experience a user of *this* provider and *this* component will come to expect.

Each provider's surface is designed **from that provider's own authoritative API and stands on its own merit.** Because every provider is brought to this bar concurrently, another provider's component is never a design reference: mirror the structural *shape* across providers, but derive what a component models -- its resources, fields, decomposition, and depth -- only from the target provider's own API. Users adopt Planton for a world-class experience of the specific cloud they care about, not for multi-cloud breadth, so a design averaged toward a cross-provider lowest common denominator fails the bar.

**A module never mutates a resource it merely references.** When a spec field references another component (an IAM role a cluster assumes, a subnet a fleet launches into), the reference is the entire relationship: the module must not reach into the referenced resource to attach policies, add rules, or otherwise change its configuration. The referenced resource owns its own configuration -- requirements it must satisfy (e.g. a managed policy the consumer's cloud API demands on a role) are expressed on that resource's own spec, stated in the consumer's field comment and docs, and left to fail loudly at the provider when unmet. Side-effect mutation hides the real dependency from the resource graph, silently rewrites nodes the user owns (including ones created outside Planton), and inevitably diverges between engines.

**Completeness Indicators:**
- ✅ Coverage accounting records every composition and exclusion with its reason; `GUIDE.md` carries the durable rationale where the component warrants one
- ✅ Proto schema is validated with real-world constraints
- ✅ Both IaC modules have feature parity
- ✅ Both IaC modules are richly commented to the authoring bar (explain why/trade-offs, not narration)
- ✅ Examples are tested and current
- ✅ Documentation answers "why these choices?"
- ✅ Presets provide ready-to-deploy starting points for common patterns

**Incompleteness Indicators:**
- ❌ Proto has fields that aren't used in IaC modules
- ❌ IaC modules reference fields not in proto
- ❌ Examples fail validation against current schema
- ❌ Silent coverage omissions (no recorded reason for anything left out)
- ❌ Missing Terraform or Pulumi implementation
- ❌ IaC modules work but are uncommented, or narrate line-by-line instead of explaining why
- ❌ No presets, or presets reference stale fields from an older spec.proto

---

## Ideal State Checklist

The following sections define the complete, ideal state of any component. This serves as both a reference for developers building components and as the specification for automated auditing.

**File-set presence and absence is machine-enforced.** The `pkg/anatomy` conformance gate (CI lane `.github/workflows/lint.component-anatomy.yaml`) checks every component folder against the one canonical anatomy, file by file -- which files must exist, which must not, and what belongs where. The checklists below therefore concentrate on **content judgment**: what makes a good spec, a good module, a good guide, a good preset. Where a checklist item is a pure existence check, treat it as a pointer to the anatomy gate.

### 1. Cloud Resource Registry

**Location:** `shared/cloudresourcekind/cloud_resource_kind.proto`

**Requirements:**

- [ ] **Enum Entry Exists** - Component has an entry in the `CloudResourceKind` enum
- [ ] **Correct Provider Band** - Enum value is within the correct provider's 1,000-wide band (each provider owns a full band so its catalog can grow without colliding with a neighbor; the authoritative allocation table lives in the enum's header comment in `shared/cloudresourcekind/cloud_resource_kind.proto`):
  - Test/dev/custom: 1-49
  - AWS: 1000-1999
  - Azure: 2000-2999
  - GCP: 3000-3999
  - Kubernetes: 4000-4999
  - DigitalOcean: 5000-5999
  - Cloudflare: 7000-7999
  - Auth0: 8000-8999
  - OpenFGA: 9000-9999
- [ ] **Unique Enum Value** - No duplicate enum numbers
- [ ] **Unique ID Prefix** - The `id_prefix` is globally unique across all providers
- [ ] **Proper Metadata** - `kind_meta` includes:
  - `provider` - Correct provider enum value
  - `version` - The served API version (e.g. `v1alpha1`; never bare `v1`)
  - `id_prefix` - Short, descriptive prefix (3-7 characters)
  - `service_group` - The provider-console service group the component is browsed under (a `CloudProviderServiceGroup` value belonging to the kind's own provider). REQUIRED for grouped providers (aws, azure, gcp, kubernetes, cloudflare, digital_ocean) and prohibited for providers without a service taxonomy — registry tests enforce both directions
- [ ] **Optional Metadata (when applicable)** - `kind_meta` may also include:
  - `prerequisites` - Other `CloudResourceKind`s that must exist first (e.g. an operator or CRD-installer like `KubernetesGatewayApiCrds`); drives resource-graph and infra-chart ordering
  - `is_service_kind` - Whether this kind is a Service Hub deployment target
  - `container_kind` - Whether this kind contains child resources in the org graph
  - `publishes_kubernetes_connection` - Set only on cluster kinds whose deploy publishes a Kubernetes provider connection (EKS, GKE, AKS). Drives ordering: a Kubernetes workload whose `planton.dev/connection` names the connection the cluster will publish orders after the cluster, in `pkg/manifestgraph` and the platform alike. Flagging a new cluster kind is a two-repo act — the platform ships its connection materializer in the same release train (its conformance test binds the two)
  - `deprecations` - Schema versions announced as deprecated (each with an optional plain-language `note`). A deprecated version keeps serving and converting; every surface that speaks versions announces it as on the way out. Each entry must name a non-served version this release ships a schema for, with an authored conversion path to the served version — the registry tests and the bundle conformance gate refuse anything else
  - Note: there is no `kubernetes_meta`/`category`/`namespace_prefix` field on `CloudResourceKindMeta`; the Kubernetes layout is flat (`catalog/kubernetes/<component>/`).

**Example:**
```protobuf
GcpCertManagerCert = 3016 [(kind_meta) = {
  provider: gcp
  version: "v1alpha1"
  id_prefix: "gcpcert"
  service_group: gcp_security
}];
```

---

### 2. Folder Structure

**Base Path:** `catalog/<provider>/<component>/`

The anatomy follows one rule: **version directories hold only the versioned API contract; the component root holds the living component.** The `pkg/anatomy` gate enforces this shape file by file.

**Requirements:**

- [ ] **Correct Provider Hierarchy** - Component folder is under the correct provider:
  - `catalog/aws/<component>/`
  - `catalog/gcp/<component>/`
  - `catalog/azure/<component>/`
  - `catalog/kubernetes/<component>/`
  - etc.

- [ ] **Lowercase Folder Naming** - Component folder name matches the `CloudResourceKind` enum value but in all lowercase
  - Enum: `GcpCertManagerCert` → Folder: `gcpcertmanagercert`
  - Enum: `KubernetesPostgres` → Folder: `kubernetespostgres`

- [ ] **Version Directory Holds the Contract Only** - The versioned API contract lives under a maturity-channel version directory (`v1alpha1`, `v1beta1`, `v1`, ... — never a bare `v` glob, and today's components serve `v1alpha1`): `api.proto`, `spec.proto`, `input.proto`, `outputs.proto`, their `.pb.go` stubs, `BUILD.bazel`, `spec_test.go`, and `reference.md`. Everything that lives and evolves with the component — docs, logo, presets, example manifests, IaC modules — sits at the component root, outside any version directory.

- [ ] **One Kind, One Glyph** - `logo.svg` is the glyph every console card, catalog row, and diagram node wears for the kind, and on a diagram it is the one identity signal nothing else supplies: color says the family, shape says the nature, only the glyph says WHICH thing. A kind wears an official mark only when it IS a product or built-in object of the provider it belongs to, unaltered, and only where the provider offers its icons for use in third-party diagrams (Google Cloud does; Kubernetes publishes its resource icon set for exactly this; AWS offers its Architecture Icons in two forms, a service tile for the kind that IS a service and a resource glyph for the kind that IS a resource it draws; Microsoft permits its Azure architecture icons in architecture diagrams, and its set draws most portal resources, so a subnet or a disk wears its own); where a provider's brand terms reserve its logos and icons (Cloudflare's and DigitalOcean's do), every kind of that provider wears a Planton-drawn glyph. An official icon is worn only when one file of it reads on both a light and a dark plinth — a catalog logo is one CDN-served file an image element cannot re-theme — so an icon a provider publishes only as a light/dark pair is not worn, and recoloring it is the alteration every provider forbids. Software of another project hosted on a provider — a database, a broker, a mesh, an operator running on Kubernetes — is not the provider's product and is always Planton-drawn, whatever that project's own terms would allow, so every catalog logo stays under one owner's terms. Planton's own kinds wear the Planton brand mark. Every other kind — a database inside an instance, a job of a service, a record in a zone, a grant, a schedule — wears a Planton-drawn glyph that says what that kind is, in the provider's icon language, never containing, modifying, or resembling a provider's or a project's mark; the one exception is tied to a license, not a provider: where a provider publishes its own icon set under terms that permit derivative work, drawn glyphs for that provider's own objects may extend the set on its base form. Two kinds of one provider never share a glyph, except Planton's own kinds sharing Planton's — and a glyph is the drawing, compared with its `<desc>`, `<title>`, comments, and whitespace removed, so a copied icon never becomes its own by describing itself differently. Every file carries its provenance in a `<desc>` (`Official <Provider> product icon: …` or `Official <Provider> resource icon: …`, `Planton-drawn glyph: …`, or `Planton brand mark`), and when a provider that offers its icons later publishes one for a kind wearing a drawn glyph, the official one replaces it. Whether a glyph reads is judged by a person on the contact sheet `tools/catalog-logo-sheet` renders, at the sizes the console draws an icon (18px on an attachment plate is the size that decides), on a light and a dark wash — and, for a set judged before its kinds have diagram cards, on the nine family washes the platform hands the tool. The `pkg/cataloglogo` gate enforces all of this (CI lane `lint.catalog-logo.yaml`); its baseline was the debt of providers judged before the law, only ever shrank, and is empty now that every provider in the catalog is judged.

**Example Structure:**
```
catalog/gcp/gcpcertmanagercert/
├── README.md                    # GitHub-facing component page
├── catalog.md                   # THE catalog page
├── logo.svg                     # Component logo
├── GUIDE.md                     # Authored operational judgment (optional)
├── cost.yaml                    # Cost profile: billing model + the spec fields that drive the bill
├── controls.yaml                # Control profile: posture per central-catalog control, with evidence
├── v1alpha1/                    # The versioned contract ONLY
│   ├── api.proto
│   ├── spec.proto
│   ├── input.proto
│   ├── outputs.proto
│   ├── api.pb.go
│   ├── spec.pb.go
│   ├── input.pb.go
│   ├── outputs.pb.go
│   ├── BUILD.bazel
│   ├── spec_test.go
│   └── reference.md
├── presets/                     # .yaml manifests + .md sidecar pairs
│   ├── 01-managed-dns-validated.yaml
│   ├── 01-managed-dns-validated.md
│   ├── 02-load-balancer-cert.yaml
│   └── 02-load-balancer-cert.md
├── e2e/                         # Validated example manifest + test assets
│   ├── manifest.yaml            # The canonical example manifest
│   ├── profile.yaml             # (optional)
│   ├── prerequisite.yaml        # (optional)
│   └── scenarios/               # (optional) manifest variants
├── conversions/                 # (optional) cross-version conversion specs
└── iac/
    ├── permissions.yaml         # Least-privilege runner permissions (derived or proven)
    ├── import-map.yaml          # (optional)
    ├── pulumi/                  # No Makefile, no .gitignore
    │   ├── main.go
    │   ├── Pulumi.yaml
    │   ├── README.md
    │   └── module/
    │       ├── main.go
    │       ├── locals.go
    │       ├── outputs.go
    │       └── <resource-specific>.go
    └── tf/                      # No Makefile, no .gitignore
        ├── provider.tf
        ├── variables.tf
        ├── locals.tf
        ├── main.tf
        ├── outputs.tf
        └── README.md
```

The runner permissions manifests are machine-verified beyond their structure: every AWS action name a `permissions.yaml` declares is checked against a committed snapshot of AWS's own machine-readable service reference, every Azure operation against a snapshot of ARM's provider-operations inventory — on its own plane, since Azure separates management-plane `actions` from data-plane `data_actions` and an operation on the wrong plane would grant nothing — and every GCP permission against a snapshot of IAM's testable-permissions inventory, keyed by the service segment of the dotted name (all snapshots refreshed by `make generate-action-inventory`; validated in CI by the `lint.catalog-data` gate). An invented or misspelled permission action cannot ship — wildcard patterns must match at least one action the provider actually defines. The AWS snapshot also records which actions the service reference declares no resource types for: IAM evaluates those against `Resource "*"` only, so the gate holds every statement carrying one to exactly the wildcard resource (an ARN-scoped grant would read tighter than required and silently deny at runtime) and, conversely, refuses a wildcard statement whose actions all support scoping. The GCP snapshot unions three scope-anchored queries — project, organization, and billing account — because IAM's inventory lists a permission only under the resource types it can be tested on, and permissions like `resourcemanager.projects.create` (org-scoped) or `billing.resourceAssociations.create` (billing-scoped) are invisible to a project-anchored query alone. Token-authenticated providers get the same treatment in their own vocabulary: every Cloudflare permission group a manifest names is checked as a (name, scope) pair against a snapshot of Cloudflare's own permission-group inventory — names are not unique across scopes, so a real group declared at the wrong scope is refused as distinctly as an invented one — and every DigitalOcean token scope against a snapshot of the provider's published scope reference, matched exactly (DigitalOcean evaluates no wildcards) with the global alias scopes (`api:read`/`api:write`) refused outright as never least-privilege. DigitalOcean's second credential plane — Spaces object storage, which speaks the S3-compatible API under a separate Spaces key pair no API token can reach — is declared as per-bucket grant levels held to the provider's own closed grant vocabulary (read, readwrite, fullaccess). Every Auth0 Management API scope is checked against a snapshot of a tenant's own Management API definition — the system resource server a tenant cannot edit, and the set Auth0 checks a grant against — keyed by the resource segment of Auth0's verb-first `verb:resource` spelling. The published API reference is deliberately not the source: it omits scopes the tenant enforces (`update:connections_options`, without which a declared connection installs once and every later apply is refused), so a snapshot drawn from it would refuse a correct manifest. `make generate-action-inventory ARMS=<provider>` refreshes one arm alone, needing only that provider's credential. An entry needed only when a manifest sets a spec field says so as data (`condition.spec_field_set`), never only in its notes: the gate proves the field exists in the component's own spec, so a catalog page can mark the entry optional, a chart can union it only when one of its manifests sets the field, and a connection check does not report its absence as a gap.

Two central homes at the catalog root complement the per-component files: `catalog/_compliance/` holds the control catalog and framework crosswalks that every `controls.yaml` references (each crosswalk may declare its provider scope in `spec.providers` -- a benchmark that names a provider in its own title, like CIS AWS, appears only on that provider's components, while an empty scope means provider-neutral and applies everywhere), and `catalog/_pricing/` holds the cost-estimate pipeline — prices are volatile, so they live in one refreshable tree instead of beside every component. A component's quantities have exactly one home, in one of three forms: `derivations/` carries machine-executable rules turning ANY manifest's spec values into metered quantities and price choices (schema `finops/componentcostderivation/v1` — conditions select which rules apply, and may range existentially over the elements of a repeated tree or match a value's version-family prefix; quantity factors multiply out consumption, including the excess over a provider's included allotment; a rule may expand once per element of a repeated list so each instance's own class picks its own rate; and prices resolve by slug or by the manifest's own values through the price book's attribute identities; configurations whose cost is unknowable refuse with the reason, never a guess), `capacity/` carries the cluster-capacity twin (schema `finops/componentcapacityderivation/v1` — workload bindings locate the manifest's ContainerResources blocks, instance counts, and per-instance volumes, falling back to the spec's own declared defaults when a manifest omits what the modules default at deploy time, and exact Kubernetes-quantity arithmetic sums them into the capacity footprint the workload reserves; cluster-capacity components create no cloud SKU, so their estimates state capacity, never dollars), while `models/` carries hand-authored per-preset quantity assumptions for components not yet derived (schema `finops/componentcostestimatemodel/v1`). `pricebook/` carries one pinned price book per provider (schema `finops/pricebook/v1` — every unit price with its source URL and retrieval date; entries with a machine selector are refreshed from the provider's public price API by `make generate-price-book`; entries carrying attributes are the value-keyed identities derivations look prices up by). `estimates/` is GENERATED by `make generate-cost-estimates` — for derived components by replaying every preset manifest through the rules, for modeled components by joining the model with the book — every line cost, total, and footprint computed exactly, never authored. The `lint.catalog-data` gate re-computes every figure, holds the committed estimates byte-identical to their inputs, and rejects any disagreement between the rules or model, the price book, and the component's `cost.yaml`.

The fact-sheets also ship: the catalog bundle packs each covered component's cost profile, control profile, permission manifest, generated estimates, and (when derived) its cost derivation as their own cargo trees (`costs/`, `controls/`, `permissions/`, `estimates/`, `derivations/`), together with the central control catalog, crosswalks, and price books (`compliance/`, `pricebooks/`) — every file byte-identical to its tree source. The derivations and price books riding together is what lets a consuming control plane compute a manifest's exact monthly estimate server-side with zero external calls. Each covered component's catalog entry additionally carries projected summaries (monthly cost range across priced presets, control-posture counts, permissions provenance), computed at bundle build time from the same documents the cargo packs and re-proven by the bundle conformance gate — a summary can never disagree with the document behind it. Components without fact-sheets ship no cargo and no summaries: absence means "not yet covered", never "free".

---

### 3. Protobuf API Definitions

**Location:** `<version>/*.proto` (e.g. `v1alpha1/`)

#### 3.1 api.proto

**Purpose:** Wires together the Kubernetes Resource Model envelope (metadata, spec, status)

**Requirements:**

- [ ] **File Exists** - `<version>/api.proto` is present (enforced by the anatomy gate)
- [ ] **Correct Package** - Package declaration matches path:
  - `package dev.planton.<provider>.<component>.<version>;`
- [ ] **Standard Imports** - Imports common proto dependencies:
  ```protobuf
  import "buf/validate/validate.proto";
  import "catalog/<provider>/<component>/<version>/spec.proto";
  import "catalog/<provider>/<component>/<version>/outputs.proto";
  import "shared/metadata.proto";
  ```
- [ ] **Resource Message** - Defines `<Kind>` message with KRM structure:
  ```protobuf
  message <Kind> {
    string api_version = 1 [(buf.validate.field).string.const = '<provider>.planton.dev/<version>'];
    string kind = 2 [(buf.validate.field).string.const = '<Kind>'];
    dev.planton.shared.CloudResourceMetadata metadata = 3 [(buf.validate.field).required = true];
    <Kind>Spec spec = 4 [(buf.validate.field).required = true];
    <Kind>Status status = 5;
  }
  ```
- [ ] **Status Message** - Defines `<Kind>Status` message wrapping the stack outputs:
  ```protobuf
  message <Kind>Status {
    // stack-outputs
    <Kind>StackOutputs outputs = 1;
  }
  ```

#### 3.2 spec.proto

**Purpose:** Defines the configuration schema (the "spec" section of the manifest)

**Requirements:**

- [ ] **File Exists** - `<version>/spec.proto` is present (enforced by the anatomy gate)
- [ ] **Correct Package** - Package declaration matches path
- [ ] **Validation Imports** - If using field validations, imports buf.validate:
  ```protobuf
  import "buf/validate/validate.proto";
  ```
- [ ] **Spec Message** - Defines `<Kind>Spec` message with provider-specific fields
- [ ] **Field Validations** - Critical fields have validation rules:
  - Required fields: `[(buf.validate.field).required = true]`
  - String patterns: `[(buf.validate.field).string.pattern = "regex"]`
  - Numeric ranges: `[(buf.validate.field).int32 = {gte: 1, lte: 100}]`
- [ ] **Documentation** - Every field has a comment explaining its purpose
- [ ] **Provider Parity** - Fields model the full configurable argument surface of the mapped provider resources at the pinned provider version, with sensible defaults
- [ ] **Enums for Choices** - Use enums for fields with fixed choices (not free-form strings)

**Example:**
```protobuf
message GcpCertManagerCertSpec {
  // GCP project ID where certificate will be created
  string gcp_project_id = 1 [(buf.validate.field).required = true];
  
  // Primary domain name for the certificate
  string primary_domain_name = 2 [(buf.validate.field).required = true];
  
  // Alternate domain names (SANs)
  repeated string alternate_domain_names = 3;
  
  // Certificate type (MANAGED or LOAD_BALANCER)
  CertificateType certificate_type = 4;
}
```

#### 3.3 input.proto

**Purpose:** Defines inputs to the IaC modules (includes spec + credentials + environment context)

**Requirements:**

- [ ] **File Exists** - `<version>/input.proto` is present (enforced by the anatomy gate)
- [ ] **Correct Package** - Package declaration matches path
- [ ] **Standard Imports** - Imports common dependencies:
  ```protobuf
  import "catalog/<provider>/<component>/<version>/api.proto";
  import "catalog/<provider>/provider.proto";
  ```
- [ ] **StackInput Message** - Defines `<Kind>StackInput` message with the target resource and the provider config:
  ```protobuf
  message <Kind>StackInput {
    // target cloud-resource
    <Kind> target = 1;
    // provider configuration / credentials
    <Provider>ProviderConfig provider_config = 2;
  }
  ```
- [ ] **Provider Config Field** - References the correct provider config type (the `<Provider>ProviderConfig` message from `catalog/<provider>/provider.proto`, consistent with the `provider_config = 2` example above):
  - AWS: `dev.planton.aws.AwsProviderConfig`
  - GCP: `dev.planton.gcp.GcpProviderConfig`
  - Kubernetes: `dev.planton.kubernetes.KubernetesProviderConfig`

#### 3.4 outputs.proto

**Purpose:** Defines outputs from the IaC deployment (what gets written to status.outputs)

**Requirements:**

- [ ] **File Exists** - `<version>/outputs.proto` is present (enforced by the anatomy gate)
- [ ] **Correct Package** - Package declaration matches path
- [ ] **StackOutputs Message** - Defines `<Kind>StackOutputs` message
- [ ] **Relevant Outputs** - Contains outputs that users actually need:
  - Resource identifiers (IDs, ARNs, names)
  - Connection information (endpoints, URLs, IPs)
  - Generated values (passwords via secrets, connection strings)
- [ ] **Documentation** - Every output field has a comment
- [ ] **No Sensitive Data** - Passwords/keys reference secret managers, not plain text

**Example:**
```protobuf
message GcpCertManagerCertStackOutputs {
  // Certificate resource ID
  string certificate_id = 1;
  
  // Certificate status (ACTIVE, PENDING, FAILED)
  string certificate_status = 2;
  
  // Expiration timestamp
  string expiration_time = 3;
}
```

#### 3.5 Generated Proto Stubs

**Requirements:**

- [ ] **Go Stubs Generated** - `.pb.go` files exist for all `.proto` files:
  - `api.pb.go`
  - `spec.pb.go`
  - `input.pb.go`
  - `outputs.pb.go`
- [ ] **Stubs Are Current** - Generated files match proto definitions (run `make protos` to regenerate)

#### 3.6 Unit Tests

**Location:** `<version>/spec_test.go`

**Purpose:** Validate that all buf.validate rules in spec.proto are syntactically and semantically correct

**Requirements:**

- [ ] **File Exists** - `<version>/spec_test.go` is present (enforced by the anatomy gate)
- [ ] **Substantial Content** - File is non-empty (>500 bytes indicates real tests)
- [ ] **Validation Tests** - Tests for ALL validation rules in `spec.proto`:
  - Test that required fields trigger validation errors when missing
  - Test that pattern validations work correctly (string patterns, regex)
  - Test that range validations enforce limits (min/max, gte/lte)
  - Test that enum validations reject invalid values
  - Test that custom CEL validations work as expected
- [ ] **Tests Execute** - All tests run successfully (no compilation errors)
- [ ] **Tests Pass** - All tests pass when running component-specific test:
  ```bash
  go test ./catalog/<provider>/<component>/v1alpha1/
  ```
- [ ] **Meaningful Coverage** - Tests cover critical validation paths:
  - Happy path (valid configurations)
  - Error paths (missing required fields, invalid patterns)
  - Edge cases (boundary values, special characters)
  - Each validation rule has at least one test

**Critical:** Test execution is part of completeness. A component with tests that fail is considered incomplete.

**Example:**
```go
func TestGcpCertManagerCertSpec_Validation(t *testing.T) {
    tests := []struct {
        name    string
        spec    *GcpCertManagerCertSpec
        wantErr bool
    }{
        {
            name: "valid spec",
            spec: &GcpCertManagerCertSpec{
                GcpProjectId:      "my-project",
                PrimaryDomainName: "example.com",
            },
            wantErr: false,
        },
        {
            name: "missing gcp_project_id",
            spec: &GcpCertManagerCertSpec{
                PrimaryDomainName: "example.com",
            },
            wantErr: true,
        },
    }
    // ... test implementation
}
```

---

### 4. IaC Modules - Pulumi

**Base Path:** `iac/pulumi/` (component root)

#### 4.1 Pulumi Module Files

**Location:** `iac/pulumi/module/`

**Purpose:** The actual deployment logic (the "recipe")

**CRITICAL:** Files must contain **actual implementation**, not empty stubs. Both audit and completion workflows must verify file content, not just existence.

**Requirements:**

- [ ] **main.go** - Controller/orchestrator that:
  - Loads `<Kind>StackInput` from environment variable
  - Sets up provider configuration (using credentials from stack input)
  - Calls resource-specific logic
  - Returns stack outputs
  - **MUST NOT** be an empty stub that just returns `nil`
  - **MUST** contain actual provider setup and resource creation calls
- [ ] **locals.go** - Data transformations and computed values:
  - Transforms spec fields into provider-specific formats
  - Generates names, labels, tags
  - Computes derived values
  - **MUST** contain actual field extraction and computation logic
  - **MUST NOT** just define empty structs
- [ ] **outputs.go** - Maps deployed resources to `<Kind>StackOutputs`:
  - Extracts resource IDs, ARNs, endpoints
  - Formats output structure matching `outputs.proto`
  - **MUST** contain actual `ctx.Export()` calls
  - **MUST** map all fields from `outputs.proto`
- [ ] **Resource-Specific Files** - One or more `.go` files containing actual resource provisioning logic
  - Example: `cert_manager_cert.go` for the certificate resource
  - Example: `dns_authorization.go` for DNS validation resources
  - **MUST** contain actual resource creation logic using provider SDK
  - **MUST NOT** be empty or return nil without creating resources

**Code Quality:**
- [ ] **Uses Generated Stubs** - Imports and uses the generated protobuf Go stubs
- [ ] **Provider Configuration** - Correctly configures the provider (AWS, GCP, etc.) using credentials
- [ ] **Error Handling** - Proper error handling and propagation
- [ ] **Resource Dependencies** - Explicit dependencies where needed (e.g., Pulumi `DependsOn`)
- [ ] **Compiles Successfully** - `go build` succeeds without errors
- [ ] **No Empty Stubs** - Functions return actual resources, not nil
- [ ] **Well-Commented Module Code** - Authoring comments explain *why*, trade-offs, provider quirks, and non-obvious ordering (not line-by-line narration), to the same density and intent as the `spec.proto` field-comment standard. This is the canonical **module-comment bar** the forge, update, fix, and audit rules reference; these modules render on the public catalog, so their comments are part of the deliverable.

#### 4.2 Pulumi Entrypoint Files

**Location:** `iac/pulumi/`

**Requirements:**

- [ ] **main.go** - Entry point that:
  ```go
  func main() {
      pulumi.Run(func(ctx *pulumi.Context) error {
          return module.Resources(ctx)
      })
  }
  ```
- [ ] **Pulumi.yaml** - Project configuration:
  ```yaml
  name: <component-name>
  runtime: go
  description: Pulumi module for <Kind>
  ```
- [ ] **README.md** - Pulumi-specific usage guide
- [ ] **No Makefile, no .gitignore** - Module directories carry no build scaffolding; the anatomy gate rejects both. Build with plain `go build`, and exercise the module against the example manifest with the CLI:
  ```bash
  # from inside iac/pulumi/
  planton pulumi preview --manifest ../../e2e/manifest.yaml --module-dir .
  ```

**Integration:**
- [ ] **Compiles Successfully** - `go build` completes without errors
- [ ] **Plugin Dependencies Listed** - `Pulumi.yaml` documents required plugins
- [ ] **Executable** - Binary can be built and run

---

### 5. IaC Modules - Terraform

**Base Path:** `iac/tf/` (component root)

**Purpose:** Feature-parity Terraform implementation

**CRITICAL:** Files must contain **actual implementation**, not empty stubs. Both audit and completion workflows must verify file content, not just existence.

**Requirements:**

- [ ] **variables.tf** - Input variables that mirror `spec.proto`:
  - Every field in `<Kind>Spec` has a corresponding Terraform variable
  - Variable types match proto field types (string, number, list, map)
  - Required fields are marked as required in Terraform
  - Optional fields have default values matching proto defaults
  - Variable descriptions match proto field comments
  - **MUST** be generated and match spec.proto exactly

**Critical:** The Planton CLI transforms the YAML manifest into Terraform variable format. If `variables.tf` doesn't match `spec.proto`, deployments will fail.

- [ ] **provider.tf** - Provider configuration:
  - Configures the appropriate provider (AWS, GCP, Azure, etc.)
  - Uses credential information passed via variables
  - Sets provider version constraints
  - **MUST NOT** be empty
  - **MUST** contain actual provider configuration block

- [ ] **locals.tf** - Local value transformations:
  - Transforms input variables into provider-specific formats
  - Computes derived values (names, labels, tags)
  - Centralizes repeated expressions
  - **MUST** contain actual local value definitions
  - **MUST NOT** be empty or missing

- [ ] **main.tf** - Resource definitions:
  - Creates the primary resources
  - Creates supporting resources (networking, IAM, etc.)
  - Manages resource dependencies
  - **MUST NOT** be empty (0 bytes) or contain only comments
  - **MUST** contain actual `resource` blocks using provider SDK
  - **MUST** implement all fields from spec.proto

- [ ] **outputs.tf** - Output values matching `outputs.proto`:
  - Every field in `<Kind>StackOutputs` has a corresponding Terraform output
  - Output descriptions match proto field comments
  - **MUST** contain actual `output` blocks
  - **MUST** extract values from created resources

- [ ] **README.md** - Terraform-specific usage guide

- [ ] **No Makefile, no .gitignore** - Module directories carry no build scaffolding; the anatomy gate rejects both. Exercise the module against the example manifest with the CLI:
  ```bash
  # from inside iac/tf/
  planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
  ```

**Code Quality:**
- [ ] **Valid HCL** - All `.tf` files are valid Terraform configuration
- [ ] **Validates Successfully** - `terraform validate` passes
- [ ] **Feature Parity with Pulumi** - Creates the same resources as Pulumi module
- [ ] **No Hardcoded Values** - All configuration comes from variables
- [ ] **Proper Dependencies** - Uses `depends_on` where needed
- [ ] **Not Empty** - main.tf has substantial content (>100 bytes minimum)
- [ ] **Functional** - Can actually deploy resources, not just validate syntax
- [ ] **Well-Commented Module Code** - Meets the same **module-comment bar** as the Pulumi module (§4.1 Code Quality): comments explain *why*/trade-offs/provider quirks/ordering, not line-by-line narration.

**Example Structure:**

`variables.tf` mirrors `spec.proto`:
```hcl
variable "gcp_project_id" {
  description = "GCP project ID where certificate will be created"
  type        = string
}

variable "primary_domain_name" {
  description = "Primary domain name for the certificate"
  type        = string
}

variable "alternate_domain_names" {
  description = "Alternate domain names (SANs)"
  type        = list(string)
  default     = []
}
```

---

### 6. Documentation - Authored Guide

**Location:** `GUIDE.md` (component root, optional)

**Purpose:** The durable-judgment home. The guide carries the authored operational judgment a component has earned: operational recipes, design rationale, and parity accounting. Research working notes are never committed -- research informs the design, and the judgment it yields is distilled into the guide, the spec comments, and the user-facing docs.

**CRITICAL:** Where a guide exists, it is the **primary source of truth** for understanding the component's judgment. It should be consulted when:
- Executing any lifecycle operation (forge, audit, update, delete)
- Making decisions about component behavior
- Understanding design rationale and scoping decisions
- Troubleshooting or debugging issues
- Evaluating whether to keep, update, or delete the component

**Requirements:**

- [ ] **Warranted Presence** - `GUIDE.md` exists at the component root where the component carries non-obvious judgment (state-ownership flags, component-vs-flag choices, operational surprises). A missing guide is not a defect by itself: guides exist where judgment earns a place
- [ ] **Substantial Relative to Complexity** - Short is fine; empty scaffolding is not
- [ ] **Operational Recipes** - The non-obvious "how do I actually run this" knowledge: sequencing, day-2 operations, what breaks and how to recover
- [ ] **Design Rationale** - Why the kind is shaped the way it is: what was composed in, what was deliberately exported to neighboring kinds, and what the trade-offs were
- [ ] **Parity Accounting** - Explicit statement of:
  - The pinned provider version parity is declared against
  - Full configurable-argument coverage of the mapped provider resources (no deliberate exclusions)
  - Any provider resources deliberately composed into this kind, with the mapping documented
  - Any recorded exclusions (deprecated or superseded surface only), each with its reason

**Content Quality:**
- [ ] **Teaches Judgment** - "When X vs when Y, and what breaks if you choose wrong" -- never feature lists
- [ ] **Grounded** - Every claim traces to `spec.proto`, the IaC modules, or recorded platform behavior
- [ ] **Opinionated** - Makes clear recommendations
- [ ] **Actionable** - Readers understand what to do
- [ ] **Well-Structured** - Uses headings, sections, tables

**Reference:** `catalog/kubernetes/kubernetesdeployment/GUIDE.md` is a worked example of the form: it opens with the judgment the guide carries and spends every section on composition choices and their consequences.

---

### 7. Documentation - User-Facing

#### 7.1 README.md

**Location:** `README.md` (component root)

**Purpose:** Concise, Planton perspective overview -- the GitHub-facing component page

**Requirements:**

- [ ] **File Exists** - `README.md` is present at the component root (enforced by the anatomy gate)
- [ ] **Moderate Length** - Typically 50-200 lines (not a deep research document)
- [ ] **Overview Section** - High-level explanation from Planton perspective:
  - What the component does
  - Why Planton created it
  - How it fits into the framework
- [ ] **Purpose Section** - Clear statement of goals:
  - What problems it solves
  - What it simplifies
- [ ] **Key Features** - Bullet points of capabilities
- [ ] **Benefits** - Why users should use this vs alternatives
- [ ] **Example Usage** - One simple, complete example showing:
  - YAML manifest
  - CLI deployment command
  - Expected outcome
- [ ] **Best Practices** - Quick tips for production use

**NOT Included:**
- Operational judgment and design rationale (that's `GUIDE.md`)
- History of the technology (not relevant to users)
- Comparison of every deployment method (too detailed)
- Every possible configuration option (that's what presets and `v1alpha1/reference.md` cover)

**Tone:**
- Helpful and encouraging
- Focused on getting started quickly
- Assumes reader knows basic concepts
- Points to other documentation for depth

#### 7.2 catalog.md

**Location:** `catalog.md` (component root)

**Purpose:** THE catalog page -- the user-facing landing page rendered by the console's public catalog; the page a person or AI agent lands on when searching for this software

**Requirements:**

- [ ] **File Exists** - `catalog.md` is present at the component root (enforced by the anatomy gate)
- [ ] **Head-Shape Contract** - The `# H1` is the human display name and the first sentence stands alone as the one-line description -- the catalog bundle machine-parses both for the kind's display metadata (and the bundle deliberately falls back rather than fail on a malformed head, which is why the shape is gated below)
- [ ] **Follows the Catalog Page Standard** - The mandatory section structure (What Gets Created / Before You Deploy / Deploy with Console + CLI + InfraChart arms / Key Configuration / Outputs and Dependencies / Common Patterns / Works With), tone, and content follow `_rules/docs/write-planton-component-catalog-md.mdc`. The structural half is machine-enforced: the `pkg/catalogpage` gate (CI lane `.github/workflows/lint.catalog-page.yaml`) checks the head shape, the H2/H3 skeleton, the InfraChart-arm law, and the validity of every embedded manifest against a shrink-only baseline -- a new page ships at the bar or fails the PR
- [ ] **Judgment-First Configuration** - `Key Configuration` teaches decisions and consequences; the exhaustive field list belongs to the generated `v1alpha1/reference.md`, never duplicated here
- [ ] **Source-Verified** - Every manifest validates against `spec.proto`, every output exists in `outputs.proto`, every What Gets Created bullet matches the IaC module
- [ ] **Current** - Field references and examples match the current `spec.proto`

**NOT Included:**
- Exhaustive field tables (that's the generated `v1alpha1/reference.md`)
- Deep operational judgment for architects (that's `GUIDE.md`)
- External URLs, marketing language, or research prose (banned by the standard)

**Tone:**
- Senior documentation writer: precise, earned, trusts the reader
- Written for three readers at once: the evaluating developer, the 2 AM engineer, and the AI agent grounding a deployment plan

---

### 8. Supporting Files

#### 8.1 Example Manifest

**Location:** `e2e/manifest.yaml` (component root)

**Purpose:** The canonical validated example manifest. It is load-bearing
twice: the reference generator embeds it as the reference page's Example
block, and the e2e framework treats its presence as the testability marker
and deploys exactly this manifest.

**Requirements:**

- [ ] **File Exists** - `e2e/manifest.yaml` is present (the anatomy
  conformance gate and e2e discovery both key on it)
- [ ] **Valid Manifest** - Complete YAML manifest with:
  - `apiVersion`, `kind`, `metadata`, `spec`
  - Concrete, working values that pass protovalidate
  - Deployable by the e2e runner and usable with
    `planton pulumi preview --manifest e2e/manifest.yaml`
- [ ] **Non-Production Values** - Uses test/dev values (not real production data)

#### 8.2 Pulumi Supporting Files

**Location:** `iac/pulumi/` (component root)

**Files:**

- [ ] **README.md** - Pulumi module usage guide:
  - How to use the module standalone
  - Required environment variables
  - How to pass credentials
  - Example deployment commands
  - Troubleshooting tips

#### 8.3 Terraform Supporting Files

**Location:** `iac/tf/` (component root)

**Files:**

- [ ] **README.md** - Terraform module usage guide:
  - How to use the module standalone
  - Required variables
  - How to pass credentials
  - Example terraform commands
  - Troubleshooting tips

---

### 9. Presets

**Location:** `presets/`

**Purpose:** Production-quality, directly deployable YAML manifests representing the most common real-world configuration patterns for the component. Each preset is a ranked starting point that users can deploy immediately (after replacing placeholders) without needing to understand every field in `spec.proto`.

**Reference:** See `architecture/presets.md` for the full convention specification and authoring guide.

**Requirements:**

- [ ] **Directory Exists** - `presets/` directory is present
- [ ] **At Least One Preset** - Minimum 1 YAML + companion MD pair (rank 01 = the "30-second decision" configuration)
- [ ] **KRM Envelope Correct** - Every preset YAML has `apiVersion` and `kind` matching the exact constants in `api.proto`
- [ ] **Metadata Convention** - `metadata.name` is prefixed with `my-` (signals a template, not a live resource)
- [ ] **StringValueOrRef Compliance** - All `StringValueOrRef` fields use the `value:` wrapper form with descriptive angle-bracket placeholders
- [ ] **Naming Convention** - Files follow `{NN}-{kebab-case-description}.yaml` + `.md` pattern:
  - Rank is a zero-padded two-digit number (`01`-`99`)
  - Description is lowercase, hyphenated, no spaces or underscores
- [ ] **Companion Markdown** - Every `.yaml` has a companion `.md` with required sections:
  - Title (H1)
  - Description (2-4 sentences)
  - When to Use (bulleted list)
  - Key Configuration Choices (bulleted list with field references)
  - Placeholders to Replace (table)
- [ ] **No Duplicate Ranks** - Each rank number is unique within the component's `presets/` directory
- [ ] **Schema Consistency** - All field names in preset YAML files exist in the current `spec.proto` (no stale references to renamed or removed fields)
- [ ] **Default Annotations Honored** - Fields with `recommended_default` or `default` proto annotations use the annotated value with a citing comment
- [ ] **No Status Section** - Presets must not include a `status` block (status is system-managed)

**Quality Guidelines:**

- **Quantity**: 1-5 presets per component. Simple components (3-5 spec fields) need only 1. Complex components with distinct deployment patterns (e.g., internal vs external, dev vs production) benefit from 2-4.
- **Ranking**: Rank 01 = the configuration you'd deploy with 30 seconds to decide. Lower ranks represent progressively more specialized patterns.
- **No Forced Patterns**: Do not create presets for hypothetical use cases. Every preset should represent a configuration that users actually deploy.
- **Deployable**: Every preset must be structurally valid and deployable after replacing angle-bracket placeholders.

**Relationship to Other Artifacts:**

| Artifact | Purpose | Presets Difference |
|----------|---------|-------------------|
| `README.md` | User-facing component overview | Presets are actionable starting points, README is explanation |
| `e2e/manifest.yaml` | The canonical validated example manifest | Presets are production-quality, not minimal validation configs |
| `GUIDE.md` | Authored operational judgment and design rationale | Presets are actionable starting points, not explanation |

**Example:**
```
presets/
├── 01-internet-facing-https.yaml    # Most common: HTTPS ALB with SSL
├── 01-internet-facing-https.md
├── 02-internal-http.yaml            # Internal-only, no SSL
└── 02-internal-http.md
```

---

### 10. Verified Fact-Sheets

Every component carries the machine-checked data layer beside its contract. These files are what the console's Cost/Posture/Permissions tabs, the compose canvas's cost figures, the CLI, and the assistant's file-first answers are built from — a component without them is honestly uncovered on every surface, never approximated.

- [ ] **cost.yaml** (component root) - Billing model, cost dimensions, and price provenance. Committed prices come from the pinned price books in `catalog/_pricing/pricebook/`; quantities have exactly one home (a derivation, a capacity derivation, or a model — see the central-homes paragraph above). Never hand-type a dollar figure here or anywhere else in the component's docs: estimates are GENERATED by `make generate-cost-estimates`, and the `lint.catalog-data` gate recomputes every figure.
- [ ] **controls.yaml** (component root) - Technical control posture with evidence, citing control IDs from the central catalog in `catalog/_compliance/`. Posture maps onto framework requirements through the crosswalks; it never claims a component is "compliant", "certified", or "authorized".
- [ ] **iac/permissions.yaml** - The least-privilege provisioning permissions the runner needs, per provider plane. Every action/scope is validated against the provider's own published inventory — an invented or misspelled permission fails CI.

Fact-sheet conformance is enforced by its own CI gates (the anatomy burn-down ledger and the catalog-data lint), so it is deliberately not folded into the audit percentage scheme below — the gates, not a score, decide whether the data layer is present and truthful.

---

## Completeness Assessment Criteria

When evaluating whether a component is "complete," assess each category:

### Critical (Must Have - 48.64%)

These are non-negotiable for a component to be considered functional:

1. ✅ Entry in `cloud_resource_kind.proto` (4.44%)
2. ✅ Correct folder structure (4.44%)
3. ✅ All four proto files (api, spec, input, outputs) (13.32%)
4. ✅ Generated proto stubs (.pb.go files) (3.33%)
5. ✅ spec_test.go with validation tests (2.77%)
6. ✅ **Tests execute and pass** (2.78%) - Component-specific `go test` succeeds
7. ✅ Pulumi module with main.go, locals.go, outputs.go (6.66%)
8. ✅ Pulumi entrypoint (main.go, Pulumi.yaml — no Makefile; the anatomy gate rejects build scaffolding) (6.66%)
9. ✅ Terraform module with all 5 core files (variables.tf, provider.tf, locals.tf, main.tf, outputs.tf) (4.24%)

**Note:** Test execution is now explicitly part of critical items. Failing tests = incomplete component.

### Important (Should Have - 41.36%)

These significantly improve quality and usability:

10. ✅ Authored operational judgment (GUIDE.md) (13.18%)
11. ✅ User-facing README (component-root README.md) (13.09%)
13. ✅ Pulumi supporting documentation (README, overview) (5.05%)
14. ✅ Terraform supporting documentation (README) (2.52%)
15. ✅ Supporting files (e2e example manifest) (2.52%)
16. ✅ Presets with companion documentation (presets/) (5.00%)

### Nice to Have (Polish - 10%)

These add polish and maintainability:

17. ✅ Extensive examples covering edge cases (3.33%)
18. ✅ Additional architecture documentation (3.33%)
19. ✅ Extra supporting files and helpers (3.34%)

### Percentage Calculation

**Completion Score:**

- Critical items: **48.64%** weight
  - Registry: 4.44%
  - Folder: 4.44%
  - Proto files: 13.32%
  - Generated stubs: 3.33%
  - Test file: 2.77%
  - **Test execution: 2.78%** ← Now explicit
  - Pulumi module: 6.66%
  - Pulumi entrypoint: 6.66%
  - Terraform module: 4.24%
  
- Important items: **41.36%** weight (7 major items)
  - Research docs: 13.18%
  - Examples: 6.55%
  - User-facing README: 6.54%
  - Pulumi supporting docs: 5.05%
  - Terraform supporting docs: 2.52%
  - Supporting files: 2.52%
  - **Presets: 5.00%** ← New
- Nice to Have: **10%** weight (polish items)

**Interpretation:**
- 100% - Fully complete, production-ready
- 80-99% - Functionally complete, minor improvements needed
- 60-79% - Partially complete, significant work remaining
- 40-59% - Skeleton exists, major implementation needed
- <40% - Early stage or abandoned

### Quality Multipliers

Beyond file existence, assess quality:

- **Proto Schema Quality** - Do fields match research findings? Are validations present?
- **IaC Implementation Quality** - Are both modules feature-complete? Do they work? Are they richly commented to the module-comment bar (why/trade-offs/quirks/ordering, not narration)?
- **Documentation Quality** - Is the research comprehensive? Are examples current?
- **Consistency Quality** - Do variables.tf match spec.proto? Do outputs match outputs.proto?

A component with all files but low quality in these dimensions should be scored lower than the raw percentage suggests.

---

## Using This Document

### For Developers

When building a new component, use this document as your checklist. Work through each section systematically, ensuring every requirement is met.

### For Reviewers

When reviewing a PR that adds or updates a component, use this document to validate completeness. Check off items and provide specific feedback on what's missing.

### For Auditing

This document serves as the specification for an automated audit tool. The tool should:

1. **Check file existence AND content** for each required file:
   - **CRITICAL:** Don't just check if file exists - verify it has actual implementation
   - Check file size (e.g., main.tf with 0 bytes is incomplete)
   - Check for empty stubs (e.g., Pulumi main.go that just returns nil)
   - Verify functions contain actual resource creation logic
2. Validate folder structure matches conventions
3. Check proto stubs are current (compare timestamps)
4. Validate terraform files with `terraform validate`
5. Check that variables.tf fields match spec.proto fields
6. Check that outputs.tf fields match outputs.proto fields
7. Run unit tests with `make test`
8. **Verify IaC module implementation completeness**:
   - Pulumi module: Check main.go has provider setup and resource calls
   - Pulumi module: Check locals.go extracts and computes values
   - Pulumi module: Check outputs.go has ctx.Export() calls
   - Terraform module: Check main.tf has resource blocks (not empty)
   - Terraform module: Check provider.tf has provider configuration
   - Terraform module: Check locals.tf has local value definitions
   - Terraform module: Check outputs.tf has output blocks
   - Both modules: Check authoring comments meet the module-comment bar (explain why/trade-offs/quirks/ordering, not merely present and not line-by-line narration)
9. **Verify preset coverage and correctness**:
   - Check `presets/` directory exists with at least one YAML + MD pair
   - Verify `apiVersion` and `kind` match `api.proto` constants
   - Verify all `StringValueOrRef` fields use `value:` wrapper
   - Verify preset field names exist in current `spec.proto` (detect stale presets)
   - Verify naming convention and companion file pairing
10. Calculate completion percentage based on **implementation**, not just file presence
11. Generate a report showing:
   - Overall completion percentage (considering implementation)
   - Missing items by category
   - Empty/stub files that need implementation
   - Quality issues (mismatches, outdated files, empty implementations)
   - Recommended next steps

**Key Principle:** A component with all files present but empty implementations should score LOW, not high. Implementation matters more than file existence.

---

## Conclusion

A "complete" component in Planton is not simply a collection of files. It's a well-researched, thoughtfully-scoped, fully-implemented package that serves real-world deployment needs with both Pulumi and Terraform, backed by comprehensive documentation that explains both "how" and "why," and equipped with ready-to-deploy presets that give users an immediate starting point.

This document provides the definitive reference for what completeness means, enabling both human developers and automated tools to assess and improve components systematically.

