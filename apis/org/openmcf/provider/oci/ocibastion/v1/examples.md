# OciBastion Examples

## Minimal Bastion

A bastion with default settings for development access:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBastion
metadata:
  name: dev-bastion
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciBastion.dev-bastion
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  targetSubnetId:
    value: "ocid1.subnet.oc1..example"
```

## Corporate Network Restriction

A bastion that only allows connections from two corporate CIDR ranges, with an 8-hour session limit. References compartment and subnet via `valueFrom`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBastion
metadata:
  name: corp-bastion
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBastion.corp-bastion
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  targetSubnetId:
    valueFrom:
      kind: OciSubnet
      name: private-subnet
      fieldPath: status.outputs.subnetId
  displayName: "corp-bastion-prod"
  clientCidrBlockAllowList:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
  maxSessionTtlInSeconds: 28800
```

## DNS Proxy Enabled

A bastion with DNS proxy for FQDN-based session targets and SOCKS5 dynamic port forwarding:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBastion
metadata:
  name: dns-bastion
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBastion.dns-bastion
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  targetSubnetId:
    value: "ocid1.subnet.oc1..example"
  clientCidrBlockAllowList:
    - "10.0.0.0/16"
  maxSessionTtlInSeconds: 14400
  isDnsProxyEnabled: true
```

## Common Operations

### Create a managed SSH session

After deploying the bastion, create a session using the OCI CLI:

```shell
oci bastion session create-managed-ssh \
  --bastion-id <bastion_id from stack outputs> \
  --target-resource-id <compute instance OCID> \
  --target-os-username opc \
  --key-type PUB \
  --ssh-public-key-file ~/.ssh/id_rsa.pub
```

### Update allowed CIDR ranges

Modify the `clientCidrBlockAllowList` in the manifest and re-apply. This field is updatable without recreation.

### Extend the maximum session TTL

Update `maxSessionTtlInSeconds` and re-apply. Existing sessions are not affected; only new sessions use the updated limit.

## Best Practices

1. **Restrict client CIDRs** — always set `clientCidrBlockAllowList` in production to limit access to known network ranges.
2. **Use the shortest practical session TTL** — reduces the window of exposure if a session is left open.
3. **Enable DNS proxy only when needed** — adds attack surface; enable only if sessions must target hosts by FQDN.
4. **Use `valueFrom` references** for `targetSubnetId` — avoids hardcoding OCIDs and ensures the subnet exists before the bastion.
5. **Do not manage sessions via IaC** — sessions are ephemeral (minutes to hours) and are better managed via CLI or automation scripts.
