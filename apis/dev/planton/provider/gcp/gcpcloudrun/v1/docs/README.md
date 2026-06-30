# Deploying GCP Cloud Run: From Platform Fundamentals to Production IaC

## Introduction: The Serverless Container Revolution

For years, the conventional wisdom around running containers in production was clear: you needed Kubernetes. You needed to understand node pools, cluster autoscaling, pod resource requests, and a constellation of YAML files. Cloud Run challenged that assumption.

Built directly on the open-source Knative project, Google Cloud Run represents a different philosophy: **you configure workloads, not infrastructure**. Instead of managing VMs, clusters, and node pools, you define your container image, its resource requirements, and how it should scale. Cloud Run handles everything else—autoscaling from zero to thousands of instances based on demand, routing traffic, and billing you only when requests are actively being processed.

This document explores the landscape of Cloud Run deployment approaches, from simple CLI commands to production-grade Infrastructure-as-Code (IaC) patterns. We'll examine why certain methods that seem convenient during development become anti-patterns in production, compare the dominant IaC tools, and explain the design philosophy behind Planton's Cloud Run API.

## Understanding Cloud Run's Place in the GCP Compute Spectrum

Before diving into deployment methods, it's worth understanding **when** to choose Cloud Run over other GCP compute options. The decision is fundamentally about the trade-off between control and management overhead.

| Compute Service | Management Model | Billing Unit | Key Feature | Ideal Workload |
|:---|:---|:---|:---|:---|
| **Compute Engine (GCE)** | IaaS | Per VM (CPU/Mem/Storage) | Full OS/kernel control | Stateful, legacy, custom OS, or complex apps |
| **GKE Standard** | CaaS | Per Node (VM) | Full Kubernetes API control | Complex, multi-container, stateful applications |
| **GKE Autopilot** | CaaS (Managed) | Per Pod Resource Request | Managed nodes, pod-based scaling | Kubernetes-native apps without node management |
| **App Engine** | PaaS | Per Usage / Resources | Language-specific runtime, minimal config | Simple web apps and mobile backends |
| **Cloud Functions** | FaaS | Per Invocation / Usage | Event-driven, single-purpose code | Small, event-triggered tasks (e.g., storage events) |
| **Cloud Run** | CaaS/PaaS (Serverless Containers) | Per Usage (Requests/CPU/Mem) | **Scale-to-zero**, stateless containers | Stateless, request-driven HTTP services & APIs |

Cloud Run's defining features—**scale-to-zero autoscaling** and **pay-per-use billing**—make it the most cost-effective option for spiky, unpredictable, or low-traffic workloads. When your service receives no traffic, it costs $0. This is fundamentally different from GKE Autopilot, which scales *nodes* to zero but requires additional tooling (like KEDA) to achieve *pod-level* scale-to-zero.

A critical architectural insight: Cloud Run is built on Knative. Every Cloud Run deployment manipulates the same declarative primitives (Service, Configuration, Revision, Route) that Knative uses. This means that declarative IaC tools aren't a secondary interface—they're a first-class method for managing Cloud Run, philosophically aligned with the platform's design.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern (Click-Ops and Source-Based Deployment)

**Method:** Using the Google Cloud Console to manually configure services, or running `gcloud run deploy --source .` directly from a local directory.

**Why It Seems Convenient:** Zero configuration files, instant feedback, and no need to understand container registries or build pipelines.

**The Problem:** Both approaches introduce non-deterministic infrastructure that's impossible to reproduce or audit.

Manual console changes create configuration drift. When the service fails six months later, you'll have no record of what settings were applied, when, or by whom. The `--source` flag is even more insidious. It triggers a Cloud Build job that uses buildpacks to automatically containerize your source code, creating a container image with a *non-deterministic digest*. You can never be certain exactly what image is running in production.

**Verdict:** Acceptable for initial discovery and local experimentation. A production-level anti-pattern.

### Level 1: Scripted CLI (Image-Based Deployment)

**Method:** Using `gcloud run deploy --image <image-uri>` in CI/CD scripts, deploying a specific, pre-built container image.

**What It Solves:** You now have a deterministic artifact (a specific container image digest) and can version your deployment commands in source control.

**The Limitation:** The deployment is still imperative. There's no single source of truth representing the service's desired state. If someone manually updates the service's environment variables in the console, your script won't detect or correct the drift. You're running commands, not managing state.

**Verdict:** A significant improvement over Level 0. Suitable for simple CI/CD pipelines but lacks the auditability and drift detection of declarative tools.

### Level 2: Lightweight Declarative (gcloud YAML)

**Method:** Defining the service in a Knative-compatible YAML file and applying it with `gcloud run services replace service.yaml`.

**What It Solves:** You now have a declarative, version-controlled file representing your service's desired state. This YAML is Knative-native, making it portable.

**The Limitation:** The `gcloud` CLI doesn't provide a planning phase (like `terraform plan`), state management, or dependency orchestration. If your architecture involves multiple services, IAM bindings, Secret Manager secrets, and VPC configuration, you're manually orchestrating multiple `gcloud` and `terraform` commands.

**Verdict:** A strong choice for teams committed to Google-native tooling and simple architectures. Not suitable for complex, multi-resource deployments.

### Level 3: Production IaC (Terraform, Pulumi, Crossplane)

**Method:** Using a dedicated IaC tool to declaratively define Cloud Run services, their dependencies (secrets, IAM, VPC), and their relationships.

**What It Solves:**

- **State Management:** A remote state file (e.g., in GCS) tracks the real-world state of every resource, enabling drift detection and collaborative deployments.
- **Planning:** Tools like Terraform provide a plan phase, showing exactly what will change before applying.
- **Dependency Orchestration:** The tool automatically determines the correct order to create, update, or destroy resources based on their dependencies.
- **Auditability:** Every change is a code change, reviewable and traceable in version control.

**The Trade-off:** Higher upfront learning curve and additional tooling to manage.

**Verdict:** The production-grade standard. Essential for any environment requiring compliance, auditability, or team collaboration.

## Comparative Analysis: Terraform vs. Pulumi vs. Cloud Deployment Manager

When you've chosen to adopt declarative IaC (Level 3), the next decision is which tool to use. Three major options exist for GCP, though one is deprecated.

### Terraform (and OpenTofu)

**Language:** HCL (HashiCorp Configuration Language), a declarative domain-specific language.

**Philosophy:** Strictly declarative. You define *what* the end state should be, not *how* to achieve it. This constraint is a feature—it leads to simple, auditable "plans" and promotes reusability through a module system.

**State Management:** Requires manual configuration of a remote backend (e.g., a GCS bucket). The state file is **unencrypted by default**, and secrets are stored in **plaintext**. This is a well-known limitation.

**Secret Handling:** Because secrets in the state file are insecure, production-grade Terraform configurations never define secrets inline. Instead, they follow a superior architectural pattern:

1. Create the secret in GCP Secret Manager (as a `google_secret_manager_secret` resource).
2. Grant the Cloud Run service's identity (service account) IAM access to that secret.
3. Reference the secret in the Cloud Run service using `secret_key_ref`, not by value.

This correctly separates application-level secrets from infrastructure-level configuration.

**Ecosystem:** The de facto standard for IaC. Massive provider ecosystem. OpenTofu is a community-driven, Linux Foundation-managed fork that serves as a drop-in replacement.

**Best For:** Platform teams, SREs, and enterprises requiring strict, declarative, auditable, ecosystem-agnostic infrastructure.

### Pulumi

**Language:** General-purpose programming languages (Python, Go, TypeScript, etc.).

**Philosophy:** Infrastructure *as* code, not infrastructure *in* a DSL. You use loops, conditionals, classes, and unit testing frameworks. This is more powerful but can lead to overly complex, imperative-style code.

**State Management:** Offers a "zero-config," managed, and **encrypted** backend by default. Secrets are encrypted in state, unlike Terraform.

**Secret Handling:** Built-in encryption. Secrets can be defined inline and will be encrypted at rest in the state file.

**Ecosystem:** Can use *all* Terraform providers via a bridge, in addition to native Pulumi providers.

**Best For:** Development teams that own their infrastructure ("you build it, you run it") and want to use their existing language skills and testing frameworks.

### Cloud Deployment Manager (Deprecated)

**Status:** Google's first-generation native IaC tool. **End of support: March 31, 2026.**

**Verdict:** Not recommended for any new projects. The deprecation of Cloud Deployment Manager solidifies Terraform and Pulumi as the industry standards for GCP infrastructure.

### The Verdict: Which Tool for Which Use Case?

- **Terraform/OpenTofu:** The best choice for platform teams, SREs, and enterprises that require a strict, declarative, auditable, and ecosystem-agnostic approach. Its "plaintext secrets" limitation inadvertently enforces a superior architectural pattern (externalizing secrets to Secret Manager).

- **Pulumi:** The best choice for development teams that prefer general-purpose languages, want built-in secret encryption, and value the power of programming constructs like loops and classes.

- **Cloud Deployment Manager:** Not recommended (deprecated).

Planton's protobuf-based API is **philosophically closer to Terraform's HCL**. Both are schema-driven, declarative languages that define *state*. The design patterns established by the Terraform community—clean resource separation, explicit dependency management, and externalized secret handling—are the most relevant model for our API design.

## Production Essentials: What "Level 3" Really Means

Reaching "production-grade" isn't just about using an IaC tool. It's about configuring the service correctly for security, reliability, and performance.

### 1. Authentication and Authorization

**Default Security Posture:** Cloud Run services are **private by default** (`--no-allow-unauthenticated`). A private service returns a `403 Forbidden` error unless the caller is authenticated and authorized.

**Making a Service Public:** Setting `allowUnauthenticated: true` grants the special `allUsers` member the `roles/run.invoker` IAM role, making the service publicly accessible.

**Service-to-Service Authentication:** For private services, the calling service must authenticate using Google-signed ID tokens:

1. The calling service's identity (its attached Service Account) is granted `roles/run.invoker` on the receiving service.
2. The caller fetches a Google-signed ID token from the local metadata server.
3. This token is sent in the `Authorization: Bearer <ID_TOKEN>` header.
4. Cloud Run automatically validates the token before forwarding the request.

**Common Anti-Pattern:** Deploying a private service, receiving a 403 error, and "fixing" it by setting `allUsers`, thereby exposing an internal API to the public internet. The correct solution is to configure the service-to-service IAM bindings.

### 2. Networking: VPC Integration and Custom Domains

**Direct VPC Egress (Recommended):** The modern approach for accessing private resources inside a VPC (e.g., a Redis instance, a private database). The Cloud Run service is simply associated with a VPC subnet. This provides lower latency, higher throughput, and a simpler API than the legacy "Serverless VPC Access Connector."

**Ingress Control:** By setting the service's `ingress` setting to `internal` or `internal-and-cloud-load-balancing`, access is restricted to internal GCP resources or an Internal Load Balancer.

**Powerful Microservice Pattern:** A "frontend" service can be public (`ingress: "all"`) but use Direct VPC Egress to call a "backend" service. The backend service can be set to `ingress: "internal"`. Because the frontend's call originates from within the VPC (due to Direct VPC Egress), the call is recognized as internal and succeeds. This creates a fully-private, secure, and serverless data plane.

**Custom Domains and SSL:** The production-grade solution is a **Global External Application Load Balancer**, which enables the use of Cloud Armor (WAF), Cloud CDN, and custom SSL certificates. The load balancer routes traffic to a Serverless Network Endpoint Group (NEG), which acts as a pointer to the Cloud Run service. For simpler applications, **Firebase Hosting** provides a low-cost alternative, offering free SSL, a global CDN, and simple rewrite rules to proxy requests to Cloud Run.

### 3. Scaling, Performance, and Health Checks

**The Concurrency Knob:** The `concurrency` setting (default: 80) is the most important performance tuning parameter. It defines the maximum number of requests a single container instance can process *simultaneously*.

**Common Concurrency Trap:** A developer deploys a single-threaded application (like a standard Python Flask server) with the default `concurrency: 80`. The first request saturates the single vCPU. The 79 other concurrent requests queue up, leading to mass timeouts. The correct fix is to load-test the application and set a *realistic* concurrency (e.g., 4 for a CPU-bound task, or 1 for a strictly single-request application).

**Scaling Limits:**

- `min-instances`: Keeps instances "warm" to reduce cold starts. Incurs idle costs.
- `max-instances`: Critical for cost control (preventing runaway scaling) and for protecting downstream dependencies (like a database) from being overwhelmed.

**Health Checks:**

- **Startup Probes:** Determine if a container has *finished starting* and is ready to accept traffic. Essential for slow-starting applications (e.g., Java services, ML models) to prevent them from being killed before they're ready.
- **Liveness Probes:** Determine if a container is *still healthy*. If a liveness probe fails, Cloud Run restarts the container. Intended to recover from unrecoverable states like deadlocks.

**The "Missing Startup Probe" Crash Loop:** A developer has a slow-starting application (45 seconds to load). They add a *liveness probe* to check `/healthz`. The liveness probe starts checking after 10 seconds and fails. Cloud Run assumes the container is dead and restarts it. The container is caught in a crash loop. The correct fix is to add a *startup probe*, which tells Cloud Run, "Do not run the liveness probe or send me traffic until *this* probe passes."

### 4. Observability

Cloud Run is deeply integrated with Google Cloud Observability:

- **Cloud Logging:** All stdout and stderr streams from the container are automatically captured and stored.
- **Cloud Monitoring:** Key service-level metrics (Request Count, Latency, Billable Instance Time) and container-level metrics (CPU/Memory Utilization, Instance Count) are collected automatically.
- **Cloud Trace:** Provides latency sampling and reporting. Automatically enabled for requests that pass through a Google Cloud Load Balancer or originate from other traced services.
- **Error Reporting:** Automatically parses container logs for unhandled exceptions (in common languages) and aggregates them into a dedicated dashboard.

## The 80/20 Configuration Principle

Just as good API design focuses on the 20% of configuration that 80% of users need, effective IaC for Cloud Run should prioritize the "core" fields.

### The "80% Core": Essential Configuration

Analysis of production-grade Terraform and Pulumi examples reveals a common set of fields that are set in the vast majority of use cases:

- **project & location:** Non-negotiable identifiers for all GCP resources.
- **name:** The service name.
- **image:** The container image URI (e.g., `us-docker.pkg.dev/...`). This is the central, defining field.
- **service_account:** The IAM identity the service runs as. Essential for any service that needs to call other Google APIs.
- **ingress:** Network access control (`ALL`, `INTERNAL_ONLY`, `INTERNAL_LOAD_BALANCER`). A top-level security decision.
- **allowUnauthenticated:** A simplified boolean controlling the `roles/run.invoker` binding for `allUsers`.
- **port:** The port the container listens on (defaults to 8080).
- **env:** Basic, non-secret environment variables.
- **secrets:** A first-class mechanism for mounting GCP Secret Manager secrets as environment variables or volumes. This is **part of the 80% Core**, not an advanced feature. No production application should use plaintext environment variables for sensitive data.

### The "20% Advanced": Scalability, Networking, and Reliability

These fields are critical for tuning a service for production, but their defaults are often sufficient for initial deployment:

- **resources:** CPU (1, 2, 4) and Memory (512Mi, 1Gi) limits.
- **scaling:** `min_instances`, `max_instances`, and `concurrency` settings.
- **timeout_seconds:** Request timeout.
- **network:** VPC access configuration (Direct VPC Egress) and Cloud SQL connections.
- **probes:** Startup and liveness probe definitions.
- **traffic:** For manually defining traffic splits between revisions (blue-green, canary deployments).
- **execution_environment:** GEN1 vs. GEN2 (Gen2 offers full Linux compatibility but has slower cold starts).

## Container Images and CI/CD: The Build-Deploy Separation

### Use Artifact Registry, Not Container Registry

**Google Container Registry (GCR)** (`gcr.io`) is a legacy service. **Google Artifact Registry (AR)** (`pkg.dev`) is the recommended service for all container image storage.

**Key Advantages of Artifact Registry:**

1. **Granular IAM:** Repository-level IAM permissions (GCR's permissions are project-wide, tied to the underlying GCS bucket).
2. **Regionality:** Regional repositories (e.g., `us-central1-docker.pkg.dev`) reduce latency and data transfer costs.
3. **Multi-format:** Unified registry for Docker images, language packages (npm, Maven, Python), and OS packages.

### The Build-Deploy Separation Principle

The `gcloud run deploy --source` workflow is a powerful development-time convenience, but it represents a production-level anti-pattern for IaC.

**Why:** IaC operates on the principle of deterministic, declarative state. The desired state should be "I want *this specific image digest* to be running." A source-based deployment is an *imperative* command ("go build this source code") that results in a *non-deterministic* image digest.

**The Production Pattern:**

1. **Build Step (CI responsibility):** A CI pipeline (GitHub Actions, Cloud Build, GitLab CI) builds the container image from source, tags it with a specific version (e.g., `v1.2.3` or a commit SHA), and pushes it to Artifact Registry.
2. **Deploy Step (IaC/CD responsibility):** The IaC tool (Terraform, Pulumi, Planton) deploys the service using the specific image URI and digest created by the build step.

This separation ensures that the image is a deterministic, versioned, and auditable artifact.

### Authentication: Use Workload Identity Federation, Not Service Account Keys

**The Anti-Pattern:** Downloading a service account JSON key and storing it as a GitHub Secret. This is a long-lived, static credential that's a major security risk if compromised.

**The Best Practice:** **Workload Identity Federation (WIF)** is the modern, secure, and keyless authentication method. The `google-github-actions/auth@v3` action exchanges a short-lived GitHub OIDC token (generated per-workflow) for a short-lived Google Cloud access token. This exchange is based on a trust relationship pre-configured in IAM (a Workload Identity Pool).

GitLab CI and other platforms also support OIDC and Workload Identity Federation, enabling the same secure, keyless authentication pattern.

## Advanced Deployment Strategies: Blue-Green and Canary

Cloud Run's immutable Revision model is the foundation for safe deployments.

### Immutable Revisions and Instant Rollbacks

When a new configuration is deployed, a *new* revision is created. The old, working revision is not touched. A "rollback" is not a new deployment—it's an atomic, near-instantaneous operation that updates the Route (traffic split) to send 100% of traffic back to the last-known-good revision.

### Blue-Green Deployment

This is a two-step traffic split:

1. Deploy the new revision with `--no-traffic` (or an IaC configuration where `traffic { percent = 0 }`).
2. Run automated tests against the new revision's direct, unique URL.
3. If tests pass, execute a second deployment (or `gcloud run services update-traffic`) to shift 100% of traffic to the new revision.

### Canary Deployment

This is a *progressive* traffic split:

1. Deploy the new revision with a 10% traffic split.
2. *Wait and monitor* key metrics (e.g., error rate, latency).
3. If metrics are healthy, apply a new configuration with a 50% traffic split.
4. Repeat until 100% of traffic is migrated.

### Orchestrating Canary Rollouts

IaC tools like Terraform and Pulumi can *execute* a single step of a canary deployment (e.g., `terraform apply` to set 10% traffic). They do *not*, by themselves, *orchestrate* the full, timed rollout (the "wait and monitor" steps). This gap is filled by:

1. **Google Cloud Deploy:** Google's enterprise-grade, managed continuous delivery service. It natively understands canary deployments for Cloud Run and can be configured to automatically execute a progressive rollout (e.g., 25%, 50%, 100%) with verification jobs between phases.
2. **Cloud Run Release Manager:** An open-source, experimental tool that runs *in* Cloud Run and watches your services. By adding a `rollout-strategy=gradual` label to a service, this tool will automatically manage the canary rollout, progressively shifting traffic, monitoring Cloud Monitoring for errors, and automatically rolling back on failure.

## Cost Optimization: When Cloud Run is (and Isn't) the Cheapest Option

Cloud Run's pricing is built on a "pay-per-use" philosophy:

- **vCPU-seconds:** Billed per 100ms when an instance is actively starting or processing a request.
- **Memory-seconds:** Billed per 100ms.
- **Requests:** A small, per-request fee ($0.40 per million after the free tier).
- **Networking:** Outbound (egress) network traffic.

### Cost Optimization Strategies

1. **Scale to Zero:** The most powerful cost-optimization feature. Unless strict latency requirements forbid cold starts, let the service scale to zero. At zero instances, cost is $0.
2. **Right-sizing:** Monitor CPU and memory utilization in Cloud Monitoring and allocate only the resources the service needs. Over-provisioning (e.g., 2 vCPU for a single-threaded app) wastes money.
3. **Concurrency Tuning:** A *higher* stable concurrency is significantly cheaper. Handling 80 concurrent requests on *one* 1-vCPU instance is far more cost-effective than handling those same 80 requests on *80* different 1-vCPU instances (with `concurrency: 1`).
4. **Regional Selection:** Costs vary by region. Hosting the service in the same region as its Artifact Registry repository and its users minimizes data egress costs.

### Cloud Run vs. GKE Autopilot: A Cost Comparison

The choice is entirely dependent on the workload pattern:

- **Cloud Run (Spiky/Unpredictable):** For spiky, unpredictable, or low-traffic workloads, Cloud Run is *always* more cost-effective. Its scale-to-zero capability means you pay $0 for idle time.

- **GKE Autopilot (Sustained/Complex):** For high-traffic, *sustained* workloads, the pay-per-request model of Cloud Run can become more expensive than the pay-per-pod-request model of GKE Autopilot. At high, constant load, the fixed cost of GKE becomes more attractive. GKE Autopilot is also the *only* option if the application *requires* Kubernetes-native features like sidecars (beyond the Cloud SQL proxy), DaemonSets, or complex networking.

**A Critical Clarification on Scale-to-Zero:**

- **Cloud Run** scales *service instances* to zero *by default* based on request traffic.
- **GKE Autopilot** scales *nodes* to zero. The *pods* (your application) do *not* scale to zero by default. A standard deployment with `replicas: 1` will run (and be billed for) 24/7.
- To achieve *pod-level* scale-to-zero on GKE (like Cloud Run), you must install and configure **KEDA (Kubernetes Event-driven Autoscaler)**.

Therefore, Cloud Run offers true, request-based scale-to-zero out of the box, making it the simpler and more cost-effective choice for event-driven or spiky workloads.

## What Planton Supports (and Why)

Planton's `GcpCloudRun` API is designed to be a production-grade, declarative interface for deploying Cloud Run services. It follows the design principles and best practices identified in this analysis:

### Declarative, State-Based Design

Like Cloud Run's native Knative API and like Terraform's HCL, Planton's protobuf-based API is **declarative**. You define the *desired state* of the service, not a series of imperative commands to achieve it.

### Image-Based Deployment (Not Source-Based)

The API requires a specific container image URI as its primary input. This enforces the separation of the "Build" step (a CI responsibility) from the "Deploy" step (an IaC/CD responsibility), ensuring deterministic, reproducible deployments.

### Secrets as First-Class Configuration

The API provides dedicated fields for mounting secrets from GCP Secret Manager. This is part of the "80% Core," not an advanced feature, reflecting the reality that no production application should use plaintext environment variables for sensitive data.

### Production-Ready Defaults

- Services are **private by default** (`allowUnauthenticated: false`).
- The API exposes `ingress`, `vpc_access`, `scaling`, `probes`, and `traffic` as first-class fields, recognizing that these are essential for production services, not optional extras.

### Designed for Artifact Registry

All documentation and examples use Artifact Registry's `pkg.dev` domain format, not the legacy `gcr.io`.

### Grounded in IaC Best Practices

The API's structure mirrors the logical separation established by Terraform and Pulumi: a top-level Service resource managing routing (traffic) and IAM, and a Template sub-resource defining the immutable Revision (container image, env, secrets, resources, probes).

## Conclusion: Workload-Centric Infrastructure

The evolution of Cloud Run—from Knative's open-source foundation to Google's fully managed platform—reflects a broader industry shift: developers increasingly expect to configure *workloads* (containers, resources, scaling policies) and *not* infrastructure (VMs, clusters, node pools).

Planton's Cloud Run API embraces this workload-centric model. By combining the declarative rigor of IaC with the production-proven patterns from the Terraform and Pulumi communities, we provide a platform that's both developer-friendly and production-ready.

The choice to use Cloud Run over GKE or Compute Engine isn't about which is "better"—it's about which is the right tool for the workload. For stateless, request-driven HTTP services and APIs, Cloud Run's scale-to-zero, pay-per-use model, and Knative-based architecture make it the simplest, most cost-effective, and most operationally efficient choice.

You define the workload. Cloud Run handles the infrastructure. That's the promise—and when deployed using declarative IaC, it's a promise that's both reliable and reproducible.

