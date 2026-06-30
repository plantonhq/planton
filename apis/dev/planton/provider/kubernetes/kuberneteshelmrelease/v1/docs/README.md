# Helm Release Deployment: From Imperative Commands to Production-Grade Automation

## Introduction

"Just run `helm install` and you're done" might work for a developer's local cluster, but it's a recipe for disaster in production. The story of Helm deployment methods is the story of learning—often painfully—that managing application state in Kubernetes requires more than a powerful CLI. It requires continuous reconciliation, automated drift correction, and a single source of truth.

This document explores how the industry evolved from manual `helm` commands to the modern GitOps controllers that power production Kubernetes platforms today. Understanding this progression isn't just historical curiosity—it directly informs why Planton's HelmRelease API is designed the way it is, and what trade-offs we deliberately chose.

## The Evolution of Helm Deployment Methods

### Level 0: The Manual CLI Pattern (The Anti-Pattern)

Every Helm deployment starts the same way: someone runs `helm install my-app my-chart`. This works beautifully... once. The problems emerge immediately after:

**The Configuration Drift Problem**: A developer runs `helm upgrade my-app --set replicas=3` to fix a production issue. The cluster now runs 3 replicas, but Git still says 2. Which is the truth? When the next deploy happens, the replica count silently reverts to 2. One study found that 44% of developers have experienced drift-related outages.

**The "Where's My Release?" Problem**: Kubernetes doesn't track Helm releases in its API. If you lose your Helm state (stored in Secrets), you lose the ability to upgrade or rollback. Manual deployments have no audit trail—who deployed what, when, and why?

**The Verdict**: Manual Helm commands are excellent for learning and local development. They are actively harmful in shared or production environments.

### Level 1: The Scripting Layer (Shell Scripts and Makefiles)

The first "fix" teams implement is wrapping Helm commands in scripts. A typical Makefile might contain:

```makefile
deploy-prod:
    helm upgrade --install my-app my-chart \
        --values values-base.yaml \
        --values values-prod.yaml \
        --wait --timeout=10m
```

This achieves repeatability—the same script produces the same result. But it fundamentally fails to solve the state management problem. A script is a **push-based** mechanism: it runs, executes commands, and exits. It has no long-running process to monitor the cluster. If someone manually deletes a resource or changes a value after the script completes, that drift persists indefinitely.

**The Shell Escaping Nightmare**: Scripts that try to pass complex values through `--set` flags quickly encounter shell escaping hell. Passing JSON configuration or strings with quotes requires navigating multiple layers of escaping—shell syntax, Helm's value parser, and YAML's rules—often resulting in cryptic template rendering failures.

**The Verdict**: Scripting adds repeatability but doesn't add resilience. It's a small step forward, not a solution.

### Level 2: Infrastructure-as-Code (The External State Trap)

The next logical evolution is to treat Helm releases as resources in a broader IaC system like Terraform or Pulumi.

#### The Terraform Helm Provider: A Three-Body Problem

Terraform's `helm_release` resource appears ideal—Terraform is declarative, stateful, and production-proven. But the community verdict is brutal: "Do not use helm via terraform, just don't! It will ruin your day everyday."

The issue is a **three-body state problem**:
1. Terraform maintains state in its `tfstate` file
2. Helm maintains state in cluster Secrets
3. Kubernetes maintains actual resource state in etcd

When these three sources of truth inevitably desynchronize—during failed upgrades, network issues, or state corruption—the system enters a painful-to-debug failure mode. State lock contention, inability to apply plans, and manual intervention become routine.

#### Pulumi's Evolution: From Template-Only to Native SDK

Pulumi's journey provides a perfect case study. Pulumi initially offered a `Chart` resource that rendered Helm templates and managed resources directly, bypassing Helm's lifecycle. This "template-only" model broke Helm hooks—pre-install scripts, post-upgrade jobs, and lifecycle management that complex charts depend on.

In response, Pulumi introduced the `Release` resource, which uses the Helm SDK natively and "directly offloads all responsibility of orchestration and lifecycle management to Helm itself." This architectural shift—from external management to native integration—is the industry's verdict: **respect Helm's authority as a package manager**.

**The Verdict**: External IaC tools create competing sources of truth. The solution must be in-cluster, not external.

### Level 3: The Production Solution (In-Cluster GitOps Controllers)

This is where the evolution culminates. Modern production platforms use **in-cluster, Kubernetes-native controllers** that watch for Custom Resources and perform continuous reconciliation.

The architecture is elegant:
- A Custom Resource Definition (like FluxCD's `HelmRelease` or ArgoCD's `Application`) defines the desired state
- An in-cluster controller watches for these CRDs
- When a CRD is created or updated (typically via a Git commit), the controller uses the Helm SDK to install or upgrade the release
- The controller runs a **continuous reconciliation loop**, detecting and correcting drift automatically

This model solves every problem from the previous levels:
- **Drift**: The controller continuously compares desired state (in the CRD) with live state (in etcd) and auto-corrects
- **State**: The single source of truth is the declarative CRD, stored in Git
- **Lifecycle**: The controller uses the native Helm SDK, respecting the full chart lifecycle including hooks and tests

## GitOps Controllers: FluxCD vs ArgoCD

The two dominant, CNCF-graduated GitOps platforms are FluxCD and ArgoCD. While both achieve continuous reconciliation, they have fundamentally different philosophies about how to treat Helm charts.

### FluxCD: The "Helm-Native Actuator"

FluxCD is a modular GitOps toolkit. For Helm deployments, it uses two specialized controllers:
- **source-controller**: Fetches and caches Helm charts from repositories (HTTP or OCI)
- **helm-controller**: Watches `HelmRelease` CRDs and uses the Helm SDK to perform native Helm operations

**Philosophy**: FluxCD treats Helm as a trusted package manager. When reconciling a `HelmRelease`, it calls `helm install`, `helm upgrade`, and `helm rollback` via the SDK. This means all Helm hooks—pre-install, post-upgrade, test—execute correctly, just as if a human ran the Helm CLI.

**Drift Detection**: FluxCD reconciles at a defined interval. Since version 2.2.0, it includes robust drift detection, allowing it to "detect and correct cluster state drift from the desired release state" by re-applying the desired configuration.

**Strengths**:
- Full Helm lifecycle fidelity (hooks work correctly)
- Modular, toolkit-based architecture
- Excellent for platform teams building custom GitOps workflows

**Trade-offs**:
- CLI-first; web UI (like Capacitor) is optional and third-party
- Multi-tenancy requires manual Kubernetes RBAC configuration

### ArgoCD: The "Helm-as-Template" Platform

ArgoCD takes the opposite approach. Its documentation states unambiguously: **"Helm is only used to inflate charts with helm template. The lifecycle of the application is handled by Argo CD instead of Helm."**

When ArgoCD reconciles an `Application` CRD pointing to a Helm chart, it:
1. Runs `helm template` to generate raw Kubernetes manifests
2. Passes those manifests to its own GitOps engine
3. Performs a diff and synchronizes resources using its own logic

**Critical Trade-off**: This model **breaks Helm hooks**. Pre-install scripts, post-upgrade jobs, and lifecycle events don't execute. For charts that depend on hooks (many database charts use pre-install hooks for permission setup), this is a dealbreaker.

**Why Do This?** ArgoCD gains two significant advantages:
1. **Superior Diffing**: By rendering manifests itself, ArgoCD provides a precise, resource-by-resource diff in its web UI before syncing
2. **Multi-Source Composition**: ArgoCD can pull a chart from a public OCI registry and `valueFiles` from a separate, private Git repository—a powerful enterprise pattern for overlaying private configuration on public charts

**Strengths**:
- Rich, full-featured web UI (the primary adoption driver)
- Built-in multi-tenancy (AppProjects, RBAC, SSO)
- Multi-cluster management from a single control plane
- Self-healing with continuous drift correction

**Trade-offs**:
- No Helm hook support
- More opinionated and monolithic than Flux

### The Comparison Table

| Feature | FluxCD | ArgoCD | Recommendation |
|---------|--------|--------|----------------|
| **Helm Philosophy** | Helm-native (uses SDK) | Helm-as-template (renders only) | Flux for fidelity, Argo for features |
| **Helm Hooks** | Full support | None (uses its own sync hooks) | Flux if charts rely on hooks |
| **Value Overrides** | Inline or ConfigMaps/Secrets | Inline, valueFiles, or multi-source | Argo for enterprise overlay patterns |
| **User Interface** | CLI-first (third-party UIs) | Rich web UI (core feature) | Argo for teams wanting visual management |
| **Drift Detection** | Yes (reconciliation interval) | Yes (self-healing mode) | Both excellent |
| **Multi-Tenancy** | Manual (Kubernetes RBAC) | Built-in (AppProjects, RBAC, SSO) | Argo for large, multi-team platforms |
| **Multi-Cluster** | Per-cluster installation | Single control plane for fleet management | Argo for platform operators |
| **Production Use Case** | Platform teams, Helm-native fidelity | Application teams, UI-driven workflows | Choose based on team needs |

**The Strategic Choice**: FluxCD is best for teams that prioritize Helm-native fidelity and want a modular toolkit. ArgoCD is best for teams that want a feature-rich, UI-driven "platform-in-a-box" for application developers.

## Production-Grade Best Practices

### Version Pinning: The Immutability Principle

The cardinal rule of production deployments: **pin all versions**. A declarative manifest must be a precise, repeatable snapshot.

Helm uses semantic versioning (SemVer 2.0) for chart versions. Both the `version` field in `Chart.yaml` and the `image.tag` in `values.yaml` must be exact, immutable values like `1.1.5`, never floating tags like `latest` or version ranges like `^1.2.0`.

**Common Anti-Patterns**:
- **Floating Image Tags**: Using `latest` breaks declarative immutability. The artifact can change without any Git commit, making rollbacks and audits impossible.
- **SemVer Ranges in Production**: While useful in development or staging to absorb patch releases automatically, production manifests must be explicitly pinned. A Git commit should be required to change versions—that's the correct GitOps pattern.

**Planton's Choice**: Our `chart_version` field is a required string that does not resolve ranges. This enforces immutability at the API level, forcing explicit version changes through Git commits.

### Secrets Management: The Decoupled Pattern

Storing plain-text secrets in `values.yaml` files and committing them to Git is a critical security violation. A robust solution allows secrets to be stored in Git, but in an encrypted state. The architectural question is: which component decrypts them?

#### Pattern 1: Coupled Decryption (Mozilla SOPS)

SOPS encrypts values within a YAML file (not the whole file). A developer runs `sops -e values.yaml` and commits the encrypted result. The GitOps controller must be SOPS-aware and have access to decryption keys (AWS KMS, GCP KMS, PGP) to decrypt the file in-memory before passing values to Helm.

**Impact**: This couples the HelmRelease controller to the decryption process, dramatically increasing complexity and the controller's security surface.

#### Pattern 2: Decoupled Materialization (Sealed Secrets and External Secrets Operator)

This is the cleaner, more composable pattern. It delegates secret management to a specialized controller:

**Bitnami Sealed Secrets**:
- Uses asymmetric encryption (public/private key pair)
- Developer encrypts a standard Secret manifest with `kubeseal` CLI using a public key
- Commits the resulting `SealedSecret` CRD to Git
- In-cluster `sealed-secrets-controller` decrypts it with its private key and materializes a native Kubernetes Secret

**External Secrets Operator (ESO)**:
- Syncs secrets from external stores (AWS Secrets Manager, Azure Key Vault, HashiCorp Vault)
- Developer commits an `ExternalSecret` CRD that defines what to fetch
- ESO controller authenticates to the external store and materializes a native Secret

In both cases, the HelmRelease controller is **completely unaware** of the encryption. The `values.yaml` simply references the materialized Secret using standard chart patterns like `existingSecretName: my-app-secret`.

| Method | Encryption Model | Impact on HelmRelease API |
|--------|------------------|---------------------------|
| **Sealed Secrets** | Asymmetric (public/private key) | **None** (references native Secret) |
| **SOPS** | Symmetric (via Cloud KMS, PGP) | **High** (controller must decrypt) |
| **External Secrets Operator** | None (syncs from external stores) | **None** (references native Secret) |

**Planton's Choice**: We follow the decoupled model. Planton's HelmRelease API does not implement SOPS-style decryption. We document integration with ESO and Sealed Secrets as the recommended patterns. This keeps our API simple, secure, and focused on deploying charts.

### Release Lifecycle: Atomic Operations and Rollbacks

Production deployments must be transactional. A partially-applied, broken release is unacceptable.

**Key Helm Flags**:
- `--wait`: Instructs Helm to wait (up to a timeout) for all resources to become "ready" before marking the release successful. Without this, Helm "fires and forgets."
- `--atomic`: Enables `--wait` and automatically triggers a rollback to the previous revision if the wait condition fails. This prevents releases from being left in a broken state.
- `--timeout`: Specifies how long to wait before failing the operation.

**Planton's Choice**: Our HelmRelease API includes these as first-class fields (`atomic`, `wait`, `timeout`). They are not "advanced" features—they are core requirements for production safety.

### Chart Repositories: OCI vs HTTP

Helm supports two repository models:

**HTTP Repositories** (Traditional): Simple HTTP servers hosting `.tgz` charts and an `index.yaml` catalog.

**OCI Registries** (Modern): Stores charts in OCI-compliant registries (Harbor, Amazon ECR, Google Artifact Registry) as standard artifacts alongside container images.

**Benefits of OCI**:
- Unified registry and authentication model for both images and charts
- Content-addressable, immutable storage
- Re-uses the battle-tested security models of container registries

**Planton's Choice**: Our `repo` field supports both `https://` and `oci://` prefixes. OCI support is non-negotiable for a modern deployment tool.

## What Planton Supports (And Why)

Planton's HelmRelease API is intentionally minimal, opinionated, and focused on the 80% use case.

### The Core Specification

Our `HelmReleaseSpec` includes:
- **`repo`**: The chart repository URL (supports `https://` and `oci://`)
- **`name`**: The chart name
- **`version`**: The exact chart version (no SemVer range resolution)
- **`values`**: A map of key-value pairs for customization

This maps directly to Helm's "Three Big Concepts": Repository (where), Chart (what), and Release (the instance).

### What We Deliberately Omit from V1

To keep the API clean and implementable, we exclude 20% "power-user" features:

**Post-Renderers**: Allows piping Helm's output through another tool (like Kustomize) for last-mile changes. This is an escape hatch for poorly-designed charts and would require the controller to manage external binaries, STDIN/STDOUT piping, and complex error handling.

**Helm Hook Customization**: Advanced debugging features like disabling specific hooks suggest a problem with the chart, not the deployment.

**`tpl` Function in Values**: Passing template strings as values leads to un-debuggable, "meta-templated" charts. Values should contain data, not code.

### The Philosophy: Simple, Secure, Production-Ready

Planton's HelmRelease API follows the **GitOps pull model**. CI pipelines do not run `helm install`. Instead, they:
1. Build and push container images
2. Commit changes to Git (updating `chart_version` or `image.tag` in the manifest)

The HelmRelease controller detects the Git commit and pulls the new configuration, triggering the Helm deployment.

This model ensures:
- **Single Source of Truth**: Git is the source of truth, not CI logs
- **Auditability**: Every deployment is a Git commit with a full history
- **Declarative State**: The cluster converges toward the desired state in Git

## Ecosystem Integration

### CI/CD Pipelines

Modern GitOps pipelines (GitHub Actions, GitLab CI, Tekton) follow the pull model. The CI system's responsibility is to:
1. Run tests and build artifacts
2. Update a Git repository with the new chart version or image tag
3. Let the HelmRelease controller handle the actual deployment

**Environment Promotion**: Promoting from dev → staging → prod is a series of Git commits or pull requests. A developer opens a PR to change the production manifest to the chart version validated in staging.

### Monitoring and Observability

**Application Monitoring**: The Helm chart should expose application-level metrics (e.g., Prometheus endpoints) configured via values.

**Release Monitoring**: The HelmRelease controller must expose a `/metrics` endpoint with Prometheus metrics like `planton_helm_release_reconciliation_status`. This allows operators to monitor the health of the deployment system itself—"Did the last 10 upgrades succeed?"

### Dependency Management

"Umbrella charts" (charts that list other charts as dependencies in `Chart.yaml`) are fully supported. The HelmRelease controller doesn't need to be aware of this—Helm itself resolves, downloads, and installs dependencies when installing the parent chart.

### Licensing

The entire Helm deployment ecosystem is open source:
- **Helm**: Apache 2.0
- **FluxCD**: Apache 2.0
- **ArgoCD**: Apache 2.0

This de-risks adoption and ensures no vendor lock-in. Both FluxCD and ArgoCD are CNCF Graduated projects with strong commercial support markets (Akuity, Codefresh, and InfraCloud for Argo; Weaveworks historically for Flux).

## Conclusion: Why Declarative, In-Cluster, Helm-Native

The evolution from manual commands to in-cluster controllers isn't arbitrary—it's the result of learning from production failures. Manual commands drift. Scripts can't reconcile. External IaC creates state conflicts. The only architecturally sound solution is a Kubernetes-native controller that runs inside the cluster, uses the Helm SDK natively, and continuously reconciles desired state with actual state.

Planton's HelmRelease API embodies this lesson. We provide a minimal, production-ready interface that:
- Enforces version pinning through a non-ranged `version` field
- Decouples secrets management (no built-in SOPS, integrate with ESO or Sealed Secrets)
- Supports modern OCI registries alongside traditional HTTP repos
- Includes safety-critical flags (`atomic`, `wait`, `timeout`) as first-class fields

This isn't the API that deploys the most charts—it's the API that deploys charts safely, predictably, and at scale.

