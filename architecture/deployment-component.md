# Deployment Component: Definition and Ideal State

## What is a Deployment Component?

A **deployment component** in Planton is a self-contained, production-ready package that enables declarative deployment of a specific infrastructure resource or application workload to a cloud provider or Kubernetes cluster.

### Technical Definition

A deployment component consists of:

1. **API Definition (Protobuf)** - A strongly-typed, language-neutral schema that defines:
   - The configuration interface (`spec.proto`)
   - The deployment inputs (`stack_input.proto`)
   - The deployment outputs (`stack_outputs.proto`)
   - Field-level validation rules

2. **Infrastructure-as-Code Modules** - Executable deployment logic in both:
   - Pulumi (Go-based, using real programming language)
   - Terraform/OpenTofu (HCL-based, declarative)

3. **Documentation** - Multi-layered documentation serving different audiences:
   - Research documentation (comprehensive landscape analysis)
   - User-facing documentation (Planton perspective)
   - Examples (copy-paste ready, validated against current API)

### Role in Planton

Deployment components are the **atomic units of deployment** in Planton. They serve as:

- **The Menu Items** - In the restaurant analogy from the main README, deployment components are the individual dishes available for order
- **Reusable Building Blocks** - Platform engineers compose multiple deployment components to build complete application stacks
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
- **Documentation-First** - Research-driven design with comprehensive landscape analysis

### Examples of Deployment Components

**Cloud Provider Resources:**
- `AwsRdsInstance` - PostgreSQL/MySQL on AWS RDS
- `GcpCloudSql` - PostgreSQL/MySQL on Google Cloud SQL
- `AzureAksCluster` - Managed Kubernetes on Azure
- `GcpCertManagerCert` - SSL/TLS certificates on GCP

**Kubernetes Workloads:**
- `PostgresKubernetes` - PostgreSQL deployed to any Kubernetes cluster
- `RedisKubernetes` - Redis deployed to any Kubernetes cluster
- `MicroserviceKubernetes` - Containerized microservice deployment

**SaaS Platform Resources:**
- `MongodbAtlas` - MongoDB Atlas cluster
- `ConfluentKafka` - Confluent Cloud Kafka cluster
- `SnowflakeDatabase` - Snowflake database

---

## What Does "Complete" Mean?

Completeness of a deployment component is **contextual and principle-driven**, not a simple checklist of every possible feature.

### Philosophy of Completeness

**The 90/10 Principle:**

A complete deployment component covers the broad majority of what real users need -- the ~90% of a provider's surface that production deployments actually reach -- benchmarked against the provider's own canonical API as the **floor** for completeness, never the ceiling for ambition. This means:

- **Schema as the Floor** - The provider's canonical API (the Terraform provider is the reference) is the minimum bar for how complete a component should be. An advanced organization should be able to reach the long tail of what the provider offers, not just the surface knobs.
- **Broad, Production-Grade Coverage** - Expose the fields production deployments actually use, across both common and advanced scenarios.
- **Research-Driven** - Coverage is grounded in the provider schema and real-world usage patterns, not guesswork.
- **Opinionated Defaults** - Provide sensible defaults so breadth never costs usability: advanced fields default well and stay out of the way until a user needs them.

**Example:** For `GcpCertManagerCert`, the most-reached fields are:
- `gcp_project_id` - Where to deploy
- `primary_domain_name` - What domain to secure
- `cloud_dns_zone_id` - Where to create validation records
- `certificate_type` - MANAGED vs LOAD_BALANCER

Advanced fields like certificate scope, location, and labels are exposed with sensible defaults so the long tail is reachable without burdening the common case.

### Contextual vs Absolute Completeness

**Contextual Completeness** means a component is complete when:

1. **Research Validates Coverage** - The `docs/README.md` research document maps the deployment landscape and the provider's surface, and justifies the coverage reached (and any genuinely-niche surface deliberately skipped, with a reason)

2. **Proto Schema Meets the Floor** - The `spec.proto` covers the provider's real surface to the 90/10 bar, benchmarked against the provider's API as the floor -- not arbitrarily trimmed below it

3. **Both IaC Modules Implement the Schema** - Every field defined in `spec.proto` is actually used in both Pulumi and Terraform modules (no unused fields, no missing implementations)

4. **Presets Validate the API** - The preset YAML files contain working, deployable configurations that demonstrate the API's capabilities and validate against the current schema

5. **Documentation Explains Decisions** - Users understand the coverage and the rationale for anything deliberately skipped, reducing support burden

**What we still avoid** (90/10 is deliberately not literal 100%):

- Genuinely beta, deprecated, or vendor-locked fields with no real-world use -- skipped *with a recorded reason*, not silently
- Supporting every possible deployment method
- Field count for its own sake -- coverage must be tested, parity-verified, and deploy-validated

### Quality At Scale

A deployment component reaches 90/10 by covering the provider's real surface **with** the same rigor -- every field researched, documented, validated, and exercised in both engines. Breadth never excuses a hastily-added, undocumented, or untested field: quality is the constant, and coverage is raised to the floor on top of it.

**Completeness Indicators:**
- ✅ Research document explains landscape and rationale
- ✅ Proto schema is validated with real-world constraints
- ✅ Both IaC modules have feature parity
- ✅ Examples are tested and current
- ✅ Documentation answers "why these choices?"
- ✅ Presets provide ready-to-deploy starting points for common patterns

**Incompleteness Indicators:**
- ❌ Proto has fields that aren't used in IaC modules
- ❌ IaC modules reference fields not in proto
- ❌ Examples fail validation against current schema
- ❌ No research justifying scope decisions
- ❌ Missing Terraform or Pulumi implementation
- ❌ No presets, or presets reference stale fields from an older spec.proto

---

## Ideal State Checklist

The following sections define the complete, ideal state of any deployment component. This serves as both a reference for developers building components and as the specification for automated auditing.

### 1. Cloud Resource Registry

**Location:** `apis/project/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

**Requirements:**

- [ ] **Enum Entry Exists** - Component has an entry in the `CloudResourceKind` enum
- [ ] **Correct Provider Range** - Enum value is within the correct provider's numeric range:
  - Test/dev/custom: 1-49
  - SaaS platforms: 50-199
  - AWS: 200-399
  - Azure: 400-599
  - GCP: 600-799
  - Kubernetes: 800-999
  - DigitalOcean: 1200-1499
  - Civo: 1500-1799
  - Cloudflare: 1800-2099
- [ ] **Unique Enum Value** - No duplicate enum numbers
- [ ] **Unique ID Prefix** - The `id_prefix` is globally unique across all providers
- [ ] **Proper Metadata** - `kind_meta` includes:
  - `provider` - Correct provider enum value
  - `version` - Currently `v1` for all components
  - `id_prefix` - Short, descriptive prefix (3-7 characters)
- [ ] **Optional Metadata (when applicable)** - `kind_meta` may also include:
  - `prerequisites` - Other `CloudResourceKind`s that must exist first (e.g. an operator or CRD-installer like `KubernetesGatewayApiCrds`); drives resource-graph and infra-chart ordering
  - `is_service_kind` - Whether this kind is a Service Hub deployment target
  - `container_kind` - Whether this kind contains child resources in the org graph
  - Note: there is no `kubernetes_meta`/`category`/`namespace_prefix` field on `CloudResourceKindMeta`; the Kubernetes layout is flat (`provider/kubernetes/<component>/v1/`).

**Example:**
```protobuf
GcpCertManagerCert = 616 [(kind_meta) = {
  provider: gcp
  version: v1
  id_prefix: "gcpcert"
}];
```

---

### 2. Folder Structure

**Base Path:** `apis/dev/planton/provider/<provider>/<component>/v1/`

**Requirements:**

- [ ] **Correct Provider Hierarchy** - Component folder is under the correct provider:
  - `apis/dev/planton/provider/aws/<component>/v1/`
  - `apis/dev/planton/provider/gcp/<component>/v1/`
  - `apis/dev/planton/provider/azure/<component>/v1/`
  - `apis/dev/planton/provider/kubernetes/<component>/v1/`
  - etc.

- [ ] **Lowercase Folder Naming** - Component folder name matches the `CloudResourceKind` enum value but in all lowercase
  - Enum: `GcpCertManagerCert` → Folder: `gcpcertmanagercert`
  - Enum: `PostgresKubernetes` → Folder: `postgreskubernetes`

- [ ] **Version Subfolder** - All files are under `v1/` subfolder (prepared for future API versioning)

**Example Structure:**
```
apis/dev/planton/provider/gcp/gcpcertmanagercert/v1/
├── api.proto
├── spec.proto
├── stack_input.proto
├── stack_outputs.proto
├── api.pb.go
├── spec.pb.go
├── stack_input.pb.go
├── stack_outputs.pb.go
├── spec_test.go
├── README.md
├── docs/
│   └── README.md
├── presets/
│   ├── 01-managed-dns-validated.yaml
│   ├── 01-managed-dns-validated.md
│   ├── 02-load-balancer-cert.yaml
│   └── 02-load-balancer-cert.md
└── iac/
    ├── hack/
    │   └── manifest.yaml
    ├── pulumi/
    │   ├── main.go
    │   ├── Pulumi.yaml
    │   ├── Makefile
    │   ├── debug.sh
    │   ├── README.md
    │   ├── overview.md
    │   └── module/
    │       ├── main.go
    │       ├── locals.go
    │       ├── outputs.go
    │       └── <resource-specific>.go
    └── tf/
        ├── provider.tf
        ├── variables.tf
        ├── locals.tf
        ├── main.tf
        ├── outputs.tf
        └── README.md
```

---

### 3. Protobuf API Definitions

**Location:** `v1/*.proto`

#### 3.1 api.proto

**Purpose:** Wires together the Kubernetes Resource Model envelope (metadata, spec, status)

**Requirements:**

- [ ] **File Exists** - `v1/api.proto` is present
- [ ] **Correct Package** - Package declaration matches path:
  - `package dev.planton.provider.<provider>.<component>.v1;`
- [ ] **Standard Imports** - Imports common proto dependencies:
  ```protobuf
  import "buf/validate/validate.proto";
  import "dev/planton/provider/<provider>/<component>/v1/spec.proto";
  import "dev/planton/provider/<provider>/<component>/v1/stack_outputs.proto";
  import "dev/planton/shared/metadata.proto";
  ```
- [ ] **Resource Message** - Defines `<Kind>` message with KRM structure:
  ```protobuf
  message <Kind> {
    string api_version = 1 [(buf.validate.field).string.const = '<provider>.planton.dev/v1'];
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

- [ ] **File Exists** - `v1/spec.proto` is present
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
- [ ] **90/10 Coverage** - Fields reach the provider's real surface (benchmarked against the schema as the floor), with sensible defaults
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

#### 3.3 stack_input.proto

**Purpose:** Defines inputs to the IaC modules (includes spec + credentials + environment context)

**Requirements:**

- [ ] **File Exists** - `v1/stack_input.proto` is present
- [ ] **Correct Package** - Package declaration matches path
- [ ] **Standard Imports** - Imports common dependencies:
  ```protobuf
  import "dev/planton/provider/<provider>/<component>/v1/api.proto";
  import "dev/planton/provider/<provider>/provider.proto";
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
- [ ] **Credential Field** - References the correct provider credential type:
  - AWS: `dev.planton.provider.aws.credential.v1.AwsCredential`
  - GCP: `dev.planton.provider.gcp.credential.v1.GcpCredential`
  - Kubernetes: `dev.planton.provider.kubernetes.provider.v1.KubernetesProvider`

#### 3.4 stack_outputs.proto

**Purpose:** Defines outputs from the IaC deployment (what gets written to status.outputs)

**Requirements:**

- [ ] **File Exists** - `v1/stack_outputs.proto` is present
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
  - `stack_input.pb.go`
  - `stack_outputs.pb.go`
- [ ] **Stubs Are Current** - Generated files match proto definitions (run `make protos` to regenerate)

#### 3.6 Unit Tests

**Location:** `v1/spec_test.go`

**Purpose:** Validate that all buf.validate rules in spec.proto are syntactically and semantically correct

**Requirements:**

- [ ] **File Exists** - `v1/spec_test.go` is present
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
  go test ./apis/dev/planton/provider/<provider>/<component>/v1/
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

**Base Path:** `v1/iac/pulumi/`

#### 4.1 Pulumi Module Files

**Location:** `v1/iac/pulumi/module/`

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
  - Formats output structure matching `stack_outputs.proto`
  - **MUST** contain actual `ctx.Export()` calls
  - **MUST** map all fields from `stack_outputs.proto`
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

#### 4.2 Pulumi Entrypoint Files

**Location:** `v1/iac/pulumi/`

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
- [ ] **Makefile** - Automation targets:
  - `build` - Compiles the Go code
  - `install-pulumi-plugins` - Installs required Pulumi provider plugins
  - `test` - Runs the module against test manifests
- [ ] **debug.sh** - Debugging helper script for local testing
- [ ] **README.md** - Pulumi-specific usage guide
- [ ] **overview.md** - Module architecture and design decisions

**Integration:**
- [ ] **Compiles Successfully** - `make build` completes without errors
- [ ] **Plugin Dependencies Listed** - `Pulumi.yaml` or `Makefile` documents required plugins
- [ ] **Executable** - Binary can be built and run

---

### 5. IaC Modules - Terraform

**Base Path:** `v1/iac/tf/`

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

- [ ] **outputs.tf** - Output values matching `stack_outputs.proto`:
  - Every field in `<Kind>StackOutputs` has a corresponding Terraform output
  - Output descriptions match proto field comments
  - **MUST** contain actual `output` blocks
  - **MUST** extract values from created resources

- [ ] **README.md** - Terraform-specific usage guide

**Code Quality:**
- [ ] **Valid HCL** - All `.tf` files are valid Terraform configuration
- [ ] **Validates Successfully** - `terraform validate` passes
- [ ] **Feature Parity with Pulumi** - Creates the same resources as Pulumi module
- [ ] **No Hardcoded Values** - All configuration comes from variables
- [ ] **Proper Dependencies** - Uses `depends_on` where needed
- [ ] **Not Empty** - main.tf has substantial content (>100 bytes minimum)
- [ ] **Functional** - Can actually deploy resources, not just validate syntax

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

### 6. Documentation - Technical Research

**Location:** `v1/docs/README.md`

**Purpose:** Comprehensive research document explaining the deployment landscape

**CRITICAL:** This document is the **primary source of truth** for understanding the component. It should be consulted when:
- Executing any lifecycle operation (forge, audit, update, delete)
- Making decisions about component behavior
- Understanding design rationale and scoping decisions
- Troubleshooting or debugging issues
- Evaluating whether to keep, update, or delete the component

**Requirements:**

- [ ] **File Exists** - `v1/docs/README.md` is present
- [ ] **Substantial Content** - Typically 300-1000+ lines (not a stub)
- [ ] **Introduction** - What the component is and why it matters
- [ ] **Landscape Analysis** - Survey of deployment methods:
  - Manual (cloud console, CLI)
  - IaC tools (Terraform, Pulumi, CloudFormation, etc.)
  - Specialized tools (Helm, Ansible, Crossplane, etc.)
  - Comparison of approaches
- [ ] **90/10 Coverage Decision** - Explicit explanation of:
  - Which provider surface is covered and why
  - Which genuinely-niche features are skipped, and the reason
  - How coverage was benchmarked against the provider schema as the floor
- [ ] **Best Practices** - Production-ready recommendations
- [ ] **Common Pitfalls** - Known issues and how to avoid them
- [ ] **Research-Backed** - References to official documentation, community discussions, real-world usage

**Content Quality:**
- [ ] **Technical Depth** - Goes beyond marketing material
- [ ] **Opinionated** - Makes clear recommendations
- [ ] **Actionable** - Readers understand what to do
- [ ] **Well-Structured** - Uses headings, sections, tables
- [ ] **Examples Included** - Shows real code/configuration snippets

**Example Sections:**
- Introduction
- The Evolution (history of the technology)
- Deployment Methods (manual → automated)
- Comparative Analysis
- Planton's Approach
- Implementation Landscape
- Production Best Practices
- Conclusion

---

### 7. Documentation - User-Facing

#### 7.1 README.md

**Location:** `v1/README.md`

**Purpose:** Concise, Planton perspective overview

**Requirements:**

- [ ] **File Exists** - `v1/README.md` is present
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
- Detailed landscape analysis (that's in `docs/README.md`)
- History of the technology (not relevant to users)
- Comparison of every deployment method (too detailed)
- Every possible configuration option (that's in examples)

**Tone:**
- Helpful and encouraging
- Focused on getting started quickly
- Assumes reader knows basic concepts
- Points to other documentation for depth

---

### 8. Supporting Files

#### 8.1 Hack Manifest

**Location:** `v1/iac/hack/manifest.yaml`

**Purpose:** Test manifest for local development and CI/CD testing

**Requirements:**

- [ ] **File Exists** - `v1/iac/hack/manifest.yaml` is present
- [ ] **Valid Manifest** - Complete YAML manifest with:
  - `apiVersion`, `kind`, `metadata`, `spec`
  - Realistic test values
  - Can be used for `make test` in Pulumi folder
- [ ] **Non-Production Values** - Uses test/dev values (not real production data)

#### 8.2 Pulumi Supporting Files

**Location:** `v1/iac/pulumi/`

**Files:**

- [ ] **README.md** - Pulumi module usage guide:
  - How to use the module standalone
  - Required environment variables
  - How to pass credentials
  - Example deployment commands
  - Troubleshooting tips

- [ ] **overview.md** - Module architecture:
  - High-level architecture diagram (text/ASCII)
  - Key design decisions
  - Resource relationships
  - Data flow

- [ ] **debug.sh** - Debugging helper script:
  - Sets up environment for local testing
  - Exports manifest as environment variable
  - Runs Pulumi commands with proper configuration

#### 8.3 Terraform Supporting Files

**Location:** `v1/iac/tf/`

**Files:**

- [ ] **README.md** - Terraform module usage guide:
  - How to use the module standalone
  - Required variables
  - How to pass credentials
  - Example terraform commands
  - Troubleshooting tips

---

### 9. Presets

**Location:** `v1/presets/`

**Purpose:** Production-quality, directly deployable YAML manifests representing the most common real-world configuration patterns for the component. Each preset is a ranked starting point that users can deploy immediately (after replacing placeholders) without needing to understand every field in `spec.proto`.

**Reference:** See `architecture/presets.md` for the full convention specification and authoring guide.

**Requirements:**

- [ ] **Directory Exists** - `v1/presets/` directory is present
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
| `iac/hack/manifest.yaml` | Minimal test manifest for CI/CD | Presets are production-quality, not minimal test configs |
| `docs/README.md` | Research and design rationale | Presets are actionable starting points, not explanation |

**Example:**
```
v1/presets/
├── 01-internet-facing-https.yaml    # Most common: HTTPS ALB with SSL
├── 01-internet-facing-https.md
├── 02-internal-http.yaml            # Internal-only, no SSL
└── 02-internal-http.md
```

---

## Completeness Assessment Criteria

When evaluating whether a deployment component is "complete," assess each category:

### Critical (Must Have - 48.64%)

These are non-negotiable for a component to be considered functional:

1. ✅ Entry in `cloud_resource_kind.proto` (4.44%)
2. ✅ Correct folder structure (4.44%)
3. ✅ All four proto files (api, spec, stack_input, stack_outputs) (13.32%)
4. ✅ Generated proto stubs (.pb.go files) (3.33%)
5. ✅ spec_test.go with validation tests (2.77%)
6. ✅ **Tests execute and pass** (2.78%) - Component-specific `go test` succeeds
7. ✅ Pulumi module with main.go, locals.go, outputs.go (6.66%)
8. ✅ Pulumi entrypoint (main.go, Pulumi.yaml, Makefile) (6.66%)
9. ✅ Terraform module with all 5 core files (variables.tf, provider.tf, locals.tf, main.tf, outputs.tf) (4.24%)

**Note:** Test execution is now explicitly part of critical items. Failing tests = incomplete component.

### Important (Should Have - 41.36%)

These significantly improve quality and usability:

10. ✅ Comprehensive research document (docs/README.md) (13.18%)
11. ✅ User-facing README (v1/README.md) (13.09%)
13. ✅ Pulumi supporting documentation (README, overview) (5.05%)
14. ✅ Terraform supporting documentation (README) (2.52%)
15. ✅ Supporting files (hack manifest, debug scripts) (2.52%)
16. ✅ Presets with companion documentation (v1/presets/) (5.00%)

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
- **IaC Implementation Quality** - Are both modules feature-complete? Do they work?
- **Documentation Quality** - Is the research comprehensive? Are examples current?
- **Consistency Quality** - Do variables.tf match spec.proto? Do outputs match stack_outputs.proto?

A component with all files but low quality in these dimensions should be scored lower than the raw percentage suggests.

---

## Using This Document

### For Developers

When building a new deployment component, use this document as your checklist. Work through each section systematically, ensuring every requirement is met.

### For Reviewers

When reviewing a PR that adds or updates a deployment component, use this document to validate completeness. Check off items and provide specific feedback on what's missing.

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
6. Check that outputs.tf fields match stack_outputs.proto fields
7. Run unit tests with `make test`
8. **Verify IaC module implementation completeness**:
   - Pulumi module: Check main.go has provider setup and resource calls
   - Pulumi module: Check locals.go extracts and computes values
   - Pulumi module: Check outputs.go has ctx.Export() calls
   - Terraform module: Check main.tf has resource blocks (not empty)
   - Terraform module: Check provider.tf has provider configuration
   - Terraform module: Check locals.tf has local value definitions
   - Terraform module: Check outputs.tf has output blocks
9. **Verify preset coverage and correctness**:
   - Check `v1/presets/` directory exists with at least one YAML + MD pair
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

A "complete" deployment component in Planton is not simply a collection of files. It's a well-researched, thoughtfully-scoped, fully-implemented package that serves real-world deployment needs with both Pulumi and Terraform, backed by comprehensive documentation that explains both "how" and "why," and equipped with ready-to-deploy presets that give users an immediate starting point.

This document provides the definitive reference for what completeness means, enabling both human developers and automated tools to assess and improve deployment components systematically.

