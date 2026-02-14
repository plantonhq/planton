---
title: "Deploying DigitalOcean Droplets: Choosing Control Over Convenience"
description: "Deploying DigitalOcean Droplets: Choosing Control Over Convenience deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceandroplet"
---

# Deploying DigitalOcean Droplets: Choosing Control Over Convenience

## Introduction

There's a persistent narrative in cloud computing that infrastructure-as-a-service is the past and platforms-as-a-service are the future. The story goes: why manage servers when you can just deploy code? Why worry about operating systems when a platform can abstract them away? It's a compelling pitch, and for many workloads, it's exactly right.

But here's the truth that marketing slides often skip: **sometimes you need control more than you need convenience.** Sometimes you need to run legacy software that doesn't fit the PaaS mold. Sometimes you need to install custom system packages, configure kernel parameters, or tune network settings that no abstraction layer exposes. Sometimes you need a virtual machine that you fully own, from the bootloader up.

That's what **DigitalOcean Droplets** provide: Linux-based virtual machines with complete root access, predictable pricing, and the flexibility to run anything that runs on Ubuntu, Debian, Fedora, or any other supported distribution. They're the foundational infrastructure-as-a-service offering in DigitalOcean's portfolio, positioned between the fully-managed App Platform (where you have zero infrastructure responsibility) and DigitalOcean Kubernetes (where you manage container orchestration but not the control plane).

Droplets aren't for everyone. If you're building a stateless web API and don't want to think about servers, App Platform is simpler. If you've embraced containers and Kubernetes, DOKS is more aligned. But if you need granular control, want to self-host databases, require custom server configurations, or are migrating from traditional hosting, Droplets are the answer.

The challenge is managing them correctly. A Droplet is not fundamentally different from any other cloud VM—an EC2 instance, a Compute Engine VM, a Linode—but the way you provision, configure, and maintain it determines whether you get a resilient production system or a brittle house of cards. This guide walks through the spectrum of deployment methods, explains what works at scale, and shows how OpenMCF abstracts the complexity into a clean, protobuf-defined API.

---

## The DigitalOcean Compute Landscape: Choosing Your Level of Abstraction

DigitalOcean offers three distinct compute models, each representing a different trade-off between control and operational overhead:

### App Platform (PaaS): Zero Ops, Maximum Convenience

You provide application code or a container image. DigitalOcean handles provisioning, scaling, load balancing, SSL certificates, and infrastructure maintenance. You never see a server. You never SSH into anything. It's the "Heroku-like" experience: push code, get running app.

**When to use it:** Stateless web apps, APIs, microservices, MVPs, or any workload where you value speed-to-market over infrastructure control.

**Limitation:** You get what the platform gives you. No custom system packages, no kernel tuning, no direct access to the underlying OS.

---

### Droplets (IaaS): Full Control, Full Responsibility

You get a virtual machine with root access and complete control over everything from the operating system upward. DigitalOcean provides the hypervisor and network infrastructure; you manage the OS, security patches, application stack, and all operational concerns.

**When to use it:** Custom server configurations, self-hosted databases, legacy applications, development/staging environments, or workloads that need OS-level control.

**Limitation:** You're responsible for security hardening, patching, backups, monitoring, and availability. It's manual unless you automate it with Infrastructure-as-Code.

---

### DigitalOcean Kubernetes (CaaS): Container Orchestration with Managed Control Plane

DigitalOcean manages the Kubernetes control plane (API server, etcd, scheduler). You manage worker nodes (which are Droplets under the hood) and containerized workloads. It's high-control, high-complexity: perfect for teams that have adopted Kubernetes but don't want to operate the control plane.

**When to use it:** Microservices architectures, container-native applications, or teams that have standardized on Kubernetes.

**Limitation:** Kubernetes is powerful but complex. Running a production cluster can require dedicated expertise—overkill if you just need to run a web server and a database.

---

### The Decision Matrix

| **Model**               | **Abstraction Level** | **User Responsibility**           | **Ideal Use Case**                         |
|-------------------------|----------------------|-----------------------------------|--------------------------------------------|
| **App Platform (PaaS)** | High                 | Application code only             | Web apps, APIs, rapid prototyping          |
| **Droplets (IaaS)**     | Low                  | OS, networking, security, apps    | Custom servers, databases, dev/test        |
| **Kubernetes (CaaS)**   | Medium               | Worker nodes, containers, workloads | Microservices, container orchestration     |

**Bottom line:** If you're using Droplets, you've chosen control over convenience. The rest of this guide assumes you've made that choice deliberately and now need to manage Droplets at production scale.

---

## The Deployment Spectrum: From ClickOps to Production Code

Not all methods of provisioning Droplets are created equal. Here's how they stack up, from what to avoid to what works in production:

### Level 0: The DigitalOcean Control Panel (Anti-Pattern for Production)

**What it is:** Using the web dashboard to click through Droplet creation, selecting region, size, image, SSH keys, and optional features like backups and monitoring.

**What it teaches you:** The basic workflow. The control panel is a good way to explore DigitalOcean's options and understand what a Droplet needs (region, size, image, VPC, SSH keys).

**What it doesn't solve:** Repeatability. A manually-created Droplet is a snowflake. You can't guarantee that your staging environment matches production. You can't codify what you did. You can't version-control it, review it in a pull request, or hand it off to another engineer.

**The security trap:** Manual creation is where most security mistakes happen:
- Forgetting to add SSH keys, defaulting to password authentication (insecure)
- Forgetting to create or apply a Cloud Firewall, leaving SSH exposed to the internet
- Forgetting to enable backups or monitoring during creation (hard to add later)
- Forgetting to assign the Droplet to a VPC, leaving it on the public internet

DigitalOcean's own "Recommended Droplet Setup" guide exists specifically because these manual oversights are so common.

**Verdict:** Use the console to learn, never for production or staging. Even dev environments benefit from code.

---

### Level 1: CLI Scripting with doctl (Better, But Still Imperative)

**What it is:** Using DigitalOcean's official CLI, `doctl`, to create Droplets in shell scripts:

```bash
doctl compute droplet create myserver \
  --region nyc3 \
  --size s-2vcpu-4gb \
  --image ubuntu-22-04-x64 \
  --ssh-keys 12345678 \
  --vpc-uuid "uuid-of-vpc" \
  --enable-monitoring \
  --enable-backups \
  --tag-names "production,web"
```

**What it solves:** Automation. You can script provisioning, commit scripts to version control, and integrate them into CI/CD pipelines. The CLI is synchronous, returns structured output (JSON mode available), and supports all Droplet operations.

**What it doesn't solve:** State management. Scripts don't track what exists. If you run the script twice, you create two Droplets. If creation fails halfway through (Droplet created, Firewall rule failed), cleanup is manual. There's no declarative model—just imperative commands executed in sequence.

**Verdict:** Acceptable for throwaway dev environments or one-time migrations. Not suitable for production, where you need idempotency, state tracking, and rollback capabilities.

---

### Level 2: Configuration Management with Ansible (Hybrid Approach)

**What it is:** Using Ansible's `community.digitalocean.digital_ocean_droplet` module to provision Droplets in YAML playbooks.

**What it solves:** Idempotency (mostly). Ansible's module can create, update, and destroy Droplets. It aims for declarative-ish behavior: "ensure this Droplet exists with these properties."

**What it doesn't solve well:** Lifecycle state. Ansible is fundamentally a *configuration management* tool designed to configure software *on* existing servers. It can provision infrastructure, but it doesn't maintain a state file mapping declared resources to real-world IDs like Terraform does.

**The production pattern:** Most teams use Ansible for what it's best at: configuration. The mature workflow is **Terraform (provision) → Ansible (configure)**:
1. Terraform provisions the Droplet, VPC, and Firewall (stateful, declarative)
2. Terraform outputs the Droplet's IP address
3. Ansible reads that IP, SSH into the Droplet, and installs packages, configures services, and hardens security

This provides a clean separation of concerns: infrastructure state (Terraform) vs. software configuration (Ansible).

**Verdict:** Powerful when paired with a stateful provisioning tool. Not the first choice for pure infrastructure provisioning on its own.

---

### Level 3: Infrastructure-as-Code with Terraform or Pulumi (Production Standard)

**What it is:** Using Terraform (with the `digitalocean` provider) or Pulumi (with the `pulumi-digitalocean` SDK) to declaratively define Droplets and their lifecycle.

**Terraform example:**

```hcl
provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_vpc" "prod" {
  name   = "prod-vpc"
  region = "nyc3"
}

resource "digitalocean_firewall" "web" {
  name = "prod-web-firewall"
  tags = ["web"]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["203.0.113.0/24"] # Office IP only
  }
}

resource "digitalocean_droplet" "web" {
  name       = "prod-web-01"
  region     = "nyc3"
  size       = "g-2vcpu-8gb"
  image      = "ubuntu-22-04-x64"
  vpc_uuid   = digitalocean_vpc.prod.id
  ssh_keys   = [var.ssh_key_id]
  monitoring = true
  backups    = true
  tags       = ["production", "web"]

  user_data = file("cloud-init.yml")

  lifecycle {
    create_before_destroy = true
  }
}
```

**Pulumi example (Python):**

```python
import pulumi_digitalocean as do

vpc = do.Vpc("prod-vpc",
    name="prod-vpc",
    region="nyc3")

web_droplet = do.Droplet("web",
    name="prod-web-01",
    region="nyc3",
    size="g-2vcpu-8gb",
    image="ubuntu-22-04-x64",
    vpc_uuid=vpc.id,
    ssh_keys=[ssh_key_id],
    monitoring=True,
    backups=True,
    tags=["production", "web"])
```

**What it solves:** Everything that matters for production:

- **Declarative configuration:** Describe the desired end state, not the steps to get there
- **State management:** Track what exists, what changed, and what needs to be created/updated/deleted
- **Idempotency:** Running the same config twice produces the same result
- **Plan/preview:** See what will change before applying
- **Version control:** Treat infrastructure as code—diffs, reviews, rollbacks
- **Lifecycle control:** Terraform's `create_before_destroy` enables zero-downtime replacements when you change immutable properties like the image

**The immutable infrastructure pattern:** When you change a Droplet's `user_data` or `image`, Terraform creates a *new* Droplet, waits for it to become healthy, then destroys the old one. This is safer than in-place updates, which can fail halfway through and leave you with a broken server.

**Terraform vs. Pulumi:** Both are production-ready. Terraform has broader adoption, more community resources, and a battle-tested HCL syntax. Pulumi lets you use real programming languages (Python, TypeScript, Go), which is better for complex logic, testing, and teams that prefer code over DSLs. For standard Droplet provisioning, they're equally capable—the choice is team preference.

**Verdict:** This is the production standard. Use Terraform if you want the most mature ecosystem. Use Pulumi if you prefer coding infrastructure in a general-purpose language. Both beat manual provisioning by miles.

---

### Level 4: Golden Images with Packer (The Image Factory)

**What it is:** Using Packer to pre-build custom Droplet images (snapshots) with all necessary software, configuration, and security hardening baked in.

**How it works:**
1. Packer launches a temporary Droplet from a base image (e.g., `ubuntu-22-04-x64`)
2. Runs provisioners (Ansible playbooks, shell scripts) to install packages, apply patches, configure services
3. Creates a snapshot of the configured disk
4. Destroys the temporary Droplet
5. Returns a snapshot ID that Terraform/Pulumi can use as the `image` for new Droplets

**Example Packer config:**

```json
{
  "builders": [{
    "type": "digitalocean",
    "api_token": "{{user `do_token`}}",
    "image": "ubuntu-22-04-x64",
    "region": "nyc3",
    "size": "s-1vcpu-1gb",
    "snapshot_name": "web-server-{{timestamp}}"
  }],
  "provisioners": [{
    "type": "ansible",
    "playbook_file": "playbooks/web-server.yml"
  }]
}
```

**What it solves:** Fast, reliable boot times and reduced `user_data` complexity. Instead of running a long cloud-init script on every Droplet boot (which can fail if package repos are down or scripts have bugs), you launch from a pre-tested "golden image" that already has everything installed and configured.

**Production pattern:**
- Use Packer to build golden images in CI/CD
- Store snapshot IDs in Terraform variables
- Terraform provisions Droplets from the snapshot ID instead of a distribution slug
- Updates? Build a new golden image, update the Terraform variable, `terraform apply` triggers a zero-downtime replacement

**Verdict:** Essential for production systems that need fast, predictable deployments and minimal boot-time configuration. Packer + Terraform is the gold standard for immutable infrastructure.

---

### Level 5: Kubernetes Control Plane with Crossplane (The Everything-Is-K8s Approach)

**What it is:** Using Crossplane (an open-source Kubernetes add-on) to manage DigitalOcean Droplets as Kubernetes Custom Resources.

**Example Droplet CRD:**

```yaml
apiVersion: compute.do.crossplane.io/v1alpha1
kind: Droplet
metadata:
  name: prod-web
spec:
  forProvider:
    name: prod-web-01
    region: nyc3
    size: g-2vcpu-8gb
    image: ubuntu-22-04-x64
    vpcUUID: vpc-12345678
    monitoring: true
    backups: true
    tags:
      - production
      - web
```

**What it solves:** Unified control plane. If your organization has standardized on Kubernetes for *everything*—not just containerized apps, but infrastructure, databases, and third-party services—Crossplane lets you manage Droplets using `kubectl` and the Kubernetes API.

**What it doesn't solve:** Complexity. This is a niche pattern for teams that are deeply invested in Kubernetes as their universal API. For most teams, Terraform or Pulumi is simpler and more portable.

**Verdict:** Use it if you're already running Crossplane and want to manage all infrastructure through Kubernetes. Skip it if you're not already in that ecosystem—Terraform is easier.

---

## Production Essentials: Networking, Security, and Data Resiliency

### VPC Networking: Isolation Is Non-Negotiable

**The anti-pattern:** Creating Droplets on DigitalOcean's default public network, where they communicate over the internet.

**The production standard:** Every Droplet belongs to a **Virtual Private Cloud (VPC)**. VPCs provide a private network layer where Droplets (and Managed Databases, Load Balancers, etc.) communicate over internal IPs without exposing traffic to the public internet.

**Key behavior:**
- VPCs are region-scoped (one VPC per region)
- Droplets in the same VPC can communicate privately
- Droplets in different VPCs (or different regions) communicate via public IPs unless you set up VPC Peering (advanced)

**Terraform pattern:**

```hcl
resource "digitalocean_vpc" "prod" {
  name   = "prod-vpc"
  region = "nyc3"
}

resource "digitalocean_droplet" "app" {
  # ...
  vpc_uuid = digitalocean_vpc.prod.id
}
```

**Critical:** The `vpc_uuid` field is the modern standard. Older tutorials reference a `private_networking` boolean—that's deprecated. Always use `vpc_uuid`.

---

### Firewalls: Defense in Depth

**Two layers:**

1. **DigitalOcean Cloud Firewalls:** Network-based, stateful firewalls managed separately from Droplets. They block traffic *before* it reaches the Droplet.

2. **Host-Based Firewalls:** Software running on the Droplet (e.g., `ufw`, `iptables`). A second layer of defense.

**The tag-based pattern:** Cloud Firewalls apply to Droplets via tags, not by Droplet ID. This is incredibly flexible:

```hcl
resource "digitalocean_firewall" "web" {
  name = "web-firewall"
  tags = ["web"]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["203.0.113.0/24"] # Office IP
  }
}

resource "digitalocean_droplet" "web" {
  # ...
  tags = ["web", "production"]
}
```

Any Droplet tagged `web` automatically gets the firewall rules. Add a new Droplet with the same tag? It's protected instantly. No need to update the firewall config.

**The most dangerous mistake:** Leaving SSH (port 22) open to `0.0.0.0/0` (the entire internet). This invites continuous brute-force attacks. Always restrict SSH to known IPs (office, VPN, or bastion host).

---

### SSH Key Management: No Passwords, Ever

**The only acceptable pattern:** SSH key-pair authentication.

**The anti-pattern:** Password-based SSH login. Passwords are vulnerable to brute-force attacks. SSH keys are cryptographically secure.

**How to do it right:**
1. Generate an SSH key pair locally (`ssh-keygen`)
2. Upload the public key to DigitalOcean (via control panel or API)
3. Reference the key in your IaC:

```hcl
resource "digitalocean_droplet" "web" {
  # ...
  ssh_keys = [data.digitalocean_ssh_key.admin.id]
}
```

4. Disable password authentication in SSH config (via cloud-init or Ansible)

**Critical gotcha:** On modern Ubuntu images, cloud-init sets `PasswordAuthentication yes` in `/etc/ssh/sshd_config.d/50-cloud-init.conf`, which overrides the main `sshd_config`. To disable password auth, your cloud-init script must modify that override file:

```yaml
#cloud-config
runcmd:
  - sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config.d/50-cloud-init.conf
  - systemctl restart sshd
```

---

### Backups vs. Snapshots: Understanding the Difference

**Automated Backups:**
- **What:** Automatic daily or weekly backups with 7-day or 4-week retention
- **Cost:** 20-30% of Droplet cost (percentage-based)
- **Use case:** Disaster recovery ("restore from yesterday")
- **Limitation:** Fixed retention, automatic (no control over when they happen)

**Manual Snapshots:**
- **What:** On-demand snapshots taken manually or via API, kept indefinitely
- **Cost:** $0.06/GB-month (size-based)
- **Use case:** Golden images (from Packer), point-in-time backups before major changes
- **Limitation:** Manual (you must trigger them)

**Best practice:** Enable backups for production Droplets (cheap insurance). Use snapshots for pre-change backups and custom images.

---

### Block Storage (Volumes): Persistent Data That Outlives Droplets

**What:** Network-attached SSD storage that persists independently of Droplets. Analogous to AWS EBS or GCP Persistent Disks.

**Why:** Droplet root disks are ephemeral. If you destroy the Droplet, the data is gone. Volumes let data outlive the Droplet—detach from old instance, attach to new one.

**Key limitation:** A Volume can only attach to **one Droplet at a time**. No shared read-write storage (like NFS or AWS EFS). For shared storage, you need a dedicated NFS server or an object storage workaround (mounting Spaces with `s3fs-fuse`, though performance is limited).

**Terraform pattern:**

```hcl
resource "digitalocean_volume" "db_data" {
  name   = "db-data"
  region = "nyc3"
  size   = 100
}

resource "digitalocean_droplet" "db" {
  # ...
  volume_ids = [digitalocean_volume.db_data.id]
}
```

---

### Cloud-Init and User Data: First-Boot Automation

**What:** The `user_data` field accepts a cloud-init script that runs on the Droplet's first boot. This is how you automate package installation, service configuration, and security hardening.

**Two common formats:**

1. **Shell script:**

```bash
#!/bin/bash
apt-get update
apt-get install -y nginx
systemctl enable nginx
systemctl start nginx
```

2. **Cloud-config YAML (declarative, preferred):**

```yaml
#cloud-config
package_update: true
package_upgrade: true
packages:
  - nginx
  - fail2ban
  - ufw

write_files:
  - path: /etc/nginx/sites-available/default
    content: |
      server {
        listen 80;
        location / {
          return 200 "Hello, world!";
        }
      }

runcmd:
  - systemctl enable nginx
  - systemctl start nginx
  - ufw allow 22/tcp
  - ufw allow 80/tcp
  - ufw --force enable
```

**Best practice:** Use cloud-config for readability and maintainability. Write idempotent scripts (safe to run multiple times). Test user_data in a dev environment before applying to production.

---

## The 80/20 Configuration: What Most Users Actually Need

Cross-tool analysis of `doctl`, Terraform, Pulumi, Ansible, and Crossplane reveals a consistent "core" of approximately **ten essential fields** that define 80% of all Droplet configurations:

### The Essential 80%

| **Field**       | **Type**         | **Description**                                                                 | **Why It's Essential**                                                  |
|------------------|------------------|---------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `name`          | string           | Hostname for the Droplet                                                        | Universally required. Must be DNS-compatible.                           |
| `region`        | string           | Datacenter location (e.g., `nyc3`, `sfo3`, `fra1`)                              | Universally required. Determines latency and compliance.                |
| `size`          | string           | Size slug (e.g., `s-2vcpu-4gb`, `g-8vcpu-32gb`, `s-1vcpu-1gb-amd`)             | Universally required. Must be a string to support Premium tiers.        |
| `image`         | string           | OS slug (e.g., `ubuntu-22-04-x64`) or snapshot ID                               | Universally required. Must be a string to support both types.           |
| `vpc_uuid`      | string           | VPC to attach the Droplet to                                                    | Essential for production networking. Replaces deprecated `private_networking`. |
| `ssh_keys`      | array of string  | SSH key fingerprints or IDs to embed in root account                            | Essential for secure access. Password auth is insecure.                 |
| `user_data`     | string           | Cloud-init script for first-boot provisioning                                   | Primary mechanism for automated initialization.                         |
| `monitoring`    | bool             | Enable DigitalOcean monitoring agent                                            | Free and recommended for production observability.                      |
| `backups`       | bool             | Enable automated backups                                                        | Essential for production data resiliency.                               |
| `tags`          | array of string  | Tags for organization and integration                                           | **Critical:** Primary integration point for Cloud Firewalls and Load Balancers. |

### The Rare/Advanced 20%

- `ipv6`: Enable IPv6 (common but not minimal)
- `resize_disk`: Lifecycle parameter for resizing existing Droplets
- `graceful_shutdown`: Deletion behavior control
- `kernel`: Custom kernel selection (extremely rare)

---

### Configuration Examples

**Development (Minimal):**

```yaml
name: dev-server
region: nyc3
size: s-1vcpu-1gb
image: ubuntu-22-04-x64
ssh_keys: [dev-key-fingerprint]
tags: [dev]
```

**Staging (Standard):**

```yaml
name: staging-app
region: sfo3
size: g-2vcpu-8gb
image: ubuntu-22-04-x64
vpc_uuid: vpc-staging-uuid
ssh_keys: [cicd-key-id]
monitoring: true
backups: true
tags: [staging, webapp, staging-firewall]
user_data: |
  #cloud-config
  package_update: true
  packages:
    - nginx
```

**Production (High-Performance):**

```yaml
name: prod-api-worker
region: fra1
size: c-4  # CPU-Optimized (4 vCPU / 8 GB RAM)
image: "12345678"  # Custom Packer snapshot ID
vpc_uuid: vpc-prod-uuid
ssh_keys: [prod-key-id]
monitoring: true
backups: true
volumes: [volume-data-uuid]
tags: [production, api-worker, prod-api-lb, prod-db-fw]
```

---

## OpenMCF's Approach: Abstraction with Pragmatism

OpenMCF abstracts DigitalOcean Droplet provisioning behind a clean, protobuf-defined API (`DigitalOceanDroplet`). This provides a consistent interface while respecting DigitalOcean's native idioms.

### What We Abstract

The `DigitalOceanDropletSpec` includes the essential 80%:

- **`droplet_name`**: DNS-compatible hostname (lowercase alphanumeric + hyphens, ≤63 chars)
- **`region`**: DigitalOcean region enum (validated against supported regions)
- **`size`**: Size slug as an opaque string (supports standard and Premium tiers like `s-1vcpu-1gb-amd`)
- **`image`**: Image slug or snapshot ID as a string (supports both distribution images and custom snapshots)
- **`vpc`**: VPC reference (can reference another DigitalOceanVpc resource or provide UUID directly)
- **`enable_ipv6`**: Optional IPv6 enablement (disabled by default)
- **`enable_backups`**: Optional automated backups (disabled by default, but recommended for production)
- **`disable_monitoring`**: Opt-out flag (monitoring is enabled by default because it's free and useful)
- **`volume_ids`**: Array of Volume references for persistent storage
- **`tags`**: Array of tags for Cloud Firewall and Load Balancer integration
- **`user_data`**: Cloud-init script (max 32 KiB)
- **`timezone`**: Optional timezone setting (UTC by default)

### Design Decisions

**Why `size` is a string:** To natively support DigitalOcean's Premium CPU tiers (e.g., `s-1vcpu-1gb-amd`, `s-1vcpu-1gb-intel`). The "Premium" designation is embedded in the slug, not a separate boolean. Treating `size` as an opaque string matches DigitalOcean's API and avoids abstraction complexity.

**Why `image` is a string:** To support both distribution slugs (`ubuntu-22-04-x64`) and numeric snapshot IDs (from Packer or manual snapshots). A single string type covers both use cases cleanly.

**Why `vpc_uuid` is required:** Modern production standard. The deprecated `private_networking` boolean is omitted to avoid legacy patterns and user confusion.

**Why `tags` is first-class:** Tags are the primary integration point for Cloud Firewalls and Load Balancers. They're not optional metadata—they're essential infrastructure glue.

**Why monitoring is on by default:** It's free, provides essential metrics (CPU, memory, disk, network), and there's no reason to disable it unless you're using external monitoring exclusively. The API uses a `disable_monitoring` flag to make the default explicit.

### Under the Hood: Pulumi

OpenMCF currently uses **Pulumi (Go)** for DigitalOcean Droplet provisioning. Why?

- **Language consistency:** Pulumi's Go SDK fits naturally with OpenMCF's broader multi-cloud orchestration (also Go-based)
- **Programming flexibility:** Pulumi's programming model makes conditional logic, multi-resource strategies, and custom integrations straightforward
- **Equivalent coverage:** Pulumi's DigitalOcean provider (bridged from Terraform) supports all Droplet operations we need

That said, Terraform would work equally well for standard provisioning. The choice is an implementation detail—the protobuf API remains the same regardless.

---

## Key Takeaways

1. **Droplets are for control, not convenience.** If you need granular infrastructure control, self-hosted databases, or custom server configurations, Droplets are the right choice. If you want zero-ops deployment, use App Platform.

2. **Manual provisioning is an anti-pattern.** The DigitalOcean control panel is for learning, not production. Use Infrastructure-as-Code (Terraform or Pulumi) for repeatability, state management, and version control.

3. **The production stack is Packer + Terraform/Pulumi.** Build golden images with Packer, provision Droplets from those snapshots with IaC, and use Terraform's `create_before_destroy` for zero-downtime updates.

4. **VPC, SSH keys, and Cloud Firewalls are non-negotiable.** Every production Droplet belongs to a VPC, uses SSH key authentication (never passwords), and has a Cloud Firewall applied via tags. Restrict SSH to known IPs.

5. **Tags are infrastructure glue.** Cloud Firewalls and Load Balancers integrate with Droplets via tags, not IDs. This decouples infrastructure components and enables dynamic scaling.

6. **The 80/20 config is name, region, size, image, VPC, SSH keys, tags, monitoring, backups, and user_data.** Advanced options (IPv6, custom kernels, resize behavior) are rare. Focus on the essentials.

7. **Backups and Volumes are different.** Backups are for disaster recovery (automatic, percentage-based pricing). Snapshots are for golden images (manual, size-based pricing). Volumes are for persistent data that outlives Droplets.

8. **Cloud-init is your first-boot automation layer.** Use cloud-config YAML for declarative provisioning. Test in dev before production. Remember the SSH config override gotcha on Ubuntu.

9. **OpenMCF abstracts the API** into a clean protobuf spec, making multi-cloud deployments consistent while respecting DigitalOcean's unique characteristics. The API prioritizes the 80% of config that 80% of users need.

---

## Further Reading

- **DigitalOcean Droplets Documentation:** [DigitalOcean Docs - Droplets](https://docs.digitalocean.com/products/droplets/)
- **Terraform DigitalOcean Provider:** [GitHub - digitalocean/terraform-provider-digitalocean](https://github.com/digitalocean/terraform-provider-digitalocean)
- **Pulumi DigitalOcean Package:** [Pulumi Registry - DigitalOcean](https://www.pulumi.com/registry/packages/digitalocean/)
- **Packer DigitalOcean Builder:** [HashiCorp Developer - Packer DigitalOcean](https://developer.hashicorp.com/packer/plugins/builders/digitalocean)
- **DigitalOcean API Reference:** [DigitalOcean API - Droplets](https://docs.digitalocean.com/reference/api/api-reference/#tag/Droplets)
- **Recommended Droplet Setup Guide:** [DigitalOcean Docs - Production-Ready Droplet](https://docs.digitalocean.com/products/droplets/getting-started/recommended-droplet-setup/)
- **Cloud-Init Documentation:** [cloud-init.readthedocs.io](https://cloudinit.readthedocs.io/)

---

**Bottom Line:** DigitalOcean Droplets give you full control over Linux VMs with predictable pricing and straightforward management. Manage them with Infrastructure-as-Code (Terraform or Pulumi), build golden images with Packer, secure them with VPCs and Cloud Firewalls, and automate initialization with cloud-init. OpenMCF makes this simple with a protobuf API that hides complexity while exposing the essential configuration you actually need—nothing more, nothing less.

