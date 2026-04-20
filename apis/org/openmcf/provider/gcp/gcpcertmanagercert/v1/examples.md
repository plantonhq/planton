# GCP Cert Manager Cert Examples

## Create Using CLI

Create a YAML manifest using one of the examples below. After the YAML is created, apply it with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

(You can also use the shorter form `planton apply -f <yaml-path>` if your environment is configured accordingly.)

---

## Basic Example (Single Domain - Certificate Manager)

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: single-cert-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: example.com
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: MANAGED
```

This manifest requests a Certificate Manager certificate for `example.com` and uses Cloud DNS zone `example-com-zone` for DNS validation.

---

## Multiple Domains Example (SANs)

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: multi-cert-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: myapp.io
  alternateDomainNames:
    - api.myapp.io
    - blog.myapp.io
    - www.myapp.io
  cloudDnsZoneId:
    value: myapp-io-zone
  certificateType: MANAGED
```

In this example, we issue a certificate for `myapp.io` as the primary domain and add three alternate domains (SANs): `api.myapp.io`, `blog.myapp.io`, and `www.myapp.io`. It automatically creates DNS records in the Cloud DNS zone `myapp-io-zone`.

---

## Wildcard Domain Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: wildcard-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: '*.wildexample.com'
  cloudDnsZoneId:
    value: wildexample-com-zone
  certificateType: MANAGED
```

This manifest requests a wildcard certificate for `*.wildexample.com`, allowing you to secure any subdomain under `wildexample.com` without creating separate certificates.

---

## Load Balancer Certificate Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: lb-cert-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: lb.example.com
  alternateDomainNames:
    - www.lb.example.com
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: LOAD_BALANCER
```

This example creates a Google-managed SSL certificate optimized for load balancers. Use this when you're attaching the certificate directly to a GCP load balancer.

---

## Multi-Domain Wildcard Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: multi-wildcard-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: '*.services.example.com'
  alternateDomainNames:
    - services.example.com
    - '*.api.example.com'
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: MANAGED
```

This advanced example combines wildcard and regular domains, covering `*.services.example.com`, `services.example.com`, and `*.api.example.com` in a single certificate.

---

## Development Environment Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: dev-cert
  org: my-org
  env:
    id: development
spec:
  gcpProjectId: my-gcp-dev-project
  primaryDomainName: dev.example.com
  alternateDomainNames:
    - '*.dev.example.com'
  cloudDnsZoneId:
    value: dev-example-com-zone
  certificateType: MANAGED
```

This example shows a typical development environment certificate setup with a dev subdomain and wildcard for all dev services.

---

## Using Foreign Key Reference for DNS Zone

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: ref-cert-example
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: example.com
  cloudDnsZoneId:
    kind: GcpDnsZone
    name: my-dns-zone-resource
    fieldPath: status.outputs.zone_name
  certificateType: MANAGED
```

This example demonstrates using a foreign key reference to another OpenMCF resource (`GcpDnsZone`) instead of hardcoding the zone name.

---

## Minimal Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: minimal-cert
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: barebones.io
  cloudDnsZoneId:
    value: barebones-io-zone
```

Here, we only specify the required fields. The certificate will be created using Certificate Manager (default) with DNS validation for `barebones.io`. Alternate domain names are not specified.

---

## External DNS (Manual Validation) Example

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: external-dns-cert
  org: my-org
  env:
    id: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: example.com
  alternateDomainNames:
    - www.example.com
  certificateType: MANAGED
```

When `cloudDnsZoneId` is omitted, the module creates DNS authorizations in GCP but does **not** insert validation records into any DNS zone. Instead, the required CNAME records are exported as the `dns-validation-records` stack output. After deployment, read these records and manually create them in your DNS provider (e.g. AWS Route 53, Cloudflare). The certificate will provision once the records propagate globally.

```shell
# Read the validation records after deployment
pulumi stack output dns-validation-records
```

---

## After Deploying

Once you've chosen one of these examples, apply it with OpenMCF using Pulumi or Terraform:

```shell
openmcf pulumi up --manifest gcp-cert-manager-cert.yaml --stack myorg/production
```

or

```shell
openmcf terraform apply --manifest gcp-cert-manager-cert.yaml --stack myorg/production
```

When the command completes successfully, your SSL certificate will be provisioned. You can confirm by checking the GCP console:

- **Certificate Manager**: Navigate to Certificate Manager in the GCP Console
- **Load Balancer Certificates**: Check the Load Balancing section

You should see your new certificate and any DNS validation records created within Cloud DNS.

---

## Verification

To verify your certificate is active:

1. **Check Certificate Status**: In GCP Console, verify the certificate shows as "Active" or "Provisioned"
2. **DNS Records**: Confirm validation records were created in Cloud DNS
3. **Domain Validation**: Ensure domains are validated successfully

---

Enjoy secure connections with OpenMCF!

