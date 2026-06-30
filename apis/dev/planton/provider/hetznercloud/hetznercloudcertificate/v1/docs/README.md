# HetznerCloud Certificate — Research Documentation

## Introduction

TLS certificates are the mechanism that load balancers use to terminate HTTPS connections. Without a certificate, a load balancer can only handle unencrypted HTTP traffic. In Hetzner Cloud, certificates are first-class API objects stored in the certificate store and referenced by load balancer HTTPS services via their numeric ID.

Hetzner Cloud supports two fundamentally different certificate types:

- **Uploaded certificates**: The operator provides a PEM-encoded certificate chain and private key. Hetzner Cloud stores them. The operator is responsible for obtaining, renewing, and rotating the certificate.
- **Managed certificates**: The operator provides a list of domain names. Hetzner Cloud obtains a Let's Encrypt certificate automatically and handles renewal. The operator is responsible for ensuring DNS records point to the load balancer so the ACME HTTP-01 challenge succeeds.

The `HetznerCloudCertificate` component unifies both types into a single Planton resource with a proto `oneof` that enforces mutual exclusivity at the schema level. The IaC modules route to the correct provider resource (`hcloud_uploaded_certificate` or `hcloud_managed_certificate`) based on which variant is set.

## Certificates in the Hetzner Cloud Ecosystem

### Uploaded vs Managed: Feature Comparison

| Aspect | Uploaded | Managed |
|--------|----------|---------|
| **Certificate source** | User-provided PEM files | Let's Encrypt (automatic) |
| **Renewal** | Manual — operator must replace before expiry | Automatic — Hetzner Cloud renews ~30 days before expiry |
| **Wildcard support** | Yes (if your CA issued one) | No — Let's Encrypt HTTP-01 does not support wildcards |
| **EV/OV certificates** | Yes | No — Let's Encrypt only issues DV certificates |
| **Prerequisites** | Certificate + private key PEM files | DNS records pointing to LB + HTTPS service on LB |
| **Provisioning time** | Immediate (certificate is stored as-is) | Minutes (ACME challenge + issuance) |
| **Immutable fields** | `certificate`, `privateKey` (ForceNew) | `domainNames` (ForceNew) |
| **Mutable fields** | `name`, `labels` | `name`, `labels` |
| **Cost** | Free (bundled with LB) | Free (bundled with LB) |

### The ACME HTTP-01 Challenge

Managed certificates use the ACME HTTP-01 challenge to prove domain ownership. The sequence:

1. Hetzner Cloud requests a certificate from Let's Encrypt for the specified domains.
2. Let's Encrypt responds with a challenge token for each domain.
3. Hetzner Cloud places the token at `http://<domain>/.well-known/acme-challenge/<token>` via the load balancer.
4. Let's Encrypt makes an HTTP request to each domain to verify the token.
5. If all challenges succeed, Let's Encrypt issues the certificate.

This means:

- Every domain must have a DNS A or AAAA record pointing to the load balancer's IP.
- The load balancer must have an HTTPS service configured that references this certificate.
- Port 80 must be reachable (Let's Encrypt uses HTTP, not HTTPS, for the challenge).

If these prerequisites are not met, the certificate stays in a `pending` state. There is no timeout — the certificate will complete once the prerequisites are satisfied.

### Renewal Behavior

**Managed certificates** are renewed automatically by Hetzner Cloud approximately 30 days before expiry. The renewal uses the same ACME HTTP-01 mechanism. If the DNS records have been removed or the load balancer no longer exists at renewal time, the renewal fails silently — the certificate expires without warning. Monitor the `not_valid_after` output to detect upcoming expirations.

**Uploaded certificates** have no automatic renewal. The operator must:
1. Obtain a new certificate from their CA.
2. Update the manifest with the new `certificate` and `privateKey` values.
3. Re-deploy. Since both fields are ForceNew, this destroys the old certificate and creates a new one with a new `certificate_id`.
4. Any load balancers referencing the old certificate ID must be updated to reference the new ID.

In an Planton workflow, load balancers that use `valueFrom` to reference the certificate will automatically pick up the new ID on their next deployment.

### Pricing

Certificates are free in Hetzner Cloud. There is no per-certificate charge and no charge for managed certificate issuance or renewal. The cost is bundled into the load balancer pricing.

This contrasts with some other providers (e.g., AWS ACM is free, but Azure App Service Certificates have per-certificate pricing). In Hetzner Cloud, cost is never a factor in certificate type selection.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

#### Uploaded Certificate

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Security** > **Certificates** in the left sidebar
3. Click **Create Certificate**
4. Select **Upload Certificate**
5. Enter a name
6. Paste the PEM certificate chain into the certificate field
7. Paste the PEM private key into the private key field
8. Add labels (optional)
9. Click **Create Certificate**

The certificate is available immediately after creation.

#### Managed Certificate

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Security** > **Certificates**
3. Click **Create Certificate**
4. Select **Managed Certificate**
5. Enter a name
6. Enter domain names (one per line)
7. Add labels (optional)
8. Click **Create Certificate**

The certificate enters a pending state. Once DNS records and a load balancer HTTPS service are in place, the ACME challenge completes and the certificate becomes active.

**Pros:**
- Visual confirmation of certificate status and validity dates
- Easy to browse all certificates in one place
- Immediate feedback on uploaded certificate parsing errors

**Cons:**
- No version control for certificate configurations
- Private keys must be pasted into a web form (security risk in shared environments)
- No way to enforce naming or labeling standards
- No audit trail for who created or modified certificates
- Cannot reproduce certificate configurations across environments

**Verdict:** Acceptable for initial exploration. Not viable for production workflows where certificates need to be tracked, rotated, and reproduced.

### Level 1: CLI (`hcloud`)

#### Uploaded Certificate

```bash
# Create an uploaded certificate from PEM files
hcloud certificate create \
  --name wildcard-cert \
  --type uploaded \
  --cert-file /path/to/fullchain.pem \
  --key-file /path/to/privkey.pem

# Verify
hcloud certificate describe wildcard-cert
```

#### Managed Certificate

```bash
# Create a managed certificate
hcloud certificate create \
  --name api-cert \
  --type managed \
  --domain example.com \
  --domain www.example.com

# Check status (managed certs may take a few minutes)
hcloud certificate describe api-cert
```

#### Common Operations

```bash
# List all certificates
hcloud certificate list

# Update name or labels
hcloud certificate update --name new-name 12345
hcloud certificate add-label 12345 env=production

# Delete
hcloud certificate delete 12345
```

**Key CLI behaviors:**
- The `--type` flag selects `uploaded` or `managed`. It defaults to `uploaded` if `--cert-file` and `--key-file` are provided.
- For managed certificates, the `--domain` flag is repeatable. Each domain is added to the SAN list.
- The CLI reads PEM files from disk (`--cert-file`, `--key-file`) rather than accepting inline PEM strings. This is safer than pasting into a terminal.
- Managed certificate creation returns immediately, but the certificate may not be usable until the ACME challenge completes.

**Pros:**
- Scriptable, can be embedded in CI/CD
- PEM files read from disk (no pasting into terminals)
- Single command per operation

**Cons:**
- No state tracking — cannot detect drift or manage lifecycle
- No declarative relationship between certificate and load balancer
- Must track certificate IDs manually for load balancer configuration
- No secrets management for private keys in scripts

**Verdict:** Good for quick operations and scripted workflows. Not suitable when certificates are part of a managed infrastructure stack.

### Level 2: IaC (Terraform)

#### Uploaded Certificate

```hcl
resource "hcloud_uploaded_certificate" "wildcard" {
  name        = "wildcard-cert"
  certificate = file("certs/fullchain.pem")
  private_key = file("certs/privkey.pem")

  labels = {
    env     = "production"
    purpose = "wildcard"
  }
}
```

#### Managed Certificate

```hcl
resource "hcloud_managed_certificate" "api" {
  name         = "api-cert"
  domain_names = ["example.com", "www.example.com"]

  labels = {
    env = "production"
  }
}
```

#### Certificate Referenced by Load Balancer

```hcl
resource "hcloud_managed_certificate" "web" {
  name         = "web-cert"
  domain_names = ["example.com", "www.example.com"]
}

resource "hcloud_load_balancer" "web" {
  name               = "web-lb"
  load_balancer_type = "lb11"
  location           = "fsn1"
}

resource "hcloud_load_balancer_service" "https" {
  load_balancer_id = hcloud_load_balancer.web.id
  protocol         = "https"
  listen_port      = 443
  destination_port = 80

  http {
    certificates = [hcloud_managed_certificate.web.id]
  }
}
```

**Key Terraform behaviors:**
- `hcloud_uploaded_certificate` and `hcloud_managed_certificate` are separate resource types. There is no unified `hcloud_certificate` resource for creation (the legacy `hcloud_certificate` is deprecated and aliases to `hcloud_uploaded_certificate`).
- Both resources mark their core fields as `ForceNew`: `certificate`/`private_key` for uploaded, `domain_names` for managed.
- The `private_key` attribute is marked `Sensitive` in the schema — it does not appear in plan output.
- The uploaded certificate resource uses `DiffSuppressFunc` with `EqualCert()` to suppress diffs when the PEM content is semantically equivalent but formatted differently.

**Pros:**
- State tracking and drift detection
- Dependency graph between certificate and load balancer
- Sensitive value handling (private key masked in plans)
- Reproducible across environments

**Cons:**
- Two separate resource types requires conditional logic for a single-component abstraction
- No built-in certificate rotation workflow (must taint/replace)
- State file contains the private key (encrypted, but present)

**Verdict:** Production-grade for certificate management. The split resource types add complexity that Planton's unified component absorbs.

### Level 3: IaC (Pulumi)

#### Uploaded Certificate

```go
cert, err := hcloud.NewUploadedCertificate(ctx, "wildcard", &hcloud.UploadedCertificateArgs{
    Name:        pulumi.String("wildcard-cert"),
    Certificate: pulumi.String(certPEM),
    PrivateKey:  pulumi.ToSecret(pulumi.String(keyPEM)).(pulumi.StringInput),
    Labels: pulumi.StringMap{
        "env": pulumi.String("production"),
    },
})
```

#### Managed Certificate

```go
cert, err := hcloud.NewManagedCertificate(ctx, "api", &hcloud.ManagedCertificateArgs{
    Name: pulumi.String("api-cert"),
    DomainNames: pulumi.StringArray{
        pulumi.String("example.com"),
        pulumi.String("www.example.com"),
    },
    Labels: pulumi.StringMap{
        "env": pulumi.String("production"),
    },
})
```

**Key Pulumi behaviors:**
- `hcloud.UploadedCertificate` and `hcloud.ManagedCertificate` mirror the Terraform split. The legacy `hcloud.Certificate` is deprecated.
- `pulumi.ToSecret()` marks the private key as a secret in the Pulumi state. It is encrypted at rest and never shown in preview output.
- Both types produce the same output properties: `Fingerprint`, `NotValidBefore`, `NotValidAfter`, `Type`.

**Pros:**
- Same benefits as Terraform (state, dependencies, drift detection)
- First-class secret handling with `pulumi.ToSecret`
- Go type system catches misconfigurations at compile time

**Cons:**
- Same split-type complexity as Terraform
- Requires Go toolchain for the HCloud provider module

**Verdict:** Equivalent to Terraform for certificate management, with the added benefit of compile-time safety in Go.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | Planton |
|--------|---------|-----|-----------|--------|---------|
| **Reproducible** | No | Partial | Yes | Yes | Yes |
| **State tracked** | No | No | Yes | Yes | Yes |
| **Secret handling** | Paste in form | File on disk | Sensitive attribute | `ToSecret` | Inherited from IaC |
| **Unified type** | Yes (UI abstracts) | Yes (`--type` flag) | No (two resources) | No (two types) | Yes (proto oneof) |
| **Drift detection** | No | No | Yes | Yes | Yes |
| **LB dependency** | Manual reference | Manual ID | Resource reference | Resource reference | `valueFrom` |

## The Planton Approach

### Why a Single Component for Two Types

Terraform and Pulumi expose uploaded and managed certificates as separate resource types. Planton unifies them into a single `HetznerCloudCertificate` component because:

1. **Same purpose**: Both types serve the same function — providing a TLS certificate for load balancer HTTPS services. The difference is in sourcing, not in consumption.
2. **Same outputs**: Both produce the same set of outputs (`certificate_id`, `type`, `fingerprint`, `not_valid_before`, `not_valid_after`). Downstream consumers (load balancers) do not care which type was used.
3. **Mutual exclusivity**: A certificate is either uploaded or managed, never both. The proto `oneof` enforces this structurally — there is no ambiguity.
4. **Simpler catalog**: One component in the catalog instead of two. Users choose the type at configuration time, not at component selection time.

### The 80/20 Field Scoping

**Included:**
- `uploaded.certificate` — the PEM certificate chain (required for uploaded)
- `uploaded.privateKey` — the PEM private key (required for uploaded)
- `managed.domainNames` — the domain list (required for managed)

**Derived (not in spec):**
- `name` — derived from `metadata.name`
- `labels` — derived from metadata following CG01

**Not exposed:**
- Certificate `type` field — implicit from which variant is set. No need to declare `type: "uploaded"` when the `uploaded` block already conveys the type.
- Explicit `domain_names` for uploaded certificates — the domains are extracted from the certificate by Hetzner Cloud automatically.

This scoping means the spec contains exactly the fields the user must provide and nothing they don't. There are no redundant or derived fields cluttering the interface.

### Proto Oneof Design

The `oneof certificate` construct in `spec.proto` enforces mutual exclusivity at the protobuf level:

```protobuf
message HetznerCloudCertificateSpec {
  oneof certificate {
    option (buf.validate.oneof).required = true;
    UploadedCertificateConfig uploaded = 1;
    ManagedCertificateConfig managed = 2;
  }
}
```

This design means:
- A manifest with both `uploaded` and `managed` blocks is a **parse error**, not a runtime error.
- A manifest with neither block fails proto validation (`oneof.required = true`).
- The Go type system generates a type switch (`HetznerCloudCertificateSpec_Uploaded` / `HetznerCloudCertificateSpec_Managed`) that the Pulumi module uses directly — no string comparison or boolean flags.

The Terraform module achieves the same routing with `count` and `local.is_uploaded` / `local.is_managed` booleans, plus a validation block that enforces exactly one variant.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module uses a Go type switch on the proto oneof:

```go
switch cert := spec.Certificate.(type) {
case *HetznerCloudCertificateSpec_Uploaded:
    return uploadedCertificate(ctx, name, cert.Uploaded, locals, provider)
case *HetznerCloudCertificateSpec_Managed:
    return managedCertificate(ctx, name, cert.Managed, locals, provider)
}
```

Both `uploadedCertificate` and `managedCertificate` functions export the same five outputs, making the module polymorphic — downstream code that reads the outputs does not need to know which type was used.

The private key is wrapped with `pulumi.ToSecret` before being passed to `UploadedCertificateArgs.PrivateKey`. This ensures the value is encrypted in the Pulumi state file and masked in all CLI output.

### Terraform Module Architecture

The Terraform module uses conditional `count`:

```hcl
resource "hcloud_uploaded_certificate" "this" {
  count = local.is_uploaded ? 1 : 0
  # ...
}

resource "hcloud_managed_certificate" "this" {
  count = local.is_managed ? 1 : 0
  # ...
}
```

Outputs use ternary expressions to select from whichever resource was created:

```hcl
output "certificate_id" {
  value = local.is_uploaded
    ? hcloud_uploaded_certificate.this[0].id
    : hcloud_managed_certificate.this[0].id
}
```

The `variables.tf` includes a validation block that enforces exactly one variant:

```hcl
validation {
  condition = (
    (var.spec.uploaded != null ? 1 : 0) +
    (var.spec.managed != null ? 1 : 0)
  ) == 1
  error_message = "Exactly one of 'uploaded' or 'managed' must be set."
}
```

## Production Best Practices

### Managed Certificate Prerequisites Checklist

Before deploying a managed certificate, verify:

1. **DNS records exist** — each domain in `domainNames` must have an A or AAAA record pointing to the load balancer's IP address.
2. **Load balancer has an HTTPS service** — the HTTPS service must reference this certificate's ID. This creates a circular dependency: the certificate needs the LB for the ACME challenge, and the LB needs the certificate for HTTPS. In practice, the certificate enters a pending state until the LB is configured, then completes.
3. **Port 80 is reachable** — the ACME HTTP-01 challenge uses plain HTTP. If the load balancer only listens on 443, the challenge fails. Ensure the LB has an HTTP service on port 80 or that the HTTPS service redirects appropriately.
4. **No conflicting certificates** — if another certificate (managed or uploaded) already covers the same domain on the same load balancer, the ACME challenge may behave unpredictably.

### Renewal Monitoring for Uploaded Certificates

Uploaded certificates have no automatic renewal. To avoid unexpected expiration:

- **Monitor `not_valid_after`** — this output provides the certificate's expiration date in ISO-8601 format. Set up alerts 30 and 7 days before expiry.
- **Automate rotation** — use an external ACME client (e.g., certbot, lego) to obtain new certificates, then update the manifest and re-deploy.
- **Use managed certificates when possible** — if you don't need wildcards, EV/OV, or a specific CA, managed certificates eliminate the renewal burden entirely.

### Certificate Rotation Without Downtime

Rotating an uploaded certificate requires replacing the resource (ForceNew). During the brief window between destroying the old certificate and creating the new one, the load balancer's HTTPS service has no valid certificate.

To minimize this window:

1. **Pulumi**: The replace happens atomically within a single `pulumi up`. The provider deletes the old certificate and creates the new one in sequence. The interruption is typically under one second.
2. **Terraform**: Same behavior — `terraform apply` handles the delete-then-create sequence.

For zero-downtime rotation:
1. Create a new certificate resource with a different name.
2. Update the load balancer to reference the new certificate.
3. Delete the old certificate.

This three-step approach avoids any period where the load balancer references a non-existent certificate.

### Immutability Awareness

| Field | Change Behavior |
|-------|----------------|
| `uploaded.certificate` | ForceNew — destroys and recreates the certificate |
| `uploaded.privateKey` | ForceNew — destroys and recreates the certificate |
| `managed.domainNames` | ForceNew — destroys and recreates the certificate |
| `name` (via metadata) | Update in place |
| `labels` (via metadata) | Update in place |

Plan changes carefully: any modification to the certificate content or domain list results in a new certificate with a new `certificate_id`. Load balancers referencing the old ID must be updated.

### When to Use Uploaded vs Managed

| Scenario | Recommendation |
|----------|---------------|
| Standard web HTTPS (DV sufficient) | Managed — zero renewal burden |
| Wildcard certificate (`*.example.com`) | Uploaded — HTTP-01 does not support wildcards |
| EV or OV certificate required | Uploaded — Let's Encrypt only issues DV |
| Internal/private CA | Uploaded — managed only works with Let's Encrypt |
| Certificate used across providers | Uploaded — same cert can be stored in Hetzner Cloud and elsewhere |
| Rapid prototyping | Managed — no need to generate certificates externally |

### DNS Propagation Timing

When creating a managed certificate for a new domain, DNS propagation can delay the ACME challenge. If the DNS record was created moments before the certificate request, Let's Encrypt may not yet see the record.

Mitigation:
- Create DNS records first and wait for propagation (typically 1-5 minutes for most DNS providers, up to 48 hours in worst cases).
- Use a DNS provider with fast propagation (Cloudflare, Route53, Hetzner DNS — all propagate within seconds to minutes).
- The managed certificate will retry the challenge. If DNS propagation completes within the retry window, the certificate will eventually succeed without intervention.

## References

- [Hetzner Cloud Certificates API](https://docs.hetzner.cloud/#certificates)
- [Hetzner Cloud Load Balancer Documentation](https://docs.hetzner.cloud/#load-balancers)
- [Let's Encrypt HTTP-01 Challenge](https://letsencrypt.org/docs/challenge-types/#http-01-challenge)
- [Terraform hcloud_uploaded_certificate](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/uploaded_certificate)
- [Terraform hcloud_managed_certificate](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/managed_certificate)
- [Pulumi hcloud.UploadedCertificate](https://www.pulumi.com/registry/packages/hcloud/api-docs/uploadedcertificate/)
- [Pulumi hcloud.ManagedCertificate](https://www.pulumi.com/registry/packages/hcloud/api-docs/managedcertificate/)
