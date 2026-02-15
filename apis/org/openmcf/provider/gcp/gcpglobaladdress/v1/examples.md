# GCP Global Address Examples

This document provides comprehensive examples for reserving GCP global addresses using OpenMCF. Each example includes the manifest YAML and explains the use case and key configuration choices.

## Table of Contents

- [Example 1: Minimal External Static IP](#example-1-minimal-external-static-ip)
- [Example 2: External IPv6 Address](#example-2-external-ipv6-address)
- [Example 3: Internal VPC Peering Range (Cloud SQL / Redis)](#example-3-internal-vpc-peering-range-cloud-sql--redis)
- [Example 4: Private Service Connect Address](#example-4-private-service-connect-address)
- [Example 5: Full Configuration with All Fields](#example-5-full-configuration-with-all-fields)
- [Example 6: Using Project and Network References](#example-6-using-project-and-network-references)
- [Presets](#presets)

---

## Example 1: Minimal External Static IP

### Use Case

Reserve a static external IPv4 address for an HTTP(S) load balancer. This is the most common use case — you need a stable public IP that persists across load balancer updates and can be referenced in DNS records, allowlists, and firewall rules.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: lb-static-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.lb-static-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: prod-lb-external-ip
  description: Static IP for production HTTPS load balancer
```

Deploy:

```shell
openmcf apply -f lb-static-ip.yaml
```

### Key Choices

- **No `addressType` or `ipVersion`** — defaults to `EXTERNAL` and `IPV4`, which is correct for a standard load balancer IP.
- **No `address`** — GCP assigns an available IP automatically. Specify this only if you need to reserve a specific IP (e.g., migrating from another provider).
- **`description`** — always include a description so the address's purpose is clear when viewed in the console or via `gcloud`.

---

## Example 2: External IPv6 Address

### Use Case

Reserve a static external IPv6 address for dual-stack load balancing. IPv6 global addresses enable your load balancer to serve traffic from IPv6 clients without NAT64 translation.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: lb-ipv6
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.lb-ipv6
spec:
  projectId:
    value: my-gcp-project-123
  addressName: prod-lb-ipv6
  addressType: EXTERNAL
  ipVersion: IPV6
  description: IPv6 static IP for dual-stack HTTPS load balancer
```

Deploy:

```shell
openmcf apply -f lb-ipv6.yaml
```

### Key Choices

- **`ipVersion: IPV6`** — explicitly requests an IPv6 address. IPv6 global addresses are only available for EXTERNAL addresses with premium network tier.
- **Pair with an IPv4 address** — for full dual-stack, deploy both an IPv4 (Example 1) and an IPv6 address, attaching each to separate forwarding rules on the same load balancer.

---

## Example 3: Internal VPC Peering Range (Cloud SQL / Redis)

### Use Case

Reserve an internal IP range for VPC network peering. Google managed services (Cloud SQL, Memorystore Redis, AlloyDB, Filestore) require a reserved private IP range to establish private connectivity via VPC peering. This is the standard prerequisite for enabling private IP on any of these managed services.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: vpc-peering-range
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.vpc-peering-range
spec:
  projectId:
    value: my-gcp-project-123
  addressName: google-managed-services-range
  addressType: INTERNAL
  purpose: VPC_PEERING
  prefixLength: 20
  network:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
  description: /20 IP range for Google managed services (Cloud SQL, Redis, AlloyDB)
```

Deploy:

```shell
openmcf apply -f vpc-peering-range.yaml
```

### Key Choices

- **`addressType: INTERNAL`** — this is a private IP range, not a public IP.
- **`purpose: VPC_PEERING`** — tells GCP this range is for VPC peering with Google-managed services.
- **`prefixLength: 20`** — reserves a /20 range (4,096 IP addresses). This is Google's recommended minimum for Cloud SQL. Use `/16` for large environments with many managed service instances.
- **`network`** — the VPC network self-link. The reserved range is scoped to this network. The network cannot be deleted while this address exists.
- **No `address`** — let GCP pick an available range from the VPC's RFC 1918 space. Specify only if you need a deterministic starting address (e.g., `10.100.0.0`).

### Follow-Up: Private Service Networking Connection

After reserving the range, you must create a private services access connection to actually establish the VPC peering. This is typically handled by a separate OpenMCF component or directly via:

```bash
gcloud services vpc-peerings connect \
  --service=servicenetworking.googleapis.com \
  --ranges=google-managed-services-range \
  --network=prod-vpc \
  --project=my-gcp-project-123
```

---

## Example 4: Private Service Connect Address

### Use Case

Reserve a single internal IP address for a Private Service Connect (PSC) endpoint. PSC enables private connectivity to Google APIs or published services without leaving the VPC. The reserved IP becomes the entry point for traffic to the PSC endpoint.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: psc-endpoint-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.psc-endpoint-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: psc-google-apis
  addressType: INTERNAL
  purpose: PRIVATE_SERVICE_CONNECT
  address: "10.10.0.100"
  network:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
  description: PSC endpoint for Google APIs
```

Deploy:

```shell
openmcf apply -f psc-endpoint-ip.yaml
```

### Key Choices

- **`purpose: PRIVATE_SERVICE_CONNECT`** — designates this address for a PSC endpoint.
- **`address: "10.10.0.100"`** — specifies an exact IP within the VPC's CIDR range. For PSC, using a specific address is common so that DNS records and firewall rules can reference a known IP.
- **No `prefixLength`** — PSC endpoints use a single IP, not a CIDR range.
- **`network`** — required because this is an INTERNAL address.

---

## Example 5: Full Configuration with All Fields

### Use Case

Demonstrates every available field in a single manifest. This is useful as a reference template — copy it and remove the fields you don't need.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: full-example
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.full-example
spec:
  # GCP project — direct value or reference
  projectId:
    value: my-gcp-project-123

  # Resource name in GCP (RFC 1035: lowercase, hyphens, 1-63 chars)
  addressName: full-example-peering-range

  # Specific starting IP for the range (optional — omit to auto-assign)
  address: "10.200.0.0"

  # EXTERNAL for public IPs, INTERNAL for private ranges
  addressType: INTERNAL

  # Human-readable description
  description: "Full example: /24 internal range for VPC peering"

  # IPV4 or IPV6 (IPV6 only valid for EXTERNAL)
  ipVersion: IPV4

  # VPC network — required for INTERNAL addresses
  network:
    value: projects/my-gcp-project-123/global/networks/prod-vpc

  # CIDR prefix length (8-29) — required for VPC_PEERING
  prefixLength: 24

  # VPC_PEERING, PRIVATE_SERVICE_CONNECT, or empty — only for INTERNAL
  purpose: VPC_PEERING
```

Deploy:

```shell
openmcf apply -f full-example.yaml
```

### Key Choices

- **All fields populated** — this manifest exercises every field for reference purposes. In practice, most deployments use a subset.
- **`address: "10.200.0.0"` with `prefixLength: 24`** — reserves the exact range `10.200.0.0/24` (256 IPs). This provides deterministic addressing, useful when you need to set up routes or firewall rules for the peered range.

---

## Example 6: Using Project and Network References

### Use Case

Reference other OpenMCF resources instead of hardcoding project IDs and network self-links. This enables composable infrastructure where resources are wired together through foreign key references.

### Manifest

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: peering-range-referenced
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.peering-range-referenced
spec:
  projectId:
    ref:
      kind: GcpProject
      name: my-gcp-project
      fieldPath: status.outputs.project_id
  addressName: managed-services-range
  addressType: INTERNAL
  purpose: VPC_PEERING
  prefixLength: 20
  network:
    ref:
      kind: GcpVpc
      name: prod-vpc
      fieldPath: status.outputs.network_self_link
  description: VPC peering range using resource references
```

Deploy:

```shell
openmcf apply -f peering-range-referenced.yaml
```

### Key Choices

- **`projectId.ref`** — resolves the project ID from a `GcpProject` resource's outputs. The global address is automatically created in the correct project without hardcoding.
- **`network.ref`** — resolves the VPC network self-link from a `GcpVpc` resource's outputs. This creates an implicit dependency: the VPC must exist before the address can be reserved.
- **`fieldPath`** — specifies the exact output field to use. These are the default paths, so you can omit `fieldPath` if using the defaults.

---

## Presets

OpenMCF presets provide pre-configured templates for common patterns. Use them as starting points and customize as needed.

### Preset: external-static-ip

Standard external static IPv4 address for load balancers and CDN.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: my-lb-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: my-lb-external-ip
  addressType: EXTERNAL
  ipVersion: IPV4
  description: Static IP for HTTPS load balancer
```

### Preset: internal-vpc-peering-range

Internal /20 range for VPC peering with Google managed services.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: my-peering-range
spec:
  projectId:
    value: my-gcp-project-123
  addressName: google-managed-svc-range
  addressType: INTERNAL
  purpose: VPC_PEERING
  prefixLength: 20
  network:
    value: projects/my-gcp-project-123/global/networks/my-vpc
  description: IP range for Google managed services VPC peering
```

### Preset: private-service-connect

Internal IP for a Private Service Connect endpoint.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: my-psc-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: psc-endpoint
  addressType: INTERNAL
  purpose: PRIVATE_SERVICE_CONNECT
  network:
    value: projects/my-gcp-project-123/global/networks/my-vpc
  description: Private Service Connect endpoint address
```

---

## Deployment Commands

### Using the OpenMCF CLI

```bash
openmcf apply -f <filename>.yaml
```

### Using Pulumi Directly

```bash
cd iac/pulumi
pulumi stack init dev
pulumi up
```

### Using Terraform Directly

```bash
cd iac/tf
terraform init
terraform plan
terraform apply
```

### Validate Manifest

```bash
openmcf validate --manifest <filename>.yaml
```

### Destroy Address

```bash
openmcf pulumi destroy --manifest <filename>.yaml
```

---

## Additional Resources

- [Reserving a Static External IP Address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address)
- [Configuring Private Services Access](https://cloud.google.com/vpc/docs/configure-private-services-access)
- [Private Service Connect Overview](https://cloud.google.com/vpc/docs/private-service-connect)
- [Global Addresses REST API](https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses)

---

For more details, see the [main README](README.md).
