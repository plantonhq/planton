# Google Cloud Functions: The Serverless Evolution

## Introduction: From FaaS to Cloud Run

Ask any developer who's been in the serverless space for more than a few years: "What's a Google Cloud Function?" You'll likely hear "It's a serverless function platform—write code, deploy it, and it scales automatically." That's true, but incomplete.

Here's what changed: **Google Cloud Functions (Gen 2) isn't a separate FaaS platform anymore. It's a developer-friendly abstraction layer on top of Cloud Run and Eventarc.**

This isn't a minor implementation detail. It's a fundamental architectural shift that affects everything from performance characteristics to pricing models to deployment patterns. When you deploy a Gen 2 function, you're actually deploying a Cloud Run service. The platform handles containerization via buildpacks, manages event routing through Eventarc, and presents a simplified "just deploy my code" interface.

**Why does this matter?**

1. **Performance**: Gen 2 functions inherit Cloud Run's superior capabilities—up to 1,000 concurrent requests per instance (vs. 1 in Gen 1), 60-minute timeouts (vs. 9 minutes), and massively more compute (32GB RAM, 8 vCPU).

2. **Strategic Clarity**: Understanding that Gen 2 *is* Cloud Run helps you make informed deployment decisions. Need custom containers? Use Cloud Run directly. Want the simplest event-driven workflow? Cloud Functions provides that abstraction.

3. **IaC Design**: When designing infrastructure-as-code APIs, recognizing this architecture means modeling the underlying reality—build configuration, service configuration, and trigger configuration—rather than inventing a new abstraction that hides complexity poorly.

This guide maps the deployment landscape for Google Cloud Functions, explains the Gen 1 vs. Gen 2 decision (spoiler: always choose Gen 2), and details why Planton's API is designed around the 80/20 principle: expose the essential 20% of configuration that 80% of production deployments need.

---

## Understanding the Generational Divide

### What is Google Cloud Functions?

At its core, Google Cloud Functions is a serverless, event-driven compute platform. You write a function—a small piece of code that responds to HTTP requests or cloud events—and the platform handles everything else: provisioning compute, scaling from zero to thousands of instances based on demand, routing events, and shutting down when idle.

**Common production use cases**:
- **HTTP/REST APIs**: Webhook handlers for Stripe, Twilio, or internal microservices
- **Event Processing**: Responding to Cloud Storage uploads (image thumbnailing, video transcoding), Pub/Sub messages (async job processing), or Firestore document changes (data validation, fan-out)
- **Serverless Backends**: Mobile app backends built on Firebase, where functions provide server-side logic without managing servers
- **Integration Glue**: Connecting Google Cloud services—triggering workflows when files land in Storage, publishing to Pub/Sub when Firestore documents change

### Gen 1 vs. Gen 2: Not an Upgrade, a Re-Platform

Google Cloud offers two co-existing generations of Cloud Functions. This isn't like software versioning where v2 is just v1 with new features. Gen 2 is a **complete architectural redesign**.

**Generation 1 (Gen 1)**: The original, proprietary FaaS platform. Runs on Google's internal infrastructure—a black box to users. It served its purpose but has fundamental limitations.

**Generation 2 (Gen 2)**: A developer experience layer built on two production services:
- **Cloud Run** provides the execution runtime (containers, scaling, networking)
- **Eventarc** provides the event routing (125+ event sources in a standardized CloudEvents format)

The Gen 2 "function" is syntactic sugar. Under the hood, your source code is packaged into a container via buildpacks and deployed as a Cloud Run service. Event triggers are configured as Eventarc subscriptions.

**The Critical Comparison**:

| Feature | Gen 1 (Legacy) | Gen 2 (Cloud Run) |
|---------|----------------|-------------------|
| **Architecture** | Google Internal | Cloud Run + Eventarc |
| **Concurrency** | 1 request/instance | Up to 1,000 requests/instance |
| **Max Timeout** | 9 minutes | **60 minutes** (HTTP) |
| **Max Memory** | 8 GB | **32 GiB** |
| **Max vCPU** | 2 vCPU | **8 vCPU** |
| **Event Sources** | Limited (Firebase, Pub/Sub, Storage) | **125+ via Eventarc** |
| **Traffic Management** | None | Revisions, traffic splitting, blue-green |
| **Pricing** | Function-specific (GB-sec, invocations) | Cloud Run pricing (vCPU-sec, GB-sec, invocations) |

### The Migration Reality

There is **no in-place upgrade** from Gen 1 to Gen 2. Migration requires:
1. Redeploying the function as a new Gen 2 resource (using the `--gen2` flag or equivalent IaC configuration)
2. Updating SDK imports in your code (e.g., `firebase-functions` → `firebase-functions/v2` in Node.js)
3. Potentially updating trigger definitions and dependency files
4. Running Gen 1 and Gen 2 side-by-side during cutover
5. Manually migrating traffic (for HTTP) or letting both receive events (for event-driven triggers)

**Strategic Choice**: Given Gen 2's overwhelming advantages in performance, scalability, event sources, and feature set, **all new development must default to Gen 2**. Gen 1 exists only for maintaining legacy applications or edge cases where a specific Gen 1-only feature is required (e.g., certain Firebase Analytics triggers).

### Runtime Support: Production-Ready Languages

All major languages are supported and production-ready: **Node.js, Python, Go, Java, .NET, Ruby, PHP**. The platform provides clear end-of-life timelines for runtime versions.

**Critical for production readiness**: Use only non-deprecated runtime versions. For example:
- **Node.js**: 20 and 22 (LTS versions)
- **Python**: 3.10, 3.11, 3.12
- **Go**: 1.21, 1.22

Deprecated runtimes (Node.js 18, Python 3.8/3.9) may continue working but receive no security patches. A well-designed IaC API should curate the `runtime` field as an enum, actively removing deprecated versions to guide users toward safe choices.

### Trigger Types: HTTP and Event-Driven

Functions respond to events, which fall into two categories:

**1. HTTP Triggers**: The function is invoked via an HTTPS endpoint. This is the most common pattern for webhooks, REST APIs, and mobile backends. The function receives an HTTP request object and returns an HTTP response.

**2. Event-Driven Triggers** (via Eventarc in Gen 2): The function is invoked when an event occurs in a Google Cloud service. Common sources include:
- **Cloud Pub/Sub**: Messages published to a topic trigger the function (asynchronous, decoupled architectures)
- **Cloud Storage**: Object creation, deletion, archiving, or metadata updates
- **Cloud Firestore**: Document write, update, delete (real-time data processing)
- **125+ Additional Sources**: Eventarc standardizes events from Google services and third-party platforms (DataDog, ForgeRock, Check Point) into CloudEvents format

---

## The Maturity Spectrum: Deployment Methods from Manual to Production IaC

The deployment landscape for Cloud Functions ranges from UI-based manual configuration to sophisticated declarative infrastructure-as-code. Your choice reflects operational maturity.

### Level 0: The Anti-Pattern (ClickOps in the Console)

**What it is**: Using the Google Cloud Console web UI to create functions by filling out forms, uploading ZIP files, and clicking "Deploy."

**Why teams do it**: It's the path of least resistance for initial exploration. The visual interface makes it easy to understand available options during proof-of-concept work.

**Why it fails in production**:
- **Non-reproducible**: Creating the same function in a different region or project requires manually replicating dozens of clicks and hoping you didn't miss a configuration field.
- **No audit trail**: When staging and production diverge, you have no record of what changed, when, or why.
- **Policy conflicts**: Organization-wide security policies (e.g., Uniform Bucket-Level Access) can block the console's upload mechanism, leading to cryptic "Access Denied" errors.

**Verdict**: Acceptable for learning the service. Unacceptable for any infrastructure that matters.

### Level 1: Imperative Scripting (gcloud CLI)

**What it is**: Using Google Cloud's command-line tool to script deployments:

```bash
gcloud functions deploy my-function \
  --gen2 \
  --region=us-central1 \
  --runtime=python311 \
  --source=. \
  --entry-point=hello_http \
  --trigger-http \
  --allow-unauthenticated
```

**What it solves**: Repeatability. Scripts are version-controlled, documented, and executable in CI/CD pipelines.

**What it doesn't solve**: Idempotency and state management. Running the script twice with different parameters modifies the function, but you're responsible for tracking what the current state is. The `gcloud` tool has no concept of "desired state"—it's purely imperative.

**gcloud vs. Cloud Run CLI**: Because Gen 2 functions *are* Cloud Run services, you can also use:

```bash
gcloud run deploy my-function \
  --source=. \
  --function=hello_http \
  --region=us-central1
```

This dual-path reveals the architecture: `gcloud functions deploy` is the FaaS abstraction, while `gcloud run deploy --function` is the underlying Cloud Run path enhanced with buildpack support.

**When to use it**: Ad-hoc scripting, debugging, or simple CI/CD for small teams. Not suitable for managing infrastructure at scale.

### Level 2: Configuration Management (Ansible)

**What it is**: Using the `google.cloud` Ansible collection to define functions as declarative YAML tasks:

```yaml
- name: Deploy production function
  google.cloud.gcp_cloudfunctions_cloud_function:
    name: my-function
    runtime: python311
    entry_point: hello_http
    source_archive_url: gs://my-bucket/code.zip
    trigger_http: true
```

**What it solves**: Idempotency (re-running the playbook converges to the desired state) and integration with server configuration workflows.

**What it doesn't solve**: Advanced state management, dependency graphing, or plan/preview workflows. Ansible is excellent for teams already using it for VM configuration, but for infrastructure-only management, dedicated IaC tools are more robust.

**When to use it**: When cloud function deployment is tightly coupled to server configuration (e.g., Ansible configures VMs and deploys the functions they call in a single playbook).

### Level 3: Production-Grade IaC (Terraform, Pulumi, OpenTofu)

**What it is**: Using stateful Infrastructure-as-Code tools to define functions as declarative resources with managed state, dependency resolution, and plan/preview workflows.

**Terraform** (HCL-based, industry standard):

```hcl
resource "google_cloudfunctions2_function" "function" {
  name     = "my-function"
  location = "us-central1"

  build_config {
    runtime     = "python311"
    entry_point = "hello_http"
    source {
      storage_source {
        bucket = google_storage_bucket.bucket.name
        object = google_storage_bucket_object.object.name
      }
    }
  }

  service_config {
    available_memory    = "512M"
    timeout_seconds     = 60
    service_account_email = google_service_account.sa.email
    vpc_connector       = google_vpc_access_connector.connector.id
    ingress_settings    = "ALLOW_INTERNAL_ONLY"

    secret_environment_variables {
      key        = "API_KEY"
      secret     = google_secret_manager_secret.api_key.secret_id
      version    = "latest"
    }
  }
}
```

**Pulumi** (code-native, general-purpose languages):

```typescript
import * as gcp from "@pulumi/gcp";

const fn = new gcp.cloudfunctionsv2.Function("function", {
  name: "my-function",
  location: "us-central1",
  buildConfig: {
    runtime: "python311",
    entryPoint: "hello_http",
    source: {
      storageSource: {
        bucket: bucket.name,
        object: object.name,
      },
    },
  },
  serviceConfig: {
    availableMemory: "512M",
    timeoutSeconds: 60,
    serviceAccountEmail: sa.email,
    vpcConnector: connector.id,
    ingressSettings: "ALLOW_INTERNAL_ONLY",
    secretEnvironmentVariables: [{
      key: "API_KEY",
      secret: apiKeySecret.secretId,
      version: "latest",
    }],
  },
});
```

**What they solve**:
- **Stateful management**: Remote state tracking enables multi-user collaboration with locking
- **Dependency graphing**: Automatically understand that the storage bucket must exist before the function, service account before IAM binding
- **Plan/preview workflows**: See exactly what will change before applying
- **Multi-environment support**: Terraform Workspaces and Pulumi Stacks enable managing dev/staging/prod from a single codebase

**OpenTofu**: A community-driven fork of Terraform, fully compatible with the same `hashicorp/google` provider. Feature parity is guaranteed—the choice between Terraform and OpenTofu is about licensing and governance, not capabilities.

**Critical architectural detail**: Pulumi's GCP provider bridges to Terraform's provider, ensuring 100% feature parity. The difference is language (HCL vs. TypeScript/Python/Go), not capability.

**When to use it**: Production infrastructure. Always. These tools are the industry standard for managing cloud resources at scale.

### Level 4: Higher-Level Abstractions

**Crossplane**: Extends the Kubernetes API to manage cloud resources as Custom Resource Definitions (CRDs). For teams managing all infrastructure from Kubernetes, Crossplane provides a unified control plane. The trade-off is complexity—you're managing infrastructure via Kubernetes operators rather than dedicated IaC tools.

**Serverless Framework**: A multi-cloud abstraction layer that uses a `serverless.yml` file to define functions and automatically deploys them to the target platform (Google Cloud, AWS, Azure). Useful for teams building portable FaaS applications across multiple clouds.

**Google Deployment Manager**: Google's original native IaC tool using YAML or Jinja templates. This is **legacy** and not a strategic choice. It's notably absent from modern community discussions and best practice guides. Google's own investment focuses on their Terraform provider.

**Cloud Foundation Toolkit**: Not a deployment tool—it's a collection of production-ready **Terraform modules** maintained by Google. This reinforces Terraform's dominance as the standard for IaC on GCP.

---

## Production Essentials: Beyond "Hello World"

Deploying a function is easy. Deploying a **production-grade** function requires configuring critical features for performance, networking, security, and reliability.

### Performance: Eliminating Cold Starts

**The Problem**: A "cold start" is the latency penalty incurred when a new function instance initializes. For a Python function with heavy dependencies, this can take several seconds. If your upstream system has a 3-second timeout (e.g., Slack webhooks have a 300ms timeout), cold starts cause failures.

**The Solution (Gen 2)**: Configure **minimum instances**. Setting `min_instance_count = 1` (or `--min-instances=1`) keeps at least one instance warm and ready to serve requests, virtually eliminating cold start latency for most traffic.

**The Trade-off**: This is a cost-vs-performance decision. You're billed for the idle compute time of warm instances. But for latency-sensitive applications (interactive APIs, user-facing webhooks), the cost is justified.

### Networking: VPC Connectivity for Private Resources

**Use Case**: Your function needs to access private resources—a Cloud SQL database, Memorystore Redis instance, or internal microservice running on a VM.

**Solution**: Attach a **Serverless VPC Access connector** to the function:

```hcl
service_config {
  vpc_connector = google_vpc_access_connector.connector.id
  vpc_connector_egress_settings = "ALL_TRAFFIC"
}
```

**Setup**: The VPC connector is a separate resource that must be created first. It lives in the target VPC and requires a dedicated /28 subnet (16 IP addresses).

**Ingress Controls (Privacy)**: Use `ingress_settings = "ALLOW_INTERNAL_ONLY"` to make the function private, accessible only from within the VPC or project. This prevents public internet access.

**Egress Controls (Firewalling)**: Use `vpc_connector_egress_settings = "ALL_TRAFFIC"` to route all outbound traffic through the VPC. This enables using Cloud NAT to provide a static egress IP for the function—critical for integrating with third-party services that require IP allowlisting.

### Security: Secrets and IAM

**Anti-Pattern: Hardcoded Secrets**
Never hardcode API keys, database passwords, or other secrets in code or plain-text environment variables. Plain environment variables are visible to anyone with `roles/viewer` permission on the project.

**Best Practice: Secret Manager Integration**
Gen 2 provides native, platform-level integration with Google Secret Manager. Instead of your function *code* fetching secrets at runtime, the *platform* securely injects them as environment variables before the function starts:

```hcl
service_config {
  secret_environment_variables {
    key     = "DB_PASSWORD"
    secret  = "my-db-password"
    version = "latest"
  }
}
```

**IAM: Two Critical Identities**

1. **Invoker Identity (Who can call the function?)**
   - **Role**: `roles/run.invoker` (Gen 2) or `roles/cloudfunctions.invoker`
   - **Public**: Grant to `allUsers` (what `--allow-unauthenticated` does)
   - **Private**: Remove `allUsers`, grant only to specific service accounts

2. **Runtime Identity (What can the function do?)**
   - **Configuration**: `service_account_email` in service config
   - **Anti-Pattern**: Using the default Compute Engine service account, which often has broad Editor-level permissions
   - **Best Practice**: Create a dedicated, per-function service account with least-privilege permissions (e.g., `roles/storage.objectViewer` to read from a bucket, `roles/pubsub.publisher` to publish messages)

### Observability: Built-In Integration

Gen 2's Cloud Run foundation provides automatic integration with Google Cloud's observability suite:

- **Cloud Logging**: stdout/stderr automatically captured and viewable
- **Cloud Monitoring**: Key metrics (invocation count, execution time, memory usage, instance count) collected automatically
- **Error Reporting**: Unhandled exceptions automatically grouped and surfaced
- **Cloud Trace**: Distributed tracing to identify latency bottlenecks across services

No additional configuration required—it's built-in.

### Common Anti-Patterns (A Summary)

❌ **Non-idempotent code**: Event-driven functions receive "at-least-once" delivery. Code that creates two database entries for one event (instead of handling duplicates) is a common failure mode.

❌ **Global scope misuse**: Failing to initialize expensive clients (database connections, API clients) in the global scope forces re-initialization on every invocation, killing performance.

❌ **Missing outbound timeouts**: Calling an external API without a timeout can cause the function to hang for the full 60-minute timeout, leading to high costs.

❌ **Public unauthenticated functions**: Deploying with `--allow-unauthenticated` when the function should be internal, exposing it to the public internet.

❌ **Secrets in environment variables**: Using plain `environment_variables` instead of `secret_environment_variables`.

❌ **Using the default service account**: Granting the function broad, project-level permissions instead of least-privilege.

---

## The 80/20 Configuration Philosophy

The `gcloud` CLI requires only a handful of flags to deploy a "Hello World" function: name, region, runtime, entry point, source, and trigger. But production functions need a distinct, secondary set of configuration—VPC connectivity, secrets, performance tuning, and security controls.

This creates a clear 80/20 split:

### The Essential 80%

These are non-negotiable, required to deploy *any* function:

- **`name`**: Resource identifier
- **`location`**: Region (e.g., `us-central1`)
- **`runtime`**: Language and version (e.g., `python311`, `nodejs22`)
- **`entry_point`**: Function name in source code
- **`source`**: Where the code lives (GCS bucket + object)
- **`trigger`**: Either HTTP or event-driven (Pub/Sub, Storage, Firestore)

### The Critical 20%

These are the "optional but common" fields needed to run in production:

- **`service_account_email`**: Least-privilege runtime identity
- **`available_memory`**: Common tuning parameter (default: 256MB; production often needs 512MB-1GB)
- **`timeout_seconds`**: Common tuning parameter (default: 60s; long-running tasks may need more)
- **`environment_variables`**: Non-sensitive configuration
- **`secret_environment_variables`**: Secure secrets from Secret Manager
- **`vpc_connector`**: For private resource access
- **`ingress_settings`**: For private functions (`ALLOW_INTERNAL_ONLY`)
- **`min_instance_count`**: For cold start mitigation
- **`max_instance_count`**: For cost/concurrency control

### The Long Tail (Advanced)

Most teams never touch these:

- **`build_environment_variables`**: Custom build-time configuration
- **`build_worker_pool`**: Using custom build infrastructure
- **`docker_repository`**: Custom container registry
- **`secret_volumes`**: Alternative to environment variables (more complex)

### Source Code Deployment: The GCS Bucket Pattern

**Historical context**: There were two "code-aware" paths—GCS bucket and Cloud Source Repositories. However, **Cloud Source Repositories was deprecated for new customers on June 17, 2024**. This solidifies the **GCS bucket pattern as the only stable, long-term path** for IaC.

**The Pattern**:
1. Package source code (e.g., ZIP file)
2. Upload to a GCS bucket
3. Reference the bucket and object in the function configuration

```hcl
build_config {
  source {
    storage_source {
      bucket = "my-code-bucket"
      object = "my-function-v1.2.3.zip"
    }
  }
}
```

**CI/CD best practice**: Hash the source directory and use the hash in the object name (e.g., `code-abcdef01.zip`). If the hash hasn't changed, the object name is the same, and Terraform/Pulumi skips redeployment—preventing unnecessary rebuilds.

### Real-World Configuration Examples

**Example 1: Dev - Basic HTTP Function**

```hcl
resource "google_cloudfunctions2_function" "dev" {
  name     = "http-hello-dev"
  location = "us-central1"

  build_config {
    runtime     = "python311"
    entry_point = "hello_http"
    source {
      storage_source {
        bucket = "planton-dev-code"
        object = "http-hello-dev.zip"
      }
    }
  }

  service_config {
    # All other fields use GCP defaults
    # 256MB memory, 60s timeout, default service account
  }
}

resource "google_cloud_run_service_iam_member" "public" {
  service  = google_cloudfunctions2_function.dev.name
  location = google_cloudfunctions2_function.dev.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}
```

**Example 2: Staging - Pub/Sub Trigger with VPC and Secrets**

```hcl
resource "google_cloudfunctions2_function" "staging" {
  name     = "pubsub-worker-staging"
  location = "europe-west1"

  build_config {
    runtime     = "go122"
    entry_point = "ProcessMessage"
    source {
      storage_source {
        bucket = "planton-staging-code"
        object = "pubsub-worker-v1.2.zip"
      }
    }
  }

  service_config {
    available_memory      = "512M"
    timeout_seconds       = 300
    service_account_email = google_service_account.worker.email
    vpc_connector         = google_vpc_access_connector.staging.id
    ingress_settings      = "ALLOW_INTERNAL_ONLY"

    secret_environment_variables {
      key     = "DB_PASSWORD"
      secret  = google_secret_manager_secret.db_password.secret_id
      version = "latest"
    }
  }

  event_trigger {
    trigger_region        = "europe-west1"
    event_type            = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic          = google_pubsub_topic.jobs.id
    retry_policy          = "RETRY_POLICY_RETRY"
  }
}
```

**Example 3: Prod - High-Availability HTTP Function**

```hcl
resource "google_cloudfunctions2_function" "prod" {
  name     = "api-gateway-prod"
  location = "us-east1"

  build_config {
    runtime     = "nodejs22"
    entry_point = "handleRequest"
    source {
      storage_source {
        bucket = "planton-prod-code"
        object = "api-gateway-v3.4.1.zip"
      }
    }
  }

  service_config {
    available_memory      = "1024M"
    timeout_seconds       = 60
    service_account_email = google_service_account.api_gateway.email
    min_instance_count    = 2  # HA: 2 warm instances
    max_instance_count    = 100  # Cost control

    secret_environment_variables {
      key     = "JWT_SECRET"
      secret  = google_secret_manager_secret.jwt.secret_id
      version = "latest"
    }

    secret_environment_variables {
      key     = "DB_CONNECTION"
      secret  = google_secret_manager_secret.db_conn.secret_id
      version = "latest"
    }
  }
}

resource "google_cloud_run_service_iam_member" "private" {
  service  = google_cloudfunctions2_function.prod.name
  location = google_cloudfunctions2_function.prod.location
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.api_client.email}"
}
```

---

## Cloud Functions vs. Cloud Run: When to Use Each

Understanding that Gen 2 functions *are* Cloud Run services raises a natural question: **When should I use Cloud Functions vs. Cloud Run directly?**

**Use Cloud Functions (Gen 2) when**:
- You want the simplest "just deploy my code" experience (source-first workflow)
- Your application is primarily **event-driven** (Pub/Sub, Storage, Firestore)
- The unit of deployment is a small function or snippet, not a complex web server
- You want the platform to handle containerization automatically

**Use Cloud Run directly when**:
- You already have a containerized application (a Dockerfile)
- Your application requires a language, library, or binary not supported by standard GCF runtimes
- You need total control over the container (custom system packages, sidecar containers, custom build process)
- You're deploying a multi-endpoint web application rather than a single function

**The strategic insight**: Cloud Functions is a *developer experience* abstraction for the most common serverless use case (event-driven, source-first, FaaS). Cloud Run is the underlying *runtime* that provides more control and flexibility when you need it.

---

## Cost and Performance

### Pricing Model

**Gen 1**: Priced on invocations, compute time (GB-sec), and networking.

**Gen 2**: Uses **Cloud Run pricing**: vCPU-seconds, memory (GB-sec), invocations, and networking.

**The Surprise**: For low-traffic functions, Gen 2 can be *more expensive* than Gen 1. The reason: a default Gen 1 function (256MB) was allocated a fractional vCPU (e.g., 1/6th). A default Gen 2 function gets **1 full vCPU** to enable high-concurrency features.

However, for high-traffic functions, Gen 2 is **vastly cheaper**. Its ability to serve 80+ concurrent requests on a single instance means it scales far more efficiently than Gen 1 (which would need 80+ instances for the same load).

### Cost Optimization

1. **Right-Sizing**: Don't over-provision memory or vCPU. Analyze utilization metrics and adjust.
2. **min_instances Trade-off**: Setting `min_instances > 0` is a conscious decision to trade cost for performance. You're billed for idle compute.
3. **Concurrency Tuning**: Default concurrency is 80. If your function is I/O-bound (waiting for APIs), high concurrency is efficient. If CPU-bound, set concurrency to 1 and potentially increase vCPU.

### Performance Tuning

- **Global Scope**: Initialize expensive objects (database clients, API clients) in the global scope, outside the function handler. They're created once per instance, not once per invocation.
- **Concurrency**: Don't set concurrency too low unless CPU-bound. Low concurrency forces more instances, increasing cold starts.
- **Timeouts**: While Gen 2 supports 60-minute timeouts, this is often an anti-pattern. For long-running tasks, trigger a Cloud Task or Cloud Workflow and return immediately, allowing asynchronous processing.

---

## The Planton Choice: Protobuf + Pulumi

Planton models GCP Cloud Functions as protobuf-defined resources (`GcpCloudFunctionSpec`) deployed via Pulumi. This design reflects strategic choices:

### Why Protobuf-Defined APIs

1. **Language-agnostic schema**: Protobuf generates client libraries for every language (Go, TypeScript, Python, Java), enabling multi-language tooling.
2. **Built-in validation**: Buf validate constraints enforce correctness at the API layer, preventing invalid configurations before deployment.
3. **Versioned evolution**: Protobuf's backward compatibility guarantees mean v1 clients continue working as the schema evolves.
4. **80/20 forcing function**: Explicit field definitions force conscious decisions about what to include, focusing on essential and common fields while omitting rarely-used parameters.

### Why Pulumi for Deployment

While Terraform is the industry standard, Planton uses Pulumi:

**Programmatic generation**: Planton generates Pulumi programs from protobuf specs. Pulumi's code-native approach (TypeScript, Go, Python) makes programmatic generation significantly simpler than HCL generation.

**Provider bridging**: Pulumi reuses Terraform providers, inheriting their maturity and feature coverage. This means 100% feature parity with Terraform while maintaining code-native flexibility.

**State management parity**: Pulumi's state management is architecturally identical to Terraform (remote backends, locking, plan/preview workflows).

**Testing**: Pulumi programs are executable code in general-purpose languages, enabling standard testing frameworks to validate infrastructure logic before deployment.

### What We Expose (and Exclude)

Planton's API will include:

**Essential (The 80%)**:
- `name`, `location`, `runtime`, `entry_point`
- `source` (GCS bucket + object)
- `trigger` (HTTP or event-driven)

**Critical (The 20%)**:
- `service_account_email`
- `available_memory`, `timeout_seconds`
- `environment_variables`, `secret_environment_variables`
- `vpc_connector`, `ingress_settings`
- `min_instance_count`, `max_instance_count`

**Excluded (Advanced 5%)**:
- `build_worker_pool`, `docker_repository`
- `secret_volumes` (environment variables are simpler)
- Advanced health check tuning

This isn't about limiting capabilities—it's about presenting the 20% of configuration that 80% of production deployments need, making the correct path the easy path.

---

## Conclusion: The FaaS Abstraction on a CaaS Foundation

The serverless landscape has matured. Google Cloud Functions (Gen 2) isn't a proprietary FaaS platform—it's a carefully crafted developer experience layer on top of Cloud Run and Eventarc. This architectural choice brings massive benefits: superior performance, broader event sources, advanced traffic management, and Cloud Run's production-proven reliability.

For teams practicing infrastructure-as-code, the deployment choice is clear: **Terraform and Pulumi are the production standard**. Both offer mature, feature-complete providers with 100% API coverage. The decision between them is about philosophy (HCL DSL vs. general-purpose languages), not capability.

Planton's approach—protobuf-defined APIs generating Pulumi deployments—optimizes for a different dimension: **API clarity over configuration completeness**. By exposing the essential 20% of configuration through a typed, validated schema, we make the correct path the easy path. Teams needing advanced features can drop to Pulumi or Terraform directly, but most teams ship faster by starting with 80/20 defaults.

The paradigm shift isn't about serverless functions—it's about treating infrastructure configuration as a versioned, typed, validated API rather than an ad-hoc collection of key-value pairs. That's the foundation for building production infrastructure at scale.

---

## Further Reading

- [Google Cloud Functions Documentation](https://cloud.google.com/functions/docs) - Official documentation
- [Cloud Run Documentation](https://cloud.google.com/run/docs) - Understanding the underlying runtime
- [Eventarc Documentation](https://cloud.google.com/eventarc/docs) - Event-driven trigger configuration
- [Terraform Google Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions2_function) - Terraform resource documentation
- [Pulumi GCP Provider](https://www.pulumi.com/registry/packages/gcp/api-docs/cloudfunctionsv2/function/) - Pulumi resource documentation
- [Cloud Functions Best Practices](https://cloud.google.com/run/docs/tips/functions-best-practices) - Google's production guidance

