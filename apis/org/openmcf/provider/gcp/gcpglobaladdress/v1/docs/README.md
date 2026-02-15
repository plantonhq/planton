# GCP Global Addresses: From Manual IP Reservations to Codified Network Identity

## Introduction

Every cloud application that communicates beyond its own boundaries needs an IP address. For many workloads, an ephemeral address is fine — the address changes when the resource is recreated, and nothing depends on its stability. But for a significant class of infrastructure — load balancers, CDN frontends, VPC-peered managed services, Private Service Connect endpoints — you need an address that persists independently of the resource it's attached to. That's what global address reservations provide.

In Google Cloud Platform, a global address (`google_compute_global_address`) reserves either a public IP at global scope or a private IP range within a VPC network. The "global" designation means the address is not tied to a specific region — it can be used by global resources like HTTP(S) load balancers, SSL proxies, and TCP proxies. For internal addresses, "global" means the reserved range spans the entire VPC network, making it available for cross-region VPC peering with managed services.

The resource itself is deceptively simple — a handful of fields that produce a single IP or CIDR range. The complexity lies in understanding *when* and *why* you need one, how the fields interact (particularly `address_type`, `purpose`, `network`, and `prefix_length`), and how to manage these reservations as code rather than as forgotten console artifacts. This document traces the evolution of address management in GCP, compares the available deployment methods, and explains how OpenMCF distills the essential configuration into a clean, validated API.

## The Three Faces of Global Addresses

Global addresses serve three distinct use cases, each with its own configuration pattern. Understanding these patterns is the key to using this resource correctly.

### Face 1: External Static IPs

The most straightforward use case. Reserve a public IPv4 or IPv6 address that you can attach to a global forwarding rule. The address persists across load balancer updates, regional failovers, and infrastructure re-provisioning.

**Why you need this:**
- DNS records (A/AAAA) point to this IP. Changing the IP means updating DNS and waiting for propagation.
- Partner and customer allowlists reference this IP. A new IP means coordinating with external parties.
- SSL certificates may be bound to this IP (for classic SSL proxy LBs).
- CDN frontends need a stable anycast IP for edge caching.

**Configuration:** `address_type: EXTERNAL`, optionally `ip_version: IPV6`. No `network`, `purpose`, or `prefix_length` needed.

### Face 2: Internal VPC Peering Ranges

Google's managed services — Cloud SQL, Memorystore (Redis/Memcached), AlloyDB, Filestore, and others — can connect to your VPC via VPC network peering. This peering requires a reserved private IP range that Google's service network uses to assign addresses to your managed service instances.

**Why you need this:**
- Without a reserved range, you can't enable private IP on Cloud SQL, Redis, AlloyDB, or Filestore.
- The range determines which RFC 1918 addresses your managed service instances receive.
- Multiple managed services can share the same reserved range.
- The reservation prevents address conflicts between your VPC subnets and the peered network.

**Configuration:** `address_type: INTERNAL`, `purpose: VPC_PEERING`, `prefix_length: 20` (or similar), `network: <vpc-self-link>`.

### Face 3: Private Service Connect Addresses

Private Service Connect (PSC) enables private connectivity to Google APIs, Google-managed services, or third-party services published by other GCP projects. A PSC endpoint requires a reserved internal IP that becomes the entry point for private traffic.

**Why you need this:**
- PSC endpoints route traffic to Google APIs (like `storage.googleapis.com`) through your VPC instead of over the internet.
- Published services (your own or third-party) get a private IP in your VPC for access without exposing them publicly.
- DNS can resolve Google API domains to these private IPs, enabling fully private access paths.

**Configuration:** `address_type: INTERNAL`, `purpose: PRIVATE_SERVICE_CONNECT`, `network: <vpc-self-link>`. No `prefix_length` (PSC uses a single IP).

## Evolution and Historical Context

### The IP Address Lifecycle Problem

In the early days of cloud computing, IP addresses were treated as disposable. Spin up a VM, get an IP. Tear it down, lose the IP. This worked for internal services but created problems as soon as external dependencies entered the picture.

The first fix was "static IP reservation" — a way to claim an IP address and keep it even when the resource using it is deleted. Cloud providers all converged on this pattern:

1. **Reserve** an address (creating the reservation resource)
2. **Attach** it to a compute resource (VM, load balancer, forwarding rule)
3. **Detach** and **re-attach** as needed (the reservation persists)
4. **Release** when truly no longer needed (returns the IP to the pool)

GCP distinguishes between regional addresses (tied to a region, used by VMs and regional load balancers) and global addresses (region-agnostic, used by global load balancers). This document focuses on global addresses because they serve the use cases with the highest management complexity: load balancer frontends and VPC peering ranges.

### The Private Networking Revolution

The second evolutionary wave came with the shift to private networking. Organizations moved from "everything talks over public IPs" to "everything stays within the VPC, and only the edge touches the internet." This created demand for:

- **VPC peering ranges**: Managed services needed private IP space to peer into customer VPCs.
- **Private Service Connect**: A more flexible alternative to VPC peering, allowing service-to-service connectivity through a single IP endpoint rather than a full network peering relationship.

Both patterns require reserved internal global addresses, but with fundamentally different semantics: VPC peering reserves a *range* (specified by `prefix_length`), while PSC reserves a *single IP*. The same GCP resource type handles both, distinguished by the `purpose` field.

### Why These Reservations Accumulate

Global address reservations are among the most "set and forget" resources in a cloud environment:

1. **External static IPs** are created once, added to DNS, shared with partners, and then left untouched for years. Nobody wants to change a production load balancer's IP.
2. **VPC peering ranges** are created when private networking is first configured and then persist for the lifetime of the VPC. Deleting the range breaks all managed service connectivity.
3. **PSC addresses** are tied to endpoint configurations that are rarely modified.

The problem isn't creating these resources — it's tracking them. In organizations with multiple projects and VPCs, address reservations become invisible infrastructure that everyone depends on but nobody manages. Moving them to code is the first step toward visibility.

## Deployment Methods Landscape

### Level 0: Manual Console Provisioning

The GCP Console provides address reservation under **VPC network → IP addresses → Reserve External Static Address** (or via the Compute Engine section). You fill in a form: name, network tier, IP version, type, and optionally a specific address.

For internal VPC peering ranges, the path is **VPC network → VPC network peering → Set up connection → Allocated IP ranges → Allocate a new IP range**.

**Why teams start here:**
- The form guides you through the interdependent fields. Selecting "Internal" shows the network and purpose dropdowns.
- Immediate feedback. The reserved address appears in the IP addresses list within seconds.
- Good for one-off setups during initial VPC configuration.

**Why teams leave:**
- **No record of intent**: The console doesn't capture *why* a range was reserved. Three months later, nobody remembers if `10.100.0.0/20` is for Cloud SQL, Redis, or something else.
- **No review process**: Reserving a /16 range (65,536 IPs) is a significant commitment — it removes that space from your VPC's available addresses permanently. There's no approval gate in the console.
- **Discovery is manual**: Finding all reserved ranges across multiple projects requires clicking through each project's IP addresses page. There's no cross-project view.
- **Drift is invisible**: If someone releases a reservation or changes its description in the console, there's no notification or diff.

**Verdict**: Use the console for learning and initial exploration. For production address management — especially VPC peering ranges that affect managed service connectivity — move to code.

### Level 1: CLI Automation with gcloud

The `gcloud` CLI provides the `gcloud compute addresses` command group for managing global addresses:

```bash
# Reserve an external static IPv4 address
gcloud compute addresses create prod-lb-ip \
  --global \
  --ip-version=IPV4 \
  --description="Production HTTPS load balancer IP"

# Reserve an internal VPC peering range
gcloud compute addresses create google-managed-services \
  --global \
  --purpose=VPC_PEERING \
  --addresses=10.100.0.0 \
  --prefix-length=20 \
  --network=prod-vpc \
  --description="IP range for Cloud SQL and Redis private networking"

# Reserve a Private Service Connect address
gcloud compute addresses create psc-google-apis \
  --global \
  --purpose=PRIVATE_SERVICE_CONNECT \
  --addresses=10.10.0.100 \
  --network=prod-vpc \
  --description="PSC endpoint for Google APIs"

# List all global addresses
gcloud compute addresses list --global --project=my-project

# Describe a specific address
gcloud compute addresses describe prod-lb-ip --global
```

**Advantages over the console:**
- **Scriptable**: Wrap in bash scripts, run in CI/CD, version-control the scripts.
- **Discovery**: `gcloud compute addresses list --global` shows all reservations in a project. Scriptable across projects.
- **Automation**: Can be part of a project bootstrap script that reserves standard ranges.

**Limitations:**
- **Imperative**: The command says "create this address." If it already exists, you get an error. No built-in idempotency.
- **No state tracking**: The script doesn't know what it previously created. If a reservation is released externally, the script can't detect or restore it.
- **No cross-resource validation**: The script can't verify that the reserved range doesn't overlap with existing subnets or other reservations (GCP's API does this, but only at creation time).

**Verdict**: CLI scripts are a good step up for project bootstrap and discovery. But for ongoing management of address reservations — especially VPC peering ranges that other infrastructure depends on — state-tracking IaC is essential.

### Level 2: Infrastructure as Code with Terraform

Terraform treats global addresses as declarative resources with full lifecycle management:

```hcl
# External static IP for HTTPS load balancer
resource "google_compute_global_address" "lb_ip" {
  name         = "prod-lb-external-ip"
  project      = var.project_id
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
  description  = "Static IP for production HTTPS load balancer"
}

# Internal VPC peering range for Cloud SQL / Redis
resource "google_compute_global_address" "peering_range" {
  name          = "google-managed-services-range"
  project       = var.project_id
  address_type  = "INTERNAL"
  purpose       = "VPC_PEERING"
  prefix_length = 20
  network       = google_compute_network.vpc.id
  description   = "/20 range for Google managed services VPC peering"
}

# Private Service Connect address
resource "google_compute_global_address" "psc_address" {
  name         = "psc-google-apis"
  project      = var.project_id
  address_type = "INTERNAL"
  purpose      = "PRIVATE_SERVICE_CONNECT"
  address      = "10.10.0.100"
  network      = google_compute_network.vpc.id
  description  = "PSC endpoint for Google APIs"
}
```

**Why Terraform works well for global addresses:**
- **Declarative**: Describe the desired state. Terraform calculates the diff.
- **Plan/apply cycle**: Before reserving a /20 range, `terraform plan` shows exactly what will be created. Critical for ranges that permanently consume VPC address space.
- **Cross-resource references**: The `network` field references a VPC resource directly, ensuring the reservation is always in the correct network.
- **Drift detection**: `terraform plan` reveals if a reservation was released or modified outside of Terraform.
- **ForceNew visibility**: Since all fields are ForceNew, Terraform clearly shows when a change will destroy and recreate the address — which means the IP *will change*. This visibility prevents accidental IP changes.

**Considerations:**
- **State dependency**: The state file must be properly managed. Losing state for an address reservation means Terraform doesn't know the address exists, potentially leading to conflicts or orphaned resources.
- **No cross-field validation in HCL**: Terraform won't prevent you from setting `purpose: VPC_PEERING` on an EXTERNAL address. The error only surfaces during `apply`.
- **Labels support**: The `google_compute_global_address` resource supports labels for organization and cost attribution.

**Verdict**: Terraform is the production standard for address reservation management. Its plan/apply cycle provides the review gate that prevents accidental address changes and range over-allocation.

### Level 3: Infrastructure as Code with Pulumi

Pulumi uses general-purpose programming languages for infrastructure. A global address in Go (matching the OpenMCF implementation):

```go
// External static IP
lbIP, err := compute.NewGlobalAddress(ctx, "lb-ip", &compute.GlobalAddressArgs{
    Name:        pulumi.String("prod-lb-external-ip"),
    Project:     pulumi.String(projectID),
    AddressType: pulumi.StringPtr("EXTERNAL"),
    IpVersion:   pulumi.StringPtr("IPV4"),
    Description: pulumi.StringPtr("Static IP for production HTTPS load balancer"),
})

// Internal VPC peering range
peeringRange, err := compute.NewGlobalAddress(ctx, "peering-range", &compute.GlobalAddressArgs{
    Name:         pulumi.String("google-managed-services-range"),
    Project:      pulumi.String(projectID),
    AddressType:  pulumi.StringPtr("INTERNAL"),
    Purpose:      pulumi.StringPtr("VPC_PEERING"),
    PrefixLength: pulumi.IntPtr(20),
    Network:      vpc.SelfLink,
    Description:  pulumi.StringPtr("/20 range for Google managed services"),
})
```

**Strengths for global addresses:**
- **Type safety**: `compute.GlobalAddressArgs` catches invalid field names at compile time.
- **Programmatic range calculation**: Use code to calculate non-overlapping ranges across multiple VPCs:

```go
for i, vpc := range vpcs {
    _, err := compute.NewGlobalAddress(ctx, fmt.Sprintf("peering-%d", i), &compute.GlobalAddressArgs{
        Address:      pulumi.StringPtr(fmt.Sprintf("10.%d.0.0", 100+i)),
        PrefixLength: pulumi.IntPtr(20),
        // ...
    })
}
```

- **Preview workflow**: `pulumi preview` shows planned changes, equivalent to `terraform plan`.
- **Output chaining**: The reserved address output can be passed directly to forwarding rule or DNS record resources.

**Considerations:**
- **Language verbosity**: Pulumi's GCP SDK uses `pulumi.String()`, `pulumi.StringPtr()`, and typed args extensively. More verbose than Terraform HCL for simple cases.
- **Same ForceNew behavior**: Changes still destroy and recreate. Pulumi's preview makes this visible, but the behavior is inherent to the GCP API.

**Verdict**: Pulumi is an excellent choice for teams that need programmatic range management or are already using Go/TypeScript/Python for infrastructure.

## Comparative Analysis

| Criterion | Console | gcloud CLI | Terraform | Pulumi |
|---|---|---|---|---|
| **Repeatability** | None | Script-dependent | Full (declarative) | Full (declarative) |
| **Drift Detection** | None | None | On-demand (`plan`) | On-demand (`preview`) |
| **Review Gate** | None | Script review | Plan output in PR | Preview output in PR |
| **State Tracking** | None | None | State file | State backend |
| **Cross-Resource Refs** | Manual | Variable substitution | Native references | Native references |
| **Cross-Field Validation** | Console UI | API-time only | API-time only | API-time only |
| **Range Overlap Prevention** | API-time | API-time | API-time | API-time + code |
| **ForceNew Visibility** | None (just recreates) | None | Plan shows destroy+create | Preview shows replace |
| **Best For** | Learning, discovery | Bootstrap scripts | Production management | Programmatic patterns |

**Key takeaway**: For global addresses, the ForceNew behavior makes the plan/preview workflow critical. Any tool that doesn't show you that a change will destroy and recreate the address — causing the IP to change — is a risk in production.

## The OpenMCF Approach

### Why a Dedicated Component?

Global address reservations could theoretically be inlined into a load balancer or Cloud SQL component. OpenMCF separates them for three reasons:

1. **Independent lifecycle**: An address often outlives the resource it's attached to. Load balancers are recreated during upgrades; the IP stays. Cloud SQL instances are migrated; the peering range stays. Managing the address independently ensures it isn't accidentally destroyed.

2. **Cross-resource sharing**: A single VPC peering range serves Cloud SQL, Redis, AlloyDB, and Filestore simultaneously. It shouldn't be owned by any single consuming component.

3. **Address-first workflows**: In many organizations, network addresses are reserved by a platform team and consumed by application teams. The platform team reserves the range; the application team deploys Cloud SQL within it.

### Schema-Level Validation

OpenMCF's protobuf schema encodes GCP's cross-field constraints directly via CEL (Common Expression Language) validations:

```protobuf
// CEL: purpose can only be set when address_type is INTERNAL
option (buf.validate.message).cel = {
  id: "purpose_requires_internal"
  expression: "this.purpose == '' || this.address_type == 'INTERNAL'"
};

// CEL: prefix_length is required when purpose is VPC_PEERING
option (buf.validate.message).cel = {
  id: "vpc_peering_requires_prefix_length"
  expression: "this.purpose != 'VPC_PEERING' || has(this.prefix_length)"
};

// CEL: network is required when address_type is INTERNAL
option (buf.validate.message).cel = {
  id: "internal_requires_network"
  expression: "this.address_type != 'INTERNAL' || this.network.value != ''"
};
```

These validations catch the three most common misconfiguration errors *at authoring time*:

1. **Setting purpose on an EXTERNAL address**: GCP rejects this at API time with a confusing error. OpenMCF catches it during manifest validation.
2. **Forgetting prefix_length for VPC_PEERING**: Without this, GCP would attempt to reserve a single IP, which isn't valid for peering ranges. OpenMCF requires it explicitly.
3. **Omitting network for INTERNAL addresses**: Internal addresses must be scoped to a VPC network. OpenMCF enforces this before any API call.

This is a significant improvement over both Terraform and Pulumi, where these cross-field validations only surface during the apply/up step — after you've already committed and potentially run through CI.

### StringValueOrRef for Composition

The `project_id` and `network` fields use OpenMCF's `StringValueOrRef` pattern:

**Direct value** — provide a literal string when you know the value:

```yaml
spec:
  projectId:
    value: "my-prod-project-123"
  network:
    value: "projects/my-prod-project-123/global/networks/prod-vpc"
```

**Foreign key reference** — reference another OpenMCF resource's output:

```yaml
spec:
  projectId:
    ref:
      kind: GcpProject
      name: my-project
  network:
    ref:
      kind: GcpVpc
      name: prod-vpc
```

The `network` reference is particularly important for VPC peering ranges. With a foreign key reference, the global address is automatically linked to the correct VPC without hardcoding self-links, and the dependency graph ensures the VPC exists before the address is reserved.

### OpenMCF Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: vpc-peering-range
spec:
  projectId:
    value: my-project
  addressName: google-managed-services
  addressType: INTERNAL
  purpose: VPC_PEERING
  prefixLength: 20
  network:
    value: projects/my-project/global/networks/prod-vpc
  description: /20 range for Cloud SQL and Redis private networking
```

**Compared to Terraform:**

```hcl
resource "google_compute_global_address" "peering_range" {
  name          = "google-managed-services"
  project       = var.project_id
  address_type  = "INTERNAL"
  purpose       = "VPC_PEERING"
  prefix_length = 20
  network       = google_compute_network.vpc.id
  description   = "/20 range for Cloud SQL and Redis private networking"
}
```

The field count is similar — this isn't a resource where abstraction reduces verbosity. The value of OpenMCF for global addresses is:

1. **Schema-level validation** — catches cross-field errors before deployment.
2. **Composability** — StringValueOrRef enables foreign key references to projects and VPCs.
3. **Dual backend** — the same manifest works with both Pulumi and Terraform, so teams aren't locked to a single IaC tool.
4. **KRM-style familiarity** — teams using Kubernetes-style YAML for application deployment get the same pattern for infrastructure.

## 80/20 Scoping: What's Included and What's Excluded

### What's Included (and Why)

OpenMCF's `GcpGlobalAddressSpec` covers the fields that appear in real-world address reservations:

| Field | Why It's Included |
|---|---|
| `project_id` | Every address needs a project. StringValueOrRef enables composition. |
| `address_name` | GCP requires a unique name per project. RFC 1035 pattern-validated. |
| `address` | Optional specific IP. Critical for deterministic VPC peering ranges and PSC endpoints. |
| `address_type` | Fundamental distinction: EXTERNAL vs. INTERNAL. Drives all other field requirements. |
| `description` | Operational necessity. Address reservations without descriptions become mysteries. |
| `ip_version` | IPV4 or IPV6. Required for IPv6 load balancer addresses. |
| `network` | Required for INTERNAL addresses. StringValueOrRef enables VPC references. |
| `prefix_length` | Required for VPC_PEERING. Determines the size of the reserved range. |
| `purpose` | Distinguishes VPC_PEERING from PRIVATE_SERVICE_CONNECT. Drives validation rules. |

### What's Excluded (and Why)

| Excluded Feature | Why It's Not in the Spec |
|---|---|
| `network_tier` | Global addresses always use PREMIUM tier. The field exists in the API but is only meaningful for regional addresses. |
| `subnetwork` | Only applicable to regional internal addresses, not global ones. |
| `users` | Read-only field showing which resources reference this address. Not a configuration input. |
| `label_fingerprint` | Internal GCP concurrency control field. Managed automatically by the provider. |
| Private services access connection | The `google_service_networking_connection` resource that establishes the actual VPC peering after range reservation is a separate lifecycle concern. It connects the reserved range to `servicenetworking.googleapis.com`. OpenMCF may add a dedicated component for this in the future. |

The goal is to cover the three primary use cases (external static IP, VPC peering range, PSC address) with a focused schema, rather than exposing every API field including those irrelevant to global addresses.

## Production Best Practices

### External Address Management

**Reserve early, release never (almost)**:
- Reserve external static IPs during initial infrastructure setup, before configuring DNS.
- Once an IP is in DNS records, partner allowlists, or SSL certificate configurations, treat it as permanent infrastructure.
- Only release an address when you've confirmed all references have been updated and propagated.

**Use descriptions religiously**:
- Every external IP should have a description that answers: "What uses this IP?" and "Who manages it?"
- Good: `"Static IP for api.example.com HTTPS LB — managed by platform team"`
- Bad: `""` (empty) or `"lb ip"` (meaningless at scale)

**Consider IPv6 early**:
- Reserve both IPv4 and IPv6 addresses for public-facing services from the start.
- Adding IPv6 later requires DNS changes and potentially new forwarding rules.
- IPv6 addresses are free in GCP (same as IPv4 static addresses).

### VPC Peering Range Planning

**Size ranges appropriately**:

| Prefix Length | IP Count | When to Use |
|---|---|---|
| /24 | 256 | Small environments, single Cloud SQL instance |
| /20 | 4,096 | Standard environments, multiple managed services |
| /16 | 65,536 | Large environments, many managed service instances across services |

Google recommends `/20` as the minimum for most production environments. This accommodates growth in Cloud SQL instances, Redis clusters, and other peered services.

**Avoid overlapping ranges**:
- Plan your VPC address space to prevent conflicts between subnet ranges and peering ranges.
- Document allocated ranges in a central location (a spreadsheet, a wiki, or — ideally — in code).
- Use the `address` field to specify the range start when you need deterministic addressing.

**One range per VPC (usually sufficient)**:
- Multiple managed services (Cloud SQL, Redis, AlloyDB, Filestore) can share a single reserved range via VPC peering.
- Only create additional ranges if you need separate address spaces for different purposes or organizational boundaries.

**Network deletion lock**:
- A VPC network cannot be deleted while reserved internal ranges reference it. This is a GCP safety mechanism, not a bug.
- To delete a VPC, first release all reserved ranges (which requires deleting all managed service instances using them).

### Private Service Connect Addresses

**Use specific addresses for PSC**:
- Unlike VPC peering ranges (where auto-assignment is fine), PSC endpoints benefit from specific addresses because you'll configure DNS to resolve API hostnames to these IPs.
- Choose addresses from a well-documented internal range that doesn't conflict with other allocations.

**One address per PSC endpoint**:
- Each PSC endpoint needs its own reserved address. Don't try to reuse addresses across endpoints.

### ForceNew Awareness

All fields on `google_compute_global_address` (except labels) are ForceNew. This means:

- Changing `address_name`, `address_type`, `ip_version`, `network`, `prefix_length`, `purpose`, or `address` destroys the existing reservation and creates a new one.
- For EXTERNAL addresses, this means the IP **will change**. DNS records, allowlists, and all references must be updated.
- For VPC peering ranges, this means all managed services using the range will **lose connectivity** during the recreation.

**Mitigation:**
- Review all changes with `openmcf plan` or `terraform plan` before applying.
- For production external IPs, consider the address reservation as immutable. If you need a different configuration, create a new address alongside the old one, migrate references, then release the old one.
- For VPC peering ranges, plan migrations carefully. Create the new range, set up a parallel peering connection, migrate managed services, then release the old range.

## Security Considerations

### External Address Exposure

- Reserved external IPs are publicly routable. Even if no resource is attached, scanners will probe the IP.
- Ensure firewall rules are in place *before* attaching the IP to a load balancer or forwarding rule.
- Monitor unused reserved IPs. GCP charges for static IPs that are reserved but not attached to a running resource.

### Internal Range Isolation

- VPC peering ranges create a network bridge between your VPC and Google's service network. Traffic flows both ways.
- Ensure VPC firewall rules account for traffic from the peered range. Managed services (Cloud SQL, Redis) will have source IPs within the reserved range.
- Use Cloud VPC flow logs to monitor traffic patterns between your VPC and peered networks.

### Least Privilege for Address Management

- Reserve address management to platform/infra teams using IAM. `roles/compute.networkAdmin` grants address create/delete permissions.
- Application teams should have read-only access to address reservations (`roles/compute.networkViewer`).
- Use OpenMCF's resource references to decouple address creation from consumption: platform team creates the `GcpGlobalAddress`, application team references its outputs.

## Common Pitfalls

1. **Forgetting that all changes are ForceNew**: The most dangerous pitfall. Renaming an address, changing its type, or modifying the prefix length destroys and recreates it. Always run plan first.

2. **Under-sizing VPC peering ranges**: Starting with a /24 (256 IPs) works initially but runs out as you add Cloud SQL replicas, Redis instances, and other managed services. Start with /20 or larger.

3. **Orphaned external IPs**: GCP charges for reserved static IPs that aren't attached to a running resource. Periodically audit `gcloud compute addresses list --global` for addresses in `RESERVED` status.

4. **Missing the private services connection step**: Reserving a VPC peering range is only half the setup. You also need a `google_service_networking_connection` to actually establish the peering. Without it, the range exists but no managed service can use it.

5. **Overlapping address spaces**: When multiple teams reserve internal ranges independently, ranges can technically overlap with subnet allocations. GCP prevents exact conflicts at the API level, but adjacent or overlapping /20 ranges can create routing confusion.

6. **IPv6 on INTERNAL addresses**: IPv6 is only supported for EXTERNAL addresses. Setting `ip_version: IPV6` on an INTERNAL address produces an API error. OpenMCF's schema could catch this in a future version.

## Conclusion

GCP global address reservations are small resources with outsized impact. An external static IP is the permanent identity of your public-facing infrastructure. A VPC peering range is the foundation of all private connectivity to managed services. A PSC address is the entry point for private API access.

The progression from console clicks to codified infrastructure follows a clear path:

1. **Stop clicking**: Move every address reservation to code. Addresses created in the console are the ones nobody remembers or can explain.
2. **Validate early**: OpenMCF's schema-level CEL validations catch the three most common cross-field errors (purpose on EXTERNAL, missing prefix_length for VPC_PEERING, missing network for INTERNAL) before any API call is made.
3. **Plan changes carefully**: Every field is ForceNew. Use `openmcf plan` to see exactly what will be destroyed and recreated. For production external IPs and VPC peering ranges, treat address reservations as effectively immutable.
4. **Size ranges for growth**: Start VPC peering ranges at /20, not /24. Adding capacity later means creating a new range and migrating all dependent services.
5. **Compose through references**: Use StringValueOrRef to wire addresses to projects and VPCs through foreign key references rather than hardcoded values. This enables address reservations to be managed independently from the resources that consume them.

OpenMCF's `GcpGlobalAddress` component captures these principles in a focused API that covers the three real-world use cases — external static IPs, VPC peering ranges, and PSC endpoints — with schema-level validation that catches the mistakes that cause the most production impact. It generates both Pulumi and Terraform, so teams aren't locked to a single IaC tool, and its KRM-style manifests integrate naturally with Kubernetes-native workflows.

For further reading:
- [Reserving a Static External IP Address (Google Cloud Documentation)](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address)
- [Configuring Private Services Access](https://cloud.google.com/vpc/docs/configure-private-services-access)
- [Private Service Connect Overview](https://cloud.google.com/vpc/docs/private-service-connect)
- [Global Addresses REST API Reference](https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses)
- [VPC Network Peering](https://cloud.google.com/vpc/docs/vpc-peering)
- [IP Address Pricing](https://cloud.google.com/vpc/network-pricing#ipaddress)
