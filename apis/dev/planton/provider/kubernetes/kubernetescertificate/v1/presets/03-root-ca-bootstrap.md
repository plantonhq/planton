# Self-Signed Root CA Certificate (CA Bootstrap)

This preset creates a self-signed root CA certificate for bootstrapping an internal PKI. The resulting certificate Secret becomes the signing key for a CA Issuer, which can then issue leaf certificates for internal services.

## When to Use

- You need internal TLS certificates not signed by a public CA
- You are building a private PKI for mTLS, service mesh, or internal services
- You want a self-managed certificate authority

## Key Configuration Choices

- **`isCa: true`** -- Tells cert-manager to add the CA basic constraint and cert-sign key usage to the certificate
- **Namespace-scoped Issuer** -- References a SelfSigned Issuer (not a ClusterIssuer), since CA bootstrap is a namespace-scoped operation
- **10-year duration** (`87600h`) -- Root CAs should be long-lived; short durations cause unnecessary disruption
- **90-day renew-before** (`2160h`) -- Early renewal buffer for the root CA
- **RSA 4096** -- Higher key strength for CA certificates (leaf certs use 2048)
- **`rotationPolicy: never`** -- Critical: rotating a CA's private key invalidates all previously issued certificates
- **PKCS1 encoding** -- Traditional RSA encoding; use PKCS8 if ECDSA is needed

## CA Bootstrap Workflow

This preset is **Step 2** of the four-step CA bootstrap chain:

1. **KubernetesIssuer** (SelfSigned) → creates the bootstrap issuer
2. **KubernetesCertificate** (this preset) → creates the root CA cert
3. **KubernetesIssuer** (CA) → creates an issuer backed by the root CA cert's Secret
4. **KubernetesCertificate** (leaf) → issues leaf certs signed by the CA

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Namespace for the CA certificate (typically `cert-manager`) | Where your cert-manager infrastructure lives |
| `<your-root-ca-name>` | Conventional SAN for the CA certificate | A descriptive name like `my-org-root-ca` |
| `<your-root-ca-secret>` | Secret name for the CA cert + key | Referenced by the CA Issuer's `caSecretName` |
| `<your-selfsigned-issuer>` | Name of the SelfSigned Issuer | KubernetesIssuer's `issuer_name` output |

## Related Presets

- **01-cluster-issuer** -- Use for public TLS certificates via ACME
- **02-wildcard** -- Use for wildcard certificates
