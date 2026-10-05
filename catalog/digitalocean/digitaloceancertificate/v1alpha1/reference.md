# DigitalOceanCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanCertificateSpec defines the fields required to create an SSL certificate in DigitalOcean.
A certificate is either issued and auto-renewed by Let's Encrypt (supply the domains) or uploaded
as user-provided PEM material (supply the key and certificate). The choice of certificate_source
branch fully determines the certificate type; both provisioners derive DigitalOcean's `type`
argument from whichever branch is set.

Every argument on this resource is create-only: any change replaces the certificate. Both
provisioners create the replacement before destroying the old certificate so a load balancer
referencing it by name never observes a gap.

## Example

```yaml
# Example DigitalOceanCertificate manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- a Let's Encrypt certificate: DigitalOcean issues and
# auto-renews it. Every listed domain must already be managed by DigitalOcean
# DNS in the same account.
#
# Document 2 -- a custom (user-provided) certificate: paste your own PEM
# material. The pair below is a THROWAWAY self-signed example (CN
# planton-e2e.invalid) -- replace all three values with your real material.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCertificate
metadata:
  name: example-docert-lets-encrypt
spec:
  certificateName: web-lets-encrypt
  letsEncrypt:
    domains:
      - example.com
      - "*.example.com"
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCertificate
metadata:
  name: example-docert-custom
spec:
  certificateName: web-custom
  custom:
    leafCertificate: |
      -----BEGIN CERTIFICATE-----
      MIIBkDCCATegAwIBAgIUZTNBcipSFS0umzBYZfnmE3UJ2YQwCgYIKoZIzj0EAwIw
      HjEcMBoGA1UEAwwTcGxhbnRvbi1lMmUuaW52YWxpZDAeFw0yNjA4MTgwNTI4MjVa
      Fw0zNjA4MTUwNTI4MjVaMB4xHDAaBgNVBAMME3BsYW50b24tZTJlLmludmFsaWQw
      WTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQosROw/D3FMQWRMKGFdcfpYUFaFuar
      eIRnJCbKT+y8i+hmk09dzw+i6Md3ruweTaXEc8P+CYPf6LbhJzR/OLqFo1MwUTAd
      BgNVHQ4EFgQU0gFX4oBzgNuZ+mSFs44uPvImF68wHwYDVR0jBBgwFoAU0gFX4oBz
      gNuZ+mSFs44uPvImF68wDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNHADBE
      AiBWQte/D5NHcDDduh/R5LYqA+A7ukhFhM6KpFzAFQnRHwIgTk1HMKYXSr+EBndV
      ktTsgT2GaJQv5Ygbwv3VidvCTcY=
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgI+JHoXfavvhJ48oX
      NY8PFbglsE3gttH0JFhDQngkWcGhRANCAAQosROw/D3FMQWRMKGFdcfpYUFaFuar
      eIRnJCbKT+y8i+hmk09dzw+i6Md3ruweTaXEc8P+CYPf6LbhJzR/OLqF
      -----END PRIVATE KEY-----
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.certificateName` | `string` | yes |  |  |
| `spec.letsEncrypt` | `DigitalOceanCertificateLetsEncryptParams` |  |  |  |
| `spec.letsEncrypt.domains` | `[]string` | yes |  |  |
| `spec.custom` | `DigitalOceanCertificateCustomParams` |  |  |  |
| `spec.custom.leafCertificate` | `string` | yes |  |  |
| `spec.custom.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.custom.certificateChain` | `string` |  |  |  |

## Field Details

### spec.certificateName

`string` · required

certificate_name is the unique, human-readable identifier of the certificate within the
DigitalOcean account. The name IS the resource identity: because a Let's Encrypt
certificate's UUID rotates on every auto-renewal, DigitalOcean addresses certificates by
their stable name, and other resources (e.g. load balancer forwarding rules) reference
certificates by name for the same reason.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64"}}

### spec.letsEncrypt

`DigitalOceanCertificateLetsEncryptParams`

A free Let's Encrypt certificate that DigitalOcean issues and auto-renews.

### spec.letsEncrypt.domains

`[]string` · required

domains is the list of fully-qualified domain names (or wildcard domains) to include
in the certificate. At least one domain is required. DigitalOcean issues one certificate
covering all listed names and renews it automatically; each renewal produces a new
certificate UUID while the name stays stable.

- rule: {"required":true,"repeated":{"unique":true,"items":{"string":{"pattern":"^(?:\\*\\.[A-Za-z0-9\\-\\.]+|[A-Za-z0-9\\-\\.]+\\.[A-Za-z]{2,})$"}}}}

### spec.custom

`DigitalOceanCertificateCustomParams`

A user-provided (custom) certificate uploaded as PEM material.

### spec.custom.leafCertificate

`string` · required

leaf_certificate is the PEM-encoded public certificate.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.custom.privateKey

`string` · required · sensitive

private_key is the PEM-encoded private key matching the leaf certificate.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.custom.certificateChain

`string`

certificate_chain is the optional PEM-encoded intermediate chain, ordered from the
issuing CA up to (but not including) the root.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | certificate_id is the certificate's resource identifier, which at the current provider pin is the certificate NAME, not a UUID: a Let's Encrypt certificate's UUID rotates on every auto-renewal, so DigitalOcean addresses certificates by their stable name. Resources that reference certificates (e.g. load balancer forwarding rules) consume this output. |
| `status.outputs.expiry_rfc3339` | `string` | expiry_rfc3339 is the expiration timestamp of the certificate in RFC 3339 format. For Let's Encrypt certificates this moves forward on every auto-renewal. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanCdn | `spec.certificate` | `status.outputs.certificate_id` |
| DigitalOceanLoadBalancer | `spec.forwardingRules[].certificateName` | `status.outputs.certificate_id` |
| DigitalOceanLoadBalancer | `spec.domains[].certificateName` | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
