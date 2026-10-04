# CloudflareOriginCaCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareOriginCaCertificateSpec provisions a Cloudflare Origin CA certificate:
a free TLS certificate that Cloudflare's edge trusts, installed on an origin
server so the Cloudflare-to-origin hop can run encrypted end-to-end (the "Full
(Strict)" SSL mode). It is NOT a public/browser-trusted certificate — it is only
valid between Cloudflare and the origin.

This component is a one-click certificate+key node. By default it generates the
private key and the certificate signing request (CSR) for you and returns the
signed certificate together with the (sensitive) private key, so a downstream
origin (a Kubernetes TLS secret, a load balancer listener, a VM) can mount both
without any out-of-band key handling. Advanced users who already manage their own
key can instead supply a `csr`, in which case no key is generated and the private
key never leaves their possession.

The certificate is account/user-scoped (there is no zone_id): it is requested
against the user's account via the API token and is valid for the hostnames listed
below, which may belong to any zone the account controls.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareOriginCaCertificate
metadata:
  name: test-origin-cert
spec:
  hostnames:
    - example.com
    - "*.example.com"
  requestType: origin-rsa
  requestedValidity: 5475
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.hostnames` | `[]string` | yes |  |  |
| `spec.requestType` | `string` |  | `origin-rsa` |  |
| `spec.requestedValidity` | `int64` |  | `5475` |  |
| `spec.csr` | `string` |  |  |  |

## Field Details

### spec.hostnames

`[]string` · required

The hostnames the certificate is valid for, e.g. "example.com" and
"*.example.com". These are the Subject Alternative Names on the issued
certificate. At least one is required. Changing them re-issues the certificate.
Every hostname must belong to an ACTIVE zone on the account: the API rejects
hostnames of pending (undelegated) zones with a misleading 400 code 1010
"zone not part of your account". Order does not matter — Cloudflare stores
the list lexicographically sorted and the IaC modules send it that way, so
list your primary hostname first purely for readability (it becomes the
CSR's common name on the generated-key path).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.requestType

`string` · optional (explicit presence)

The signature type of the requested certificate, which also selects the key
algorithm generated when `csr` is omitted: "origin-rsa" (RSA, the default and
the broadest-compatibility choice), "origin-ecc" (ECDSA, smaller/faster), or
"keyless-certificate" (for Keyless SSL setups). Defaults to "origin-rsa".

- default: `origin-rsa`
- rule: request_type must be one of origin-rsa, origin-ecc, keyless-certificate

### spec.requestedValidity

`int64` · optional (explicit presence)

How long the certificate is valid, in days. One of 7, 30, 90, 365, 730, 1095,
or 5475 (15 years). Defaults to 5475. Changing it re-issues the certificate.

- default: `5475`
- rule: requested_validity must be one of 7, 30, 90, 365, 730, 1095, 5475 (days)

### spec.csr

`string`

Optional certificate signing request (CSR) in PEM format. Supply this to use
your own key material: the module then requests the certificate for this exact
CSR and does NOT generate a key, so `status.outputs.private_key` is empty
(your key never leaves your control). When omitted (recommended), the module
generates a key + CSR for `hostnames` keyed to `request_type` and returns both
the certificate and the generated private key. A CSR is public material, not a
secret (the private key never appears in it).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareOriginCaCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | The Origin CA certificate identifier. |
| `status.outputs.certificate` | `string` | The issued Origin CA certificate in PEM format. This is public certificate material (not a secret) — install it on the origin alongside the private key. |
| `status.outputs.private_key` | `string` (sensitive) | The PEM-encoded private key for the certificate. Populated only when the module generated the key (i.e. when `spec.csr` was omitted); empty when a user-supplied CSR was used. Sensitive — exported as a secret so a downstream origin can mount it without the key ever living in plaintext. |
| `status.outputs.expires_on` | `string` | RFC3339 timestamp of when the certificate expires. |

## See Also

- [Overview](../README.md)
