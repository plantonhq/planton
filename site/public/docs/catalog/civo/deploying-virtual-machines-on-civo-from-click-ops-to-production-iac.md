---
title: "Deploying Virtual Machines on Civo: From Click-Ops to Production IaC"
description: "Deploying Virtual Machines on Civo: From Click-Ops to Production IaC deployment documentation"
icon: "package"
order: 100
componentName: "civocomputeinstance"
---

# Deploying Virtual Machines on Civo: From Click-Ops to Production IaC

## The Cloud for Developers Who Value Simplicity

For years, the conventional wisdom in cloud computing was clear: if you wanted production-grade infrastructure, you needed the complexity of AWS, the learning curve of GCP, or the enterprise pricing of Azure. Virtual machines were either expensive and feature-rich (hyperscalers) or cheap but lacking (budget VPS providers). The middle ground—affordable, fast, and developer-friendly—seemed like a myth.

Civo Cloud challenges that assumption. Born from a simple observation that most developers don't need 200+ services and don't want surprise egress bills, Civo offers compute instances that boot in under 60 seconds, cost up to 75% less than hyperscalers, and include unlimited bandwidth at no extra charge. It's optimized for the 80% use case: dev environments, staging clusters, production workloads that value simplicity over every possible managed service.

This document explores how to deploy Civo compute instances—from quick manual provisioning to production-grade Infrastructure as Code. We'll examine the deployment method spectrum, compare IaC tooling, and explain why OpenMCF supports declarative provisioning on Civo as a first-class citizen in our multi-cloud framework.

## The Deployment Maturity Spectrum

Not all deployment approaches are created equal. Here's how Civo instance deployment methods progress from quick experiments to production-ready automation.

### Level 0: Manual Dashboard Provisioning

**What it is:** Civo's web console offers point-and-click VM creation. Select a region (London, New York, Frankfurt, Phoenix, Mumbai), choose an instance size (`g3.small` for 1 vCPU/2GB up to `g3.2xlarge` for 8 vCPU/32GB), pick an OS image (Ubuntu, Debian, Rocky Linux), configure networking, attach SSH keys, and click create. Instances boot in under 60 seconds.

**Why it exists:** Perfect for exploration and one-off testing. The UI is intuitive—far simpler than AWS Console's labyrinth of options.

**The problem:** Manual provisioning doesn't scale. Common pitfalls include:

- **Security gaps:** Civo's default firewall allows all traffic on all ports. It's easy to spin up an instance and forget to lock it down.
- **Configuration drift:** Clicking through options means no versioned record of what you built. Rebuilding identical instances becomes guesswork.
- **No repeatability:** Staging and production environments diverge because they were created at different times by different people clicking different options.

**Verdict:** Use the dashboard for quick experiments or initial exploration. For anything that needs to be repeatable or secure, move up the maturity ladder.

### Level 1: CLI and Scripting

**What it is:** The Civo CLI (`civo`) is a production-ready, open-source Go tool that wraps Civo's REST API. After installing (one-line script, Homebrew, or Chocolatey) and saving your API key (`civo apikey save`), you can create instances via command:

```bash
civo instance create --hostname=web1 --size=g3.small --region=NYC1 \
     --diskimage=ubuntu-jammy --network=default --ssh-key=team-key
```

The CLI supports all major operations (instances, firewalls, volumes, reserved IPs, Kubernetes) and outputs JSON for scripting.

**Why it's better:** Commands are repeatable. Wrap them in shell scripts and you've documented your infrastructure. Store scripts in Git and you have version control. Use them in CI pipelines and you have automation.

**The limitations:**

- **Imperative, not declarative:** Scripts describe steps to create resources, not desired end state. If a step fails midway, you're left with partial infrastructure.
- **No dependency management:** If instance A needs a firewall and instance B needs a reserved IP, your script must handle ordering manually.
- **State management is manual:** Running the same script twice creates duplicate resources unless you add complex checks.

**Verdict:** CLI scripting is a significant upgrade from manual work. It's ideal for ephemeral environments (spin up a test VM, use it, destroy it). But for production environments with complex dependencies, declarative IaC is a better fit.

### Level 2: Direct API Integration

**What it is:** Civo's REST API (`https://api.civo.com/v2`) powers everything. Authenticate with a Bearer token, POST JSON payloads to create instances, GET to query state. Official SDKs exist for Go (`civogo`) and Python (`civo`), plus the API is simple enough for custom clients in any language.

**When it's useful:** Building custom tooling, integrating Civo into existing automation frameworks, or when you need programmatic control beyond what CLI offers.

**The trade-offs:** You're essentially rebuilding what IaC tools already provide—idempotency, dependency resolution, state management. Unless you have unique requirements (like embedding Civo provisioning into a custom SaaS control plane), using battle-tested IaC tools is more efficient.

**Verdict:** Direct API use is powerful for specialized integration but overkill for standard infrastructure provisioning. Terraform and Pulumi offer better abstractions for most teams.

### Level 3: Configuration Management (Ansible)

**What it is:** Use Ansible playbooks to orchestrate instance creation (via the `uri` module calling Civo's API or shell tasks invoking the CLI), then configure the OS and deploy applications—all in one workflow.

**The appeal:** Ansible can provision the infrastructure *and* configure it. Create an instance, wait for SSH, install packages, deploy your app, configure monitoring—all declarative YAML.

**The reality:** Civo doesn't have an official Ansible collection yet. You're writing custom tasks that hit the API. Ensuring idempotency (don't create duplicate instances) requires manual state checking. Terraform or Pulumi handle infrastructure creation better; Ansible shines at post-boot configuration.

**Verdict:** Ansible is excellent for configuring instances after they exist (software installation, user management, security hardening). For creating the instances themselves, pair Ansible with a dedicated IaC tool rather than reimplementing resource management.

### Level 4: Production IaC (Terraform & Pulumi)

**What it is:** Declarative infrastructure as code using purpose-built tools. Define desired state (instances, networks, firewalls, volumes) in configuration files. The tool handles creating, updating, and destroying resources to match your declaration.

**Terraform:**

- **Provider:** Official `civo/civo` provider on Terraform Registry, actively maintained by Civo.
- **Coverage:** Compute instances, Kubernetes clusters, networks, firewalls, volumes, load balancers, reserved IPs, DNS.
- **Maturity:** Production-ready. Used by companies like Defense.com to manage Civo workloads at scale.
- **Ecosystem:** Works with OpenTofu (Terraform's open-source fork) identically.

Example Terraform resource:

```hcl
resource "civo_instance" "web" {
  hostname = "web-server-1"
  size     = data.civo_size.small.id
  region   = "NYC1"
  disk_image = data.civo_disk_image.ubuntu.id
  network_id = data.civo_network.default.id
  firewall_id = civo_firewall.web.id
  sshkey_id = data.civo_ssh_key.team.id
  
  script = <<-EOT
    #!/bin/bash
    apt-get update
    apt-get install -y docker.io
  EOT
}
```

**Pulumi:**

- **Provider:** Bridges the Terraform provider, allowing you to define Civo infrastructure in TypeScript, Python, Go, C#, or Java.
- **Maturity:** Production-ready (v1.1.7 as of late 2025, maintained by Civo).
- **Benefits:** Use real programming languages (loops, conditionals, functions) instead of HCL. Pulumi's secret management encrypts sensitive values automatically.

Example Pulumi code (Python):

```python
import pulumi_civo as civo

instance = civo.Instance("web",
    hostname="web-server-1",
    size="g3.small",
    region="NYC1",
    disk_image="ubuntu-jammy",
    network_id=network.id,
    firewall_id=firewall.id,
    sshkey_id=ssh_key.id,
    script="""#!/bin/bash
        apt-get update
        apt-get install -y docker.io
    """)
```

**Why this is production-ready:**

1. **State management:** Terraform and Pulumi track what resources exist. Running `apply` twice doesn't duplicate resources—it reconciles state.
2. **Dependency resolution:** If a firewall references a network, the tool creates the network first automatically.
3. **Multi-environment support:** Use Terraform workspaces or Pulumi stacks to manage dev/staging/prod with the same code.
4. **Change preview:** See what will change before applying (`terraform plan`, `pulumi preview`).
5. **Secret management:** Store API tokens in environment variables or encrypted backends, not in code.

**Terraform vs Pulumi:**

- **Community:** Terraform has broader adoption on Civo (more examples, tutorials). Pulumi is newer but growing.
- **Language:** Terraform uses HCL (declarative DSL). Pulumi uses familiar programming languages (TypeScript, Python, Go).
- **Workflow:** Both support GitOps, CI/CD integration, and team collaboration via state backends.

**Verdict:** For production Civo infrastructure, use Terraform or Pulumi. Choose Terraform if your team prefers declarative DSLs and wants the largest community. Choose Pulumi if you want full programming language power and already use Pulumi elsewhere.

### Level 5: Kubernetes-Native IaC (Crossplane)

**What it is:** Crossplane turns Kubernetes into a control plane for infrastructure. Install the Crossplane Civo provider, then create Civo instances by applying Kubernetes manifests:

```yaml
apiVersion: compute.civo.crossplane.io/v1alpha1
kind: Instance
metadata:
  name: web-server-1
spec:
  forProvider:
    hostname: web-server-1
    size: g3.small
    region: NYC1
    diskImage: ubuntu-jammy
    networkRef:
      name: production-network
    firewallRef:
      name: web-firewall
```

**Why it's powerful:**

- **GitOps-native:** Manage infrastructure and applications with the same workflow (ArgoCD, FluxCD).
- **Kubernetes CRDs:** Your Civo instances become Kubernetes resources. Use `kubectl` to manage cloud VMs.
- **Composition:** Create high-level abstractions (a "WebService" resource that provisions instance + firewall + DNS in one).

**The trade-offs:**

- **Maturity:** The Crossplane Civo provider is early-stage (v0.1.x). Not as battle-tested as Terraform.
- **Complexity:** Requires a Kubernetes cluster to run Crossplane. Adds operational overhead.
- **Coverage:** Supports instances and Kubernetes clusters, but fewer resources than Terraform provider.

**Verdict:** Crossplane is compelling if you're already running Kubernetes and want unified GitOps workflows for infra + apps. For teams not deeply invested in Kubernetes operators, Terraform/Pulumi offer simpler paths to production.

## Comparing Terraform and Pulumi on Civo

Both tools are production-ready for Civo. Here's how to choose:

| Criterion | Terraform | Pulumi |
|-----------|-----------|--------|
| **Community Adoption** | Larger (more examples, tutorials, Stack Overflow answers) | Smaller but growing |
| **Language** | HCL (declarative DSL) | TypeScript, Python, Go, C#, Java |
| **Resource Coverage** | Comprehensive (instances, K8s, networking, storage, DNS) | Identical (uses Terraform provider under the hood) |
| **Secret Management** | Environment variables, Vault, Terraform Cloud | Built-in encrypted config (`pulumi config set --secret`) |
| **State Management** | Explicit state files (local, S3, Terraform Cloud) | Built-in encrypted state (Pulumi Service or self-hosted) |
| **Multi-Environment** | Workspaces or separate directories | Stacks (built-in concept) |
| **Learning Curve** | Learn HCL syntax | Use language you already know |
| **Best For** | Teams wanting declarative simplicity, largest community | Teams preferring real programming constructs (loops, functions) |

**Multi-cloud note:** Both tools manage multiple clouds in one configuration. Terraform uses provider blocks; Pulumi imports multiple SDKs. For OpenMCF's multi-cloud use case, either works—choose based on team preference.

## Production-Ready Instance Configuration

Deploying instances for production requires more than picking a size and OS. Here's what production-grade Civo deployments include:

### Networking

- **Private Networks:** Create dedicated networks for isolation (e.g., separate networks for web tier and database tier). Instances in the same network communicate privately.
- **Public IPs:** Only assign public IPs to instances that need internet exposure (web servers, bastion hosts). Backend services (databases, workers) should be private-only.
- **Reserved IPs:** Use Civo's reserved IPs for production endpoints. If an instance fails, reassign the IP to a replacement—no DNS changes needed.

### Security

- **SSH Keys, Not Passwords:** Always use SSH key authentication. Upload keys to Civo (`civo sshkey upload`) and reference them during instance creation.
- **Firewalls:** Never use Civo's default firewall (allows all ports). Create custom firewalls per role:
  - Web servers: Allow 80/443 from `0.0.0.0/0`, SSH from office IPs only
  - Databases: Allow 5432 (Postgres) or 3306 (MySQL) from web server IPs only, deny everything else
- **Principle of Least Privilege:** Lock down both ingress and egress. If an instance only needs to download apt packages, restrict it to that.

### Storage

- **Ephemeral Root Disks:** Instance root disks are destroyed when instances are deleted. For production data, use Civo volumes (persistent block storage).
- **Volumes:** Create volumes separately, attach to instances. If an instance dies, detach the volume and attach to a replacement.
- **Backups:** Use Civo's snapshot feature for instance backups. Schedule automated snapshots (nightly, weekly). Store critical data backups off-instance (object storage, external backup service).

### Monitoring and Logging

- **Civo Statistics:** Enable Civo's lightweight monitoring agent (`civostatsd`) for CPU/RAM/disk metrics visible in the dashboard.
- **External Monitoring:** For production, integrate third-party tools (Prometheus, Datadog, New Relic). Install agents via cloud-init.
- **Logging:** Civo doesn't provide centralized logging. Ship logs to external services (Elasticsearch, CloudWatch Logs, Loki) using Filebeat or Fluent Bit.

### Bootstrapping

- **Cloud-Init Scripts:** Use the `user_data` field to provision instances on first boot:
  - Install security updates
  - Configure firewall rules at OS level
  - Install Docker or application dependencies
  - Enable monitoring agents
  - Join configuration management (Ansible, Chef, Puppet)

Example cloud-init script:

```bash
#!/bin/bash
set -euo pipefail

# Enable Civo monitoring
curl -s https://www.civo.com/civostatsd.sh | sudo bash

# Update system and install Docker
apt-get update
apt-get upgrade -y
apt-get install -y docker.io

# Configure Docker to start on boot
systemctl enable docker
systemctl start docker

# Deploy application
docker pull myregistry/app:latest
docker run -d --restart=always -p 80:80 myregistry/app:latest
```

### High Availability

Civo regions are single data centers. Build HA at the application level:

- **Multiple Instances:** Run at least two instances per service.
- **Load Balancers:** Use Civo's managed load balancers to distribute traffic. They provide health checks and failover.
- **Reserved IP Failover:** For database pairs (master-slave), use a reserved IP that can be reassigned to the active instance.
- **Placement Rules:** Use Civo's instance placement groups to ensure redundant instances run on different physical hosts.
- **Multi-Region Deployments:** For critical services, deploy in multiple Civo regions (e.g., London + New York) and use DNS failover.

## The 80/20 Configuration Principle

Most Civo instance deployments use a core set of parameters:

**Essential fields (95% of deployments):**

- `hostname` – Instance name
- `size` – Instance type (`g3.small`, `g3.medium`, `g3.large`)
- `region` – Data center location
- `image` – OS template (`ubuntu-jammy`, `debian-11`, `rocky-9`)
- `network_id` – Private network for the instance
- `ssh_key_ids` – SSH public keys for access

**Common optional fields (60% of deployments):**

- `firewall_ids` – Security groups to apply
- `reserved_ip_id` – Static IP for production endpoints
- `volume_ids` – Persistent storage volumes
- `user_data` – Cloud-init script for bootstrapping
- `tags` – Organizational labels

**Rarely used fields (<20%):**

- Custom disk images
- GPU-specific configuration
- Advanced placement rules

OpenMCF's Civo API focuses on the 80/20 configuration, covering the vast majority of real-world use cases without cluttering the interface with rarely-used options.

## Civo's Unique Value Propositions

### Free, Unlimited Bandwidth

Unlike AWS (which charges ~$0.09/GB egress), GCP, or even DigitalOcean (which caps free bandwidth), Civo includes unlimited ingress and egress bandwidth on all instances. For applications serving user content, streaming video, or handling large data transfers, this eliminates a major cost variable.

**Cost example:** Serving 1TB of data per month costs $90 in AWS bandwidth fees, $0 on Civo.

### Fast Provisioning

Instances boot in under 60 seconds. Kubernetes clusters (K3s) launch in ~2 minutes. For ephemeral environments (CI/CD runners, preview environments, temporary test clusters), this speed advantage compounds over time.

### Simplified Pricing

Civo uses hourly billing with no hidden fees. A `g3.small` (1 vCPU, 2GB) costs ~$10.86/month. No charges for data transfer, no IOPS pricing, no NAT gateway fees. What you see is what you pay.

### Developer-Focused Ecosystem

- **Managed Kubernetes:** K3s clusters with Traefik ingress, storage classes, and load balancer integration out of the box.
- **Managed Databases:** PostgreSQL, MySQL, Redis with automated backups (cheaper than AWS RDS).
- **Object Storage:** S3-compatible storage at $10/TB (vs $20/TB on DigitalOcean).

## OpenMCF's Approach

OpenMCF treats Civo as a first-class cloud provider in our multi-cloud IaC framework. We support declarative instance provisioning via Pulumi, abstracting Civo's resources behind a consistent API.

**Why Pulumi over Terraform for OpenMCF:**

- **Multi-language support:** Users can define infrastructure in TypeScript, Python, Go—languages they already know.
- **Programmatic composition:** Create reusable components that combine instances + networking + security.
- **Unified workflow:** Manage Civo instances alongside AWS, GCP, Azure, Kubernetes resources in one Pulumi program.

**Our design principles:**

1. **80/20 Configuration:** Our Civo Compute Instance API exposes essential fields (hostname, size, region, image, network, SSH keys) as required, common options (firewalls, reserved IPs, volumes, user-data) as optional.

2. **Secure Defaults:** We encourage users to specify firewalls, use reserved IPs for production, and provide cloud-init scripts rather than relying on Civo's permissive defaults.

3. **Multi-Cloud Consistency:** The same conceptual model (compute instance with networking, security, storage) applies across clouds. A team familiar with our AWS instance API will recognize the patterns in our Civo API.

4. **Abstraction, Not Lock-In:** We abstract Civo's specifics (like size names `g3.small`) while allowing direct access for advanced use cases. Users aren't locked into proprietary APIs—our generated Pulumi code is readable and modifiable.

## When to Use Civo

**Ideal for:**

- **Startups and SMBs:** Affordable infrastructure without hyperscaler complexity.
- **Dev/Test Environments:** Fast provisioning, free bandwidth, hourly billing make ephemeral environments cost-effective.
- **Bandwidth-Heavy Workloads:** Serving user content, APIs with high data transfer, media streaming.
- **Kubernetes-First Teams:** Civo's managed K3s is one of the fastest, cheapest Kubernetes offerings.
- **Multi-Cloud Strategies:** Civo as a cost-optimized cloud for non-critical workloads, hyperscalers for specialized services.

**Consider alternatives for:**

- **Global Scale:** Civo has 5 regions (vs AWS's 60+). If you need presence in South America, Africa, or Asia-Pacific beyond Mumbai, you'll need additional providers.
- **Specialized Managed Services:** No equivalents to AWS Lambda, BigQuery, Azure Active Directory. Use Civo for compute, hyperscalers for niche services.
- **Strict Compliance Requirements:** Smaller clouds may not have certifications (HIPAA, FedRAMP) that regulated industries require.

## Conclusion

Deploying virtual machines on Civo represents a shift in cloud economics and developer experience. The platform proves that production infrastructure doesn't require navigating AWS's maze of services or accepting surprise bandwidth bills. By offering fast, affordable compute with unlimited bandwidth and simple pricing, Civo lowers the barrier to cloud-native development.

For teams adopting Infrastructure as Code, Civo's mature Terraform and Pulumi support means you can treat it like any other cloud—declaring desired state, tracking changes in Git, automating deployments via CI/CD. The production patterns are familiar: private networks, security groups (firewalls), persistent volumes, load balancers, and automated backups.

OpenMCF integrates Civo as a strategic option in our multi-cloud framework because it excels at the 80% use case: teams that want to run workloads efficiently without overpaying for features they'll never use. Whether you're spinning up dev environments, running production APIs, or hosting Kubernetes clusters, Civo provides a compelling combination of speed, simplicity, and cost-effectiveness that traditional clouds struggle to match.

The future of cloud isn't one provider dominating all use cases. It's choosing the right cloud for each workload—and for many workloads, Civo's developer-first approach is exactly right.

