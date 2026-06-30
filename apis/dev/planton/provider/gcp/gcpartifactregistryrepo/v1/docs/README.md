# GCP Artifact Registry: Deployment Approaches and Design Philosophy

## Introduction: Beyond Container Registries

When Google Cloud deprecated Google Container Registry (GCR) in favor of Artifact Registry (AR), it wasn't simply rebranding a service—it was a strategic architectural upgrade that transformed a limited container image store into a comprehensive artifact management platform. Understanding this evolution is essential for making informed deployment decisions.

GCR was never a standalone service. It was an abstraction layer over Google Cloud Storage buckets, which imposed critical constraints: project-level IAM granularity, mandatory multi-regional storage, and Docker-only support. Artifact Registry broke free from these limitations, offering repository-level permissions, flexible regional deployment, and support for Maven, npm, Python, Go, and generic packages alongside container images.

But this expansion of capabilities introduces a new question for platform teams: **How should we manage Artifact Registry repositories—manually through the console, imperatively with scripts, or declaratively through Infrastructure as Code?**

This document explores the full spectrum of deployment methods, from anti-patterns to production-ready solutions, and explains the architectural principles that guide Planton's design.

## The Deployment Maturity Spectrum

### Level 0: The Console Anti-Pattern

**What it is:** Creating repositories through Google Cloud Console's web interface—clicking "Create Repository," filling out forms, and manually configuring settings.

**Why it exists:** The console is excellent for discovery, learning the API surface, and debugging permission issues.

**Why it fails in production:**
- **No repeatability:** Cannot recreate infrastructure reliably across environments
- **No auditability:** No version control or change history
- **No scalability:** Managing dozens of repositories across multiple projects becomes unmanageable
- **Drift vulnerability:** Manual changes create configuration inconsistencies

**Verdict:** Suitable only for learning and one-off experiments, never for production infrastructure.

### Level 1: Imperative Scripting with gcloud CLI

**What it is:** Using `gcloud artifacts repositories create` commands in shell scripts to automate repository creation.

**What it solves:** 
- Enables automation and repeatability
- Provides programmatic control
- Works well for simple migrations or batch operations

**What it doesn't solve:**
The gcloud CLI reveals a critical characteristic of Artifact Registry's API through its command structure:

```bash
# Creation command
gcloud artifacts repositories create REPO_ID \
  --location=us-central1 \
  --repository-format=docker \
  --mode=STANDARD_REPOSITORY \
  --kms-key=projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key

# Update command (limited scope)
gcloud artifacts repositories update REPO_ID \
  --description="New description" \
  --update-labels=env=prod
```

Notice what's **missing** from the update command: `location`, `repository-format`, `mode`, and `kms-key`. These fields are **immutable**—they can only be set during creation and require a "destroy and recreate" operation to change. This makes imperative scripting fragile and error-prone in production environments where infrastructure state must be carefully tracked.

**Limitations:**
- No built-in state management
- Manual reconciliation of desired vs. actual state
- Difficult to handle immutable field changes
- No dependency management for related resources (IAM, cleanup policies, etc.)

**Verdict:** Useful for migrations and one-time operations, but insufficient for managing production infrastructure lifecycles.

### Level 2: Cloud SDKs for Custom Applications

**What it is:** Using Google Cloud client libraries (Python, Go, Java, Node.js) to embed repository management into custom tooling.

**Use case:** When building custom platform engineering tools that need to provision repositories as part of a larger workflow.

**Key insight:** The ecosystem maintains a clear separation:
- **Resource management libraries** (`google-cloud-artifactregistry`) create and configure repositories
- **Client authentication helpers** (`keyrings.google-artifactregistry-auth`, `gcloud auth configure-docker`) configure local tools to authenticate

This separation is important: managing infrastructure is distinct from consuming it.

**Limitations:**
- Requires custom code and maintenance
- Still lacks declarative state management
- Reinvents patterns already solved by IaC tools

**Verdict:** Only justified for highly specialized platform automation; most teams should use established IaC tools instead.

### Level 3: Declarative Infrastructure as Code (Production Standard)

**What it is:** Using declarative tools like Terraform, Pulumi, or Config Connector to define repositories as code, with automatic state management and reconciliation.

**Why this is the production standard:**

1. **Declarative model:** Define the desired end state, let the tool determine actions
2. **State management:** Automatic tracking of current vs. desired configuration
3. **Immutability handling:** IaC tools handle "replace" operations for immutable fields
4. **Modularity:** Separate resources for the repository, IAM bindings, and cleanup policies
5. **Version control:** Infrastructure changes tracked in Git with review workflows
6. **Multi-cloud patterns:** Same workflows extend to other providers

**The convergence evidence:** Analysis of the three leading IaC tools reveals they've independently converged on an **identical API abstraction**:

| API Field | Terraform | Pulumi | Config Connector |
|-----------|-----------|---------|------------------|
| Repository ID | `repository_id` | `repositoryId` | `metadata.name` |
| Location | `location` (immutable) | `location` (immutable) | `spec.location` (immutable) |
| Format | `format` (immutable) | `format` (immutable) | `spec.format` (immutable) |
| Mode | `mode` (immutable) | `mode` (immutable) | `spec.mode` (immutable) |
| Docker Config | `docker_config` | `dockerConfig` | `spec.dockerConfig` |
| Cleanup Policies | `cleanup_policies` (list) | `cleanupPolicies` (list) | `spec.cleanupPolicies` (list) |
| CMEK | `kms_key_name` (immutable) | `kmsKeyName` (immutable) | `spec.kmsKeyRef` (immutable) |
| Remote Config | `remote_repository_config` | `remoteRepositoryConfig` | `spec.remoteRepositoryConfig` |
| Virtual Config | `virtual_repository_config` | `virtualRepositoryConfig` | `spec.virtualRepositoryConfig` |

This convergence is powerful validation: the "correct" API shape has been independently discovered and vetted by three different tool ecosystems.

**The critical IAM pattern:**

All production IaC tools follow a mandatory pattern: **separate resource provisioning from IAM management**. 

**Anti-pattern:**
```hcl
# Using an authoritative IAM policy on the repository resource
resource "google_artifact_registry_repository_iam_policy" "repo_policy" {
  # This OVERWRITES all other permissions - breaks modularity!
}
```

**Best practice:**
```hcl
# Repository creation
resource "google_artifact_registry_repository" "repo" {
  location      = "us-central1"
  repository_id = "prod-app-images"
  format        = "DOCKER"
}

# Additive IAM binding (separate resource)
resource "google_artifact_registry_repository_iam_member" "ci_writer" {
  repository = google_artifact_registry_repository.repo.id
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:ci-pipeline@project.iam.gserviceaccount.com"
}

# Another component can independently add its binding
resource "google_artifact_registry_repository_iam_member" "gke_reader" {
  repository = google_artifact_registry_repository.repo.id
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:gke-runtime@project.iam.gserviceaccount.com"
}
```

This additive approach enables modular IaC where CI/CD modules and runtime modules can independently manage their permissions without overwriting each other.

**Verdict:** This is the only production-ready approach for managing Artifact Registry at scale.

## Advanced Repository Modes: The Platform Features

Artifact Registry's true differentiator isn't just multi-format support—it's the three repository modes that transform it from a simple registry into an artifact management platform.

### Standard Repository (Default)

Your typical private repository: you push artifacts, you pull artifacts, you control access.

### Remote Repository (Caching Proxy)

A **caching proxy** for upstream public repositories (Docker Hub, PyPI, Maven Central, npm).

**How it works:**
1. Request an artifact from the remote repository
2. AR fetches it from the public source and caches it locally
3. Future requests are served from the cache

**Why this matters:**
- **Security:** All proxied dependencies can be scanned for vulnerabilities by Artifact Analysis
- **Performance:** Cached artifacts in the same region eliminate download latency
- **Reliability:** Insulates builds from public repository outages or rate limits

### Virtual Repository (Aggregator)

A **unified endpoint** that aggregates multiple upstream repositories (both Standard and Remote).

**The killer feature:** Defense against **dependency confusion attacks**.

Virtual repositories can be configured with upstream priority:
1. Check private "Standard" repository first
2. If not found, check public "Remote" repository second

This means if you publish an internal package called `auth-utils`, it will **always** be served from your private repository, even if an attacker publishes a malicious `auth-utils` package to PyPI. The virtual repository prioritizes your internal source.

**Example architecture:**
```
prod-python-virtual (Virtual Repository)
├─ Priority 1: prod-python-private (Standard)
└─ Priority 2: pypi-remote-cache (Remote → PyPI)
```

Developers configure a single endpoint (`prod-python-virtual`), and the platform handles routing, security, and caching automatically.

## Planton's Design Philosophy

Planton's `GcpArtifactRegistryRepo` resource is designed around three core principles:

### Principle 1: 80/20 Configuration Surface

Most Artifact Registry repositories need the same core settings. The API prioritizes the 20% of configuration that covers 80% of use cases:

**Essential fields:**
- `project_id` - The GCP project
- `region` - Deployment location (deliberately named "region" to encourage regional-first architecture)
- `repo_format` - The artifact type (DOCKER, MAVEN, PYTHON, etc.)

**Production-critical fields:**
- `enable_public_access` - Explicit control over unauthenticated access (critical security boundary)

**What's intentionally minimal:** The current spec reflects a deliberate focus on the most common scenario: regional Docker registries for private GKE deployments.

### Principle 2: Regional-First Architecture

The pricing model for Artifact Registry provides a clear architectural signal:

| Scenario | Cost |
|----------|------|
| Regional repo → Same-region GKE | **FREE egress** |
| Regional repo → Same-continent GKE | Paid egress |
| Multi-region repo → Same-continent GKE | FREE egress |
| Multi-region repo → Cross-continent GKE | Paid egress |
| Storage (all types) | $0.10/GB/month |

**Key insight:** The storage cost is intentionally higher than GCS ($0.10 vs $0.026), but this is offset by **free regional egress**. Google Cloud is deliberately incentivizing co-located architecture.

**Additional benefit:** GKE Image Streaming (which dramatically reduces container startup time) **only works** when the repository is in the same region as the GKE nodes.

**Planton guidance:** Always create regional repositories. Always co-locate them with their consumers. Multi-regional repositories should be reserved for artifacts consumed outside Google Cloud.

### Principle 3: Security by Default

The `enable_public_access` flag is deliberately explicit rather than defaulting to any value. This forces infrastructure authors to make a conscious decision about repository visibility, preventing accidental exposure of private artifacts.

**Secure integration patterns:**

For **GKE workloads**, the only supported pattern is **Workload Identity**:
1. GKE Kubernetes Service Account (KSA) → bound to → GCP Service Account (GSA)
2. GSA granted `roles/artifactregistry.reader` on the specific repository
3. No JSON keys, no secrets—automatic, keyless authentication

For **CI/CD pipelines**, the pattern is **Workload Identity Federation** (OIDC):
1. GitHub/GitLab → OIDC token → trusted by GCP Workload Identity Pool
2. Pool impersonates GSA with `roles/artifactregistry.writer` on the specific repository
3. No exported keys, no credentials management

## Production Best Practices

Based on analysis of real-world GCP deployments and the research findings, these are the non-negotiable patterns for production use:

### 1. Enforce Cleanup Policies

**The problem:** Every CI build pushes a new image tag. Storage costs grow indefinitely at $0.10/GB/month.

**The solution:** Cleanup policies with a **two-policy pattern**:

```json
[
  {
    "name": "delete-old",
    "action": {"type": "Delete"},
    "condition": {"olderThan": "30d", "tagState": "ANY"}
  },
  {
    "name": "keep-recent",
    "action": {"type": "Keep"},
    "mostRecentVersions": {"keepCount": 10}
  }
]
```

**How it works:** "Keep" policies always override "Delete" policies. This translates to: *"Delete everything older than 30 days, except the 10 most recent versions."*

### 2. Enable Immutable Tags (Docker Only)

For Docker repositories, enable `immutable_tags` to prevent tag mutation:

```hcl
docker_config {
  immutable_tags = true
}
```

This prevents production incidents where a `:latest` tag is accidentally overwritten, breaking deployments.

### 3. Use Repository-Level IAM

Grant the **minimum necessary permissions** at the repository level:

| Role | Use Case |
|------|----------|
| `roles/artifactregistry.reader` | GKE/Cloud Run runtime service accounts |
| `roles/artifactregistry.writer` | CI/CD pipeline service accounts |
| `roles/artifactregistry.repoAdmin` | Cleanup automation, human administrators |

Never grant project-level `roles/artifactregistry.admin` to workloads.

### 4. Co-locate with Consumers

Create repositories in the **same region** as their primary consumers:
- GKE cluster in `us-central1` → Artifact Registry in `us-central1`
- Cloud Run in `europe-west1` → Artifact Registry in `europe-west1`

This ensures free egress and enables GKE Image Streaming.

## Migration from GCR: The IaC-Friendly Path

If migrating from Google Container Registry, there are two paths. Only one is compatible with declarative Infrastructure as Code:

**❌ Anti-pattern:** Using `gcloud artifacts docker upgrade migrate`
- Creates "magic" `gcr.io` repositories outside IaC state management
- Causes immediate state drift
- Breaks declarative workflows

**✅ Best practice:** Manual, declarative migration
1. **Define** new `pkg.dev` repositories using Planton (or Terraform/Pulumi)
2. **Copy** existing images using `gcrane` or `docker pull/tag/push` scripts
3. **Update** all references (Kubernetes manifests, Cloud Build configs, CI/CD pipelines) to use `pkg.dev` URLs
4. **Lock down** GCR permissions after migration completes

This keeps infrastructure state clean and auditable.

## Evolution of Planton's API

The current `GcpArtifactRegistryRepoSpec` is intentionally minimal, focused on the most common use case: regional Docker registries for private workloads. This design reflects the "80/20 principle"—cover the 80% use case with 20% of the configuration surface.

**What's currently supported:**
- Repository format selection (Docker, Maven, npm, etc.)
- Regional deployment
- Public access control

**What's intentionally deferred:**
- Cleanup policies (should be added as a critical production feature)
- Docker-specific configuration (immutable tags)
- Remote and Virtual repository modes (platform features for advanced use cases)
- CMEK support (compliance/security requirement for regulated industries)

As Planton evolves, these capabilities can be added incrementally, following the proven API shape established by Terraform, Pulumi, and Config Connector.

## Conclusion: Artifact Registry as Infrastructure

The transition from GCR to Artifact Registry represents a broader shift in cloud architecture: from storage-backed abstractions to first-class, policy-driven services. The repository-level IAM, flexible location strategies, and advanced modes (Remote, Virtual) transform Artifact Registry from a simple Docker registry into a comprehensive artifact management platform.

For platform teams, this means the tool you choose to manage Artifact Registry matters immensely. Console-based workflows and imperative scripts may suffice for small-scale experiments, but production infrastructure demands declarative IaC with proper state management, immutability handling, and modular IAM patterns.

Planton embraces this declarative model, providing a clean abstraction that guides users toward secure, cost-effective, regional-first architecture while maintaining the flexibility to evolve toward advanced platform features as needs grow.

**The bottom line:** Deploy regionally, automate cleanup, authenticate keylessly, and manage it all declaratively. That's the path to production-ready Artifact Registry infrastructure.

