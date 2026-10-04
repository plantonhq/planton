# GcpWorkloadIdentityPoolProvider

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpWorkloadIdentityPoolProviderSpec defines one external identity issuer
trusted by a Workload Identity Pool — the piece that turns a pool into a
working keyless-auth path. A pool holds many providers, one per issuer
(a GitHub org, an AWS account, a SAML IdP, an X.509 estate), each with its
own attribute mapping and trust conditions.

The issuer type is an exclusive choice (aws | oidc | saml | x509) fixed at
design time; attribute_mapping translates the issuer's claims into Google
attributes, and attribute_condition gates which of the issuer's otherwise
valid credentials are accepted (e.g. only one repository, only one branch).

Lifecycle note: like the pool, GCP soft-deletes providers for ~30 days and
blocks ID reuse while soft-deleted; a create against a soft-deleted ID
fails. Prefer disabling over deleting during rotation or investigation.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkloadIdentityPoolProvider
metadata:
  name: my-sample-pool-provider
spec:
  # The pool this provider belongs to (bare pool ID). Reference a
  # GcpWorkloadIdentityPool resource — its workload_identity_pool_id output
  # is exactly this value.
  workloadIdentityPoolId:
    value: github-actions

  # Provider ID (4-32 chars; lowercase letters, digits, hyphens; the "gcp-"
  # prefix is reserved)
  workloadIdentityPoolProviderId: github-oidc

  # GCP project that owns the pool.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Name shown in the GCP console
  displayName: GitHub Actions OIDC

  # Which issuer this trusts and any operational notes
  description: Trusts GitHub-minted OIDC tokens for the engineering org

  # Claim -> Google-attribute mappings. google.subject is required for OIDC;
  # custom attribute.* entries become IAM-targetable principal sets.
  attributeMapping:
    google.subject: assertion.sub
    attribute.repository: assertion.repository

  # Gate which of the issuer's otherwise valid tokens are accepted — for
  # multi-tenant issuers like GitHub Actions, always restrict to your org.
  attributeCondition: assertion.repository_owner == "my-org"

  # The issuer — exactly one of aws / oidc / saml / x509
  oidc:
    issuerUri: https://token.actions.githubusercontent.com

  # DELETE keeps destroys real; PREVENT belongs on providers live
  # pipelines federate through (deletion also locks the ID ~30 days).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.workloadIdentityPoolId` | `string \| valueFrom` | yes |  | GcpWorkloadIdentityPool (`status.outputs.workload_identity_pool_id`) |
| `spec.workloadIdentityPoolProviderId` | `string` | yes |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.attributeMapping` | `map<string, string>` |  |  |  |
| `spec.attributeCondition` | `string` |  |  |  |
| `spec.aws` | `GcpWorkloadIdentityPoolProviderAws` |  |  |  |
| `spec.aws.accountId` | `string` | yes |  |  |
| `spec.oidc` | `GcpWorkloadIdentityPoolProviderOidc` |  |  |  |
| `spec.oidc.issuerUri` | `string` | yes |  |  |
| `spec.oidc.allowedAudiences` | `[]string` |  |  |  |
| `spec.oidc.jwksJson` | `string` |  |  |  |
| `spec.saml` | `GcpWorkloadIdentityPoolProviderSaml` |  |  |  |
| `spec.saml.idpMetadataXml` | `string` | yes |  |  |
| `spec.x509` | `GcpWorkloadIdentityPoolProviderX509` |  |  |  |
| `spec.x509.trustStore` | `GcpWorkloadIdentityPoolProviderTrustStore` | yes |  |  |
| `spec.x509.trustStore.trustAnchors` | `[]GcpWorkloadIdentityPoolProviderTrustAnchor` | yes |  |  |
| `spec.x509.trustStore.trustAnchors[].pemCertificate` | `string` | yes |  |  |
| `spec.x509.trustStore.intermediateCas` | `[]GcpWorkloadIdentityPoolProviderIntermediateCa` |  |  |  |
| `spec.x509.trustStore.intermediateCas[].pemCertificate` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.workloadIdentityPoolId

`string | valueFrom` · required

The pool this provider belongs to — the bare pool ID (the final component
of the pool's resource name), not the full path.
Reference a GcpWorkloadIdentityPool resource — its
workload_identity_pool_id output is exactly this value.
Immutable: a provider cannot move between pools.

- references: GcpWorkloadIdentityPool (`status.outputs.workload_identity_pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpWorkloadIdentityPool, name: <that resource's name>, fieldPath: status.outputs.workload_identity_pool_id}} -- a bare string does not parse

### spec.workloadIdentityPoolProviderId

`string` · required

The ID for the provider, which becomes the final component of its
resource name. 4-32 characters of lowercase letters, digits, and hyphens;
the prefix "gcp-" is reserved by Google. Immutable: changing it destroys
and recreates the provider, invalidating tokens minted for the old
audience.

- rule: the prefix 'gcp-' is reserved by Google — choose a provider ID that does not start with it
- rule: {"required":true,"string":{"pattern":"^[a-z0-9-]{4,32}$"}}

### spec.projectId

`string | valueFrom`

The GCP project that owns the pool (and therefore this provider).
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the provider.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string`

Human-readable name shown in the GCP console (max 32 characters). Mutable.

- rule: {"string":{"maxLen":"32"}}

### spec.description

`string`

Which issuer this provider trusts and any operational notes — write it
for the operator auditing trust boundaries later (max 256 characters).
Mutable.

- rule: {"string":{"maxLen":"256"}}

### spec.disabled

`bool`

Emergency kill switch: a disabled provider rejects new token exchanges
(already-issued Google credentials remain valid until they expire).
Prefer disabling over deleting — deletion starts the 30-day soft-delete
clock and blocks ID reuse. Mutable.

### spec.attributeMapping

`map<string, string>`

Maps claims from the issuer's credential to Google attributes. Keys are
"google.subject" (required — the principal IAM authenticates),
"google.groups", or custom "attribute.<name>" entries (max 50); values
are CEL expressions over the credential's claims via the `assertion`
keyword, e.g. {"google.subject": "assertion.sub",
"attribute.repository": "assertion.repository"}.
Mapped attributes are what IAM bindings can target:
principalSet://iam.googleapis.com/<pool>/attribute.<name>/<value>.
OIDC providers must define it explicitly; AWS, SAML, and X.509 providers
fall back to sensible issuer-specific defaults when omitted.

### spec.attributeCondition

`string`

A CEL expression gating which otherwise valid credentials are accepted,
over the `assertion`, `google`, and `attribute` keywords (max 4096
characters), e.g. restricting a GitHub provider to one org:
assertion.repository_owner == "my-org".
Without a condition, ANY identity the issuer vouches for can federate —
for multi-tenant issuers like GitHub Actions, always set one.

- rule: {"string":{"maxLen":"4096"}}

### spec.aws

`GcpWorkloadIdentityPoolProviderAws`

Trust an Amazon Web Services account: workloads presenting AWS
credentials from this account can federate.

### spec.aws.accountId

`string` · required

The 12-digit AWS account ID whose workloads may federate.

- rule: {"required":true,"string":{"pattern":"^[0-9]{12}$"}}

### spec.oidc

`GcpWorkloadIdentityPoolProviderOidc`

Trust an OpenID Connect issuer — the workhorse for keyless CI/CD
(GitHub Actions, GitLab CI, Kubernetes clusters, custom issuers).

### spec.oidc.issuerUri

`string` · required

The OIDC issuer URL, e.g. https://token.actions.githubusercontent.com
for GitHub Actions. Must match the `iss` claim of incoming tokens.

- rule: {"required":true,"string":{"uri":true}}

### spec.oidc.allowedAudiences

`[]string`

Acceptable `aud` (audience) values in incoming tokens; each at most 256
characters, at most 10 entries. When empty, the audience must equal the
provider's full canonical resource name (with or without the https:
prefix) — the safest default, since tokens minted for anything else are
rejected.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"maxLen":"256"}}}}

### spec.oidc.jwksJson

`string`

OIDC JWKS (public signing keys) in JSON format, for issuers whose keys
cannot be fetched from the issuer_uri's .well-known discovery document
(e.g. issuers behind a private network). Leave unset to use discovery —
the normal path. These are public keys, not a secret.

### spec.saml

`GcpWorkloadIdentityPoolProviderSaml`

Trust a SAML 2.0 identity provider, typically an enterprise IdP.

### spec.saml.idpMetadataXml

`string` · required

The SAML identity provider's configuration metadata XML document, as
exported by the IdP. Public IdP metadata (entity ID, SSO endpoints,
signing certificates), not a secret.

- rule: {"required":true}

### spec.x509

`GcpWorkloadIdentityPoolProviderX509`

Trust an X.509 certificate authority: clients presenting certificates
that chain to the configured trust store can federate (certificate-based
workload identity without any token issuer).

### spec.x509.trustStore

`GcpWorkloadIdentityPoolProviderTrustStore` · required

The trust store validating incoming end-entity certificates. Exactly one
trust store is supported per provider.

- rule: {"required":true}

### spec.x509.trustStore.trustAnchors

`[]GcpWorkloadIdentityPoolProviderTrustAnchor` · required

Trust anchors: an incoming end-entity certificate must chain up to one of
these.

- rule: {"repeated":{"minItems":"1"}}

### spec.x509.trustStore.trustAnchors[].pemCertificate

`string` · required

PEM certificate of the PKI used for validation. Must contain exactly one
CA certificate (root or intermediate). Public key material, not a secret.

- rule: {"required":true}

### spec.x509.trustStore.intermediateCas

`[]GcpWorkloadIdentityPoolProviderIntermediateCa`

Intermediate CA certificates available for building the chain from an
end-entity certificate to a trust anchor.

### spec.x509.trustStore.intermediateCas[].pemCertificate

`string` · required

PEM certificate of the PKI used for validation. Must contain exactly one
CA certificate (either root or intermediate cert). Public key material,
not a secret.

- rule: {"required":true}

### spec.deletionPolicy

`string`

Deletion policy for the provider — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the provider is deleted (starts the ~30-day soft-delete
               clock and blocks the ID from reuse until it expires)
  "PREVENT" -- destroy FAILS; protects the keyless-auth path every
               pipeline federating through this issuer depends on
  "ABANDON" -- the provider is removed from management but keeps
               exchanging tokens in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `oidc_requires_subject_mapping`: OIDC providers must map google.subject in attribute_mapping (e.g. {"google.subject": "assertion.sub"}) — GCP rejects the provider without it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpWorkloadIdentityPoolProvider, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The full resource name: projects/<project_number>/locations/global/workloadIdentityPools/<pool_id>/providers/<provider_id>. This is the AUDIENCE for the keyless-auth handshake: OIDC tokens minted for this provider set `aud` to this value (with an //iam.googleapis.com/ prefix), and GCP provider configurations that authenticate via web identity consume exactly this string as their audience. |
| `status.outputs.workload_identity_pool_provider_id` | `string` | The bare provider ID (the spec's workload_identity_pool_provider_id, echoed for composing tooling that addresses the provider by short ID). |
| `status.outputs.state` | `string` | The provider lifecycle state: ACTIVE, or DELETED while soft-deleted (GCP retains deleted providers for ~30 days; a DELETED provider rejects token exchanges and blocks reuse of its ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.workloadIdentityPoolId` | GcpWorkloadIdentityPool | `status.outputs.workload_identity_pool_id` |
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
