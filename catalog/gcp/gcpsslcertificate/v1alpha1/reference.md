# GcpSslCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSslCertificateSpec defines a self-managed Compute Engine SSL certificate
— you bring the PEM certificate chain and private key (issued by your own
CA, purchased commercially, or automated via ACME outside GCP), and the
load balancer presents it to clients. Attach it to a target HTTPS (or SSL)
proxy the same way as a Google-managed certificate; the two kinds share
one name namespace and one API collection in GCP.

Choose self-managed when you need what Google-managed certificates cannot
do: wildcard domains, your own CA or EV/OV issuance, certificates for
internal load balancers, or serving TLS before public DNS cutover. Prefer
GcpManagedSslCertificate when hands-off issuance and renewal fits — a
self-managed certificate does NOT renew itself.

One kind covers both scopes. Leave `region` empty for a GLOBAL certificate
(global external load balancer proxies); set it for a REGIONAL one
(regional external and internal ALB proxies). The two scopes expose an
identical surface in GCP.

Every field is immutable in GCP: rotation is create-replacement, repoint
the proxy's certificate list, then destroy the old certificate
(create-before-destroy) — a certificate attached to a proxy cannot be
deleted, so the destroy fails rather than dropping TLS.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslCertificate
metadata:
  name: my-sample-ssl-cert
spec:
  # GCP project that owns the certificate.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name. Shares one namespace
  # with Google-managed certificates.
  certificateName: hack-self-managed-cert

  # What this certificate secures — shown in the GCP console.
  description: Hack-manifest self-managed certificate for module smoke tests

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE

  # THROWAWAY self-signed test material (CN=e2e-test.invalid, trusted by
  # nothing, generated solely for offline plans and module smoke tests).
  # Replace BOTH blocks with your real chain and key for any real use.
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDNDCCAhygAwIBAgIUYIKLPDuIarXycSz6rx3SRFmvf1swDQYJKoZIhvcNAQEL
    BQAwGzEZMBcGA1UEAwwQZTJlLXRlc3QuaW52YWxpZDAeFw0yNjA3MDMxMjM2Mjda
    Fw0zNjA2MzAxMjM2MjdaMBsxGTAXBgNVBAMMEGUyZS10ZXN0LmludmFsaWQwggEi
    MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCO53WppNv9zZyXpUH5f+V1OXs6
    XcxoDsllHN4IiDiZ1BbLhVMUM6z1cC8equXXdN3ngY9SUeNXbhJxm91MThIr+Sym
    kTzZKR50TK7NUXKtVs0Fmgn8bs4EJ9CZWKhLsnEYxXgjxid2I1u60ui0RUjuJ8Ci
    VOdCZUCmpxlns6e0AFoCK+zW8akqyBFbv/0pdjTXRMqOjG7v2lGOIYg7sA4pDtYr
    ZPkQdflWgnIt+LjLUlOlhk+r06M+oyh25NvrK5u4YBFRid3pB2Bury2qv262QsEA
    h4slDFpXRn8CX3fvFOBjYYkhZD3oXNPC8EAV+8gKcDF56NAD57UUzK28qYRlAgMB
    AAGjcDBuMB0GA1UdDgQWBBS4KxWIaaUL66edh24eRgYgGHPEuTAfBgNVHSMEGDAW
    gBS4KxWIaaUL66edh24eRgYgGHPEuTAPBgNVHRMBAf8EBTADAQH/MBsGA1UdEQQU
    MBKCEGUyZS10ZXN0LmludmFsaWQwDQYJKoZIhvcNAQELBQADggEBADfwJLhYuN1f
    BMATSy0fGz0fLc0Wt/41I24FGMdi6DdnemnQ+wKOKw/gw8l7neN4Wg+8yv7iWNdW
    Ib+uFl9Ua6ALxDQSciMjoPKVxZ7mfeg2iO6iRu+ozckJIf7wLtLbDb1st5sj78ZL
    5c5KFACkh0zp7yfGNQInzCs+lPz3qFqqv9aY6YzQJ7H/wSzMMsgJZyZiWE6oTFvK
    PvYEJJb8Q3Uw8TF0ABMvc8U5lUGqZtmtI+i8MIHY/XTIQizZdjgUXnHmCwOG6wM+
    +Jf9De4XMfo7gtzA7dcYU0ONb07Ylfv7EfqdFb6MkwhIpLm5KyTeS1dFJSzWFDcR
    WSukHyAQgzg=
    -----END CERTIFICATE-----
  privateKey: |
    -----BEGIN PRIVATE KEY-----
    MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCO53WppNv9zZyX
    pUH5f+V1OXs6XcxoDsllHN4IiDiZ1BbLhVMUM6z1cC8equXXdN3ngY9SUeNXbhJx
    m91MThIr+SymkTzZKR50TK7NUXKtVs0Fmgn8bs4EJ9CZWKhLsnEYxXgjxid2I1u6
    0ui0RUjuJ8CiVOdCZUCmpxlns6e0AFoCK+zW8akqyBFbv/0pdjTXRMqOjG7v2lGO
    IYg7sA4pDtYrZPkQdflWgnIt+LjLUlOlhk+r06M+oyh25NvrK5u4YBFRid3pB2Bu
    ry2qv262QsEAh4slDFpXRn8CX3fvFOBjYYkhZD3oXNPC8EAV+8gKcDF56NAD57UU
    zK28qYRlAgMBAAECggEANCYrPhk3Xstl1cEs7kvKBJlRat9H9MFQpWF/dUWgwiIv
    n12sD7c76uVhjKN49MNlJ1KUZsoTWJiGiocCnxHud7Waob5moijrQC2rrKmIW9FN
    SDoKYuBctg+BhDRiVh1sQEnvqb5qMCZ/FxJYcVDHaIGBPrwVGJmymh5omvtou7rJ
    YPlJrqV5j9Dmy9jCjpEy+21arDcvJ5oPWZqT550Gniz1sX1Ja/5tSr0Kg5B9Yfvn
    B4e5casYvIAE2/1Vzkq//rWikfbWWJKOMfMZsPt1KHw+YzqB5Yn8/lwFsTtQNNA3
    mmH3jBQ81K50l/Nd/KxnZLYEiBKGFDN2ODoRyvwTkQKBgQDIBMVJi8a32h2t8IW5
    26sfltD8eW4GVOlYYUFlic725DbdMhb+fP7k7MZUjrUDOE/8F/p+AY70croe2/Fp
    ubYTBL7mB657qt2BE+nrz0zVTLg3WCA/gLo/JEIGDiIl9PxHK5p3ISazpkN7vCiu
    D88SjxonQ1llih8ET8HbnkbKwwKBgQC25nct7yn/mhzEIiUAC9vDzttFI3FD7KpA
    bPOD6F/ur/mgs+jxhy2trwehcp4VtrRRx1yf8LIYmbij6nsG2y2I8WV4Gwse8QqW
    N/JjxFt0lzYdmDWSzW0Cv1D8LFx5qp/EDKUCsYvBkTtIoS0nPUEA07G9VciYhGLN
    2G9gC0/xtwKBgH4Ns5/T/Rpk1YuHJ1+oNsIjs/VJObO3048lS6eIH+ysin8AUEl1
    0NXI+nzTqvQqiw3etriulr8rhmxoRE5TAZIezYf+k1HQruPn/uXjsRJD1Vzbpwce
    Q0IDwbA7O/4b1Nmtex1UwSU6xRC31hNMVz3k/aB861v4ne+DrDKSHx8tAoGAZs6o
    0xMKQnh4Du86aQpBX5EYw4Ymlo2jLU+Qmea2dc5IvMIkAA+B54zo9yEcJwxp00YC
    lIyRLy7JEKouuS3eLIm0BYz99Uh8MPAFuXqYBbMxYfU6t+fsjIzJktXErUbxQxvw
    bNErw4RFFJA0d0gBD9vunoRnmwNfHmG4SP5S04UCgYBPOGnkhyP8fdb4uX2zlL/C
    Tmek77ijRd6QUxzR9+zj0Ez7AILotS3Xz9drjfT4ZyUANURPA27BAXO8Kr7Z7w5L
    zKbqDydM8ncIk1Kvqog5k8Zqb2Y6Co+XRVagvwmnyoCHVDewZbR+GF+R+mdailTF
    Xw2qlSb0EUhTQTMHRcz0sA==
    -----END PRIVATE KEY-----
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.certificateName` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.certificate` | `string` | yes |  |  |
| `spec.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the certificate.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the certificate.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.certificateName

`string`

Name of the certificate in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Self-managed and Google-managed certificates share one namespace per
scope — a name used by either kind is taken. Immutable: changing it
destroys and recreates the certificate, briefly breaking every proxy
that references the old self_link (rotate create-before-destroy
instead).

- rule: certificate_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.region

`string`

Region for a REGIONAL certificate (e.g. "us-central1"), used by regional
external and internal Application Load Balancer proxies. Leave empty for
a GLOBAL certificate — the right scope for global external load
balancers. Immutable: a certificate cannot move between scopes or
regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global SSL certificate

### spec.description

`string`

What this certificate secures and where it came from (issuing CA,
rotation cadence) — write it for the operator planning the next
rotation. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.certificate

`string` · required

The certificate chain in PEM format: the leaf certificate first,
followed by intermediates (at least one intermediate; at most 5
certificates total — GCP rejects longer chains). This is PUBLIC
handshake material presented to every client, so it is deliberately not
marked sensitive; only the private key is secret. Immutable.

- rule: certificate must be a PEM-encoded certificate chain (-----BEGIN CERTIFICATE-----)
- rule: {"string":{"minLen":"1"}}

### spec.privateKey

`string` · required · sensitive

The private key matching the certificate, in PEM format (-----BEGIN
PRIVATE KEY----- / -----BEGIN RSA PRIVATE KEY----- / -----BEGIN EC
PRIVATE KEY-----). GCP accepts RSA-2048 (and larger) and ECDSA P-256
keys; the key must be unencrypted (no passphrase). Write-only in GCP —
the API never returns it, and it never appears in stack outputs.
Immutable. The PEM framing is taught here rather than enforced by a
validation rule, because sensitive fields hold a managed-secret
reference on consuming platforms and a content-shape rule would
reject every reference.

- rule: {"string":{"minLen":"1"}}

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the certificate is deleted (GCP refuses while any
               proxy still references it, so destroy fails rather
               than dropping TLS)
  "PREVENT" -- destroy FAILS; a guard rail for a certificate whose
               replacement is not yet serving
  "ABANDON" -- the certificate is removed from management but left
               in GCP (useful mid-rotation handoff)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSslCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the SSL certificate. This is the value a target HTTPS (or SSL) proxy references in its ssl_certificates list — the composition handle that terminates TLS at the load balancer. Global: https://www.googleapis.com/compute/v1/projects/{project}/global/sslCertificates/{name} Regional: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/sslCertificates/{name} |
| `status.outputs.certificate_name` | `string` | Name of the SSL certificate as it exists in GCP. |
| `status.outputs.certificate_id` | `string` | Server-assigned numeric ID of the certificate. |
| `status.outputs.expire_time` | `string` | Expiry time of the certificate in RFC3339 format, parsed by GCP from the uploaded chain. Self-managed certificates do NOT renew themselves — plan the create-before-destroy rotation off this timestamp. |
| `status.outputs.region` | `string` | Region of a regional certificate; empty for a global one. Downstream composition can use this to confirm scope compatibility (regional proxies require certificates in their own region). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTargetHttpsProxy | `spec.sslCertificates` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
