# Planton Architecture

**A multi-cloud deployment framework that brings Kubernetes-style consistency to infrastructure deployments across any cloud provider.**

---

## Table of Contents

- [What is Planton?](#what-is-planton)
- [Core Architecture](#core-architecture)
  - [The Three Pillars](#the-three-pillars)
  - [The Deployment Component Concept](#the-deployment-component-concept)
- [Repository Structure](#repository-structure)
- [Technology Stack](#technology-stack)
- [Deployment Component Lifecycle](#deployment-component-lifecycle)
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
│        Deployment Components (100+)            │
│  PostgresKubernetes | AwsRdsInstance | etc.   │
└────────────────────────────────────────────────┘
```

#### 1. APIs: Standardized Configuration Schema

**Technology:** Protocol Buffers  
**Inspiration:** Kubernetes Resource Model

Every deployment component follows the same structure:

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

Every deployment component has **both** a Pulumi module and a Terraform module. You choose which IaC engine to use.

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
planton validate --manifest postgres.yaml

# Deploy with Pulumi
planton pulumi up --manifest postgres.yaml --stack org/project/env

# Deploy with Terraform/OpenTofu
planton tofu apply --manifest postgres.yaml

# Override specific values (useful for CI/CD)
planton pulumi up \
  --manifest postgres.yaml \
  --set spec.container.cpu=500m \
  --stack org/project/env
```

---

### The Deployment Component Concept

A **deployment component** is a complete, production-ready package for deploying a specific type of infrastructure or application. Think of it as a "recipe" that includes everything needed to deploy that resource.

#### What's in a Deployment Component?

Every deployment component contains:

```
<provider>/<component>/v1/
├── api.proto                    # Main API definition (KRM structure)
├── spec.proto                   # Spec section (configuration options)
├── spec_test.go                 # Unit tests for validation rules
├── stack_input.proto            # Input to IaC modules
├── stack_outputs.proto          # Output from IaC modules
├── README.md                    # User-facing documentation
├── docs/
│   └── README.md                # Deep research and design rationale
└── iac/
    ├── pulumi/                  # Pulumi module
    │   ├── main.go
    │   ├── locals.go
    │   ├── outputs.go
    │   └── docs/
    │       └── README.md        # Pulumi architecture overview
    ├── terraform/               # Terraform module
    │   ├── main.tf
    │   ├── variables.tf
    │   ├── outputs.tf
    │   └── README.md
    └── hack/
        └── manifest.yaml        # Test manifest for local development
```

#### Categories of Deployment Components

**1. Kubernetes Components**

Deploy applications and addons to any Kubernetes cluster:
- `PostgresKubernetes` - PostgreSQL with operator (CloudNativePG)
- `RedisKubernetes` - Redis with Helm chart
- `KafkaKubernetes` - Apache Kafka with Strimzi operator
- `CertManagerKubernetes` - cert-manager addon
- `MicroserviceKubernetes` - Containerized applications

**2. Cloud Provider Managed Services**

Deploy managed services on cloud providers:
- **AWS:** `AwsRdsInstance`, `AwsRdsCluster`, `AwsEksCluster`, `AwsS3Bucket`, `AwsAlb`
- **GCP:** `GcpCloudSql`, `GcpGkeCluster`, `GcpStorageBucket`, `GcpCloudRun`
- **Azure:** `AzureAksCluster`, `AzureSqlDatabase`, `AzureStorageAccount`

**3. SaaS Platform Integrations**

Provision and manage third-party SaaS platforms:
- `MongodbAtlas` - MongoDB Atlas clusters
- `ConfluentKafka` - Confluent Cloud Kafka
- `SnowflakeDatabase` - Snowflake data warehouse

---

## Repository Structure

```
planton/
├── apis/                        # Protocol Buffer definitions
│   └── org/
│       └── planton/
│           ├── shared/          # Shared types and enums
│           │   └── cloudresourcekind/
│           │       └── cloud_resource_kind.proto  # Registry of all components
│           └── provider/        # Provider-specific components
│               ├── aws/
│               │   ├── awsrdsinstance/v1/
│               │   ├── awsekscluster/v1/
│               │   └── ...
│               ├── gcp/
│               │   ├── gcpcloudsql/v1/
│               │   ├── gcpgkecluster/v1/
│               │   └── ...
│               ├── azure/
│               │   └── ...
│               └── kubernetes/
│                   ├── workload/
│                   │   ├── postgreskubernetes/v1/
│                   │   ├── rediskubernetes/v1/
│                   │   └── ...
│                   └── addon/
│                       ├── certmanager/v1/
│                       └── ...
├── architecture/                # Architecture documentation
│   ├── README.md               # This file
│   └── deployment-component.md # Deployment component ideal state
├── .cursor/                     # Cursor AI rules
│   └── rules/
│       └── deployment-component/
│           ├── forge/          # Create new components
│           ├── audit/          # Assess completeness
│           ├── update/         # Enhance existing
│           ├── complete/       # Auto-improve workflow
│           ├── fix/            # Targeted fixes
│           └── delete/         # Remove components
├── cli/                         # CLI implementation (Go)
├── buf.yaml                     # Buf configuration
├── buf.gen.yaml                # Buf code generation
├── Makefile                    # Build automation
└── README.md                   # Project README
```

### Key Directories Explained

#### `/apis`

All Protocol Buffer definitions organized by provider. Each component's v1 directory contains:
- API definitions (api.proto, spec.proto)
- Validation tests (spec_test.go)
- IaC modules (iac/pulumi, iac/terraform)
- Documentation (README.md, docs/README.md)

#### `/architecture`

High-level architecture documentation:
- **README.md** (this file): Complete architecture overview
- **deployment-component.md**: Ideal state definition for components

#### `/_rules/deployment-component`

Cursor AI rules for managing deployment components:
- **forge**: Create new components from scratch (21-step workflow)
- **audit**: Assess component completeness (9-category scoring)
- **update**: Enhance existing components (6 scenarios)
- **complete**: Automated workflow (audit + fill gaps + verify)
- **fix**: Targeted fixes with cascading updates
- **delete**: Safe component removal

#### `/cli`

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
- **Buf Schema Registry** - Proto module hosting and documentation
- **buf-validate** - Field-level validation rules (based on protovalidate-go)
- **CEL (Common Expression Language)** - Complex validation logic

**Generated Artifacts:**
- Go stubs (for CLI and modules)
- Java stubs (for JVM integrations)
- Python stubs (for Python tooling)
- TypeScript stubs (for web UIs)

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
- Every Pulumi module has a corresponding Terraform module
- Same functionality, same defaults, same behavior
- Users choose based on team preference

### CLI Layer

**Implementation:**
- **Language:** Go
- **Binary Distribution:** Homebrew (`brew install plantonhq/tap/planton`)
- **Configuration:** Environment variables and flags
- **Module Caching:** `~/.planton/modules/` (Git-based)

**Dependencies:**
- Git (required for cloning modules)
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

## Deployment Component Lifecycle

Planton provides a sophisticated lifecycle management system for deployment components. This system ensures that all components are consistently high-quality, well-documented, and production-ready.

### The Six Operations

```
┌─────────────────────────────────────────────────────────────┐
│               Deployment Component Lifecycle                │
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

**Purpose:** Bootstrap complete, production-ready deployment components from scratch.

**What It Creates:**
- ✅ Proto API definitions (4 files with validations)
- ✅ Validation tests (spec_test.go)
- ✅ IaC modules (both Pulumi and Terraform)
- ✅ Documentation (user-facing, research, technical)
- ✅ Supporting files (test manifests, debug scripts)
- ✅ Registry entry (cloud_resource_kind.proto)

**Result:** 95-100% completion score

**Workflow:** 21-step automated process organized in 7 phases:
1. **Proto API** (6 rules): spec.proto, validations, tests, stack_outputs, api, stack_input
2. **Registration** (2 rules): cloud_resource_kind enum, proto stubs
3. **Documentation** (2 rules): user-facing docs, research docs
4. **Test Infrastructure** (1 rule): hack manifest
5. **Pulumi** (5 rules): module, entrypoint, e2e, docs, overview
6. **Terraform** (3 rules): module, e2e, docs
7. **Validation** (2 rules): build verification, test execution

**Example:**
```bash
@forge-planton-component MongodbAtlas --provider atlas
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
- Timestamped reports saved to `<component>/v1/docs/audit/<timestamp>.md`

**Example:**
```bash
@audit-planton-component MongodbAtlas
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
@update-planton-component MongodbAtlas --scenario fill-gaps
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
@complete-planton-component MongodbAtlas
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

**Purpose:** Completely remove deployment components with safety features to prevent accidents.

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

All lifecycle operations reference a single source of truth: **`architecture/deployment-component.md`**

This document defines:
- Complete checklist of required artifacts
- Quality standards for each category
- Scoring weights and rationale
- 90/10 principle (provider schema as the floor)
- Examples of complete components

**Key Insight:** The ideal state is **intentionally pragmatic**—it focuses effort on the highest-leverage work. Not every component needs every possible artifact, but every production component should reach 95%+ completion, covering the provider's real surface to the floor.

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
planton validate --manifest config.yaml
# Catches errors before calling cloud APIs
```

3. **Cloud provider validation** (during deployment):
- Final validation by actual cloud provider APIs
- Catches provider-specific constraints

**Result:** 90%+ of errors caught before making any cloud API calls.

### The 90/10 Principle

**Coverage is benchmarked against the provider, not trimmed below it.**

For every cloud resource, Planton covers the broad majority of what real users need -- the ~90% of the provider's surface that production deployments actually reach -- using the provider's own API as the floor for completeness, never the ceiling.

**Example: PostgreSQL on Kubernetes**

**Most-reached (the common path):**
- Replicas (1 or 3)
- Storage size (10Gi, 50Gi, 100Gi)
- CPU and memory limits
- Database name and credentials

**Long tail (covered, with sensible defaults):**
- Custom WAL configuration
- Replication topologies
- Fine-grained operator settings

**Planton's approach:**
- Default modules cover the real surface to the provider floor
- Sensible defaults keep the common path simple
- Genuinely beta/niche knobs are skipped with a recorded reason (or left to an escape hatch)

### Provider-Specific vs. Generic

**Planton is intentionally NOT a cloud abstraction layer.**

**Example: Postgres Deployment**

Three different deployment components:
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
- CLI commands (`planton pulumi up`)
- Deployment workflow (validate → deploy → outputs)

**What's different:**
- Configuration options (provider-specific)
- Deployment target (AWS vs GCP vs K8s)

This preserves cloud-specific power while providing experience consistency.

---

## IaC Module Design

### Dual IaC Engine Support

Every deployment component has **both** Pulumi and Terraform implementations with **feature parity**.

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

#### 3. Environment Variable-Based Input

The CLI exports the manifest as an environment variable:
```bash
export PLANTON_MANIFEST="$(cat manifest.yaml)"
```

Modules read this variable:

**Pulumi (Go):**
```go
manifestYaml := os.Getenv("PLANTON_MANIFEST")
config := &PostgresKubernetes{}
yaml.Unmarshal([]byte(manifestYaml), config)
```

**Terraform (HCL):**
The CLI transforms YAML to `.tfvars` format and passes as variables.

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
├── validate           # Validate manifest (proto-validate)
├── pulumi            # Pulumi commands
│   ├── up            # Deploy/update
│   ├── destroy       # Tear down
│   ├── preview       # Preview changes
│   └── outputs       # Show outputs
├── tofu              # Terraform/OpenTofu commands
│   ├── apply         # Deploy/update
│   ├── destroy       # Tear down
│   └── plan          # Preview changes
└── version           # Show version
```

### Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   User runs CLI command                     │
│     planton pulumi up --manifest config.yaml        │
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
│ 3. Map kind → Module                                        │
│    - Lookup: PostgresKubernetes → kubernetes/postgres/v1    │
│    - Determine module URL                                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Clone/Pull Module                                        │
│    - Check cache: ~/.planton/modules/               │
│    - Git clone (if not cached)                              │
│    - Git pull (if cached)                                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Setup Environment                                        │
│    - Export PLANTON_MANIFEST=<yaml>                 │
│    - CD to module directory                                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Delegate to IaC Engine                                   │
│    - Pulumi: exec "pulumi up"                               │
│    - Terraform: exec "terraform apply"                      │
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

**Mapping kind to module URL:**

```go
// Internal registry
var moduleRegistry = map[string]string{
  "PostgresKubernetes": "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/postgreskubernetes/v1/iac",
  "MongodbAtlas": "github.com/plantonhq/planton/apis/dev/planton/provider/atlas/mongodbatlas/v1/iac",
  "AwsRdsInstance": "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsrdsinstance/v1/iac",
  // ... 100+ components
}
```

**Override with custom modules:**
```bash
planton pulumi up \
  --manifest config.yaml \
  --module-url github.com/myorg/custom-postgres-module
```

### Module Caching

**Location:** `~/.planton/modules/<kind>/`

**Flow:**
1. CLI checks if module cached
2. If not cached: `git clone <module-url>`
3. If cached: `git pull` (update to latest)
4. Use cached module

**Benefits:**
- Fast repeat deployments
- Offline support (after first clone)
- Version control (Git tags)

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

### Adding a New Deployment Component

**High-Level Process:**

```
Research → Forge → Audit → (Complete) → Deploy & Test → Commit
```

**Detailed Steps:**

1. **Research Phase**
   - Understand the resource (AWS RDS, GKE Cluster, etc.)
   - Research deployment methods (manual, Terraform, Pulumi)
   - Identify 90/10 coverage (the provider's real surface, benchmarked against the schema as the floor)
   - Document findings

2. **Forge Phase**
   ```bash
   @forge-planton-component MongodbAtlas --provider atlas
   ```
   - Creates complete component (95-100% complete)
   - Proto definitions with validations
   - Both Pulumi and Terraform modules
   - Documentation and examples
   - Test manifests

3. **Audit Phase**
   ```bash
   @audit-planton-component MongodbAtlas
   ```
   - Verify forge created everything
   - Check completion score (should be 95-100%)
   - Identify any gaps

4. **Complete Phase (if needed)**
   ```bash
   @complete-planton-component MongodbAtlas
   ```
   - Fill any remaining gaps
   - Re-audit to verify 100%

5. **Local Testing**
   ```bash
   cd apis/dev/planton/provider/atlas/mongodbatlas/v1/iac/pulumi
   export PLANTON_MANIFEST="$(cat ../hack/manifest.yaml)"
   pulumi up
   ```
   - Test Pulumi module locally
   - Repeat for Terraform module

6. **Validation**
   ```bash
   make protos  # Regenerate stubs
   make build   # Verify compilation
   make test    # Run all tests
   ```

7. **Commit**
   ```bash
   git add -A
   git commit -m "feat(atlas): add MongodbAtlas deployment component"
   git push origin main
   ```

### Updating an Existing Component

**Scenario 1: Adding a Field to spec.proto**

```bash
# 1. Edit spec.proto
vim apis/dev/planton/provider/atlas/mongodbatlas/v1/spec.proto

# 2. Add field with validation
# message Spec {
#   int32 backup_retention_days = 5 [(buf.validate.field).int32 = {gte: 1, lte: 365}];
# }

# 3. Propagate changes
@update-planton-component MongodbAtlas --scenario proto-changed

# This will:
# - Regenerate proto stubs
# - Update Pulumi module to use new field
# - Update Terraform module to use new field
# - Update documentation and examples
# - Add test for new validation rule

# 4. Verify
@audit-planton-component MongodbAtlas
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
  "MongodbAtlas"
  "ConfluentKafka"
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
1. Fork default module to private repository
2. Add organizational defaults
3. Point CLI to custom module

**Example:**

```bash
# Fork
git clone https://github.com/plantonhq/planton
cd planton/apis/dev/planton/provider/aws/awsrdsinstance/v1/iac/pulumi
# Edit main.go to add organizational defaults

# Push to your private repo
git remote add myorg git@github.com:myorg/custom-aws-rds-module.git
git push myorg main

# Use custom module
planton pulumi up \
  --manifest rds.yaml \
  --module-url github.com/myorg/custom-aws-rds-module
```

### Pattern 2: Reusing APIs for Internal Tooling

**Scenario:** You're building an internal self-service portal where developers request databases.

**Approach:**
1. Import Planton proto SDKs
2. Use strongly-typed APIs in your tool
3. Generate manifests from user input
4. Call Planton CLI programmatically

**Example (Python):**

```python
from planton.apis.dev.planton.provider.kubernetes.workload.postgreskubernetes.v1 import api_pb2
import yaml
import subprocess

def create_database_from_ui(user_input):
    # Create strongly-typed config
    config = api_pb2.PostgresKubernetes()
    config.api_version = "kubernetes.planton.dev/v1"
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
            planton validate --manifest $manifest
          done
      
      - name: Deploy
        run: |
          for manifest in infrastructure/*.yaml; do
            planton pulumi up \
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
1. Install proto SDKs in your preferred language
2. Write custom module consuming the APIs
3. Point CLI to your module

**Example (Python Pulumi module):**

```python
# custom_postgres_module.py
import os
import yaml
import pulumi
import pulumi_kubernetes as k8s
from planton.apis.dev.planton.provider.kubernetes.workload.postgreskubernetes.v1 import api_pb2

# Read manifest from environment
manifest_yaml = os.getenv("PLANTON_MANIFEST")
manifest_dict = yaml.safe_load(manifest_yaml)

# Parse into proto
config = api_pb2.PostgresKubernetes()
# ... populate from manifest_dict ...

# Deploy using Pulumi (Python style)
namespace = k8s.core.v1.Namespace(
    config.metadata.name,
    metadata=k8s.meta.v1.ObjectMetaArgs(
        name=config.metadata.name
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
        "replicas": config.spec.container.replicas,
        "resources": {
            "limits": {
                "cpu": config.spec.container.resources.limits.cpu,
                "memory": config.spec.container.resources.limits.memory
            }
        }
    }
)
```

---

## Contributing

### Ways to Contribute

**1. Add New Deployment Components**

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
feat(aws): add AwsLambdaFunction deployment component
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
- Research docs/README.md (why and design rationale)
- Examples.md (real-world use cases)
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
cd apis/dev/planton/provider/atlas/mongodbatlas/v1
go test -v
```

**2. Integration Tests**

Test IaC modules locally:
```bash
cd apis/dev/planton/provider/atlas/mongodbatlas/v1/iac/pulumi
export PLANTON_MANIFEST="$(cat ../hack/manifest.yaml)"
pulumi preview
```

**3. End-to-End Tests**

Test full deployment workflow:
```bash
planton validate --manifest test.yaml
planton pulumi up --manifest test.yaml --stack test
planton pulumi destroy --manifest test.yaml --stack test
```

### Getting Help

**Documentation:**
- Architecture overview (this file)
- Deployment component ideal state (`architecture/deployment-component.md`)
- Lifecycle operation READMEs (`_rules/deployment-component/*/README.md`)

**Examples:**
- Browse complete components (`apis/dev/planton/provider/`)
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
- ✅ 100+ deployment components (AWS, GCP, Azure, K8s, SaaS)

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
- 90/10 principle (cover the provider surface to the floor, with sensible defaults)
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
- Browse deployment components in `apis/dev/planton/provider/`
- Read lifecycle management guides in `_rules/deployment-component/`
- Try deploying a component locally
- Contribute new components or improvements

---

**Ready to contribute?** Start with the Forge workflow to create a new deployment component, or use Complete to improve an existing one to 100% quality!

