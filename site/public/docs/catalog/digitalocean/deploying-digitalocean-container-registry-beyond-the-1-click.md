---
title: "Deploying DigitalOcean Container Registry: Beyond the \"1-Click\""
description: "Deploying DigitalOcean Container Registry: Beyond the \"1-Click\" deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceancontainerregistry"
---

# Deploying DigitalOcean Container Registry: Beyond the "1-Click"

## Introduction

Container registries have become the unsung heroes of modern cloud infrastructure. They're not flashy—no one writes blog posts celebrating the perfect registry configuration—but they're absolutely critical. A container registry sits at the junction of development and deployment, storing the artifacts that represent your application's entire lifecycle. When it works well, it's invisible. When it fails or becomes expensive due to misconfiguration, it brings CI/CD pipelines to a halt and racks up surprise cloud bills.

**DigitalOcean Container Registry (DOCR)** is DigitalOcean's answer to this need—a managed, private container registry designed not to compete on features with enterprise-grade solutions like Harbor, but to **reduce friction** for teams already running workloads on DigitalOcean Kubernetes (DOKS) or the App Platform. The value proposition is simple: co-located storage means faster image pulls, transparent pricing means no surprises, and zero operational overhead means your team spends time building applications, not babysitting registries.

DOCR is what you might call a "glue service"—it exists to make other DigitalOcean services work better. It's OCI-compliant, supports Helm charts as first-class artifacts, and integrates natively with DOKS through a "1-click" feature that patches imagePullSecrets across all namespaces. For App Platform users, it offers something even more compelling: **auto-deployment**, where a `docker push` to DOCR automatically triggers a new deployment—no intermediate CI/CD pipeline required. This feature is exclusive to DOCR; Docker Hub and GitHub Container Registry don't get it.

But here's the catch: that simplicity comes with sharp edges. DOCR has no native image signing. Vulnerability scanning requires external Snyk integration. Garbage collection isn't automatic—it's a manual, on-demand operation that puts your registry into **read-only mode**, breaking CI/CD pipelines if you run it during peak hours. And the most powerful integrations—the DOKS "1-click" button and scheduled garbage collection—are completely missing from every major Infrastructure-as-Code (IaC) tool. Terraform and Pulumi can create the registry, but they can't replicate what the web console does with a single button click.

This guide explains the landscape of deployment methods for DOCR, from the anti-patterns to avoid to the production-ready approaches that actually work at scale. We'll show how OpenMCF bridges the IaC gap by providing declarative abstractions for the features Terraform and Pulumi can't handle: automated garbage collection scheduling and DOKS integration. The goal isn't just to provision a registry—it's to provision a registry that stays healthy, secure, and cost-effective over time.

---

## The Deployment Spectrum: From ClickOps to Production IaC

Not all approaches to managing container registries are created equal. Here's how the methods stack up, from what to avoid to what works at scale:

### Level 0: The Web Console (The Anti-Pattern)

**What it is:** Using the DigitalOcean control panel to click through registry creation, selecting a subscription tier, choosing a region, and hitting "Create."

**What it solves:** Learning how DOCR works. The console is beautifully simple: name your registry, pick a tier (Starter/Basic/Professional), select a region, and you're done. For exploring the platform or running a quick proof-of-concept, it's perfectly acceptable.

**What it doesn't solve:** Repeatability, auditability, and operational safety. The real danger isn't the initial creation—it's the **stateful drift** created by post-provisioning actions. That "Integrate with Kubernetes" button in the UI? It patches service accounts across all namespaces, creating imagePullSecrets that are invisible to any IaC tool. That "Start Garbage Collection" button? It's a manual operation you'll forget to run, leading to uncontrolled storage growth and surprise bills. And when you inevitably run GC during business hours, it puts the registry into read-only mode, breaking every CI/CD pipeline trying to push images.

**The Operational Hazard:** Storage bloat is the default behavior. Every time you push a new image with an existing tag (like `latest`), the old image becomes an "untagged manifest" that **still consumes storage**. Without scheduled garbage collection, you're paying for every orphaned layer, and DOCR won't warn you until your bill jumps.

**Verdict:** Use the console to understand the workflow and test features, but **never for staging or production**. If you can't codify it, you can't reproduce it, audit it, or safely hand it off to another engineer.

---

### Level 1: CLI Scripting (Automation Without State)

**What it is:** Using the `doctl` CLI to script registry provisioning and operations:

```bash
# Create a registry
doctl registry create my-app-registry --region nyc3 --subscription-tier professional

# Login to Docker
doctl registry login

# Integrate with DOKS
doctl kubernetes cluster registry add <cluster-uuid>

# Run garbage collection
doctl registry garbage-collection start --include-untagged-manifests
```

**What it solves:** Automation. You can script every lifecycle operation: create, login, integrate, garbage collect. The CLI is well-designed, supports JSON output for parsing, and exposes features that the IaC tools don't (like the DOKS integration command `cluster registry add`).

**What it doesn't solve:** State management and idempotency. Scripts are imperative, not declarative. Run them twice and you'll get errors because resources already exist. Cleanup on failure is manual. There's no state file to track what was created or what changed. And you're still orchestrating multiple, unrelated commands (create registry, integrate with DOKS, schedule GC) with no unified abstraction.

**The Integration Challenge:** The `doctl kubernetes cluster registry add` command is the only way to replicate the "1-click" UI integration via automation. But it's a one-time, imperative action. If you run it in a script, then later modify the registry or cluster outside that script, your script has no way to detect or correct drift.

**Verdict:** Acceptable for throwaway dev environments or one-off migrations. **Not suitable for production**, where you need plan/preview workflows, rollback capabilities, and reliable drift detection.

---

### Level 2: Direct API Integration (Maximum Flexibility, Maximum Complexity)

**What it is:** Calling the DigitalOcean REST API directly from custom tooling, configuration management systems (Ansible), or internal orchestration frameworks.

**Example (cURL):**

```bash
curl -X POST "https://api.digitalocean.com/v2/registry" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app-registry",
    "subscription_tier_slug": "professional",
    "region": "nyc3"
  }'
```

**What it solves:** Total control. You can integrate DOCR provisioning into any HTTP-capable tool. The API is well-documented, returns structured JSON, and exposes endpoints for every operation: create registry, generate Docker credentials, start garbage collection, integrate with DOKS.

**What it doesn't solve:** Abstraction, state management, and idempotency. You're handling HTTP calls, authentication (API tokens in headers), error handling (rate limits, transient failures), and sequencing (create registry → wait for it to be ready → integrate with DOKS → schedule GC). You're essentially building your own IaC layer from scratch.

**The SDK Alternative:** DigitalOcean provides official SDKs (`godo` for Go, `pydo` for Python) that wrap the API and handle authentication and retries. But they're still low-level—you're calling individual methods, not declaring desired state. And the API is split across services: `client.Registry` manages the registry, `client.Kubernetes` manages DOKS integration, and garbage collection has its own endpoints. Any robust controller must orchestrate all three.

**Verdict:** Useful if you're building a custom provisioning platform or integrating DOCR into a broader orchestration framework (like OpenMCF). For most teams, higher-level IaC tools (Terraform, Pulumi) handle the API complexity for you.

---

### Level 3: Infrastructure-as-Code (Production-Ready, With Gaps)

**What it is:** Using Terraform, Pulumi, or OpenTofu with the official DigitalOcean provider to declaratively define and manage DOCR.

**Terraform example:**

```hcl
provider "digitalocean" {
  token = var.digitalocean_token
}

resource "digitalocean_container_registry" "main" {
  name                   = "my-app-registry"
  subscription_tier_slug = "professional"
  region                 = "nyc3"
}

resource "digitalocean_container_registry_docker_credentials" "main" {
  registry_name = digitalocean_container_registry.main.name
  write         = true
}
```

**Pulumi example (TypeScript):**

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const registry = new digitalocean.ContainerRegistry("main", {
    name: "my-app-registry",
    subscriptionTierSlug: "professional",
    region: "nyc3",
});
```

**What it solves:** Everything you'd expect from production IaC:
- **Declarative state**: Describe what you want, not how to get there
- **Plan/preview**: See changes before applying them
- **State management**: Track resource IDs, detect drift, enable safe updates
- **Version control**: Treat infrastructure as code, with diffs, reviews, and rollbacks
- **Multi-environment support**: Reuse configs for dev/staging/prod with different parameters

**What it doesn't solve: The IaC Gap.** This is the critical finding from the research. The Terraform and Pulumi providers only expose the **trivial provisioning API**: name, subscription tier, and region. Three fields. That's it. The two most important operational features are completely missing:

1. **DOKS Integration**: There is no resource that replicates the "1-click" DOKS integration. The workaround is a fragile, multi-resource chain:
   - Create `digitalocean_container_registry`
   - Create `digitalocean_container_registry_docker_credentials` (which is misleadingly named—it just fetches temporary credentials, it doesn't integrate with DOKS)
   - Create a `kubernetes_secret` resource using those credentials
   - Manually patch the `default` service account in **every namespace** to use that secret as an `imagePullSecret`
   
   This is complex, error-prone, and breaks if you later integrate via the UI or CLI, creating invisible drift.

2. **Garbage Collection**: There is no resource or function to manage garbage collection. GC is an imperative, on-demand API call (`POST /v2/registry/{name}/garbage-collection`). You can't enable it, disable it, or schedule it via Terraform or Pulumi. If you click "Start GC" in the UI, your IaC tool has no idea it happened.

**The Parity Breakdown:**

| Function                        | UI/Console | doctl CLI | Terraform/Pulumi |
|---------------------------------|------------|-----------|------------------|
| Create Registry                 | ✅         | ✅        | ✅               |
| Get Docker Credentials          | ✅         | ✅        | ✅ (misleading)  |
| DOKS "1-Click" Integration      | ✅         | ✅        | ❌               |
| Start Garbage Collection        | ✅         | ✅        | ❌               |
| View GC Status                  | ✅         | ✅        | ❌               |

**Verdict:** The production standard for **provisioning**, but critically incomplete for **operations**. Terraform and Pulumi are mature, reliable, and well-supported, but they only solve 50% of the problem. The other 50%—DOKS integration and automated GC—requires custom tooling or manual intervention.

---

## IaC Tool Comparison: Terraform vs. Pulumi vs. OpenTofu

For the core task of provisioning DOCR (name, tier, region), there's no meaningful difference in capability. All three tools are production-ready.

### Terraform: The Battle-Tested Standard

**Maturity:** The `digitalocean/digitalocean` provider is mature, heavily used, and officially maintained. It covers all basic DOCR operations.

**Configuration Model:** Declarative HCL. You define resources, Terraform builds the dependency graph and execution order.

**State Management:** Local or remote backends (S3, Terraform Cloud). State tracks resource IDs for safe updates.

**Strengths:**
- Broad ecosystem and community support
- Familiar to most ops teams
- Clear plan/apply workflow
- Excellent documentation

**Limitations:**
- HCL is less expressive than a full programming language (limited conditionals, no complex loops)
- The DOCR provider is incomplete (missing DOKS integration and GC resources)

**Verdict:** The default choice for teams already using Terraform or prioritizing ecosystem maturity.

---

### Pulumi: The Programmer's IaC

**Maturity:** Newer than Terraform, but production-ready. The `@pulumi/digitalocean` package is **based on the Terraform provider** (via a bridge), so it has equivalent resource coverage and inherits the same gaps.

**Configuration Model:** Real programming languages (TypeScript, Python, Go). Write infrastructure as code with loops, conditionals, and unit tests.

**State Management:** Pulumi Cloud or self-managed backends (S3, Azure Blob). Similar to Terraform.

**Strengths:**
- Full programming language expressiveness (easier to build dynamic configs)
- Better for complex orchestration logic
- Native testing frameworks

**Limitations:**
- Smaller community than Terraform
- Requires a runtime (Node.js, Python, etc.)
- **Bridged provider means it inherits Terraform's DOCR gaps** (no DOKS integration, no GC resources)

**Verdict:** Great if your team prefers coding infrastructure in familiar languages. Slightly more overhead for simple use cases, but more powerful for complex provisioning logic.

---

### OpenTofu: The Open-Source Fork

**Maturity:** OpenTofu is a community-driven fork of Terraform 1.5, created in response to Terraform's license change. It uses the **exact same DigitalOcean provider** as Terraform, so resource coverage and gaps are identical.

**Why it exists:** Terraform switched to a Business Source License (BSL) in 2023, which restricts commercial use. OpenTofu is fully open-source (MPL 2.0), making it a safer choice for teams concerned about license changes or vendor lock-in.

**Verdict:** Functionally equivalent to Terraform for DOCR provisioning. Choose OpenTofu if open-source licensing is a priority.

---

### Which Should You Choose?

- **Default to Terraform** if you want the most mature, widely-adopted solution with the largest community.
- **Choose Pulumi** if you prefer writing infrastructure in TypeScript/Python/Go and need advanced orchestration logic.
- **Choose OpenTofu** if open-source licensing is critical or you want to avoid future Terraform license risk.
- **All three have the same IaC gap**: None provide declarative resources for DOKS integration or automated garbage collection.

---

## The IaC Gap: What Terraform and Pulumi Can't Do

Let's be precise about what's missing and why it matters:

### Gap 1: DOKS Integration

**What the UI does:** You click "Integrate with Kubernetes" in the DigitalOcean console. Behind the scenes, DigitalOcean:
1. Generates a Kubernetes secret (`kubernetes.io/dockerconfigjson`) containing registry credentials
2. Creates that secret in **all namespaces** (existing and future)
3. Patches the `default` service account in **all namespaces** to use the secret as an `imagePullSecret`

**Result:** Every pod in every namespace can pull images from DOCR without any additional configuration.

**What the CLI does:** `doctl kubernetes cluster registry add <cluster-uuid>` does the exact same thing.

**What Terraform/Pulumi do:** Nothing. There's no resource for this. The documented workaround is:
1. Create `digitalocean_container_registry_docker_credentials` (fetches temporary credentials)
2. Create a `kubernetes_secret` resource with those credentials in a **specific namespace**
3. Manually configure each Deployment to use `imagePullSecrets`, or manually patch service accounts

**Why this is a problem:**
- It's complex and error-prone
- It only works for a single namespace unless you duplicate the resource for every namespace
- It doesn't handle future namespaces (the UI patches **all** namespaces, including ones created later)
- If you ever click the UI button or run the `doctl` command, your IaC tool has no way to detect or manage that state

**The GitHub issue:** [Terraform Issue #660](https://github.com/digitalocean/terraform-provider-digitalocean/issues/660) confirms this feature is missing and has been requested for years.

---

### Gap 2: Garbage Collection

**What the problem is:** Every time you push a new image with an existing tag (like `latest`), DOCR creates a new image manifest and marks the old one as "untagged." The untagged manifest and its unreferenced blob layers **still consume storage**. This is the default behavior. Garbage collection (GC) deletes these orphaned resources, but it's **not automatic**.

**What the UI does:** You click "Start Garbage Collection" in the console. DOCR:
1. Puts the registry into **read-only mode** (all `docker push` commands fail)
2. Scans for untagged manifests and unreferenced blobs
3. Deletes them
4. Returns the registry to normal mode

**What the CLI does:** `doctl registry garbage-collection start --include-untagged-manifests` does the same thing.

**What Terraform/Pulumi do:** Nothing. There's no resource or function to trigger GC, schedule GC, or even check GC status. It's a completely imperative, on-demand API call that isn't modeled declaratively.

**Why this is a problem:**
- **Storage bloat is the default.** Without scheduled GC, storage usage grows unbounded, and costs spiral.
- **Manual GC breaks CI/CD.** If you run GC during business hours, the read-only mode breaks every pipeline trying to push images.
- **No state tracking.** If you click "Start GC" in the UI, your IaC tool has no idea. Drift is invisible.

**The real requirement:** Production teams need **scheduled, automated GC** that runs during low-traffic windows (e.g., 3 AM Sunday). The `--include-untagged-manifests` flag is critical—without it, GC does nothing useful. But none of this is possible with Terraform or Pulumi alone.

---

## Production Essentials: Getting DOCR Right

### Subscription Tier Selection

For any production workload, the **Professional tier** ($20/month) is the only appropriate choice. Here's why:

| Tier         | Monthly Cost | Storage | Repositories | Max Registries | When to Use                          |
|--------------|--------------|---------|--------------|----------------|--------------------------------------|
| **Starter**  | $0           | 500 MiB | 1            | 1              | Hobby projects, initial trials       |
| **Basic**    | $5           | 5 GiB   | 5            | 1              | Dev/staging (single registry)        |
| **Professional** | $20      | 100 GiB | Unlimited    | **Up to 10**   | Production, multi-region, multi-env  |

The **multi-registry capability** (Professional only) is the key enabler for production architectures:
- **Multi-region deployments:** Create separate registries co-located with DOKS clusters in different regions (e.g., `prod-nyc3` in NYC, `prod-sfo3` in SFO) to minimize image pull latency
- **Environment separation:** Use distinct registries for different environments (`prod-registry`, `staging-registry`) or image types (`base-images`, `app-images`)

The Basic tier's 5-repository limit is insufficient for microservice architectures, and it doesn't support multiple registries, making it a non-starter for production.

---

### Region Selection and Data Locality

DOCR is a **regional service**. To achieve the low-latency promise, the registry **must** be co-located in the same datacenter region as the DOKS clusters or App Platform apps that consume its images.

**Anti-Pattern:** Using a single, central registry (e.g., in `nyc3`) for DOKS clusters in multiple regions (e.g., `nyc3` and `sfo3`). This introduces high-latency image pulls and will likely incur significant cross-region data transfer costs when bandwidth metering is enabled (currently free, but DigitalOcean has warned it will be charged at $0.02/GB in the future).

**Best Practice:** Use the Professional tier to create **region-locked** registries (e.g., `prod-registry-nyc3`, `prod-registry-sfo3`) and ensure DOKS clusters only pull from their local registry.

---

### Access Control and Authentication

The primary authentication mechanism is a DigitalOcean API token. Several patterns exist for CI/CD:

**Recommended: `doctl registry login`**

This command uses a long-lived API token to generate **short-lived, temporary** Docker credentials and automatically configures the local `config.json` file. The credentials expire after a configurable period (`--expiry-seconds`), reducing the risk of leaked tokens.

```bash
doctl registry login --expiry-seconds 3600
```

**High Risk: `docker login` with API token**

```bash
docker login -u $DO_TOKEN -p $DO_TOKEN registry.digitalocean.com
```

This is operationally simple but a security risk, as it uses the **long-lived, all-powerful** API token as the password, which could be leaked in CI logs or terminal history.

**Static Credentials File**

You can download a `config.json` file from the UI or generate it via `doctl registry docker-config --read-write`, then inject it into CI environments as a secret. This is safe if the secret is properly managed (encrypted at rest, scoped to specific repos).

---

### Image Security: Scanning and Signing

DOCR adopts a "bring-your-own-security" model.

**Vulnerability Scanning:**

The "built-in security scanning" is actually an **integration with Snyk**. It's not automatic—you must separately configure your Snyk account to scan DOCR. Third-party tools like Trivy or Anchore can also be integrated.

**Image Signing:**

DOCR has **no native support** for image signing verification, Notary, or Docker Content Trust. As an OCI-compliant registry, it can **store** Cosign signatures (which are just OCI artifacts), but **verification** is 100% a client-side responsibility, typically handled by admission controllers (like Kyverno or OPA Gatekeeper) in the DOKS cluster. DOCR itself will not block the pull of an unsigned or untrusted image.

---

### Garbage Collection Strategy

This is the most critical operational task for managing DOCR and controlling costs.

**The Strategy:**
- **Automate GC on a schedule** to prevent storage bloat
- **Run during low-traffic windows** (e.g., 3 AM Sunday) to avoid breaking CI/CD
- **Always include `--include-untagged-manifests`** flag (otherwise GC does nothing useful)

**The Hazard:**

When GC runs, the registry enters **read-only mode**. Any `docker push` during this window will fail. This is why scheduling is critical—run GC when CI/CD pipelines are idle.

**The Problem:**

Neither Terraform nor Pulumi can automate this. You need custom tooling (like a cron job calling `doctl registry garbage-collection start`) or a higher-level controller (like OpenMCF) that manages scheduling.

---

### Monitoring and Cost Control

**Storage Consumption:**

Monitor via:
1. The DigitalOcean control panel (storage usage is displayed)
2. The API (exposed in Terraform data source `digitalocean_container_registry` as `storage_usage_bytes`)

**Bandwidth:**

Currently **not metered or charged**. However, DigitalOcean has explicitly warned: "In the future, bandwidth limits will be applied and overages will be charged at $0.02/GB." Plan for this future cost.

**Logging:**

DOCR does not provide audit logs for image push/pull events. If you need this for compliance, you'll need external tooling.

---

### Disaster Recovery

DOCR is a regional service with no native cross-region replication. A full regional outage will make the registry unavailable. The only DR strategy is to:
1. Use the Professional plan to create registries in two separate regions
2. Modify CI/CD pipelines to push all critical production images to **both** registries
3. Configure DOKS clusters to fall back to the alternate region's registry if the primary is unavailable

---

### Common Anti-Patterns (and How to Fix Them)

| Anti-Pattern                          | Why It's Bad                                                                 | Fix                                                                 |
|---------------------------------------|------------------------------------------------------------------------------|---------------------------------------------------------------------|
| **Using the `latest` tag**            | Creates untagged manifests on every push, causing storage bloat             | Use unique, immutable tags (Git SHA, SemVer)                        |
| **No garbage collection**             | Default behavior leads to unbounded storage growth and cost overruns         | Automate GC on a schedule (3 AM Sunday, include untagged manifests) |
| **Using Starter/Basic for production**| Repository and registry limits are technically prohibitive                   | Use Professional tier                                               |
| **Cross-region image pulls**          | High latency and future bandwidth costs                                      | Use Professional plan for co-located registries                     |
| **Manual DOKS integration via UI**    | Breaks IaC state, creates invisible drift                                    | Automate via declarative tooling (OpenMCF)                  |

---

## OpenMCF's Approach: Bridging the IaC Gap

OpenMCF provides a declarative, protobuf-defined API for DOCR that **solves the IaC gap** identified in this guide. We don't just provision the registry—we automate the operational features that Terraform and Pulumi can't handle.

### What We Abstract

**The `DigitalOceanContainerRegistrySpec` includes:**

- **`name`** (string, required): The globally unique registry name (1-63 characters, lowercase alphanumeric + hyphens)
- **`subscription_tier`** (enum, required): `STARTER`, `BASIC`, or `PROFESSIONAL`
- **`region`** (string, required): The datacenter region (e.g., `nyc3`, `sfo3`)
- **`garbage_collection_enabled`** (boolean): Toggle automated garbage collection on/off

This follows the **80/20 principle**: 80% of users need only these four fields. The other 20%—advanced token scopes, multi-region mirroring—are handled at the application or orchestration layer.

### The Operational Abstraction: Garbage Collection

The `garbage_collection_enabled` field is not a simple flag—it's an **operational abstraction**. When set to `true`, OpenMCF's controller:
1. Runs its **own scheduler** (a controller-managed cronjob)
2. Periodically calls the imperative `POST /v2/registry/{name}/garbage-collection` endpoint
3. Schedules GC for **low-traffic windows** (e.g., 3 AM Sunday) to avoid breaking CI/CD pipelines
4. Always includes the `--include-untagged-manifests` flag

This solves the cost-management problem that Terraform and Pulumi can't.

### The Future: DOKS Integration Abstraction

The research report recommends adding a `doks_integration` field to the spec:

```protobuf
message DoksIntegration {
  repeated string cluster_uuids = 1;
}
```

When populated, the controller would call the `POST /v2/kubernetes/clusters/registry` endpoint, perfectly replicating the "1-click" UI feature. This would solve the other half of the IaC gap.

### Under the Hood: Pulumi

OpenMCF currently uses **Pulumi (Go)** for DOCR provisioning. Why?

- **Equivalent Coverage:** Pulumi's `digitalocean.ContainerRegistry` resource supports all provisioning fields we need
- **Language Flexibility:** Pulumi's Go SDK fits naturally into our broader multi-cloud orchestration
- **Future-Proofing:** Pulumi's programming model makes it easier to add conditional logic, multi-registry strategies, or custom integrations (like automated GC scheduling)

Terraform would work equally well for basic provisioning, but Pulumi gives us the flexibility to implement the operational abstractions (GC scheduling, DOKS integration) that pure HCL can't handle.

---

## Configuration Examples: Dev, Staging, Production

### Development: Starter Tier with Weekly GC

**Use Case:** Small test registry for a developer's sandbox.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: dev-registry
spec:
  name: "my-app-dev"
  subscription_tier: STARTER
  region: "nyc3"
  garbage_collection_enabled: true
```

**Rationale:**
- Starter tier (free) is sufficient for dev
- GC enabled to prevent storage bloat (even on free tier, good practice)
- OpenMCF schedules GC automatically (e.g., Monday 5 AM)

---

### Staging: Basic Tier with Daily GC

**Use Case:** Staging registry for pre-production testing.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: staging-registry
spec:
  name: "my-app-staging"
  subscription_tier: BASIC
  region: "sfo3"
  garbage_collection_enabled: true
```

**Rationale:**
- Basic tier ($5/month) provides 5 GiB storage and 5 repositories
- Region matches staging DOKS cluster location
- Daily GC (scheduled by controller) prevents storage growth

---

### Production: Professional Tier with Scheduled GC

**Use Case:** Production registry with multiple environments and multi-region support.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry-nyc3
spec:
  name: "my-app-prod-nyc3"
  subscription_tier: PROFESSIONAL
  region: "nyc3"
  garbage_collection_enabled: true
```

**Rationale:**
- Professional tier ($20/month) supports up to 10 registries and unlimited repositories
- Region-locked to `nyc3` for co-location with DOKS cluster
- GC enabled and scheduled for Sunday 3 AM (low-traffic window)

**Multi-Region Pattern:**

For DR or multi-region deployments, create a second registry in a different region:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry-sfo3
spec:
  name: "my-app-prod-sfo3"
  subscription_tier: PROFESSIONAL
  region: "sfo3"
  garbage_collection_enabled: true
```

CI/CD pipelines push images to both `prod-nyc3` and `prod-sfo3`, and DOKS clusters pull from their local registry.

---

## Key Takeaways

1. **DOCR is a "glue service"** designed to reduce friction for DOKS and App Platform users. Its value comes from co-location, transparent pricing, and zero operational overhead—not from a rich feature set.

2. **Manual management (console or CLI scripting) is an anti-pattern** for production. Use IaC (Terraform, Pulumi, or OpenTofu) for provisioning, but recognize that they can't handle DOKS integration or automated garbage collection.

3. **The IaC gap is real and significant.** Terraform and Pulumi can create the registry, but they can't replicate the "1-click" DOKS integration or schedule garbage collection. Teams end up using a mix of IaC (for provisioning) and manual operations (for integration and GC), creating invisible drift.

4. **Garbage collection is critical and non-optional** for production. Without scheduled GC, storage bloat is the default behavior, leading to uncontrolled costs. GC must run during low-traffic windows to avoid breaking CI/CD pipelines (because the registry enters read-only mode).

5. **The 80/20 config is name, subscription tier, region, and GC toggle.** Everything else—advanced token scopes, multi-region replication, image signing—is handled at the application layer or via external tooling.

6. **Subscription tier selection matters.** Professional tier ($20/month) is the only choice for production because it supports multiple registries (up to 10), enabling multi-region architectures and environment separation.

7. **OpenMCF bridges the IaC gap** by providing declarative abstractions for the features Terraform and Pulumi can't handle: automated garbage collection scheduling (and, in the future, DOKS integration). This makes DOCR management truly declarative, cost-effective, and safe for production.

---

## Further Reading

- **DigitalOcean Container Registry Documentation:** [DOCR Overview](https://docs.digitalocean.com/products/container-registry/)
- **DOCR Pricing and Tiers:** [Pricing Guide](https://www.digitalocean.com/pricing/container-registry)
- **Garbage Collection Deep Dive:** [Understanding GC in DOCR](https://www.digitalocean.com/blog/garbage-collection-digitalocean-container-registry)
- **Terraform DigitalOcean Provider:** [digitalocean_container_registry](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/container_registry)
- **Pulumi DigitalOcean Package:** [@pulumi/digitalocean](https://www.pulumi.com/registry/packages/digitalocean/)
- **DOKS Integration Guide:** [Integrate with Container Registry](https://docs.digitalocean.com/products/kubernetes/how-to/integrate-with-docr/)
- **CI/CD Integration Patterns:** [Set Up CI/CD with DOCR](https://docs.digitalocean.com/products/container-registry/how-to/set-up-ci-cd/)

---

**Bottom Line:** DigitalOcean Container Registry is simple, co-located, and cost-effective for teams already on DigitalOcean. But simplicity doesn't mean hands-off. Without scheduled garbage collection, costs spiral. Without proper IaC, drift happens. Terraform and Pulumi can provision the registry, but they can't automate the operational features that matter. OpenMCF fills that gap with declarative abstractions that make DOCR truly production-ready: automated GC scheduling, region-aware defaults, and (soon) DOKS integration. This is the registry management you wish the console had—declarative, cost-controlled, and safe for production.

