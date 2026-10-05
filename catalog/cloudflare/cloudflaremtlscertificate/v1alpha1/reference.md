# CloudflareMtlsCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareMtlsCertificateSpec uploads a certificate to the account-level
mTLS certificate store. A CA certificate (ca: true) is what Authenticated
Origin Pulls and Workers mTLS bindings reference to validate client
certificates; a leaf certificate (ca: false) is presented BY Cloudflare as
a client. Self-signed CAs are the normal case here -- this store holds
YOUR trust material, not publicly trusted certificates.

Every field is create-only at the API: changing any of them replaces the
upload (the certificate id changes). Rotate by replacing the resource and
re-pointing consumers at the new certificate id.

## Example

```yaml
# A complete, protovalidate-valid CloudflareMtlsCertificate example: a CA
# certificate uploaded to the account store for Authenticated Origin Pulls
# rows, zone TLS CA associations, and Workers mTLS bindings to reference.
# The PEM value is a placeholder -- CA uploads validating clients carry no
# private key.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareMtlsCertificate
metadata:
  name: origin-pull-ca
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-pull-ca
  ca: true
  certificates: |
    -----BEGIN CERTIFICATE-----
    REPLACE_WITH_CA_CERTIFICATE_PEM
    -----END CERTIFICATE-----
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.ca` | `bool` | yes |  |  |
| `spec.certificates` | `string` | yes |  |  |
| `spec.privateKey` | `string \| valueFrom` (sensitive) |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the certificate is uploaded to.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string`

A display name for the certificate (shown in the dashboard).

### spec.ca

`bool` · required · optional (explicit presence)

Whether this is a CA certificate (true) or a leaf certificate (false).
Must be stated explicitly -- the two are consumed by different surfaces
and the API cannot change it after upload.

- rule: {"required":true}

### spec.certificates

`string` · required

The certificate (or CA chain) in PEM form. Keep the PEM byte-stable --
a formatting-only change still replaces the upload. Cloudflare stores
the PEM canonicalized to end with exactly one trailing newline and the
store is content-addressed (measured 2026-08-28): uploading identical
content answers the EXISTING certificate id, and a duplicate upload
while one is live is rejected (400 code 1471 "This certificate already
exists for this account"). The modules canonicalize the trailing
newline before sending, so trailing whitespace differences never churn.

- rule: {"required":true}

### spec.privateKey

`string | valueFrom` · sensitive

The certificate's private key in PEM form, only needed when Cloudflare
must present this certificate itself (leaf certificates). CA uploads used
for validating clients need no key. Provide a managed-secret reference;
the platform resolves it just-in-time at deploy and never stores it in
plaintext. The API never returns the key.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareMtlsCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | The ID of the uploaded certificate -- what Authenticated Origin Pulls rows, zone TLS CA associations, and Workers mTLS bindings reference. |
| `status.outputs.expires_on` | `string` | When the certificate expires (RFC3339). |
| `status.outputs.serial_number` | `string` | The certificate's serial number. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareZoneTlsSettings | `spec.caHostnameAssociations[].mtlsCertificateId` | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
