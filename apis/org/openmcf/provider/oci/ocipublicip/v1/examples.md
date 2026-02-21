# OCI Public IP — Examples

Complete YAML manifests for common OciPublicIp configurations. Each example is self-contained and can be deployed with `openmcf apply -f <file>`.

## Table of Contents

- [Reserved Unassigned IP](#reserved-unassigned-ip)
- [Reserved IP Assigned to a Private IP](#reserved-ip-assigned-to-a-private-ip)
- [Reserved IP from a BYOIP Pool](#reserved-ip-from-a-byoip-pool)
- [Ephemeral IP for a Compute Instance](#ephemeral-ip-for-a-compute-instance)
- [Using Foreign Key References](#using-foreign-key-references)

---

## Reserved Unassigned IP

Pre-allocate a stable public IP address without assigning it to any resource. The IP can be assigned later via the OCI Console, API, or by updating the manifest with `privateIpId`.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: reserved-unassigned
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPublicIp.reserved-unassigned
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
```

**When to use**: DNS record pre-provisioning, firewall allowlist preparation, or reserving an address before the target instance exists.

---

## Reserved IP Assigned to a Private IP

Assign a reserved public IP to a specific private IP at creation time. The IP persists independently of the instance — if the instance is terminated and recreated, the same public IP can be reassigned to the new private IP.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: web-server-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciPublicIp.web-server-ip
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
  displayName: "Web Server Public IP"
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

**When to use**: Production instances that need a stable internet-facing address for DNS A records, TLS certificates, or partner integrations.

---

## Reserved IP from a BYOIP Pool

Allocate a public IP from a customer-owned IP pool instead of Oracle's default pool. The `publicIpPoolId` field is immutable after creation.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: byoip-address
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPublicIp.byoip-address
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
  displayName: "Corporate BYOIP Address"
  publicIpPoolId:
    value: "ocid1.publicippool.oc1.iad.example"
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

**When to use**: Organizations that have imported their own IP address ranges into OCI and need to allocate from those ranges for regulatory or reputation reasons.

---

## Ephemeral IP for a Compute Instance

Create an ephemeral public IP tied to a specific private IP. The IP is automatically released when the associated instance or VNIC is terminated. The `privateIpId` field is required for ephemeral IPs.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: dev-instance-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPublicIp.dev-instance-ip
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: EPHEMERAL
  displayName: "Dev Instance Ephemeral IP"
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

**When to use**: Development or test instances that need temporary internet access and don't require a stable address.

---

## Using Foreign Key References

Reference an OpenMCF-managed compartment via `valueFrom` instead of hardcoding the OCID. This creates a dependency graph so the compartment is provisioned before the public IP.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: ref-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPublicIp.ref-ip
  env: prod
  org: acme
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  lifetime: RESERVED
  displayName: "Referenced Compartment IP"
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

**When to use**: Multi-resource deployments where the compartment is also managed by OpenMCF and you want to avoid hardcoded OCIDs.

---

## Common Operations

### Check the allocated IP address

After deployment, inspect the stack outputs:

```shell
openmcf get OciPublicIp my-public-ip -o yaml
```

The `status.outputs.ipAddress` field contains the allocated IPv4 address.

### Reassign a reserved IP

Update the `privateIpId` field in the manifest and re-apply:

```shell
openmcf apply -f public-ip.yaml
```

The reserved IP moves to the new private IP. Ephemeral IPs cannot be reassigned — they must be destroyed and recreated.

---

## Best Practices

1. **Use reserved IPs for production workloads** — ephemeral IPs are released on instance termination, which can break DNS records and firewall rules.
2. **Pre-allocate IPs before instance creation** — create a reserved unassigned IP first, configure DNS and firewall rules, then assign it when the instance is ready.
3. **Set `displayName` explicitly in shared environments** — this makes it easier to identify IPs in the OCI Console when multiple teams share a compartment.
4. **Use `valueFrom` for compartment references** — avoids hardcoded OCIDs and creates explicit dependency ordering between resources.
5. **Tag with `env` and `org` metadata** — the component propagates these as freeform tags for cost tracking and governance.
