---
title: "Deploying Apps on DigitalOcean App Platform: A Production Guide"
description: "Deploying Apps on DigitalOcean App Platform: A Production Guide deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceanappplatformservice"
---

# Deploying Apps on DigitalOcean App Platform: A Production Guide

## Introduction

The Platform-as-a-Service (PaaS) model promised to free developers from infrastructure concerns: just push code and let the cloud handle the rest. Heroku pioneered this experience over a decade ago, making deployment as simple as `git push heroku main`. But as Heroku's pricing ballooned and its feature set stagnated, developers sought alternatives that offered similar simplicity without the vendor lock-in or cost premium.

**DigitalOcean App Platform** emerged as one of the most compelling Heroku alternatives—a fully-managed PaaS that handles containerized deployments, autoscaling, SSL, and CDN integration while maintaining transparent, predictable pricing. Unlike managed Kubernetes (which demands orchestration expertise) or bare Droplets (which require server management), App Platform abstracts infrastructure entirely. You define what to deploy (source code or container images), how to run it (instance sizes and scaling rules), and where to route it (domains and paths). DigitalOcean handles provisioning, load balancing, SSL certificates, health checks, and zero-downtime deployments.

What makes App Platform strategically interesting is its positioning: simpler than Kubernetes, more flexible than traditional PaaS, and tightly integrated with DigitalOcean's ecosystem (Managed Databases, Container Registry, Spaces object storage, private VPC networking). It supports multiple deployment patterns—Git-based builds with Cloud Native Buildpacks, custom Dockerfiles, or pre-built container images from registries. It offers both web services (HTTP traffic), background workers (queue processing), scheduled jobs (migrations, cron tasks), and static sites (CDN-backed frontends).

But as with any abstraction, knowing what's under the hood matters. How you deploy to App Platform—manually through a web console, via CLI scripts, through Infrastructure-as-Code tools, or with higher-level frameworks like OpenMCF—determines whether you build a maintainable, production-grade system or a fragile snowflake that crumbles under real-world demands.

This guide explains the deployment landscape for DigitalOcean App Platform, the maturity progression from anti-patterns to production-ready approaches, and how OpenMCF abstracts these complexities into a clean, protobuf-defined API that follows the 80/20 principle.

---

## The Deployment Spectrum: From Manual to Production

Not all deployment methods are created equal. Here's the progression from what to avoid to what works at scale:

### Level 0: The Web Console (Anti-Pattern for Repeatable Deployments)

**What it is:** Clicking through DigitalOcean's dashboard to configure apps—selecting source repos, setting instance sizes, adding environment variables, and configuring domains through a graphical wizard.

**What it solves:** Discoverability. The console is excellent for learning App Platform's capabilities, understanding the relationship between concepts (services vs workers vs jobs), and experimenting with configurations. The UI also generates an App Spec YAML after configuration, which you can download and use as a starting point for Infrastructure-as-Code.

**What it doesn't solve:** Repeatability, version control, or team collaboration. Manual configurations drift over time as different team members make ad-hoc changes through the UI. You can't code-review a button click. When something breaks in production, reconstructing the exact configuration becomes archaeology. Secret rotation, multi-environment deployments (dev/staging/prod), and disaster recovery are all painful.

**Common mistakes:**
- Configuring secrets directly in the UI without documenting them elsewhere (good luck rotating them consistently)
- Manually deploying different configurations to staging and production, leading to unpredictable behavior when promoting code
- Forgetting to enable auto-deploy or health checks because there's no checklist or peer review

**Verdict:** Use the console to explore and learn, but **never** for staging or production deployments. It's an educational tool, not an operational one.

---

### Level 1: CLI Scripting (Automation Without State)

**What it is:** Using DigitalOcean's `doctl` CLI to create and update apps by supplying an App Spec YAML:

```bash
doctl apps create --spec app.yaml
doctl apps update $APP_ID --spec app-updated.yaml
```

**What it solves:** Automation and scriptability. You can version-control the App Spec YAML, integrate deployments into CI/CD pipelines, and use scripting logic (loops, conditionals, variable substitution) to generate configurations dynamically. The `doctl` CLI is well-documented, supports JSON output for parsing, and handles authentication via API tokens.

**What it doesn't solve:** State management and idempotency. Shell scripts don't track what was created or changed. If a script runs twice, you might create duplicate apps or fail because resources already exist. Secrets handling is awkward—you have to decide whether to embed encrypted secrets in the YAML (which DO provides as `EV[...]` after first deploy) or inject them at runtime via environment variables. Rollback requires manual intervention.

**When it's acceptable:** CI/CD pipelines where the entire lifecycle is scripted (create, deploy, test, destroy). For example, ephemeral preview environments for pull requests where you create a fresh app, run tests, and tear it down. The CLI's synchronous nature (it waits for deployments to complete or fail) makes it suitable for integration tests.

**Verdict:** Better than the console for automation, but insufficient for production infrastructure that needs state tracking, drift detection, and declarative management.

---

### Level 2: Infrastructure-as-Code with Terraform (Production-Grade, HCL Flavor)

**What it is:** Using the official DigitalOcean Terraform provider's `digitalocean_app` resource to declaratively define App Platform deployments:

```hcl
provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_app" "api" {
  spec {
    name   = "production-api"
    region = "nyc"

    service {
      name               = "api"
      instance_size_slug = "professional-xs"
      instance_count     = 2

      git {
        repo_clone_url = "https://github.com/myorg/myapp.git"
        branch         = "main"
      }

      env {
        key   = "DATABASE_URL"
        value = "${digitalocean_database_cluster.postgres.uri}"
        type  = "SECRET"
      }

      autoscaling {
        min_instance_count = 2
        max_instance_count = 10
        metrics {
          cpu {
            percent = 70
          }
        }
      }
    }

    domain {
      name = "api.myapp.com"
      type = "PRIMARY"
    }
  }
}
```

**What it solves:** Everything that matters for production:

- **State management**: Terraform tracks what exists in a state file. Running `terraform plan` shows exactly what will change before you apply it.
- **Drift detection**: If someone manually edits the app through the console, `terraform plan` detects the difference and can reconcile it.
- **Declarative configuration**: You describe the desired state, not the steps to get there. Terraform figures out dependencies (create DB before app that references it).
- **Secret management**: Terraform can mark sensitive variables so they don't appear in logs. DigitalOcean encrypts secret env vars server-side, returning opaque `EV[...]` tokens on read (recent provider versions handle this properly without false diffs).
- **Multi-environment consistency**: Use Terraform workspaces or separate state files (dev/staging/prod) with shared modules to ensure consistent configurations across environments.
- **Integration with other resources**: Define `digitalocean_database_cluster`, `digitalocean_domain`, and `digitalocean_container_registry` in the same Terraform config and reference them in the app spec.

**Challenges:**
- **Secret lifecycle quirks**: DigitalOcean encrypts secrets after first deployment, returning encrypted values (`EV[1:...]`) on subsequent reads. Earlier Terraform provider versions would see this as a diff and try to re-apply secrets on every plan. Modern provider versions ignore encrypted value changes unless the plaintext input changes. The takeaway: mark secret variables as `sensitive = true` in Terraform and trust the provider to handle encryption.
- **Auto-deploy interplay**: If you enable `deploy_on_push` for Git sources or `deploy_on_push.enabled = true` for DOCR images, App Platform will auto-deploy when code/images change. Terraform won't see these deployments as config drifts (because the spec hasn't changed), which is usually what you want. If you need Terraform to explicitly deploy specific versions, disable auto-deploy and update the spec's image tag or git commit SHA in Terraform.

**When to use:** Production environments where you need audit trails, peer-reviewed changes (via PR review of Terraform code), and the ability to reproduce infrastructure from scratch. Terraform fits well in GitOps workflows: commit `.tf` files, run CI checks (`terraform plan`), merge, and apply in CD.

**Verdict:** Production-ready. Terraform is the industry standard for IaC, has excellent community support, and DigitalOcean's provider is well-maintained. The downside: HCL can be verbose for complex nested structures (App Spec has many optional fields), and managing Terraform state requires coordination (remote backends like S3 or Terraform Cloud).

---

### Level 3: Infrastructure-as-Code with Pulumi (Production-Grade, Code-First Flavor)

**What it is:** Using Pulumi's DigitalOcean provider to define App Platform resources in a real programming language (TypeScript, Python, Go, etc.):

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const app = new digitalocean.App("production-api", {
    spec: {
        name: "production-api",
        region: "nyc",
        services: [{
            name: "api",
            instanceSizeSlug: "professional-xs",
            instanceCount: 2,
            github: {
                repo: "myorg/myapp",
                branch: "main",
                deployOnPush: true,
            },
            envs: [{
                key: "DATABASE_URL",
                value: dbCluster.uri,
                type: "SECRET",
            }],
            autoscaling: {
                minInstanceCount: 2,
                maxInstanceCount: 10,
                metrics: {
                    cpu: { percent: 70 },
                },
            },
        }],
        domains: [{
            name: "api.myapp.com",
            type: "PRIMARY",
        }],
    },
});

export const appUrl = app.defaultIngress;
```

**What it solves:** Everything Terraform solves, with additional developer ergonomics:

- **Programming language features**: Use loops, conditionals, functions, and classes to generate configurations. For example, deploy 10 microservices by iterating over an array instead of copy-pasting HCL blocks.
- **First-class secret management**: Pulumi automatically encrypts secrets in its state using a secrets provider (KMS, HashiCorp Vault, or Pulumi Service encryption). Mark values as secrets with `pulumi.secret()` and they never appear in logs or state files in plaintext.
- **Type safety**: TypeScript/Python IDEs provide autocomplete and type checking for App Platform fields, catching configuration errors before deployment.
- **Pulumi Automation API**: Programmatically trigger deployments from your own applications (useful for SaaS platforms provisioning customer environments).

**Challenges:**
- **Smaller community**: While Pulumi's DigitalOcean provider is robust (built on the same Terraform provider bridge), you'll find fewer community modules and examples compared to Terraform.
- **State management**: Pulumi defaults to its cloud-managed state (convenient for teams), but some organizations prefer self-hosted state. Pulumi supports local, S3, or Azure Blob backends, but requires explicit configuration.

**When to use:** Teams that prefer code-first infrastructure, need dynamic configuration generation, or want integrated secret management. Pulumi shines when infrastructure provisioning is embedded in larger systems (for example, a control plane that spins up customer environments on demand).

**Verdict:** Production-ready and arguably more developer-friendly than Terraform, especially for teams already comfortable with TypeScript, Python, or Go. The trade-off is a smaller ecosystem and dependency on Pulumi's tooling.

---

### Level 4: Multi-Cloud Abstraction with OpenMCF (Cloud-Agnostic Production)

**What it is:** Defining DigitalOcean App Platform deployments using OpenMCF's protobuf-based API, which abstracts cloud-specific details into a unified, Kubernetes-style manifest:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanAppPlatformService
metadata:
  name: production-api
spec:
  service_name: production-api
  region: nyc1
  service_type: web_service
  
  git_source:
    repo_url: https://github.com/myorg/myapp.git
    branch: main
    run_command: "npm start"
  
  instance_size_slug: professional_xs
  instance_count: 2
  enable_autoscale: true
  min_instance_count: 2
  max_instance_count: 10
  
  env:
    ENVIRONMENT: production
    DATABASE_URL: ${digitalocean-postgres-prod.connection-uri}
  
  custom_domain:
    kind: DigitalOceanDnsZone
    field_path: spec.domain_name
    value: api.myapp.com
```

**What it solves:** Cloud portability and configuration simplification:

- **80/20 API surface**: OpenMCF's protobuf spec exposes only the essential fields most apps actually use (service name, region, source, instance size, scaling, env vars, domain). Advanced fields (custom health check intervals, fine-grained ingress rules, alert policies) are omitted in favor of sane defaults. This reduces cognitive load—developers configure 10 fields instead of 50.
- **Multi-cloud consistency**: The same Planton manifest pattern works across AWS App Runner, Google Cloud Run, Azure App Service, and DigitalOcean App Platform. Teams can switch cloud providers without rewriting deployment configs.
- **Foreign key references**: The `custom_domain` field uses Planton's foreign key pattern to reference other resources (DNS zones, databases) by kind and field path, avoiding hardcoded values. If the referenced database's connection URI changes, Planton automatically propagates the update.
- **Validated by design**: Protobuf schema validation catches misconfigurations before deployment (invalid region names, missing required fields, malformed domain names).
- **Pulumi under the hood**: Planton generates Pulumi code from protobuf specs, meaning you get Pulumi's state management and preview capabilities without writing Pulumi code directly.

**What it doesn't solve:** Escape hatches for advanced configurations. If you need App Platform features not exposed in Planton's minimal API (for example, custom CORS rules, specific buildpack stack versions, or multiple workers with different scaling rules), you may need to drop down to raw Pulumi/Terraform or request Planton enhancements.

**When to use:** Multi-cloud deployments where you want a single IaC experience across providers, or when you value configuration simplicity over exhaustive control. Planton is particularly effective for platform teams building internal developer platforms (IDPs) that abstract cloud complexity from application developers.

**Verdict:** Production-ready for apps that fit Planton's 80/20 model. The abstraction trades off flexibility for simplicity and portability. If your deployment needs are standard (web service, env vars, autoscaling, domain), Planton provides the best developer experience. If you have exotic requirements, Terraform/Pulumi give you full control.

---

## Source Configuration Patterns: Git vs Container Images

DigitalOcean App Platform supports two fundamentally different deployment patterns, each with distinct trade-offs:

### Git-Based Deployment (Build on Platform)

**How it works:** Point App Platform at a GitHub/GitLab repository and branch. On every push (if `deploy_on_push: true`), DigitalOcean clones the repo and builds it using Cloud Native Buildpacks (auto-detecting Node.js, Python, Ruby, Go, etc.) or a provided Dockerfile. The built container image is then deployed.

**Pros:**
- **Zero build infrastructure**: No need to maintain CI runners or container registries. DigitalOcean provides the build environment and doesn't charge for build minutes.
- **Heroku-like simplicity**: Developers push code, and deployments happen automatically. No Docker expertise required.
- **Integrated CI/CD**: The git webhook triggers builds immediately on merge to main/production branch.

**Cons:**
- **Slower builds**: Building from source on every deploy can be slower than deploying pre-built images, especially with large dependency trees (though DO caches build layers).
- **Less control over build environment**: Buildpacks are opinionated. If you need custom system packages or non-standard runtime configurations, you'll fight the platform or resort to Dockerfiles.
- **Branch-based deployment**: You're always deploying the latest commit on a branch. For immutable, version-tagged deployments, this is too dynamic.

**Best for:** Rapid prototyping, internal apps, teams without DevOps infrastructure, and scenarios where the default buildpack environment matches your needs.

---

### Container Image Deployment (Build Externally)

**How it works:** Build container images in your own CI pipeline (GitHub Actions, GitLab CI, CircleCI) and push them to DigitalOcean Container Registry (DOCR), Docker Hub, or GitHub Container Registry. Point App Platform at a specific image tag. Optionally enable `deploy_on_push` for DOCR images to auto-deploy when a new image with the same tag is pushed.

**Pros:**
- **Full control over build**: Custom base images, multi-stage builds, specialized dependencies—anything that runs in a Linux container.
- **Immutable deployments**: Deploy specific image tags (e.g., `v1.2.3`) to staging and production. Promote the same tested artifact through environments, ensuring consistency.
- **Faster deploys**: App Platform skips the build phase and only pulls and starts the image.
- **External testing**: Run unit tests, integration tests, and security scans in CI before building the image. Only push images that pass all checks.

**Cons:**
- **Requires CI/CD setup**: You need a pipeline to build and push images, plus credentials for the container registry.
- **Registry management**: Pull rate limits (for Docker Hub), credential rotation, and image garbage collection become your responsibility (though DOCR integration is seamless for DO).

**Best for:** Production environments, complex build requirements, teams with existing CI/CD pipelines, and scenarios where you need deployment gating (only deploy images that passed tests).

---

### Recommendation

**Use Git-based deployment for:**
- Development and staging environments where iteration speed matters
- Small apps with standard tech stacks (Node.js/Python/Ruby web apps)
- Teams without dedicated DevOps resources

**Use container image deployment for:**
- Production environments requiring immutable, auditable deployments
- Apps with custom build requirements (multi-stage Docker builds, non-standard dependencies)
- Multi-environment promotion workflows (dev → staging → prod with the same image)

**OpenMCF supports both patterns** via the `oneof source` field in the protobuf spec: choose either `git_source` or `image_source`.

---

## Production Essentials

### Scaling and Instance Sizing

App Platform offers two tiers:
- **Basic**: Shared CPU, lower cost (`basic-xxs` through `basic-l`). Suitable for low-traffic services, background workers, and dev/staging environments.
- **Professional**: Dedicated CPU, consistent performance (`professional-xs` through `professional-xl`). Required for production web services, autoscaling, and CPU-intensive workloads.

**Autoscaling** is configured with min/max instance counts and a CPU utilization target. App Platform monitors CPU and adjusts instance count to keep utilization near the threshold (e.g., 70%). Set `min_instance_count >= 2` for high availability—multiple instances run behind DigitalOcean's managed load balancer, enabling zero-downtime deployments and resilience to single-instance failures.

**Scale-to-zero** (setting `min_instance_count = 0`) is available for non-critical workers or cron jobs, reducing cost during idle periods. This introduces cold-start latency (10-30 seconds), so it's unsuitable for user-facing APIs.

---

### Networking, SSL, and Domains

Every app receives a free HTTPS endpoint at `<app-name>.ondigitalocean.app`. For production, add custom domains via the `domains` section of the spec. App Platform automatically provisions Let's Encrypt SSL certificates and renews them. You can configure:
- **Primary domain**: The main entry point (e.g., `api.myapp.com`)
- **Aliases**: Additional domains pointing to the same app (e.g., `www.api.myapp.com`)
- **Wildcard**: Cover all subdomains (e.g., `*.myapp.com`)

All traffic is TLS-terminated at the edge, with HTTP automatically redirected to HTTPS. Minimum TLS version defaults to 1.2 but can be configured for compliance.

**CDN integration**: Static sites are automatically served through DigitalOcean's global CDN (powered by Fastly), caching content at edge locations for low-latency worldwide access. Dynamic services (APIs) don't use the CDN by default, but you can front them with Cloudflare or a similar CDN if needed.

---

### Environment Variables and Secrets

Environment variables are the primary configuration mechanism. Define them in the `env` map with optional `type: SECRET` to mark sensitive values. DigitalOcean encrypts secrets server-side, returning opaque encrypted tokens (`EV[...]`) on read.

**Best practices:**
- Never commit plaintext secrets to version control. Use Terraform/Pulumi variable inputs marked as sensitive.
- For database connections, use DigitalOcean's built-in interpolation (`$${db.DATABASE_URL}`) when attaching managed databases—App Platform injects credentials automatically and rotates them if the DB password changes.
- Store third-party API keys (Stripe, SendGrid, etc.) as secrets and rotate them regularly.

---

### Health Checks and Reliability

App Platform supports HTTP health checks (GET requests to a specified path). If health checks fail, the container is restarted. Configure:
- `path`: Health check endpoint (e.g., `/health` returning 200 OK)
- `initial_delay_seconds`: Grace period after container start before checking (default 5s)
- `period_seconds`: Check interval (default 10s)
- `timeout_seconds`: Max time to wait for a response (default 5s)

**Liveness checks** ensure crashed or deadlocked containers are replaced quickly. Ensure your app can handle restarts gracefully (no local state, use external storage for persistence).

---

### Database Integration

App Platform integrates tightly with **DigitalOcean Managed Databases** (PostgreSQL, MySQL, Redis). Two approaches:

1. **Embedded dev database** (`production: false` in spec): App Platform provisions a free, small database for development. Not suitable for production.
2. **Managed database attachment** (`production: true`, provide `cluster_name`): App Platform connects to an existing managed database cluster, injecting connection credentials as environment variables (`DATABASE_URL`, `DB_HOST`, etc.).

Both use **private VPC networking** when the app and database are in the same region, avoiding internet exposure and reducing latency.

**Database migrations**: Use App Platform's `PRE_DEPLOY` jobs to run schema migrations before new code deploys. This ensures the database schema is updated atomically with code changes, preventing incompatibilities.

---

### Cost Optimization

- **Right-size instances**: Start small (`basic-xxs` or `professional-xs`) and scale up based on metrics. Autoscaling reduces cost during low-traffic periods.
- **Leverage CDN**: Static assets served via CDN consume no app server resources and incur no bandwidth charges beyond the included quota.
- **Use dev databases for non-production**: Save $15-50/month per environment by using embedded dev databases in staging.
- **Tear down ephemeral environments**: Automatically destroy preview apps (per-PR environments) when PRs are merged or closed.

---

## Anti-Patterns to Avoid

1. **Relying on ephemeral disk storage**: App Platform containers have ephemeral filesystems that reset on each deployment and aren't shared across instances. Use DigitalOcean Spaces (S3-compatible object storage) for file uploads, or managed databases for structured data. Storing files locally leads to data loss and broken functionality in multi-instance deployments.

2. **Skipping health checks**: Without health checks, failed containers may remain in the load balancer rotation, returning errors to users. Always configure health checks for production services.

3. **Manual secret management**: Hardcoding secrets, committing them to git, or manually updating them in the console leads to security vulnerabilities and operational pain. Use IaC tools with encrypted secret variables and leverage App Platform's secret injection.

4. **Ignoring deployment failures silently**: Configure `DEPLOYMENT_FAILED` alerts to notify via email or webhook when deployments fail. Catching failures early prevents cascading issues.

5. **Over-provisioning instances**: Running 10 large instances "just in case" wastes money. Use autoscaling and metrics to right-size based on actual load.

---

## The OpenMCF Choice: Simplicity Without Compromise

OpenMCF's `DigitalOceanAppPlatformService` API exposes **only the essential fields** that 80% of apps actually configure:

- **Service name and region**: Where and what to deploy
- **Service type**: Web service, worker, or job
- **Source configuration**: Git repo + branch OR container image + tag
- **Instance sizing and scaling**: Size slug, instance count, and optional autoscaling with min/max bounds
- **Environment variables**: Key-value map for configuration and secrets
- **Custom domain**: Optional domain reference via foreign key pattern

Advanced features (custom health check intervals, fine-grained CORS rules, alert policies, static sites, multi-component apps) are deliberately omitted to keep the API surface small and understandable. Sane defaults handle the rest:
- Health checks default to checking the service's HTTP port
- SSL certificates are provisioned automatically
- Load balancing is implicit for multi-instance deployments
- Auto-deploy defaults to enabled for Git sources

This design philosophy aligns with DigitalOcean App Platform's own ethos: **make simple things simple, not everything possible**. If your needs exceed the 80/20 model, drop down to Pulumi or Terraform for full control.

**Why this matters:** In a multi-cloud world, most apps share the same basic requirements: run code, scale on demand, expose via HTTPS, inject configuration. OpenMCF abstracts these commonalities, letting you define deployments once and target any cloud provider without rewriting configs. DigitalOcean App Platform is the deployment target, but the Planton API is the interface—cloud-agnostic, validated, and optimized for clarity.

---

## Conclusion

DigitalOcean App Platform represents a maturation of the PaaS model: the ease of Heroku without the cost, the control of containers without the complexity of Kubernetes. Whether you deploy via Git-based builds for rapid iteration or container images for production rigor, the platform handles the undifferentiated heavy lifting—load balancing, SSL, health checks, autoscaling, and private networking.

The deployment method you choose defines your operational maturity. Manual console clicks are for learning. CLI scripts are for automation. Terraform and Pulumi are for production. And OpenMCF is for teams that want cloud portability without sacrificing production-readiness.

If you're building web apps, APIs, background workers, or microservices on DigitalOcean, App Platform eliminates the infrastructure toil so you can focus on code. And if you're adopting OpenMCF, the `DigitalOceanAppPlatformService` API gives you a clean, protobuf-validated interface to that power—no App Spec YAML wrangling required.

Deploy with confidence. Scale on demand. Sleep soundly knowing your infrastructure is code, your configs are validated, and your deployments are repeatable.

