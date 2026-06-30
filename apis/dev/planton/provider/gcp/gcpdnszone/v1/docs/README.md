# GCP Cloud DNS: From Manual Clicks to Production-Grade Automation

## Introduction

DNS management has long operated under a paradox. It is simultaneously **critical and boring**—critical because a misconfigured DNS record can take down an entire production system, boring because it seems like a solved problem that shouldn't require much attention. This mindset leads organizations to manage DNS through a patchwork of manual console clicks, ad-hoc scripts, and fragile automation that breaks the moment someone needs to integrate with Kubernetes or automate certificate renewals.

Google Cloud DNS is a high-performance, globally distributed authoritative DNS service built on the same infrastructure that powers Google's own services. It provides two fundamental capabilities: **public zones** that serve domain records to the internet, and **private zones** that enable service discovery within GCP Virtual Private Clouds. The service abstracts away the operational complexity of running DNS servers while providing enterprise-grade reliability and performance.

The challenge isn't whether Cloud DNS is capable—it absolutely is—but rather **how to manage it** in a way that supports modern cloud-native workflows. Teams deploying workloads to Google Kubernetes Engine need DNS records to appear automatically when they create Ingresses. Teams using cert-manager to provision TLS certificates need DNS automation to solve ACME challenges. Teams operating in hybrid-cloud environments need DNS to resolve seamlessly between GCP and on-premises data centers. None of these scenarios work well when DNS is managed through manual console operations.

This document presents a structured view of Cloud DNS deployment methods, from anti-patterns that create operational debt to production-ready solutions that integrate seamlessly with Infrastructure-as-Code and Kubernetes-native automation. It explains what deployment approaches exist, how they've evolved, and why Planton's API is designed the way it is.

---

## The Maturity Spectrum: How DNS Management Evolved

### Level 0: The Console-Driven Anti-Pattern

The Google Cloud Console provides a straightforward web interface for creating DNS zones and adding records. You navigate to the "Create a DNS zone" page, select whether it's public or private, enter a domain name, and optionally enable DNSSEC. For simple scenarios—like pointing a single domain to a static IP for a personal project—this workflow is perfectly reasonable.

**Why this breaks down in production:**

The console workflow is **inherently non-reproducible**. When you create a DNS zone through the console, there is no artifact, no version control, and no audit trail beyond GCP's activity logs. If someone forgets to enable DNSSEC for a public zone, or if they set a 30-second TTL that quadruples your Cloud DNS bill (because DNS charges per query), there's no way to catch these mistakes before they happen.

The real killer is **scale and integration**. The moment your team needs to automate DNS updates—for Kubernetes Ingresses, for certificate renewals, for failover—the console becomes a bottleneck. You cannot integrate button-clicking into a CI/CD pipeline. You cannot test DNS configurations in a staging environment and promote them to production with confidence.

**Verdict:** Console-driven DNS management is acceptable for prototypes and personal projects but is an anti-pattern for any environment where reproducibility, automation, or multi-environment consistency matters.

---

### Level 1: Scripting with gcloud CLI

The `gcloud` command-line tool provides programmatic access to Cloud DNS. You can create zones with `gcloud dns managed-zones create`, add records using a transactional workflow (`gcloud dns record-sets transaction start`, `add`, `execute`), and enable DNSSEC with flags like `--dnssec-state=on`.

This is a meaningful step forward. Scripts written with `gcloud` can be version-controlled, reviewed, and executed in CI/CD pipelines. A team can maintain a repository of scripts that provision their DNS infrastructure, ensuring that every environment (dev, staging, production) is configured identically.

**The limitations:**

Scripts written in Bash or Python that shell out to `gcloud` are **imperative**, not **declarative**. They describe *how* to create resources step-by-step, but they don't capture a desired end-state. If a script runs twice, it might fail because the resource already exists, or worse, it might create duplicate records. You have to write custom logic to check whether resources exist, whether they've changed, and whether they need to be updated or deleted.

There's also no concept of **state management**. If someone manually modifies a DNS record through the console, your script has no way of detecting that drift. The script only knows what it last executed, not what the actual current state of Cloud DNS is.

**Verdict:** CLI scripting is a necessary tool for ad-hoc administrative tasks, but it's insufficient as a foundation for managing production DNS infrastructure. It lacks the declarative semantics and state management that modern infrastructure management requires.

---

### Level 2: Configuration Management with Ansible

Ansible provides a step toward declarative infrastructure through its `google.cloud` collection, which includes modules for `gcp_dns_managed_zone` and `gcp_dns_resource_record_set`. These modules allow you to define DNS zones and records in YAML playbooks, and Ansible's idempotency ensures that running the same playbook multiple times produces the same result.

This is a valid approach for organizations that have already standardized on Ansible for configuration management. It provides version control, repeatability, and a declarative syntax that's easier to reason about than Bash scripts.

**Why it's still not ideal:**

Ansible operates on a **push-based** model. It doesn't maintain a persistent state file that tracks what resources have been created. Instead, it queries the GCP API during each run to determine what exists and what needs to change. This works, but it's slower and less efficient than tools designed specifically for infrastructure state management.

More critically, Ansible doesn't provide the same level of **ecosystem integration** as tools like Terraform or Pulumi. There's no native way to output the nameservers assigned to a DNS zone and use them as input to another stack. There's no way to manage the lifecycle of DNS zones alongside the lifecycle of GKE clusters, load balancers, and other GCP resources in a unified dependency graph.

**Verdict:** Ansible is a workable solution for teams already committed to it, but for greenfield projects, Infrastructure-as-Code tools purpose-built for cloud resource management provide better state handling, dependency management, and ecosystem integration.

---

### Level 3: Production-Grade Infrastructure-as-Code

This is where DNS management becomes genuinely production-ready: treating DNS zones and records as **declarative infrastructure** managed by tools like **Terraform**, **Pulumi**, or **OpenTofu**. These tools provide:

1. **Declarative configuration:** You define the desired state (e.g., "a public zone named `example.com` with DNSSEC enabled"), and the tool figures out what API calls are needed to achieve it.
2. **State management:** The tool maintains a state file that tracks what resources exist and detects drift if someone makes manual changes.
3. **Dependency resolution:** You can reference outputs from one resource (like the IP address of a load balancer) as inputs to another (like an A record in a DNS zone).
4. **Multi-environment support:** You can use the same codebase to manage dev, staging, and production environments with different variable values.

#### Terraform: The Industry Standard

Terraform is the de facto standard for IaC, and its `google_dns_managed_zone` and `google_dns_record_set` resources are battle-tested. A typical Terraform configuration might look like this:

```hcl
resource "google_dns_managed_zone" "public_zone" {
  project     = "my-gcp-project"
  name        = "example-com-public"
  dns_name    = "example.com."
  visibility  = "public"
  description = "Public zone for example.com"

  dnssec_config {
    state = "on"
  }
}

resource "google_dns_record_set" "a_root" {
  project      = google_dns_managed_zone.public_zone.project
  managed_zone = google_dns_managed_zone.public_zone.name
  name         = "example.com."
  type         = "A"
  ttl          = 300
  rrdatas      = ["104.198.14.52"]
}
```

This configuration is **declarative** (it describes what should exist), **version-controlled** (it can be committed to Git), and **auditable** (changes require code review and CI approval). Terraform's state file tracks the resources it manages, and running `terraform plan` shows exactly what will change before any API calls are made.

#### Pulumi: General-Purpose Languages

Pulumi offers the same declarative model as Terraform but allows you to write infrastructure code in **general-purpose programming languages** like Go, Python, and TypeScript. This is powerful for scenarios where you need to generate DNS records dynamically based on complex logic, or when you want to integrate infrastructure management directly into your application codebase.

A Pulumi program in Python might look like:

```python
import pulumi_gcp as gcp

public_zone = gcp.dns.ManagedZone("public-zone",
    project="my-gcp-project",
    dns_name="example.com.",
    visibility="public",
    dnssec_config=gcp.dns.ManagedZoneDnssecConfigArgs(
        state="on",
    ))
```

Pulumi uses a managed state backend by default, which simplifies team collaboration by handling state locking and concurrency automatically.

#### OpenTofu: The Open-Source Fork

Following HashiCorp's license change for Terraform, **OpenTofu** emerged as a community-driven, open-source fork. It is fully compatible with Terraform's HCL syntax and uses the same `google_dns_managed_zone` resources. For organizations committed to open-source infrastructure tools, OpenTofu provides a path forward without vendor lock-in.

**Verdict:** This is the production-ready tier. Terraform, Pulumi, and OpenTofu provide the declarative semantics, state management, and ecosystem integration required to manage DNS at scale. **This is the foundation Planton is built on.**

---

### Level 4: Kubernetes-Native IaC (GCP Config Connector)

For teams that have adopted Kubernetes as their platform abstraction layer, **Config Connector (KCC)** offers a compelling alternative. Config Connector is a Kubernetes operator that allows you to manage GCP resources—including Cloud DNS zones and records—using Kubernetes Custom Resource Definitions (CRDs).

A DNS zone in Config Connector looks like this:

```yaml
apiVersion: dns.cnrm.cloud.google.com/v1beta1
kind: DNSManagedZone
metadata:
  name: example-com-public
spec:
  dnsName: "example.com."
  visibility: public
  dnssecConfig:
    state: "on"
```

The key differentiator is **continuous reconciliation**. Unlike Terraform, which applies changes when you run `terraform apply`, Config Connector runs as a controller inside your Kubernetes cluster and **continuously monitors** the state of GCP resources. If someone manually deletes a DNS record through the GCP console, Config Connector detects the drift and automatically recreates it to match the desired state defined in the Kubernetes API.

This model aligns perfectly with **GitOps workflows**, where infrastructure definitions are stored in Git, and tools like ArgoCD or Flux automatically apply changes when commits are pushed.

**When to choose Config Connector:**

- Your team has deep Kubernetes expertise and wants to manage everything through the Kubernetes API.
- You're building a Kubernetes-based internal platform where application teams define infrastructure as CRDs.
- You want continuous drift detection and auto-remediation without running scheduled CI/CD jobs.

**Trade-off:** Config Connector requires a GKE cluster to run the operator, which adds operational complexity. For teams managing GCP resources outside of Kubernetes contexts (like Cloud Run, BigQuery, or IAM), Terraform may be a more pragmatic choice.

**Verdict:** For GKE-native organizations, Config Connector is a production-ready solution that integrates infrastructure management directly into Kubernetes workflows. For multi-cloud or non-Kubernetes-centric platforms, Terraform or Pulumi remain the better choice.

---

## The "Split State" Pattern: Why DNS Zones and Records Must Be Separate

One of the most critical insights from production DNS management is the **"split state" pattern**. This pattern recognizes that DNS resources have two distinct lifecycle owners:

1. **The Infrastructure Layer:** DNS zones themselves, along with foundational records like MX (mail servers), TXT (SPF, DKIM, domain verification), and NS (subdomain delegation), are **long-lived, stable resources** managed by platform administrators.

2. **The Application Layer:** DNS records for application workloads—particularly A and CNAME records for Kubernetes Ingresses and Services—are **dynamic, ephemeral resources** that need to be created and destroyed automatically as applications are deployed and torn down.

In Kubernetes environments, teams use **external-dns**—a controller that watches Ingress and Service resources and automatically creates DNS records to match. When a developer deploys an app with an Ingress for `app.example.com`, external-dns detects the Ingress, retrieves the assigned load balancer IP, and creates the corresponding A record in Cloud DNS. When the app is deleted, the DNS record is automatically cleaned up.

**The critical mistake:** Modeling DNS records as an inline array inside the DNS zone resource. If your IaC tool defines both the zone and its records in a single resource, then every time the IaC tool runs (e.g., in CI/CD), it will try to "enforce" its view of what records should exist. Since external-dns is creating records dynamically, the IaC tool will attempt to delete them, creating a perpetual "war" between the two systems.

**The correct pattern:** Model the DNS zone and DNS records as **separate, independent resources**. This is exactly how Terraform structures its provider:

- `google_dns_managed_zone` provisions the zone.
- `google_dns_record_set` provisions individual records that reference the zone by name.

This separation allows infrastructure teams to manage zones and static records with Terraform while allowing external-dns or cert-manager to manage dynamic application records within the same zone. The systems coexist peacefully because each manages its own distinct set of resources.

**Planton's API design reflects this principle.** The `GcpDnsZone` resource provisions the zone and static records (defined in the spec). Dynamic records managed by external-dns are handled outside of the IaC lifecycle, as they should be.

---

## Ecosystem Integration: DNS as a Platform Service

Cloud DNS doesn't exist in isolation. In production environments, it's deeply integrated with Kubernetes, certificate management, and multi-region architectures.

### External-DNS: Automating Application DNS

**external-dns** is a Kubernetes controller that solves the problem of DNS for ephemeral workloads. It runs in-cluster, watches Ingress and Service resources, and automatically creates DNS records in Cloud DNS to match.

When a developer creates an Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-app
            port:
              number: 80
```

external-dns detects this Ingress, retrieves the IP address of the load balancer provisioned by the GKE Ingress controller, and creates an A record for `app.example.com` pointing to that IP. The developer never touches the GCP console or writes Terraform for DNS—the record appears automatically.

To prevent accidental data loss, external-dns creates a corresponding TXT record for each A record to "claim ownership." It only manages records that it owns, allowing IaC-managed static records to coexist safely.

**IAM Requirements:** external-dns needs a GCP service account with the following permissions:

- `dns.managedZones.list`
- `dns.resourceRecordSets.*` (or more granular `create`, `update`, `delete`)
- `dns.changes.*`

In GKE, this is typically configured using **Workload Identity**, which binds a Kubernetes ServiceAccount to a GCP service account without needing to manage JSON key files.

---

### cert-manager: Automated TLS Certificate Provisioning

**cert-manager** is a Kubernetes controller that automates the provisioning of TLS certificates from ACME-based certificate authorities like Let's Encrypt. For wildcard certificates (e.g., `*.example.com`), cert-manager must solve a **DNS-01 challenge**, which requires programmatic access to DNS.

The workflow:

1. cert-manager requests a wildcard certificate from Let's Encrypt.
2. Let's Encrypt challenges cert-manager to prove domain ownership by creating a specific TXT record.
3. cert-manager uses its GCP service account credentials to create the TXT record at `_acme-challenge.example.com`.
4. Let's Encrypt performs a DNS lookup, verifies the record, and issues the certificate.
5. cert-manager stores the certificate as a Kubernetes Secret and deletes the temporary TXT record.

**IAM Requirements:** cert-manager needs the same permissions as external-dns:

- `dns.managedZones.list`
- `dns.resourceRecordSets.*`
- `dns.changes.*`

The IaC layer provisions the DNS zone and grants these permissions to the cert-manager service account. Once this setup is complete, developers can request TLS certificates declaratively by creating `Certificate` resources in Kubernetes, and the entire ACME workflow happens automatically.

---

### Multi-Region Failover: DNS-Based Traffic Routing

Cloud DNS supports **advanced routing policies** that enable sophisticated traffic management patterns. These are the "20%" features that only a subset of users need, but they're critical for global, high-availability architectures.

#### Geolocation Routing

A company with users in both North America and Europe might deploy their application to GKE clusters in `us-west1` and `europe-west1`. Using a **geolocation routing policy**, they create a single DNS record for `app.example.com` that routes:

- European traffic to the `europe-west1` load balancer IP.
- All other traffic to the `us-west1` load balancer IP.

This reduces latency for global users by directing them to the geographically nearest deployment.

#### Active-Passive Failover

A company with a primary deployment in `us-central1` and a disaster recovery (DR) deployment in `us-east1` can use a **failover routing policy**. The primary target is the `us-central1` load balancer, which is monitored by GCP health checks. If the health checks fail (indicating an outage), Cloud DNS automatically switches to serving the `us-east1` IP, redirecting all traffic to the DR region.

This provides **automated, DNS-based failover** without requiring changes to application code or client configurations.

---

### Hybrid Cloud: Integrating On-Premises DNS

Many enterprises operate in **hybrid-cloud** environments, where some workloads run in GCP and others run in on-premises data centers. DNS must resolve seamlessly across both environments.

**Outbound Forwarding (GCP VMs resolving on-prem names):**

Create a **Forwarding Zone** in Cloud DNS for the on-premises domain (e.g., `onprem.corp.com`). This zone is configured with the IP addresses of the on-premises DNS servers as forwarding targets. When a GCP VM queries `server.onprem.corp.com`, Cloud DNS forwards the query over a Cloud VPN or Interconnect connection to the on-prem DNS servers.

**Inbound Forwarding (on-prem servers resolving GCP private names):**

Create a **DNS Inbound Server Policy** on the GCP VPC. This policy provides a stable "inbound forwarder" IP address (from the `35.199.192.0/19` range). The on-premises DNS administrator configures their BIND or Active Directory server with a conditional forwarder for the GCP domain (e.g., `gcp.internal`) pointing to this IP. Queries from on-prem servers for GCP resources are forwarded to Cloud DNS, which resolves them using the appropriate private zone.

This bidirectional DNS integration is foundational for hybrid-cloud architectures, enabling seamless service discovery across network boundaries.

---

## DNSSEC: Security Without Complexity

**DNSSEC** (DNS Security Extensions) protects domains from cache poisoning and man-in-the-middle attacks by providing cryptographic authentication of DNS responses. Traditionally, DNSSEC has been operationally complex—requiring manual key generation, rotation, and management.

Cloud DNS simplifies this dramatically. Enabling DNSSEC is as simple as setting `state = "on"` in your configuration. Cloud DNS automatically generates DNSSEC keys (using modern algorithms like `ecdsap256sha256`), signs all zone data, and handles key rotation.

**The user's only responsibility:** Retrieve the **DS (Delegation Signer) record** from Cloud DNS and add it to your domain registrar (the entity where you purchased the domain, like Google Domains or Cloudflare). This completes the "chain of trust" that allows DNS resolvers to validate your zone's signatures.

**Best Practice:** Enable DNSSEC for **all** public-facing zones. It's a straightforward security enhancement with virtually no operational overhead in Cloud DNS.

**Common Failure:** Enabling DNSSEC in Cloud DNS but forgetting to add the DS record to the registrar—or disabling it in one place but not the other—will cause validation failures for DNSSEC-validating resolvers, effectively taking your domain offline for users with strict validators. Always validate the full chain using tools like the **Verisign DNSSEC Debugger** or **DNSViz**.

---

## Production Essentials: IAM, Logging, and Anti-Patterns

### Least-Privilege IAM

While it's tempting to grant `roles/dns.admin` to service accounts for simplicity, this role is overly permissive and includes dangerous permissions like `dns.managedZones.delete`.

**Production-grade approach:** Create custom IAM roles with least-privilege permissions:

| Actor | Purpose | Recommended Permissions |
|-------|---------|------------------------|
| **CI/CD (Terraform)** | Manage zone lifecycle | `dns.managedZones.*`, `dns.resourceRecordSets.*`, `dns.changes.*` |
| **external-dns** | Manage application records | `dns.managedZones.list`, `dns.resourceRecordSets.*`, `dns.changes.*` |
| **cert-manager** | DNS-01 ACME challenges | `dns.managedZones.list`, `dns.resourceRecordSets.*`, `dns.changes.*` |

Note that external-dns and cert-manager require nearly identical permissions—both need to list zones (to find the right one) and manage record sets within those zones.

### Logging and Monitoring

Cloud DNS provides two types of logs:

1. **Query Logs (Data Access):** Logs every DNS query sent to your zones. For **public zones**, enable per-zone with `gcloud dns managed-zones update [ZONE_NAME] --log-dns-queries`. For **private zones**, enable via a DNS Policy attached to the VPC network.

2. **Audit Logs (Admin Activity):** Logs administrative actions like zone creation, record changes, and deletions.

**Log-Based Metrics:** Create custom log-based metrics to track query volume, NXDOMAIN rates (non-existent domain errors), and response latencies. These metrics can be used to build dashboards and alerts for DNS health monitoring.

### Common Anti-Patterns to Avoid

1. **Overly Short TTLs:** Setting TTLs to 30 or 60 seconds disables DNS caching and forces resolvers to query Cloud DNS for every request. Since Cloud DNS charges **per query**, this can dramatically inflate costs. For fast failover, use **DNS Routing Policies with health checks** instead of relying on low TTLs.

2. **Managing Dynamic Records with IaC:** Using Terraform or Pulumi to manage the A records for Kubernetes Ingresses. This violates the "split state" principle and creates conflicts with external-dns.

3. **Hardcoded IPs in Automation:** Applications and scripts that use hardcoded IP addresses instead of DNS names. This defeats the purpose of DNS as a service discovery mechanism and makes infrastructure brittle.

4. **DNSSEC Configuration Mismatch:** Enabling DNSSEC in Cloud DNS but forgetting to add the DS record to the registrar, or vice versa. This will cause resolution failures.

---

## What Planton Provides: An 80/20 API for Production DNS

The `GcpDnsZone` API resource in Planton is designed around the principle that **most users need 20% of the configuration options 80% of the time**. 

### Core Design Decisions

1. **Separate Zone and Record Lifecycle:** The `GcpDnsZoneSpec` includes a `records` field for static, foundational DNS records (like MX, TXT, and NS). Dynamic application records are expected to be managed by external-dns, not by the IaC layer. This respects the "split state" pattern.

2. **IAM Service Account Integration:** The `iam_service_accounts` field allows platform administrators to grant DNS management permissions to workload identities like cert-manager and external-dns. This is provisioned as part of the zone lifecycle, ensuring that automation tools have the access they need from day one.

3. **Simplified Record Definitions:** DNS records are defined with four essential fields:
   - `record_type` (A, CNAME, TXT, MX, NS, etc.)
   - `name` (the FQDN, e.g., `www.example.com.`)
   - `values` (the record data, e.g., IP addresses)
   - `ttl_seconds` (with a sensible default of 60 seconds)

   Advanced routing policies (geolocation, failover, weighted round-robin) are intentionally omitted from the core API because they represent the "20%" use case. Teams that need these capabilities can manage them through direct Terraform resources or Config Connector CRDs.

4. **Public Zones Only (for Now):** The current API focuses on **public zones**, which are the most common use case. Private zones for VPC-internal service discovery and advanced zone types (forwarding zones, peering zones) are intentionally scoped out of the initial API. These features may be added in future versions based on user demand.

### Example Configuration

Here's a typical `GcpDnsZone` resource that provisions a public zone with DNSSEC (assumed enabled by default in the underlying Pulumi/Terraform module), grants permissions to cert-manager, and defines foundational DNS records:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example-com
spec:
  project_id: my-gcp-project
  iam_service_accounts:
  - cert-manager@my-gcp-project.iam.gserviceaccount.com
  records:
  - record_type: A
    name: example.com.
    values:
    - 104.198.14.52
    ttl_seconds: 300
  - record_type: CNAME
    name: www.example.com.
    values:
    - example.com.
    ttl_seconds: 300
  - record_type: TXT
    name: example.com.
    values:
    - google-site-verification=abc123xyz789
    ttl_seconds: 3600
  - record_type: MX
    name: example.com.
    values:
    - 10 mail.example.com.
    ttl_seconds: 3600
```

This configuration is **declarative**, **version-controlled**, and **production-ready**. The underlying Pulumi module (or Terraform module, depending on the stack runner) provisions the Cloud DNS managed zone, enables DNSSEC, creates the specified records, and grants the cert-manager service account the permissions it needs to solve DNS-01 challenges.

---

## Conclusion

Managing DNS in Google Cloud has evolved from manual console operations to fully automated, Kubernetes-integrated workflows. The production-ready approach centers on three principles:

1. **Declarative Infrastructure-as-Code:** Terraform, Pulumi, or Config Connector provide the foundation for reproducible, version-controlled DNS management.

2. **The Split State Pattern:** DNS zones and static foundational records are managed by the IaC layer. Dynamic application records are managed by automation tools like external-dns and cert-manager. These systems coexist by managing distinct sets of resources.

3. **Ecosystem Integration:** DNS is not a standalone service. It's deeply integrated with Kubernetes for service discovery, with cert-manager for certificate automation, and with hybrid-cloud architectures for cross-network resolution.

Planton's `GcpDnsZone` API embodies these principles by providing a **minimal, production-grade interface** for the 80% use case—provisioning public DNS zones, enabling DNSSEC by default, defining foundational records, and granting IAM permissions to automation tools—while leaving advanced features (routing policies, private zones, forwarding zones) to be managed through direct Terraform or Config Connector when needed.

This is DNS management designed for modern cloud-native platforms: declarative, automated, and secure by default.

