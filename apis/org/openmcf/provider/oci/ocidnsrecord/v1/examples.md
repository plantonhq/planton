# OciDnsRecord Examples

## Single A Record

An A record pointing a subdomain to a single IP address:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: app-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDnsRecord.app-a-record
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "app.example.com"
  rtype: "A"
  items:
    - rdata: "192.0.2.1"
      ttl: 300
```

## Round-Robin A Records with Zone Reference

Multiple A records for load distribution, referencing the zone via `valueFrom`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: web-a-records
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsRecord.web-a-records
spec:
  zoneNameOrId:
    valueFrom:
      kind: OciDnsZone
      name: prod-zone
      fieldPath: status.outputs.zoneId
  domain: "web.example.com"
  rtype: "A"
  items:
    - rdata: "192.0.2.1"
      ttl: 300
    - rdata: "192.0.2.2"
      ttl: 300
    - rdata: "192.0.2.3"
      ttl: 300
```

## MX Records for Email

Mail exchange records with priority values:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: mail-mx-records
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsRecord.mail-mx-records
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "example.com"
  rtype: "MX"
  items:
    - rdata: "10 mail1.example.com."
      ttl: 3600
    - rdata: "20 mail2.example.com."
      ttl: 3600
```

## CNAME Alias

A CNAME record pointing a subdomain to a load balancer hostname:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: api-cname
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsRecord.api-cname
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "api.example.com"
  rtype: "CNAME"
  items:
    - rdata: "lb.example.com."
      ttl: 300
```

## TXT Record for SPF

A TXT record for email sender policy framework:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDnsRecord
metadata:
  name: spf-txt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDnsRecord.spf-txt
spec:
  zoneNameOrId:
    value: "example.com"
  domain: "example.com"
  rtype: "TXT"
  items:
    - rdata: "\"v=spf1 include:_spf.google.com ~all\""
      ttl: 3600
```

## Common Operations

### Update record values

Modify the `items` list (change rdata or TTL) and re-apply. The entire record set is replaced atomically.

### Add records to an existing set

Append new entries to the `items` list and re-apply. Existing records are preserved alongside new ones.

### Remove a record

Remove the entry from the `items` list and re-apply. The remaining records replace the old set.

## Best Practices

1. **Use zone OCIDs via `valueFrom`** — referencing by OCID avoids ambiguity when multiple zones share the same name across compartments.
2. **Set reasonable TTLs** — 300s (5 min) for records that may change, 3600s (1 hour) for stable records. OCI does not recommend values below 30s.
3. **Include trailing dots on hostnames** — CNAME, MX, NS, and SRV rdata should end with a period (e.g., `mail.example.com.`) to indicate a fully qualified domain name.
4. **One manifest per (domain, rtype) tuple** — each OciDnsRecord manages exactly one record set. Multiple manifests for the same tuple will conflict.
5. **Quote TXT record data** — TXT rdata must be enclosed in double quotes within the YAML string (e.g., `"\"v=spf1 ...\""`).
