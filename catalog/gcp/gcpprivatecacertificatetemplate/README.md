# GCP Private CA Certificate Template

A reusable certificate shape in Certificate Authority Service -- a TLS server leaf, an mTLS client -- that certificates from any pool in the same project and region reference. The template stamps predefined X.509 values onto every certificate issued with it, limits the identities and extensions a request may carry, and caps the lifetime; callers need `privateca.templateUser` on it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `privateca.googleapis.com` on the project (never disabled on destroy)
- **Certificate template** -- a `privateca_certificate_template` with its predefined values, identity constraints, passthrough extensions, and maximum lifetime

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service admin permissions (`roles/privateca.templateAdmin` or `roles/privateca.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- None. Certificates (`GcpPrivateCaCertificate`) reference the template.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaCertificateTemplate
metadata:
  name: tls-server
spec:
  location: us-central1
  maximumLifetime: 2592000s
  predefinedValues:
    caOptions:
      isCa: false
    keyUsage:
      baseKeyUsage:
        digitalSignature: true
        keyEncipherment: true
      extendedKeyUsage:
        serverAuth: true
```

```shell
planton apply -f certificate-template.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The CA Service region; certificates using the template must be in it. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `templateId` | `string` | `metadata.name` | 1-63 letters, digits, `-`, `_`. Immutable. |
| `description` | `string` | -- | What the template is for. |
| `maximumLifetime` | `string` | the pool's | The longest lifetime a certificate issued with it may have; the pool's cap still applies. |
| `predefinedValues` | `object` | -- | X.509 values stamped on every certificate: key usage, CA options, policy IDs, OCSP servers, extensions, name constraints. |
| `identityConstraints` | `object` | no limits | Subject and SAN passthrough, and a CEL expression over the identities. |
| `passthroughExtensions` | `object` | -- | Named (`knownExtensions`) and custom extensions a request may carry through. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `knownExtensions` entries are `BASE_KEY_USAGE`, `EXTENDED_KEY_USAGE`, `CA_OPTIONS`, `POLICY_IDS`, `AIA_OCSP_SERVERS`, or `NAME_CONSTRAINTS`.
- Durations are seconds with an `s` suffix; OIDs have at least one non-negative arc.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/certificateTemplates/{id}` -- what certificates reference |
| `template_id` | `string` | The template's ID |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Templates conflict-fail, they do not override.** A predefined value that conflicts with the pool's baseline values fails the request; a pool baseline value missing from `passthroughExtensions` fails it too.
- **`isCa` is presence-based.** Unset leaves the CA flag out of the certificate; `false` states CA:FALSE. The modules send the provider's `null_ca` for unset, because this resource would otherwise send false.
- **A template bills nothing**; certificates issued with it bill at their pool's tier.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpPrivateCaCertificate** -- certificates issued with the template
- **GcpPrivateCaPool** -- the pools whose policies the template must agree with

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
