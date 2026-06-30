# Deploying DigitalOcean Functions: The Path to Production-Ready Serverless

## Introduction

When serverless first emerged, the prevailing wisdom was simple: deploy your code, configure a trigger, and let the cloud handle the rest. No servers, no infrastructure, just pure business logic. For most Function-as-a-Service (FaaS) platforms, this story holds reasonably true. But DigitalOcean Functions challenges that assumption in a subtle yet critical way: **what you think is "DigitalOcean Functions" is actually two fundamentally different products with vastly different capabilities.**

This distinction matters strategically. The deployment method you choose doesn't just affect your workflow—it determines whether your functions can securely access a database, whether you get production-grade monitoring, and whether you can manage them declaratively with Infrastructure-as-Code. Get it wrong, and you'll end up with brittle CI/CD scripts, security vulnerabilities, and a serverless platform that can't scale beyond hobbyist projects.

**DigitalOcean Functions** is DigitalOcean's answer to AWS Lambda, GCP Cloud Functions, and Azure Functions. It's a FaaS platform designed for lightweight, event-driven workloads: building serverless APIs, processing webhooks, running scheduled tasks. Unlike the hyperscalers, DigitalOcean emphasizes simplicity over feature breadth. You get a curated list of runtimes (Node.js, Python, Go, PHP), predictable pricing, and a streamlined developer experience. No complex concurrency tuning, no provisioned capacity modes, no multi-tiered pricing calculators. Just functions.

But here's where it gets interesting: DigitalOcean Functions exists in two distinct and mutually exclusive deployment models, each with dramatically different capabilities:

1. **Standalone Functions**: Deployed via the `doctl serverless` CLI into a namespace. Simple, CLI-driven, suitable for development and testing. But critically: **no VPC networking, no production monitoring, and zero Infrastructure-as-Code support.**

2. **App Platform Functions**: Deployed as a component within a DigitalOcean App Platform application, managed via an `app-spec.yml`. This is the **only production-viable path**, providing VPC integration for secure database access, full monitoring via DigitalOcean Insights, and native IaC support through Terraform and Pulumi.

This guide explains the deployment landscape for DigitalOcean Functions, why the Standalone model is an anti-pattern for production, and how Planton abstracts the correct approach (App Platform) into a clean, protobuf-defined API.

---

## The Deployment Spectrum: From CLI Scripts to Production IaC

Not all deployment methods are created equal. Here's how the approaches stack up, from what to avoid to what works at scale:

### Level 0: The Web Console (Learning Tool Only)

**What it is:** Creating functions directly in the DigitalOcean Control Panel using the built-in code editor.

**What it solves:** Exploration. You can quickly prototype a function, test a runtime, and understand the platform's capabilities. The UI provides a friendly interface for setting environment variables, configuring memory and timeout limits, and triggering test executions.

**What it doesn't solve:** Repeatability, version control, or deployment safety. Any settings you configure in the console will be **overwritten** the next time you deploy from the CLI or App Platform. The documentation explicitly warns against relying on the console for configuration management. Treat it as a read-only inspection tool, not a deployment method.

**Verdict:** Use it to learn the platform and inspect existing functions. Never rely on it for staging or production.

---

### Level 1: Standalone Functions via `doctl serverless` (The Anti-Pattern)

**What it is:** The primary CLI workflow for creating and managing functions. You create a namespace, initialize a project (which generates a `project.yml` manifest), write your function code, and deploy it with `doctl serverless deploy`.

**What it solves:** Automation. You can script deployments, integrate them into CI/CD pipelines (GitHub Actions, GitLab CI), and version-control your `project.yml` configuration. The CLI handles building, packaging, and uploading your function code.

**What it doesn't solve:** Production-grade requirements. Standalone Functions are crippled by three critical gaps:

1. **No VPC Networking**: Functions run in a public, multi-tenant environment. To connect to a DigitalOcean Managed Database, you must expose that database to the public internet (0.0.0.0/0), which is a **severe security anti-pattern**. There is no way to attach Standalone Functions to a Virtual Private Cloud.

2. **No Monitoring**: There are no metrics. No dashboards. No integration with DigitalOcean Insights. Logs must be manually polled after execution using `doctl serverless activations logs <function-name>`. You get no time-series graphs for CPU, memory, latency, or request counts.

3. **No IaC Support**: The official Terraform and Pulumi providers have **zero resources** for managing functions, namespaces, or triggers. This is not an oversight or documentation gap—the resources do not exist. There is no public, documented API for deploying Standalone Functions. The `doctl` CLI calls an undocumented internal API that is not exposed to third-party tools.

**CI/CD Reality**: All automated deployments for Standalone Functions follow the same brittle pattern: install the `doctl` CLI in a GitHub Action or GitLab runner, authenticate with an API token, and execute `doctl serverless deploy`. This is **imperative scripting, not declarative Infrastructure-as-Code**. It's fragile, couples your deployment to a CLI tool's behavior, and provides no state tracking or idempotency.

**Verdict:** Acceptable for local development, throwaway prototypes, or learning the platform. **Entirely unsuitable for production.** Any function that accesses a database or requires monitoring must use a different approach.

---

### Level 2: App Platform Functions (The Production Solution)

**What it is:** Deploying functions as a component within a DigitalOcean App Platform application. The app is defined by an `app-spec.yml` manifest (or created via Terraform/Pulumi's `digitalocean_app` resource). The `app-spec` includes a `functions` component that points to a Git repository containing your function code and `project.yml` file. On deployment, App Platform's build system clones the repo, builds the function, and deploys it as part of the app.

**Example `app-spec.yml`:**

```yaml
name: my-function-app
region: nyc1
functions:
  - name: api-handler
    github:
      repo: myorg/my-functions
      branch: main
      deploy_on_push: true
    source_dir: /functions/api-handler
    envs:
      - key: DB_URL
        value: ${db.DATABASE_URL}  # Secret templating from App Platform
```

**What it solves:** Everything that Standalone Functions lack:

1. **VPC Integration**: App Platform natively supports outbound VPC networking. Your function can securely access a Managed Database or Droplet via its private network address, with zero public exposure.

2. **Production Monitoring**: Full integration with DigitalOcean Insights. You get time-series graphs for CPU, memory, request latency, restart counts, and throughput. Logs are streamed and can be forwarded to third-party services (Logtail, Datadog).

3. **Declarative IaC**: The `digitalocean_app` resource is fully supported by Terraform and Pulumi. You define the app-spec declaratively, version-control it, and manage the entire lifecycle (create, update, rollback) with `terraform apply` or `pulumi up`.

4. **Secret Management**: App Platform supports encrypted environment variables. Secrets are stored securely and templated into your `project.yml` at build time (e.g., `DB_URL: "${DB_URL}"`). This is vastly superior to hardcoding plaintext secrets in a `project.yml` file.

5. **Versioning and Rollbacks**: Every deployment is atomic and tied to a Git commit. The App Platform dashboard provides **one-click rollbacks** to any previous successful deployment. All deployments are zero-downtime.

6. **Multi-Environment Support**: The App Platform model is designed for dev/staging/prod workflows. You can link different apps to different Git branches and use the same IaC code with environment-specific variables.

**What it doesn't solve:** The underlying platform limitations (curated runtimes only, no custom Docker images, 48 MB bundle size limit, cold starts). But it makes those constraints manageable within a production-ready deployment framework.

**Verdict:** This is the **only correct approach for production**. If your function accesses a database, requires monitoring, or needs to be managed with IaC, it must be deployed via App Platform.

---

## The Infrastructure-as-Code Landscape

The IaC story for DigitalOcean Functions is stark: **Standalone Functions are an IaC black hole. App Platform Functions are fully supported.**

### The Standalone Function Gap

- **Terraform**: The official `digitalocean/digitalocean` provider has no resources for `digitalocean_function`, `digitalocean_function_namespace`, or `digitalocean_function_trigger`. A GitHub issue (terraform-provider-digitalocean#1167) remains open, explicitly requesting these resources.

- **Pulumi**: The `@pulumi/digitalocean` package is a bridged provider generated from the Terraform provider. It inherits the exact same limitations. No function resources exist.

**Conclusion**: It is **impossible** to declaratively manage the full lifecycle of a Standalone Function using any mainstream IaC tool. The only automation path is imperative CLI scripting in CI/CD pipelines—a brittle workaround, not a solution.

---

### App Platform: The IaC Solution

Both Terraform and Pulumi provide mature, production-ready support for the `digitalocean_app` resource. This resource accepts an `app-spec` (in YAML or JSON format) that can declaratively define a `functions` component.

**Terraform Example:**

```hcl
provider "digitalocean" {
  token = var.do_api_token
}

resource "digitalocean_app" "my_function" {
  spec {
    name   = "my-function-app"
    region = "nyc1"

    function {
      name = "api-handler"
      
      github {
        repo           = "myorg/my-functions"
        branch         = "main"
        deploy_on_push = true
      }

      source_dir = "/functions/api-handler"

      env {
        key   = "DB_URL"
        value = digitalocean_database_cluster.pg.uri
        type  = "SECRET"
      }
    }
  }
}
```

**Pulumi Example (TypeScript):**

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const app = new digitalocean.App("my-function", {
  spec: {
    name: "my-function-app",
    region: "nyc1",
    functions: [{
      name: "api-handler",
      github: {
        repo: "myorg/my-functions",
        branch: "main",
        deployOnPush: true,
      },
      sourceDir: "/functions/api-handler",
      envs: [{
        key: "DB_URL",
        value: dbCluster.uri,
        type: "SECRET",
      }],
    }],
  },
});
```

### IaC Tool Comparison

Both tools are production-ready. The choice is primarily a matter of team preference and existing tooling:

**Terraform: The Default Choice**

- **Maturity**: Terraform has been the IaC standard for years. Broad ecosystem, extensive community support, well-documented.
- **Configuration Model**: HCL is declarative and straightforward for standard use cases.
- **Strengths**: Familiarity across ops teams, clear plan/apply workflow, wide adoption.
- **Limitations**: HCL is less expressive than a full programming language. Complex conditional logic can be verbose.

**Pulumi: The Programmer's IaC**

- **Maturity**: Newer than Terraform, but production-ready. Actively maintained.
- **Configuration Model**: Real programming languages (TypeScript, Python, Go). Full language expressiveness (loops, conditionals, unit tests).
- **Strengths**: Better for complex orchestration logic, dynamic resource generation, integration with application code.
- **Limitations**: Smaller community, requires a language runtime (Node.js, Python, etc.).

**Recommendation**: Default to Terraform for simplicity and ecosystem maturity. Choose Pulumi if your team prefers coding infrastructure in familiar languages or needs advanced orchestration logic. Both provide equivalent resource coverage for App Platform.

---

## Production Essentials: Runtimes, Triggers, and Observability

### Supported Runtimes

DigitalOcean Functions officially supports:

- **Node.js**: 14, 18
- **Python**: 3.9, 3.11
- **Go**: 1.17, 1.20
- **PHP**: 8.0, 8.2

**Critical Limitation**: DigitalOcean Functions **does not support custom runtimes via container images**. Unlike AWS Lambda, GCP Cloud Functions, and Azure Functions, you cannot package your code in a Docker container. If you need Rust, Deno, or any unsupported runtime, you must deploy a Docker image as a **Web Service component on App Platform** (a different product with different semantics).

---

### Execution Model and Cold Starts

Functions are **stateless**. When invoked, the platform provisions an execution environment, runs your code, and then tears down the resources. If a function hasn't been called recently (a "cold" function), provisioning a new instance introduces latency.

**Typical Cold Start Times**: 400-600ms. Language choice matters—a compiled Go binary will have shorter cold starts than an interpreted Python or Node.js script with heavy dependencies.

**Optimization Strategies**:

- Use lightweight runtimes (Go > Node.js > Python for cold start speed)
- Minimize dependencies in `package.json`, `requirements.txt`, or `go.mod`
- Keep function code small (the platform has a 48 MB bundle size limit)

---

### Triggers

Functions are executed in response to triggers:

1. **HTTP Triggers (Web Functions)**: The primary use case. A function can be exposed as a "web function," which assigns it a unique, public URL for invocation via standard HTTP methods (GET, POST, etc.). This is configured in `project.yml` with `web: true`.

2. **Scheduled Triggers (Cron)**: Functions can be invoked on a recurring schedule using standard cron syntax. This is ideal for cleanup jobs, nightly reports, or periodic data syncing.

**Example `project.yml` (HTTP Function):**

```yaml
packages:
  - name: api
    functions:
      - name: hello
        runtime: 'nodejs:18'
        web: true
        limits:
          timeout: 15000  # 15 seconds
          memory: 256     # 256 MB
```

**Example `project.yml` (Scheduled Task):**

```yaml
packages:
  - name: cron-jobs
    functions:
      - name: cleanup
        runtime: 'python:3.11'
        web: false  # Not exposed to HTTP
        limits:
          timeout: 60000  # 60 seconds
          memory: 512
        triggers:
          - name: nightly
            sourceType: scheduler
            sourceDetails:
              cron: "0 0 * * *"  # At midnight
```

---

### Observability: The Critical Divide

**Standalone Functions**:
- **Logging**: Primitive. Logs must be manually polled after execution using `doctl serverless activations logs <function-name>`.
- **Monitoring**: Non-existent. No dashboards, no metrics, no integration with DigitalOcean Insights.

**App Platform Functions**:
- **Logging**: Production-grade. Logs are streamed and available in the App Platform dashboard. They can be forwarded to third-party logging services (Logtail, Datadog).
- **Monitoring**: Full-featured. Functions appear in the "Insights" tab with time-series graphs for CPU usage, memory usage, restart counts, request latency, and throughput.

**Conclusion**: Any function that runs in production needs observability. This alone disqualifies Standalone Functions.

---

## Configuration Analysis: The 80/20 Principle

Most users only need to configure a small subset of available settings. Here's the essential 80%:

### Essential Fields (80%)

- **`function_name`**: The name of the function (e.g., `api-handler`)
- **`package_name`**: The package (group) name (e.g., `default` or `main`)
- **`runtime`**: The language runtime (e.g., `nodejs:18`, `python:3.11`, `go:1.20`)
- **`web`**: Expose as an HTTP endpoint? (`true` or `false`)

### Typical Fields (15%)

- **`memory`**: Memory allocation in MB (default: 256 MB, range: 128-512 MB)
- **`timeout`**: Timeout in milliseconds (default: 3000 ms, range: 5,000-60,000 ms)
- **`environment`**: Key/value environment variables

### Advanced Fields (5%)

- **`main`**: Specify the entrypoint function (e.g., `main` for Go, `main` for Python)
- **`triggers`**: Scheduled (cron) triggers
- **`webSecure`**: Secure a web function with a token (for private APIs)

---

## Configuration Examples

### Example 1: Simple HTTP Function (Node.js)

**Use Case**: A public API endpoint.

```yaml
packages:
  - name: simple-api
    functions:
      - name: hello
        runtime: 'nodejs:18'
        web: true
        limits:
          timeout: 15000  # 15 seconds
          memory: 256     # 256 MB
```

---

### Example 2: Scheduled Task (Python)

**Use Case**: A nightly cleanup script running at midnight.

```yaml
packages:
  - name: cron-jobs
    functions:
      - name: cleanup-task
        runtime: 'python:3.11'
        web: false  # Not exposed to HTTP
        limits:
          timeout: 60000  # 60 seconds
          memory: 512
        triggers:
          - name: nightly-cron
            sourceType: scheduler
            sourceDetails:
              cron: "0 0 * * *"  # At midnight
```

---

### Example 3: Production API Endpoint (Go) with Secrets

**Use Case**: A high-performance API backend that connects to a database.

```yaml
environment:
  # This ${...} syntax templates the variable from App Platform's encrypted secret store
  DB_URL: "${DB_URL}"

packages:
  - name: go-api
    functions:
      - name: get-users
        runtime: 'go:1.20'
        web: true
        limits:
          timeout: 30000  # 30 seconds
          memory: 1024    # 1 GB
```

---

## Planton's Approach: Abstraction Over App Platform

Planton's `DigitalOceanFunction` resource is designed as a **high-level abstraction over the `digitalocean_app` resource**. This provides users with a declarative, production-ready API without forcing them to understand the nuances of App Platform specs or the pitfalls of Standalone Functions.

### What We Abstract

When a user defines a `DigitalOceanFunction` resource, Planton's provider:

1. **Synthesizes an `app-spec.yml`**: Programmatically generates the app specification with a single `functions` component, populated from the user's input (source code path, runtime, region, environment variables, secrets).

2. **Creates a Dedicated App**: Provisions a **new, dedicated DigitalOcean App Platform application** for this function. The name of the Planton resource maps to the name of the `digitalocean_app`.

3. **Manages the Lifecycle**: Updates, rollbacks, and deletions are handled declaratively through the underlying `digitalocean_app` resource.

### The Protobuf API

The `DigitalOceanFunctionSpec` includes the essential 80/20 fields:

```protobuf
message DigitalOceanFunctionSpec {
  // Essential (80%)
  string name = 1;                    // Unique name (becomes App Platform app name)
  string region = 2;                  // Region (e.g., "nyc1", "ams3")
  string function_name = 3;           // Function name (maps to project.yml)
  string runtime = 4;                 // Runtime (e.g., "nodejs:18", "python:3.11")
  
  // Source Code (via GitHub integration)
  string github_repo = 5;             // GitHub repo (e.g., "myorg/my-functions")
  string github_branch = 6;           // Branch (e.g., "main")
  string source_directory = 7;        // Path to function project (e.g., "/src/my-function")
  
  // Typical (15%)
  int32 memory_mb = 8;                // Memory in MB (default: 256)
  int32 timeout_ms = 9;               // Timeout in milliseconds (default: 3000)
  map<string, string> environment_variables = 10;  // Non-secret env vars
  map<string, string> secret_environment_variables = 11;  // Secrets (encrypted)
  
  // Advanced (5%)
  string entrypoint = 12;             // Entrypoint function name
  string cron_schedule = 13;          // Cron schedule (e.g., "0 * * * *")
}
```

### Default Choices

- **App Platform by Default**: We always provision via App Platform, ensuring VPC support, monitoring, and IaC compatibility.
- **Secret Management**: We use App Platform's encrypted environment variables for secrets, not plaintext in `project.yml`.
- **Region Matching**: We enforce that functions and dependent resources (databases, VPCs) are in the same region.

### Why This Approach?

- **Production-Ready by Default**: Users get VPC networking, monitoring, and rollback capabilities without needing to understand App Platform internals.
- **Declarative and Future-Proof**: The abstraction is built on stable, supported IaC primitives (`digitalocean_app`), not brittle CLI scripting.
- **Simplicity**: Users provide a minimal spec (name, runtime, source repo) and get a fully functional, production-ready serverless function.

---

## Key Takeaways

1. **DigitalOcean Functions is two products**: Standalone Functions (CLI-based, development-only) and App Platform Functions (production-ready, IaC-supported). They are mutually exclusive.

2. **Standalone Functions lack critical production features**: No VPC networking, no monitoring, no IaC support. They are unsuitable for any function that accesses a database or requires observability.

3. **App Platform Functions are the production solution**: They provide VPC integration, DigitalOcean Insights monitoring, declarative IaC via Terraform/Pulumi, encrypted secret management, and one-click rollbacks.

4. **There is no IaC for Standalone Functions**: The Terraform and Pulumi providers have zero resources for managing functions, namespaces, or triggers. All automation relies on brittle CLI scripting.

5. **The 80/20 config is name, runtime, source repo, and memory/timeout**: Advanced features (custom runtimes, complex triggers) are either not supported or handled at the application level.

6. **Planton abstracts the correct approach**: Our `DigitalOceanFunction` resource provisions functions via App Platform, giving users a production-ready serverless platform with a clean, protobuf-defined API.

---

## Further Reading

- **DigitalOcean Functions Documentation**: [DigitalOcean Docs - Functions](https://docs.digitalocean.com/products/functions/)
- **App Platform Functions**: [How to Manage Functions in App Platform](https://docs.digitalocean.com/products/app-platform/how-to/manage-functions/)
- **App Specification Reference**: [App Spec YAML](https://docs.digitalocean.com/products/app-platform/reference/app-spec/)
- **Terraform DigitalOcean Provider**: [DigitalOcean Terraform Reference](https://docs.digitalocean.com/reference/terraform/)
- **Pulumi DigitalOcean Provider**: [Pulumi Registry - DigitalOcean](https://www.pulumi.com/registry/packages/digitalocean/)

---

**Bottom Line**: DigitalOcean Functions is a compelling serverless platform when deployed correctly. The Standalone model is a dead end for production. App Platform Functions give you VPC security, production monitoring, and declarative IaC. Planton wraps this into a simple, protobuf-defined API that abstracts away the complexity and gives you a production-ready serverless function by default.

