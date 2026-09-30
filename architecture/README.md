# Planton Architecture

**A multi-cloud deployment framework that brings Kubernetes-style consistency to infrastructure deployments across any cloud provider.**

---

## Table of Contents

- [What is Planton?](#what-is-planton)
- [Core Architecture](#core-architecture)
  - [The Three Pillars](#the-three-pillars)
  - [The Component Concept](#the-component-concept)
- [Repository Structure](#repository-structure)
- [Technology Stack](#technology-stack)
- [Component Lifecycle](#component-lifecycle)
- [API Design Philosophy](#api-design-philosophy)
- [IaC Module Design](#iac-module-design)
- [CLI Architecture](#cli-architecture)
- [Development Workflows](#development-workflows)
- [Extension Patterns](#extension-patterns)
- [Contributing](#contributing)

---

## What is Planton?

Planton is an open-source framework that provides a unified, declarative approach to deploying infrastructure and applications across cloud providers. It solves a fundamental problem in modern cloud-native development: **the chaos of managing deployments across different clouds, each with their own tools, APIs, and mental models**.

### The Core Promise

**One structure. One workflow. Any cloud.**

Whether you're deploying a PostgreSQL database to AWS RDS, Google Cloud SQL, or a Kubernetes cluster, Planton provides the same consistent experience:
- Write a YAML manifest following the Kubernetes Resource Model
- Validate it before deployment
- Deploy using a single CLI command
- Get back structured outputs

The **manifests are provider-specific** (AWS RDS has different configuration than GCP Cloud SQL), but the **experience is identical**: same structure, same validation approach, same CLI, same workflow.

### Design Philosophy

**Consistency Without Abstraction**

Planton does NOT abstract away cloud provider differences. Instead, it provides:
- ✅ **Consistent structure:** Every resource uses KRM (apiVersion, kind, metadata, spec)
- ✅ **Consistent workflow:** Same CLI commands, same validation process
- ✅ **Consistent developer experience:** Same documentation approach, same error patterns
- ✅ **Provider-specific manifests:** Each deployment target has its own manifest with provider-specific configuration

**Why avoid abstraction?**

Cloud providers are fundamentally different. AWS RDS has `instance_class` and `security_group_ids`. GCP Cloud SQL has `tier` and `vpc_id`. Attempting to abstract these differences would either:
1. Force a "lowest common denominator" approach (losing provider-specific capabilities)
2. Create a leaky abstraction that's harder to understand than learning the providers directly

**Planton's philosophy:** Provide **consistency of experience** without **sacrificing provider-specific power**.

---

## Core Architecture

### The Three Pillars

Planton is built on three foundational components that work together seamlessly:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Planton CLI                         │
│              (Orchestration & Validation Layer)                 │
└───────────────────────┬─────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   APIs       │ │ IaC Modules  │ │   CLI Core   │
│ (Proto Defs) │ │ (Pulumi/TF)  │ │  (Go Binary) │
└──────────────┘ └──────────────┘ └──────────────┘
        │               │               │
        ▼               ▼               ▼
┌────────────────────────────────────────────────┐
│        Components (700+)            │
│  KubernetesPostgres | AwsRdsInstance | etc.   │
└────────────────────────────────────────────────┘
```

#### 1. APIs: Standardized Configuration Schema

**Technology:** Protocol Buffers  
**Inspiration:** Kubernetes Resource Model

Every component follows the same structure:

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
  # System-managed status (read-only)
```

**Why Protocol Buffers?**

Unlike Kubernetes (which uses Go structs), Planton uses Protocol Buffers to enable:

- **Language Neutrality:** Auto-generate SDKs in Go, Java, Python, TypeScript, and more
- **Beautiful Documentation:** Publish to Buf Schema Registry for instant, navigable documentation
- **Field-Level Validations:** Define validation rules directly in the API schema
- **Early Error Detection:** Catch configuration errors before deployment
- **Platform Engineering:** Import SDKs to build custom internal tools without reinventing schemas

Example validation in protobuf:
```protobuf
message PostgresKubernetesSpec {
  string cpu = 1 [(buf.validate.field).string.pattern = "^[0-9]+m$"];
  int32 replicas = 2 [(buf.validate.field).int32 = {gte: 1, lte: 10}];
}
```

The `planton validate` command checks these rules **before** calling any cloud APIs, providing instant feedback.

#### 2. IaC Modules: The "Recipes"

**Technology:** Pulumi and Terraform/OpenTofu  
**Approach:** Provider-specific, deliberately simple

Every component has **both** a Pulumi module and a Terraform module, and you choose which IaC engine to use -- unless its kind declares fewer engines (`kind_meta.provisioners`), because its provider publishes no Pulumi provider or its kinds are proven on OpenTofu alone. Such a component ships only the modules its engines run, and the CLI and the platform refuse any other engine before anything runs.

**Why Both Pulumi and Terraform?**

Different teams have different preferences and investments:

- **Pulumi:** Real programming languages (Go, Python, TypeScript), better for complex logic, type safety
- **Terraform/OpenTofu:** Mature ecosystem, HashiCorp Configuration Language, familiar to many DevOps teams

Planton doesn't force a choice—it supports both, maintaining feature parity between them.

**Design Philosophy: Deliberately Simple**

The default modules are intentionally designed to be **Terraform-like** even when written in Pulumi:

- Simple, straightforward code (no aggressive SOLID principles or DRY patterns)
- Single directory structure (like Terraform modules)
- Familiar file names (`main.go` similar to `main.tf`, `locals.go` for transformations)
- Minimal language features (only what's necessary)

**Why?** Because **adoption matters more than perfect code**. A Terraform engineer should be able to fork a Pulumi module and immediately understand the flow.

#### 3. CLI: The Orchestration Layer

**Distribution:** Homebrew  
**Role:** The "conductor" that brings everything together

Installation:
```bash
brew install plantonhq/tap/planton
```

**What the CLI does:**

1. **Reads your manifest** (local file or GitHub raw URL)
2. **Validates inputs** using proto-validate rules (catches errors early)
3. **Maps `kind` to IaC module** (knows which module deploys which component)
4. **Clones/pulls the module** from GitHub (with smart caching)
5. **Sets up the environment** (exports manifest for the module to consume)
6. **Delegates to IaC engine** (Pulumi or Terraform/OpenTofu)
7. **Streams output** to the developer

**Core commands:**

```bash
# Validate a manifest (optional but recommended)
planton validate-manifest postgres.yaml

# Deploy with Pulumi
planton pulumi update --manifest postgres.yaml --stack org/project/env

# Deploy with Terraform/OpenTofu
planton tofu apply --manifest postgres.yaml

# Override specific values (useful for CI/CD)
planton pulumi update \
  --manifest postgres.yaml \
  --set spec.container.cpu=500m \
  --stack org/project/env
```

---

### The Component Concept

A **component** is a complete, production-ready package for deploying a specific type of infrastructure or application. Think of it as a "recipe" that includes everything needed to deploy that resource.

#### What's in a Component?

The authoritative, gate-enforced component anatomy — the versioned contract under the version directory, the living module set, docs, presets, e2e profile, and the verified fact-sheets (`cost.yaml`, `controls.yaml`, `iac/permissions.yaml`) — is defined once, in [component.md](component.md) (see "Folder Structure" and "Verified Fact-Sheets" in its Ideal State Checklist). This overview deliberately does not duplicate that tree: the anatomy is machine-enforced by CI, and `component.md` is its single written home.

In one sentence: a component is the versioned Protobuf contract, both IaC engines, layered documentation, ready-to-deploy presets, an e2e profile, and machine-checked cost/controls/permissions data — everything a surface needs to deploy it and to answer what it costs, what it enforces, and what its runner needs.

#### Categories of Components

**1. Kubernetes Components**

Deploy applications and addons to any Kubernetes cluster:
- `KubernetesPostgres` - PostgreSQL with operator (CloudNativePG)
- `KubernetesValkey` - Valkey (Redis-compatible) workload
- `KubernetesKafka` - Apache Kafka with Strimzi operator
- `KubernetesCertManager` - cert-manager addon
- `KubernetesDeployment` - Containerized applications

**2. Cloud Provider Managed Services**

Deploy managed services on cloud providers:
- **AWS:** `AwsRdsInstance`, `AwsRdsCluster`, `AwsEksCluster`, `AwsS3Bucket`, `AwsAlb`
- **GCP:** `GcpCloudSql`, `GcpGkeCluster`, `GcpCloudRun`
- **Azure:** `AzureAksCluster`, `AzureStorageAccount`

**3. SaaS Platform Integrations**

Provision and manage third-party SaaS platforms:
- `Auth0Client` - Auth0 application client
- `Auth0Connection` - Auth0 identity connection
- `OpenFgaStore` - OpenFGA authorization store

---

## Repository Structure

```
planton/
├── catalog/                     # The multi-cloud component catalog
│   ├── <provider>/<kind>/       # One folder per component (aws, gcp, azure,
│   │                            #   kubernetes, digitalocean, cloudflare,
│   │                            #   auth0, openfga)
│   ├── _compliance/             # Central control catalog + framework crosswalks
│   ├── _pricing/                # Price books, derivations, generated estimates
│   ├── _patterns/               # Multi-component architecture patterns
│   └── _docs/                   # Catalog-wide guides and conventions
├── shared/                      # Shared types and enums
│   └── cloudresourcekind/       # Registry of all components (cloud_resource_kind.proto)
├── charts/                      # Ready-made infra charts composed from components
├── cmd/planton/                 # The open-source CLI and IaC engine (Go)
├── pkg/                         # Go libraries (generators, gates, engines)
├── operator/                    # The Planton Kubernetes operator (own Go module; self-hosted installs)
├── helm/                        # Helm charts that install Planton itself (operator, umbrella, runner)
├── skills/                      # Agent skills (incl. the multi-cloud-catalog skill)
├── site/                        # planton.ai website, docs, and blog
├── architecture/                # Architecture documentation
│   ├── README.md                # This file
│   └── component.md             # Component ideal state (the anatomy's one home)
├── _rules/                      # AI workflow rules (component forge/audit/update, docs, protos)
├── buf.yaml                     # Buf configuration
├── Makefile                     # Build automation
└── README.md                    # Project README
```

### Key Directories Explained

#### `/catalog`

The component catalog, organized by provider. Each component's root carries its IaC modules (`iac/pulumi`, `iac/tf`), documentation (`README.md`, `catalog.md`, `GUIDE.md`), presets, e2e profile, and verified fact-sheets (`cost.yaml`, `controls.yaml`, `iac/permissions.yaml`); the version directory carries only the versioned contract (api.proto, spec.proto, outputs.proto, spec_test.go, generated reference.md). The catalog root's central trees hold what spans components: `_compliance/` (the control catalog and the HIPAA/SOC 2/FedRAMP/CIS crosswalks) and `_pricing/` (pinned price books, derivation rules, and generated estimates).

#### `/architecture`

High-level architecture documentation:
- **README.md** (this file): Complete architecture overview
- **component.md**: Ideal state definition for components

#### `/_rules/component`

AI workflow rules for managing components:
- **forge**: Create new components from scratch (step-by-step flow)
- **audit**: Assess component completeness
- **update**: Enhance existing components
- **complete**: Automated workflow (audit + fill gaps + verify)
- **fix**: Targeted fixes with cascading updates
- **delete**: Safe component removal

#### `/cmd/planton`

Go implementation of the CLI:
- Command structure
- Manifest parsing
- Validation logic
- IaC engine integration
- Module caching

---

## Technology Stack

### API Layer

**Core Technologies:**
- **Protocol Buffers** - Schema definition language
- **buf CLI** - Proto compilation, linting, breaking change detection
- **buf-validate** - Field-level validation rules (based on protovalidate-go)
- **CEL (Common Expression Language)** - Complex validation logic

**Generated Artifacts:**
- Go stubs (for CLI and modules; committed alongside the protos)
- Java stubs (compiled as a build-time gate to catch invalid generated Java early; not committed)

Code generators run locally as pinned plugins from the git-ignored `apis/.tools/`
directory, provisioned automatically by `make proto-tools`. buf orchestrates the
generation (config, managed mode); the generator binaries themselves execute on
the developer's machine, keeping stub generation deterministic and independent
of any hosted code-generation service.

**Validation Flow:**
```
YAML Manifest → Parse → Unmarshal to Proto → Validate Rules → Deploy or Error
```

### IaC Layer

**Pulumi Stack:**
- **Language:** Go (default), Python/TypeScript supported
- **Providers:** AWS, GCP, Azure, Kubernetes, and 150+ others
- **State:** Supports local, S3, GCS, Azure Blob, Pulumi Cloud backends
- **Philosophy:** Simple, Terraform-like structure for adoption

**Terraform Stack:**
- **Language:** HCL (HashiCorp Configuration Language)
- **Providers:** AWS, GCP, Azure, Kubernetes, and 3000+ others
- **State:** Supports local, S3, GCS, Azure Storage backends
- **Philosophy:** Idiomatic Terraform module structure

**Feature Parity:**
- Every Pulumi module has a corresponding Terraform module (a kind that declares fewer engines ships only the modules they run)
- Same functionality, same defaults, same behavior
- Users choose based on team preference

### CLI Layer

**Implementation:**
- **Language:** Go
- **Binary Distribution:** GitHub Release assets (self-updating via `planton upgrade`)
- **Configuration:** Environment variables and flags
- **Module Caching:** released artifacts under `~/.planton/terraform/modules/` and `~/.planton/pulumi/binaries/`, keyed by release; staging git clone under `~/.planton/staging/` as fallback

**Dependencies:**
- Git (required only for the staging fallback and `--local-module`)
- Pulumi CLI (for `pulumi` commands)
- Terraform/OpenTofu CLI (for `tofu` commands)

### Build System

**Make Targets:**
```makefile
make protos        # Generate proto stubs
make build         # Build CLI and run tests
make test          # Run all tests
make install       # Install CLI locally
```

**Proto Generation:**
- `buf generate` creates Go, Java, Python, TypeScript stubs
- Automated by `make protos`
- Version-controlled in repository

---

## Component Lifecycle

Planton provides a sophisticated lifecycle management system for components. This system ensures that all components are consistently high-quality, well-documented, and production-ready.

### The Six Operations

```
┌─────────────────────────────────────────────────────────────┐
│               Component Lifecycle                │
└─────────────────────────────────────────────────────────────┘
         │
         ├─► 🔨 FORGE      Create new component (95-100% complete)
         │
         ├─► 🔍 AUDIT      Assess completeness (9 categories, weighted scoring)
         │
         ├─► 🔄 UPDATE     Enhance existing (6 scenarios: fill-gaps, proto-changed, etc.)
         │
         ├─► ✨ COMPLETE   Auto-improve (audit + fill gaps + verify)
         │
         ├─► 🔧 FIX        Targeted fixes (with cascading updates)
         │
         └─► 🗑️  DELETE     Safe removal (dry-run, backup, confirmation)
```

### 1. Forge: Create New Components

**Purpose:** Bootstrap complete, production-ready components from scratch.

**What It Creates:**
- ✅ Proto API definitions (4 files with validations)
- ✅ Validation tests (spec_test.go)
- ✅ IaC modules (both Pulumi and Terraform)
- ✅ Documentation (user-facing, research, technical)
- ✅ Supporting files (test manifests, debug scripts)
- ✅ Registry entry (cloud_resource_kind.proto)

**Result:** 95-100% completion score

**Workflow:** 21-step automated process organized in 7 phases:
1. **Proto API** (6 rules): spec.proto, validations, tests, outputs, api, input
2. **Registration** (2 rules): cloud_resource_kind enum, proto stubs
3. **Documentation** (2 rules): user-facing docs, research docs
4. **Test Infrastructure** (1 rule): hack manifest
5. **Pulumi** (5 rules): module, entrypoint, e2e, docs, overview
6. **Terraform** (3 rules): module, e2e, docs
7. **Validation** (2 rules): build verification, test execution

**Example:**
```bash
@forge-planton-component CloudflareD1Database --provider cloudflare
```

### 2. Audit: Assess Component Completeness

**Purpose:** Evaluate components against the ideal state and generate actionable completion reports.

**What It Checks:** 9 categories with weighted scoring:
1. Cloud Resource Registry (4.44%)
2. Folder Structure (4.44%)
3. Protobuf API Definitions (17.76%)
4. IaC Modules - Pulumi (13.32%)
5. IaC Modules - Terraform (4.44%)
6. Documentation - Research (13.34%)
7. Documentation - User-Facing (13.33%)
8. Supporting Files (13.33%)
9. Nice to Have Items (20%)

**Scoring System:**
- **100%** = Perfect, production-ready
- **95-99%** = Excellent, minor polish possible
- **80-94%** = Good, some improvements recommended
- **60-79%** = Fair, significant work needed
- **<60%** = Poor, major work required

**Report Output:**
- Overall completion percentage
- Category-by-category breakdown
- Quick wins (easy improvements)
- Critical gaps (blocking issues)
- Prioritized recommendations
- Reports are delivered in the session that runs the audit -- never saved
  into the component (the component file set is closed; the anatomy
  conformance gate enforces it)

**Example:**
```bash
@audit-planton-component CloudflareD1Database
```

### 3. Update: Enhance Existing Components

**Purpose:** Improve existing components by filling gaps, adding features, refreshing docs, or fixing issues.

**Six Update Scenarios:**
1. **Fill Gaps** - Audit-driven completion (missing files, incomplete docs)
2. **Proto Changed** - Propagate schema changes through all artifacts
3. **Refresh Docs** - Update documentation to match current state
4. **Update IaC** - Modify deployment logic in Pulumi/Terraform
5. **Fix Issue** - Targeted fixes with explanation
6. **Auto** - Intelligent scenario detection

**Safety Features:**
- Dry-run mode (preview changes)
- Backup creation (safety net)
- Validation checkpoints (verify after changes)
- Automatic retry (up to 3 times with fixes)
- Conflict detection

**Example:**
```bash
@update-planton-component CloudflareD1Database --scenario fill-gaps
```

### 4. Complete: Auto-Improve Workflow

**Purpose:** One-command workflow that audits and automatically fills all gaps to reach target completion score (default 95%).

**Three-Step Automated Workflow:**
1. **Audit** - Assess current state and identify all gaps
2. **Fill Gaps** - Automatically run update --fill-gaps
3. **Verify** - Re-audit to confirm improvement

**What Gets Filled:**
- Terraform module (if missing)
- Research documentation (if missing)
- User-facing docs (if incomplete)
- Examples (if missing/incomplete)
- Pulumi overview (if missing)
- Supporting files (manifests, debug scripts)

**Typical Results:**
- 40-60% starting → 95-98% (30-40 min)
- 60-80% starting → 95-98% (15-25 min)
- 80-94% starting → 95-100% (5-15 min)

**Example:**
```bash
@complete-planton-component CloudflareD1Database
```

### 5. Fix: Targeted Fixes with Cascading Updates

**Purpose:** Make targeted fixes to components and automatically propagate changes to all related artifacts.

**Core Philosophy:** Source code is the ultimate source of truth. Documentation describes code, code doesn't describe documentation.

**Six-Step Workflow:**
1. **Analyze** - Understand the fix needed, read current source
2. **Fix Source Code** - Make changes to proto, IaC, tests
3. **Propagate to Docs** - Update all documentation to match
4. **Validate Consistency** - Run 5 consistency checks
5. **Execute Tests** - Component tests, build, full suite
6. **Report** - Show what was fixed and propagated

**Five Consistency Checks:**
- Proto ↔ Terraform variables
- Proto ↔ Examples (examples must validate)
- Pulumi ↔ Terraform (feature parity)
- Validations ↔ Tests (every rule tested)
- Documentation ↔ Implementation (docs match reality)

**Example:**
```bash
@fix-planton-component GcpCertManagerCert \
  --explain "primaryDomainName validation should allow wildcards like *.example.com"
```

### 6. Delete: Safe Component Removal

**Purpose:** Completely remove components with safety features to prevent accidents.

**Safety Features:**
- 🔍 Dry-run mode (preview deletion)
- 💾 Automatic backup (timestamped)
- 🔎 Reference check (warns if referenced)
- ✋ Confirmation required (must type component name)
- 📋 Detailed report (shows what was deleted)

**What Gets Deleted:**
- Component folder (all files)
- Registry entry (cloud_resource_kind.proto enum)
- Generated proto stubs (regenerated after)

**Example:**
```bash
# Preview
@delete-planton-component ObsoleteComponent --dry-run

# Delete with backup
@delete-planton-component ObsoleteComponent --backup
```

### Ideal State Definition

All lifecycle operations reference a single source of truth: **`architecture/component.md`**

This document defines:
- Complete checklist of required artifacts
- Quality standards for each category
- Scoring weights and rationale
- Provider parity standard (100% of the provider's configurable surface at the pinned version)
- Examples of complete components

**Key Insight:** The ideal state is **intentionally pragmatic**—it focuses effort on the highest-leverage work. Not every component needs every possible artifact, but every production component should reach 95%+ completion, modeling the full configurable surface of its mapped provider resources at the pinned provider version.

---

## API Design Philosophy

### Kubernetes Resource Model (KRM)

Planton adopts the Kubernetes Resource Model as its API structure:

```yaml
apiVersion: <provider>.planton.dev/<version>
kind: <ComponentType>
metadata:
  name: <resource-name>
  org: <organization>
  env: <environment>
  labels: {}
  tags: {}
spec:
  # Configuration
status:
  # Output (read-only)
```

**Why KRM?**
- Familiar to millions of developers
- Clear separation: metadata vs spec vs status
- Standard conventions (apiVersion, kind)
- Extensible with labels and annotations

### Protocol Buffers Over Go Structs

**Kubernetes uses Go structs.** Planton uses Protocol Buffers. Why?

**Language Neutrality:**
```
proto definitions → buf generate → Go/Java/Python/TypeScript stubs
```

This enables:
- CLI in Go
- Backend services in Java
- Scripts in Python
- Web UIs in TypeScript
- All consuming the same schema

**Documentation as Code:**

Proto definitions published to Buf Schema Registry become navigable, searchable documentation automatically. No manual doc generation required.

### Validation Strategy

**Three Layers:**

1. **Schema-level validation** (in proto):
```protobuf
message PostgresKubernetesSpec {
  string cpu = 1 [(buf.validate.field).string.pattern = "^[0-9]+m$"];
  int32 replicas = 2 [(buf.validate.field).int32 = {gte: 1, lte: 10}];
}
```

2. **Pre-deployment validation** (CLI):
```bash
planton validate-manifest config.yaml
# Catches errors before calling cloud APIs
```

3. **Cloud provider validation** (during deployment):
- Final validation by actual cloud provider APIs
- Catches provider-specific constraints

**Result:** 90%+ of errors caught before making any cloud API calls.

### The Provider Parity Standard

**Coverage is the provider's full configurable surface, never a trimmed subset.**

Every component models 100% of the configurable arguments of the provider resources it maps to, at the pinned provider version. Partial coverage is not a design choice: where something is not covered, that is an explicitly recorded decision with a reason (deprecated or superseded surface only), never a silent omission.

**Example: PostgreSQL on Kubernetes**

**Most-reached (the common path):**
- Replicas (1 or 3)
- Storage size (10Gi, 50Gi, 100Gi)
- CPU and memory limits
- Database name and credentials

**Advanced surface (equally modeled, with sensible defaults):**
- Custom WAL configuration
- Replication topologies
- Fine-grained operator settings

**Planton's approach:**
- Modules model the full upstream surface at the pinned version
- Sensible defaults keep the common path simple
- Escape hatches (e.g. Helm values overrides) are safety valves on top of a fully modeled spec, never a substitute for modeling
- Deprecated or superseded surface may be excluded, always with a recorded reason -- omission is a decision, and every decision is recorded

### Provider-Specific vs. Generic

**Planton is intentionally NOT a cloud abstraction layer.**

**Example: Postgres Deployment**

Three different components:
- `PostgresKubernetes` - Deploy to any K8s cluster (uses CloudNativePG operator)
- `AwsRdsInstance` - Deploy to AWS RDS (managed service)
- `GcpCloudSql` - Deploy to GCP Cloud SQL (managed service)

Each has provider-specific configuration:
- AWS RDS: `instance_class`, `security_group_ids`, `db_subnet_group`
- GCP Cloud SQL: `tier`, `authorized_networks`, `database_flags`
- Kubernetes: `replicas`, `storage_class`, `resources`

**What's consistent:**
- YAML structure (KRM)
- Validation approach (proto-validate)
- CLI commands (`planton pulumi update`, `planton tofu apply`)
- Deployment workflow (validate → deploy → outputs)

**What's different:**
- Configuration options (provider-specific)
- Deployment target (AWS vs GCP vs K8s)

This preserves cloud-specific power while providing experience consistency.

---

## IaC Module Design

### Dual IaC Engine Support

Every component has **both** Pulumi and Terraform implementations with **feature parity**.

**Why both?**

Different organizations have different investments:
- Terraform: Mature, large ecosystem, familiar HCL syntax
- Pulumi: Real programming languages, type safety, easier testing

**Feature parity enforcement:**
- Audit checks for both modules
- Fix operations maintain parity
- Update operations apply to both
- Same defaults, same behavior

### Module Structure

#### Pulumi Module (Go)

```
iac/pulumi/
├── main.go              # Entry point, parses manifest
├── locals.go            # Local transformations
├── resources.go         # Resource definitions
├── outputs.go           # Stack outputs
├── go.mod               # Go dependencies
├── Pulumi.yaml          # Pulumi project config
└── docs/
    └── README.md        # Architecture overview
```

**Key Files:**
- **main.go**: Parses manifest from environment variable, calls resource creation
- **locals.go**: Transforms manifest into Pulumi-friendly structures
- **resources.go**: Creates cloud resources using Pulumi SDKs
- **outputs.go**: Exports outputs (connection strings, IDs, etc.)

#### Terraform Module (HCL)

```
iac/terraform/
├── main.tf              # Resource definitions
├── variables.tf         # Input variables
├── outputs.tf           # Output values
├── locals.tf            # Local transformations
├── versions.tf          # Provider versions
└── README.md            # Usage documentation
```

**Key Files:**
- **main.tf**: Resource definitions
- **variables.tf**: Inputs (populated by CLI from manifest)
- **outputs.tf**: Outputs (connection strings, IDs, etc.)
- **locals.tf**: Transformations and computed values

### Design Principles

#### 1. Deliberately Simple

**Pulumi modules are written like Terraform modules.**

Avoid:
- ❌ Deep class hierarchies
- ❌ Aggressive abstraction
- ❌ Complex DRY patterns
- ❌ Advanced language features

Prefer:
- ✅ Flat structure
- ✅ Explicit code
- ✅ Simple transformations
- ✅ Minimal dependencies

**Why?** Adoption. A Terraform engineer should be able to read a Pulumi module and immediately understand the flow.

#### 2. Terraform-Like File Names

Even in Pulumi (Go), use familiar names:
- `main.go` ≈ `main.tf`
- `locals.go` ≈ `locals.tf`
- `outputs.go` ≈ `outputs.tf`

This reduces cognitive friction for engineers familiar with Terraform.

#### 3. Typed Stack Input, Not Ad-Hoc Parsing

The CLI hands each module a **stack input** — a typed document carrying the
manifest (`target`) plus provisioner and provider configuration — and the
module loads it through generated types, never by hand-parsing YAML.

**Pulumi (Go):** the module calls the SDK loader, which reads the
`planton:stack-input` Pulumi config key or, when unset, the
`STACK_INPUT_YAML` (inline content) / `STACK_INPUT_YAML_FILE` (file path)
environment variables — and validates the spec before any resource is
created:

```go
stackInput := &postgreskubernetesv1.PostgresKubernetesStackInput{}
if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
    return err
}
```

**Terraform/OpenTofu (HCL):** the CLI generates a `.tfvars` file from the
manifest (spec fields become variables), so the module consumes plain
Terraform variables — no environment parsing at all.

#### 4. Battle-Tested Defaults

Modules include production-ready defaults:
- Multi-AZ for databases
- Encryption at rest
- Secure networking
- Backup enabled
- Monitoring configured

Users override only what they need to customize.

### Customization Patterns

**For Individual Developers:**
- Use default modules without modification
- Override values via manifest or CLI flags

**For Platform Engineers:**
- Fork default modules to private repos
- Customize for organizational standards
- Point CLI to custom modules via flags

**For Advanced Users:**
- Rewrite modules in different languages (Python, TypeScript)
- Use auto-generated proto SDKs
- Build entirely custom implementations

---

## CLI Architecture

### Command Structure

```
planton
├── validate-manifest   # Validate manifest (proto-validate)
├── validate-outputs    # Validate a module's output transformation overrides
├── apply / init / refresh  # Auto-routed: picks the engine from the manifest's provisioner annotation
├── pulumi              # Pulumi commands
│   ├── init            # Initialize the stack
│   ├── preview         # Preview changes
│   ├── update          # Deploy/update
│   ├── refresh         # Sync state with reality
│   └── destroy         # Tear down
├── tofu                # OpenTofu commands (terraform group mirrors it)
│   ├── init            # Initialize backend/providers
│   ├── plan            # Preview changes
│   ├── apply           # Deploy/update
│   ├── refresh         # Sync state with reality
│   └── destroy         # Tear down
└── version             # Show version
```

### Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   User runs CLI command                     │
│       planton pulumi update --manifest config.yaml          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Parse Manifest                                           │
│    - Read YAML file                                         │
│    - Extract apiVersion, kind                               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Validate (proto-validate)                                │
│    - Unmarshal to proto                                     │
│    - Run validation rules                                   │
│    - Exit on error                                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Map kind → Module (kind registry)                        │
│    - Provider, component, and declared version all derive   │
│      from the kind's registry metadata — no lookup tables   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Resolve Module (first match wins)                        │
│    - Explicit --module-dir (validated, fails loudly)        │
│    - Current directory, when it contains a valid module     │
│    - Released artifact from downloads.planton.dev (cached)  │
│    - Staging clone of the git repo (fallback)               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Build Inputs                                             │
│    - Tofu/Terraform: generate .tfvars from the manifest     │
│    - Pulumi: write the stack input, point the module at it  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Delegate to IaC Engine                                   │
│    - Pulumi: automation API drives the module               │
│    - Tofu/Terraform: exec plan/apply against the module     │
│    - Stream output to user                                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Return to User                                           │
│    - Success: Show outputs                                  │
│    - Failure: Show errors                                   │
└─────────────────────────────────────────────────────────────┘
```

### Module Resolution

There is no hand-maintained kind→module table. The kind registry (proto
metadata compiled into the CLI) declares each kind's provider, component
name, and served API version, and every module path or artifact URL derives
from those three facts.

**Resolution order (first match wins):**

1. **`--module-dir <path>`** — an explicitly chosen local module. The
   directory must actually be a module (`*.tf` files for tofu/terraform,
   `Pulumi.yaml` for pulumi); anything else fails with an error naming the
   path and what was missing. An explicit choice never falls through
   silently.
2. **The current directory** — probed as a convenience when no explicit
   choice was made; running the CLI from inside a module just works.
3. **Released artifact download (fast path)** — terraform module zips and
   pre-built pulumi binaries from `downloads.planton.dev`, keyed by release
   tag, component, and the kind's declared version. Released CLIs use this
   path; dev builds skip it.
4. **Staging clone (fallback)** — a git checkout of the modules repository
   under `~/.planton/staging/`, used when no released artifact is available.

**Custom modules:**

```bash
# Start from the official module: eject a customizable copy you own
# (a generated CONTRACT.md inside explains the contract it must keep)
planton module eject AwsS3Bucket --provisioner tofu --output-dir ./my-custom-module

# Prove a customization still conforms to the kind's contract
planton module verify --kind AwsS3Bucket --module-dir ./my-custom-module

# Point a deployment at any local module directory:
planton tofu apply --manifest config.yaml --module-dir ./my-custom-module

# Or run against a local checkout/fork of the modules repository:
planton pulumi update --manifest config.yaml --local-module
# (the checkout is found via --planton-git-repo or $PLANTON_GIT_REPO)
```

### Module Caching

**Locations:**

| Content | Path |
|---------|------|
| Terraform module zips | `~/.planton/terraform/modules/{release}/` |
| Pulumi pre-built binaries | `~/.planton/pulumi/binaries/{release}/` |
| Staging git clone | `~/.planton/staging/` |

**Benefits:**
- Fast repeat deployments (artifacts download once per release)
- Cache keyed by release version — upgrading the CLI never serves stale modules

### Validation Integration

**CLI embeds proto-validate:**

```go
import (
  "github.com/bufbuild/protovalidate-go"
  pb "github.com/plantonhq/planton/apis/..."
)

func validateManifest(manifestYaml string) error {
  // Unmarshal YAML to proto
  config := &pb.PostgresKubernetes{}
  yaml.Unmarshal([]byte(manifestYaml), config)
  
  // Validate
  validator := protovalidate.New()
  err := validator.Validate(config)
  if err != nil {
    return fmt.Errorf("validation failed: %w", err)
  }
  
  return nil
}
```

**Result:** Validation happens **before** cloning modules, **before** calling cloud APIs. Fast feedback.

---

## Development Workflows

### Adding a New Component

**High-Level Process:**

```
Research → Forge → Audit → (Complete) → Deploy & Test → Commit
```

**Detailed Steps:**

1. **Research Phase**
   - Understand the resource (AWS RDS, GKE Cluster, etc.)
   - Research deployment methods (manual, Terraform, Pulumi)
   - Map the full configurable surface of the mapped provider resources at the pinned provider version (the parity target)
   - Document findings

2. **Forge Phase**
   ```bash
   @forge-planton-component CloudflareD1Database --provider cloudflare
   ```
   - Creates complete component (95-100% complete)
   - Proto definitions with validations
   - Both Pulumi and Terraform modules
   - Documentation and examples
   - Test manifests

3. **Audit Phase**
   ```bash
   @audit-planton-component CloudflareD1Database
   ```
   - Verify forge created everything
   - Check completion score (should be 95-100%)
   - Identify any gaps

4. **Complete Phase (if needed)**
   ```bash
   @complete-planton-component CloudflareD1Database
   ```
   - Fill any remaining gaps
   - Re-audit to verify 100%

5. **Local Testing**
   ```bash
   cd catalog/cloudflare/cloudflared1database/iac/pulumi
   planton pulumi update --manifest ../../e2e/manifest.yaml --module-dir .
   ```
   - Test Pulumi module locally (the CLI builds the stack input and drives the module)
   - Repeat for the Terraform module with `planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir ../tf`

6. **Validation**
   ```bash
   make protos  # Regenerate stubs
   make build   # Verify compilation
   make test    # Run all tests
   ```

7. **Commit**
   ```bash
   git add -A
   git commit -m "feat(cloudflare): add CloudflareD1Database component"
   git push origin main
   ```

### Updating an Existing Component

**Scenario 1: Adding a Field to spec.proto**

```bash
# 1. Edit spec.proto
vim catalog/cloudflare/cloudflared1database/v1alpha1/spec.proto

# 2. Add field with validation
# message Spec {
#   int32 time_travel_retention_days = 5 [(buf.validate.field).int32 = {gte: 1, lte: 30}];
# }

# 3. Propagate changes
@update-planton-component CloudflareD1Database --scenario proto-changed

# This will:
# - Regenerate proto stubs
# - Update Pulumi module to use new field
# - Update Terraform module to use new field
# - Update documentation and examples
# - Add test for new validation rule

# 4. Verify
@audit-planton-component CloudflareD1Database
```

**Scenario 2: Fixing a Bug**

```bash
@fix-planton-component GcpCertManagerCert \
  --explain "primaryDomainName validation should allow wildcards like *.example.com"

# This will:
# - Update validation rule in spec.proto
# - Update spec_test.go with new test case
# - Update README.md to document wildcard support
# - Run consistency checks
# - Execute tests
```

**Scenario 3: Refreshing Outdated Docs**

```bash
@update-planton-component PostgresKubernetes --scenario refresh-docs

# This will:
# - Read current source code (proto, IaC)
# - Regenerate README.md to match current behavior
# - Update presets with current field names
# - Verify presets validate
```

### Quality Assurance Workflow

**Pre-Commit Checklist:**

```bash
# 1. Audit modified components
@audit-planton-component <ComponentName>

# 2. Ensure score ≥ 95%
# If score dropped, investigate and fix

# 3. Run validation
make protos
make build
make test

# 4. If all pass, commit
git add -A
git commit -m "feat: enhance <ComponentName>"
```

### Batch Improvement Workflow

**Improving multiple components to production-ready state:**

```bash
# List of components to improve
components=(
  "CloudflareD1Database"
  "Auth0Client"
  "PostgresKubernetes"
)

for component in "${components[@]}"; do
  echo "Processing $component..."
  
  # Auto-improve to 95%
  @complete-planton-component "$component"
  
  # Verify
  @audit-planton-component "$component"
done

# Commit all improvements
git add -A
git commit -m "chore: improve component completeness to 95%+"
git push origin main
```

---

## Extension Patterns

### Pattern 1: Custom Modules for Internal Standards

**Scenario:** Your organization has specific security policies (always encrypt, always multi-AZ, specific tagging).

**Approach:**
1. Eject the official module into a directory you own
2. Add organizational defaults
3. Prove the module still conforms to the kind's contract
4. Deploy with the CLI pointed at your module

**Example:**

```bash
# Eject a customizable copy of the official module (the generated
# CONTRACT.md inside explains the contract it must keep honoring)
planton module eject AwsRdsInstance --provisioner tofu --output-dir ./my-rds-module
# Edit to add organizational defaults, then version it in your own repository

# Prove the module still conforms — the input surface against the kind's
# schema, the outputs contract, and (when tofu is on PATH) engine validation
planton module verify --kind AwsRdsInstance --module-dir ./my-rds-module

# Deploy with your module (same manifest, your deployment logic)
planton tofu apply \
  --manifest rds.yaml \
  --module-dir ./my-rds-module
```

For pulumi modules, `module eject` additionally makes the copy a standalone
Go module: it generates a `go.mod` pinned to the release the module was
ejected from and rewrites the module's self-imports onto a module path you
choose with `--go-module` — the copy compiles and deploys outside this
repository from the moment it lands.

When a custom module's raw outputs use different names than the official
one, ship an output transformation alongside it — a declarative
`output_transform.yaml` mapping raw output names to the component's outputs
schema, or a `transform-outputs` executable for logic a mapping cannot
express. The CLI discovers either automatically in the module directory, and
`planton module verify` checks it as part of the full contract check
(`planton validate-outputs` remains the focused low-level tool; add
`--sample-outputs` to either for a full dry-run).

**Working directly in a repo checkout instead:** pass `--local-module` and
point `--planton-git-repo` (or `$PLANTON_GIT_REPO`) at the checkout — the CLI
derives the module directory for the manifest's kind from the tree.

### Pattern 2: Reusing APIs for Internal Tooling

**Scenario:** You're building an internal self-service portal where developers request databases.

**Approach:**
1. Import Planton proto SDKs
2. Use strongly-typed APIs in your tool
3. Generate manifests from user input
4. Call Planton CLI programmatically

**Example (Python):**

```python
from planton.apis.dev.planton.kubernetes.workload.postgreskubernetes.v1 import api_pb2
import yaml
import subprocess

def create_database_from_ui(user_input):
    # Create strongly-typed config
    config = api_pb2.PostgresKubernetes()
    config.api_version = "kubernetes.planton.dev/v1alpha1"
    config.kind = "PostgresKubernetes"
    config.metadata.name = user_input["name"]
    config.metadata.org = user_input["org"]
    config.metadata.env = user_input["env"]
    config.spec.container.replicas = user_input["replicas"]
    config.spec.container.resources.limits.cpu = user_input["cpu"]
    config.spec.container.resources.limits.memory = user_input["memory"]
    
    # Convert to YAML
    manifest_yaml = yaml.dump(config)
    
    # Write to file
    with open("manifest.yaml", "w") as f:
        f.write(manifest_yaml)
    
    # Call CLI
    subprocess.run([
        "planton", "pulumi", "up",
        "--manifest", "manifest.yaml",
        "--stack", f"{user_input['org']}/{user_input['project']}/{user_input['env']}"
    ])
```

**Benefits:**
- Don't reinvent schemas
- Get validation for free
- Type safety in your language
- Reuse Planton deployment logic

### Pattern 3: CI/CD Integration

**Scenario:** Automate deployments via CI/CD pipelines.

**Approach:**
1. Store manifests in git
2. Run validation on PR
3. Deploy on merge to main

**Example (GitHub Actions):**

```yaml
name: Deploy Infrastructure

on:
  push:
    branches: [main]
    paths:
      - 'infrastructure/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install Planton
        run: |
          brew install plantonhq/tap/planton
      
      - name: Validate Manifests
        run: |
          for manifest in infrastructure/*.yaml; do
            planton validate-manifest $manifest
          done
      
      - name: Deploy
        run: |
          for manifest in infrastructure/*.yaml; do
            planton pulumi update \
              --manifest $manifest \
              --stack prod \
              --yes  # Non-interactive
          done
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

### Pattern 4: Multi-Language Custom Modules

**Scenario:** Your team prefers Python/TypeScript over Go for IaC.

**Approach:**
1. Write a Pulumi module in any language Pulumi supports
2. Read the stack input the CLI hands every pulumi module — the
   `planton:stack-input` Pulumi config key, or the `STACK_INPUT_YAML` /
   `STACK_INPUT_YAML_FILE` environment variables when the config key is
   unset; the manifest is its `target` field
3. Deploy with `--module-dir` pointed at your module

**Example (Python Pulumi module):**

```python
# __main__.py — deployed with:
#   planton pulumi update --manifest postgres.yaml --module-dir ./my-postgres-module
import os
import yaml
import pulumi
import pulumi_kubernetes as k8s

# Load the stack input; the manifest is its `target` field
config = pulumi.Config("planton")
stack_input_yaml = config.get("stack-input") or os.getenv("STACK_INPUT_YAML")
if not stack_input_yaml:
    with open(os.environ["STACK_INPUT_YAML_FILE"]) as f:
        stack_input_yaml = f.read()

stack_input = yaml.safe_load(stack_input_yaml)
target = stack_input["target"]

# Deploy using Pulumi (Python style)
namespace = k8s.core.v1.Namespace(
    target["metadata"]["name"],
    metadata=k8s.meta.v1.ObjectMetaArgs(
        name=target["metadata"]["name"]
    )
)

postgres = k8s.helm.v3.Release(
    "postgres",
    chart="postgresql",
    repository_opts=k8s.helm.v3.RepositoryOptsArgs(
        repo="https://charts.bitnami.com/bitnami"
    ),
    namespace=namespace.metadata.name,
    values={
        "replicas": target["spec"]["container"]["replicas"],
        "resources": {
            "limits": {
                "cpu": target["spec"]["container"]["resources"]["limits"]["cpu"],
                "memory": target["spec"]["container"]["resources"]["limits"]["memory"]
            }
        }
    }
)
```

The manifest inside the stack input is already validated by the CLI before
the module runs. For Terraform-preferring teams the same pattern needs no
code at all: the CLI generates `.tfvars` from the manifest, so any HCL module
whose variables match the component's generated `variables.tf` works with
`--module-dir`.

---

## Contributing

### Ways to Contribute

**1. Add New Components**

Use the Forge workflow to create new cloud resources:
```bash
@forge-planton-component <ComponentName> --provider <provider>
```

Submit PRs with:
- Protobuf APIs with validation rules
- Both Pulumi and Terraform modules
- Documentation and examples
- Test manifests

**2. Improve Existing Components**

Use the Complete workflow to bring components to 100%:
```bash
@complete-planton-component <ComponentName>
```

Submit PRs with:
- Filled gaps (missing docs, Terraform modules, etc.)
- Updated documentation
- Additional examples
- Bug fixes

**3. Fix Issues**

Use the Fix workflow for targeted improvements:
```bash
@fix-planton-component <ComponentName> --explain "<fix description>"
```

Submit PRs with:
- Source code fixes
- Propagated documentation updates
- Test coverage
- Consistency validation

**4. Build Ecosystem Tools**

- Create language-specific helpers
- Build UI/dashboard projects
- Develop CI/CD integrations
- Write tutorials and guides

### Contribution Guidelines

**Pull Requests:**
- One component per PR (keeps reviews focused)
- Run audit before submitting (ensure ≥95% completion)
- Include test results (`make build && make test`)
- Update documentation

**Commit Messages:**
Follow Conventional Commits:
```
feat(aws): add AwsLambdaFunction component
fix(gcp): correct validation rule for GcpStorageBucket
docs(kubernetes): update PostgresKubernetes examples
chore: improve component completeness to 95%+
```

**Code Quality:**
- Run `make protos` after proto changes
- Run `make build` to verify compilation
- Run `make test` to verify tests pass
- Use audit to verify completeness

**Documentation:**
- User-facing README.md (what and how)
- GUIDE.md (why and design rationale)
- The validated example manifest at e2e/manifest.yaml (a real, deployable use case)
- IaC module docs (architecture overview)

### Development Environment Setup

**Prerequisites:**
```bash
# Install required tools
brew install go
brew install buf
brew install pulumi
brew install opentofu
brew install make

# Clone repository
git clone https://github.com/plantonhq/planton.git
cd planton
```

**Build and Test:**
```bash
# Generate proto stubs
make protos

# Build CLI and run tests
make build

# Run tests
make test

# Install CLI locally
make install
```

**Verify Installation:**
```bash
planton version
```

### Testing Your Changes

**1. Unit Tests**

Test validation rules:
```bash
cd catalog/cloudflare/cloudflared1database/v1alpha1
go test -v
```

**2. Integration Tests**

Test IaC modules locally:
```bash
cd catalog/cloudflare/cloudflared1database/iac/pulumi
planton pulumi preview --manifest ../../e2e/manifest.yaml --module-dir .
```

**3. End-to-End Tests**

Test full deployment workflow:
```bash
planton validate-manifest test.yaml
planton pulumi update --manifest test.yaml --stack test
planton pulumi destroy --manifest test.yaml --stack test
```

### Getting Help

**Documentation:**
- Architecture overview (this file)
- Component ideal state (`architecture/component.md`)
- Lifecycle operation READMEs (`_rules/component/*/README.md`)

**Examples:**
- Browse complete components (`catalog/`)
- Run audit on gold-standard components
- Compare incomplete vs complete components

**Community:**
- GitHub Discussions for questions and ideas
- GitHub Issues for bug reports and feature requests
- Documentation site for guides and references

---

## Summary

Planton is a multi-cloud deployment framework that provides **consistency without abstraction**. It brings the Kubernetes Resource Model philosophy to the entire cloud infrastructure landscape, offering:

**Core Value:**
- ✅ Standardized YAML manifests (KRM structure)
- ✅ Pre-deployment validation (proto-validate)
- ✅ Single CLI for all clouds (`planton`)
- ✅ Provider-specific power (no artificial abstraction)
- ✅ Dual IaC support (Pulumi and Terraform)
- ✅ Language-neutral APIs (Protocol Buffers)
- ✅ 700+ components (AWS, GCP, Azure, K8s, DigitalOcean, Cloudflare, Auth0, OpenFGA)

**Architecture:**
1. **APIs** - Proto definitions with validations (buf.build)
2. **IaC Modules** - Pulumi and Terraform (feature parity)
3. **CLI** - Go binary (Homebrew distribution)

**Lifecycle Management:**
1. **Forge** - Create components (95-100% complete)
2. **Audit** - Assess completeness (9 categories)
3. **Update** - Enhance components (6 scenarios)
4. **Complete** - Auto-improve workflow
5. **Fix** - Targeted fixes with propagation
6. **Delete** - Safe removal with backups

**Philosophy:**
- Consistency of experience, not abstraction of providers
- Provider parity (model the provider's full configurable surface at the pinned version, with sensible defaults)
- Deliberately simple IaC modules (adoption over perfection)
- Language neutrality (build tools in any language)
- Open source foundation (transparent, forkable, extendable)

**Use Cases:**
- Individual developers deploying infrastructure
- Teams standardizing multi-cloud deployments
- Platform engineers building internal developer platforms
- Organizations requiring consistent deployment workflows

**Getting Started:**
```bash
brew install plantonhq/tap/planton
planton version
```

**Next Steps:**
- Browse components in `catalog/`
- Read lifecycle management guides in `_rules/component/`
- Try deploying a component locally
- Contribute new components or improvements

---

**Ready to contribute?** Start with the Forge workflow to create a new component, or use Complete to improve an existing one to 100% quality!

