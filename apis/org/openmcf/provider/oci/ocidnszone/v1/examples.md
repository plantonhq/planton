# OciDnsZone Examples

## Public Primary Zone

A standard public DNS zone for hosting domain records:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsZone
metadata:
  name: example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsZone.example-com
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  zoneType: primary
```

## DNSSEC-Signed Zone

A public zone with DNSSEC enabled for DNS response authentication:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsZone
metadata:
  name: secure.example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsZone.secure-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  isDnssecEnabled: true
```

## Private Zone for Internal Resolution

A private DNS zone resolvable only within VCNs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsZone
metadata:
  name: internal.example.local
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsZone.internal-example-local
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  scope: private
  viewId:
    value: "ocid1.dnsview.oc1..example"
```

## Secondary Zone from On-Premises Masters

A secondary zone replicating from two on-premises DNS servers with TSIG authentication:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsZone
metadata:
  name: corp.example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsZone.corp-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: secondary
  externalMasters:
    - address: "198.51.100.1"
      tsigKeyId: "ocid1.tsigkey.oc1..example"
    - address: "198.51.100.2"
      tsigKeyId: "ocid1.tsigkey.oc1..example"
```

## Primary Zone with External Downstreams

A primary zone pushing zone transfers to external DNS servers for multi-site distribution:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsZone
metadata:
  name: distributed.example.com
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsZone.distributed-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  externalDownstreams:
    - address: "203.0.113.1"
    - address: "203.0.113.2"
```

## Common Operations

### Enable DNSSEC on an existing zone

Set `isDnssecEnabled: true` and re-apply. OCI generates KSK and ZSK key pairs and begins signing records. You must then add DS records at your registrar to complete the DNSSEC chain of trust.

### Update external masters

Modify the `externalMasters` list and re-apply. The secondary zone updates its replication sources without recreation.

### Move zone to a different compartment

Update `compartmentId` and re-apply. OCI supports in-place compartment moves for DNS zones.

### Configure nameservers at your registrar

After deployment, read the `nameservers` stack output and set these as NS records at your domain registrar. The output is a comma-separated list of OCI nameserver hostnames.

## Best Practices

1. **Configure registrar NS records immediately** — until your registrar points to OCI nameservers, the zone is not resolving.
2. **Enable DNSSEC for public zones** — protects against DNS spoofing and cache poisoning.
3. **Use TSIG keys for zone transfers** — authenticates transfer requests between OCI and external DNS servers.
4. **Use `valueFrom` references** for `compartmentId` — avoids hardcoding OCIDs and maintains dependency ordering in infra charts.
5. **Choose zone type carefully** — `zoneType` is ForceNew; switching between PRIMARY and SECONDARY requires zone recreation.
