# ExternalDNS Deployment on Kubernetes

## Introduction

For years, managing DNS records for Kubernetes services was a manual chore. Developers would spin up a service, grab its external IP or load balancer hostname, then context-switch to their DNS provider's console to create or update records. This workflow was error-prone, slow, and fundamentally at odds with the declarative nature of Kubernetes.

ExternalDNS changed this paradigm. By watching Kubernetes resources (Ingresses, Services, and more) and automatically synchronizing DNS records to external providers like AWS Route53, Google Cloud DNS, Azure DNS, and Cloudflare, it eliminates the manual DNS management loop entirely. You declare your desired DNS name in an annotation or hostname field, and ExternalDNS ensures it exists and points to the right target.

But here's the challenge: ExternalDNS itself needs to be deployed, configured with provider credentials, scoped to appropriate zones, and managed across potentially dozens of clusters spanning multiple cloud providers. This document explores the landscape of ExternalDNS deployment methods, from anti-patterns to production-ready approaches, and explains the design philosophy behind Planton's ExternalDNS module.

## The Maturity Spectrum: From Manual YAML to Production Automation

### Level 0: The Anti-Pattern (Raw YAML Manifests)

The simplest way to deploy ExternalDNS is to copy a YAML manifest from the documentation, fill in your provider credentials, and `kubectl apply` it. This works for a single-cluster proof-of-concept, but it falls apart at scale:

- **No version control or reusability**: Each cluster gets a slightly different configuration as manual edits accumulate
- **Credential sprawl**: Hard-coded secrets in YAML files (or worse, in Git)
- **No upgrade path**: When ExternalDNS releases a new version, you manually edit every cluster's manifest
- **Configuration drift**: Zone filters, RBAC permissions, and provider-specific flags diverge across environments

Raw manifests are brittle. They're fine for learning, but they don't belong in production multi-cluster environments.

### Level 1: Helm Charts (The Production Standard)

Helm emerged as the de facto packaging format for Kubernetes applications, and ExternalDNS is no exception. Two Helm charts dominate:

**Official kubernetes-sigs/kubernetes-external-dns chart:**
- Published by the Kubernetes SIG ExternalDNS maintainers
- Reflects upstream defaults and latest features
- Minimal opinions on secret management (you create secrets separately)
- Close alignment with ExternalDNS CLI flags

**Bitnami's kubernetes-external-dns chart:**
- Enterprise-focused with extensive configuration options
- Includes security hardening (Bitnami Secure Images with zero CVEs)
- Built-in Prometheus metrics, ServiceMonitor support
- Richer templating for AWS, Azure, GCP-specific features

Both charts are **production-ready and Apache 2.0 licensed**. The choice often comes down to whether you prefer upstream simplicity (official chart) or enterprise conveniences (Bitnami). In practice, either works well—what matters is consistency across your fleet.

Helm solves the reusability problem: you define values once, version them in Git, and apply them across clusters. Upgrades become `helm upgrade`. This is the baseline for serious deployments.

### Level 2: GitOps (Declarative Fleet Management)

Helm gets you to reusable configurations, but GitOps takes it further: treat your cluster's entire desired state as code in Git, and let a tool like ArgoCD or Flux continuously reconcile reality with that declaration.

For ExternalDNS, this means:
- Store Helm values (or raw manifests) in Git per cluster or environment
- Let ArgoCD/Flux detect drift and automatically sync changes
- Gain audit trails, rollback capabilities, and multi-cluster orchestration

GitOps is a **best practice** for multi-cloud deployments. You might have `prod-us-west-eks/kubernetes-external-dns-values.yaml` and `prod-eu-gke/kubernetes-external-dns-values.yaml`, each pointing to the appropriate DNS zone and using cloud-native credentials. Changes go through pull requests, not `kubectl apply` sessions.

ExternalDNS requires no special integration here—it's just another Kubernetes application managed declaratively. The power is in the operational model: your entire DNS automation layer is versioned, reviewable, and reproducible.

### Level 3: IaC Integration (Terraform/Pulumi)

For teams that provision clusters with infrastructure-as-code tools like Terraform or Pulumi, deploying ExternalDNS via the same tool is a natural extension. Both tools have Kubernetes and Helm providers, allowing you to:

- Create the DNS zone (e.g., Route53 hosted zone, Cloud DNS managed zone)
- Provision the cluster (EKS, GKE, AKS)
- Set up cloud-native identity bindings (IRSA role, Workload Identity, Managed Identity)
- Deploy ExternalDNS via Helm with the right configuration—all in one workflow

Published modules (like Gruntwork's EKS modules) often encapsulate this pattern, ensuring that ExternalDNS is deployed with secure defaults (IRSA, zone scoping, upsert-only policy) as part of cluster bootstrapping.

**Caveat:** IaC deployments require careful secret management (avoid committing credentials) and awareness that ExternalDNS configuration changes may require a deployment rollout. But for multi-cluster automation, IaC integration is extremely powerful—it ties DNS lifecycle to infrastructure lifecycle.

### Verdict: Helm + GitOps or IaC

For production, the maturity spectrum converges on **Helm as the deployment mechanism**, with GitOps or IaC providing the orchestration layer. Raw YAML manifests are for learning only. The choice between GitOps and IaC often depends on your team's existing tooling—either is production-grade.

## Multi-Cloud Authentication: The Cloud-Native Way

One of ExternalDNS's strengths is its broad provider support, but each cloud has distinct authentication mechanisms. Modern best practice is to use **cloud-native identity** (short-lived tokens, no long-lived secrets) whenever possible.

### AWS Route53: IRSA (IAM Roles for Service Accounts)

On EKS, ExternalDNS should use **IRSA** to assume an IAM role scoped to specific Route53 zones. No AWS access keys in the cluster—just an annotation on the ExternalDNS service account mapping it to a role ARN. The IAM policy grants `route53:ChangeResourceRecordSets` on specific hosted zone ARNs (not wildcard), enforcing least privilege.

**Why this matters:** Static credentials can be compromised or leaked. IRSA tokens are short-lived and automatically rotated, and the role can be restricted to exactly the zones ExternalDNS manages.

### Google Cloud DNS: Workload Identity

On GKE, **Workload Identity** links a Kubernetes service account to a Google Cloud service account with DNS Administrator role on the target Cloud DNS project. ExternalDNS pods acquire credentials automatically via metadata server tokens—no JSON keys needed.

The setup involves:
- Creating a GCP service account with `roles/dns.admin` on the DNS project
- Binding it to the Kubernetes service account via IAM policy (`roles/iam.workloadIdentityUser`)
- Annotating the KSA with `iam.gke.io/gcp-service-account=<GSA_EMAIL>`

This eliminates key sprawl and ties DNS permissions to the pod's identity.

### Azure DNS: Managed Identity

On AKS, **Azure Managed Identity** (especially user-assigned) is the equivalent. You create a managed identity, grant it "DNS Zone Contributor" role on the target Azure DNS zone, and establish a federated identity credential linking it to the ExternalDNS service account.

Azure AD issues tokens to the pod, allowing it to modify DNS records without any stored secrets. This is the modern approach—service principals with client secrets are fallback options, not the default.

### Cloudflare: Scoped API Tokens

Cloudflare doesn't have ambient cluster identity, so you'll use an **API token**. The key is to scope it tightly: grant only `Zone:DNS:Edit` and `Zone:Zone:Read` permissions, and limit the token to specific zones.

Store the token in a Kubernetes Secret and mount it as an environment variable (`CF_API_TOKEN`). While not as dynamic as IRSA/Workload Identity, a scoped token significantly limits blast radius compared to a global API key.

### The Pattern: Ambient Identity First, Secrets as Fallback

Across all providers, the production pattern is:
1. Use cloud-native identity (IRSA, Workload Identity, Managed Identity) if available
2. Scope credentials to specific DNS zones, never `*`
3. Store any required secrets in Kubernetes Secrets (or external secret managers)
4. Rotate and audit credentials regularly

Planton's API reflects this philosophy: it asks for the zone ID (scoping) and, where applicable, the identity binding (IRSA role ARN, managed identity client ID). Static credentials are not the default path.

## The 80/20 Configuration Principle

ExternalDNS exposes dozens of flags and configuration options, but **most production deployments use a small core subset**. Planton's API is designed around these essentials:

### Essential Fields (The 20% That Covers 80%)

**DNS Zone Identifier:**
- AWS: Route53 hosted zone ID
- GCP: Cloud DNS managed zone ID (plus project ID)
- Azure: DNS zone resource name
- Cloudflare: Zone ID

This single reference tells ExternalDNS what domain space it's responsible for, and drives automatic domain filtering.

**Authentication Mechanism:**
- AWS: IRSA role ARN (or auto-created)
- GCP: Workload Identity (inferred from project)
- Azure: Managed Identity client ID
- Cloudflare: API token

These tie ExternalDNS to cloud credentials. In most cases, the platform can auto-configure these (e.g., create an IRSA role automatically), so the user only provides overrides if needed.

**Deployment Metadata:**
- Namespace (default: `kubernetes-external-dns`)
- ExternalDNS version (default: `v0.19.0`)
- Helm chart version (default: `1.19.0`)

These allow version pinning and namespace isolation but have sane defaults.

**Provider-Specific Toggles:**
- Cloudflare: `is_proxied` (enable orange-cloud CDN proxy)
- Others: generally no toggles needed—safe defaults apply

### What's Not Exposed (The 80% You Don't Need)

**Policy and Registry:** Planton always uses `upsert-only` policy (no accidental deletions) and TXT registry (ownership tracking). These are best practices, not configurable knobs.

**Domain Filters:** Automatically derived from the zone ID—no need to manually specify domain patterns.

**TXT Owner ID:** Auto-generated from cluster name to prevent conflicts.

**Sources:** Default to Service and Ingress (the 99% use case). Advanced sources (Gateway API, CRDs) can be added via Helm value overrides if needed.

**RBAC:** Automatically configured with minimal necessary permissions.

By exposing only the essential fields, Planton's API achieves the 80/20 goal: simple for the common case, extensible for edge cases via Helm values.

## Production Best Practices Baked In

### Zone Scoping and Ownership

Every ExternalDNS instance is filtered to a specific DNS zone (via `--zone-id-filter` or `--domain-filter`), preventing it from managing unintended records. TXT ownership records (`kubernetes-external-dns/owner=<cluster-id>`) ensure that multiple ExternalDNS instances never conflict—one won't delete records created by another.

### Least Privilege RBAC

ExternalDNS only needs read access to Kubernetes resources (Services, Ingresses, Endpoints). It should never have write access to cluster objects. The Helm chart includes a ClusterRole with minimal permissions, bound to a dedicated service account.

### Observability

Prometheus metrics are enabled by default (exposed on port 7979), covering sync status, error counts, and record changes. Logging is set to `info` level, capturing every DNS operation for audit trails.

### Failure Modes and Alerts

If ExternalDNS crashes or loses cloud permissions, it stops syncing DNS—but existing records remain intact (no deletions with upsert-only policy). Monitoring should alert on:
- Pod unavailability
- Cloud API errors in logs (e.g., `AccessDenied`)
- Sync failures (via metrics)

A common canary test: deploy a dummy Ingress and verify DNS appears within the sync interval.

### Multi-Cluster Isolation

In multi-cluster setups, each cluster runs its own ExternalDNS, scoped to a distinct subdomain or zone:
- Prod-US cluster manages `*.us.prod.example.com`
- Prod-EU cluster manages `*.eu.prod.example.com`

This prevents cross-cluster interference and limits the blast radius of misconfigurations.

## The Planton Approach

Planton's `KubernetesExternalDns` module encapsulates the production patterns described above:

**Deployment Method:** Helm (official kubernetes-sigs chart) deployed via Pulumi/Terraform, with GitOps-friendly output.

**Authentication:** Cloud-native by default (IRSA on EKS, Workload Identity on GKE, Managed Identity on AKS), with automatic role/identity creation if not provided.

**Configuration Surface:** Minimal—just the zone ID and any provider-specific overrides. Policy, filters, and RBAC are pre-configured with best practices.

**Multi-Cloud:** Provider-specific configurations (EKS, GKE, AKS, Cloudflare) are expressed as a `oneof` in the protobuf spec—clear, type-safe, and scoped to what each provider actually needs.

**Extensibility:** While the API is minimal, it generates Helm values that can be extended. Power users can layer additional customizations via Helm value overrides without polluting the base API.

### Example Configurations

**GKE with Workload Identity:**
```yaml
spec:
  target_cluster: <cluster-reference>
  namespace: kubernetes-external-dns
  kubernetes_external_dns_version: v0.19.0
  gke:
    project_id: my-gcp-project
    dns_zone_id: my-zone-id
```

Planton will:
- Create a GCP service account with DNS Administrator role
- Bind it to the ExternalDNS Kubernetes service account via Workload Identity
- Deploy ExternalDNS with `--provider=google`, `--google-project=my-gcp-project`, `--zone-id-filter=my-zone-id`
- Set upsert-only policy, enable metrics, and configure TXT ownership

**EKS with IRSA:**
```yaml
spec:
  target_cluster: <cluster-reference>
  eks:
    route53_zone_id: Z1234567890ABC
    # irsa_role_arn_override: <optional-custom-role>
```

If `irsa_role_arn_override` is omitted, Planton creates an IAM role with a policy scoped to that zone, annotates the service account, and deploys ExternalDNS.

**Cloudflare:**
```yaml
spec:
  target_cluster: <cluster-reference>
  cloudflare:
    api_token: <scoped-token>
    dns_zone_id: <zone-id>
    is_proxied: true
```

Deploys ExternalDNS with Cloudflare provider, proxied mode enabled (traffic through Cloudflare's edge), filtered to the specified zone.

## Conclusion: DNS Automation as Infrastructure Code

ExternalDNS transformed Kubernetes DNS from a manual bottleneck to an automated, declarative process. But deploying ExternalDNS itself at scale—across clouds, clusters, and zones—requires careful attention to authentication, scoping, and operational best practices.

The maturity spectrum runs from raw YAML (a learning tool, not a production strategy) to Helm-based deployments orchestrated via GitOps or IaC (the production standard). Cloud-native authentication (IRSA, Workload Identity, Managed Identity) eliminates credential sprawl, and tight zone scoping prevents cross-cluster conflicts.

Planton's approach is to **expose the essential 20% of configuration** that covers 80% of real-world deployments, while embedding production best practices (upsert-only policy, TXT ownership, least-privilege RBAC, observability) as non-negotiable defaults. The result is a minimal API surface that's powerful enough for multi-cloud production environments, yet simple enough that a developer can deploy ExternalDNS to a new cluster with just a zone ID and a provider type.

DNS automation should be invisible infrastructure—reliable, secure, and boring in the best way. This module aims to make it exactly that.

---

For detailed configuration options and advanced scenarios, see the protobuf spec in `spec.proto` and the generated Pulumi/Terraform modules in the `iac/` directory.

