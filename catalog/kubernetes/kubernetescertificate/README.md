# Kubernetes Certificate

## When NOT to Use This

**If nothing on the cluster should manage the certificate's lifecycle, don't create one here.** A KubernetesCertificate asks cert-manager to keep a certificate issued and RENEWED for as long as the resource exists. Certificates minted outside the cluster (uploaded wildcard certs, appliance-managed certs) belong in a KubernetesSecret instead — consumers reference a TLS Secret either way.

## Overview

**KubernetesCertificate** requests one signed X.509 certificate from a cert-manager issuer and keeps it renewed. The signed certificate, private key, and CA chain land in a Kubernetes TLS Secret (`secret_name`) — the handle every consumer references: Ingress `tls.secretName`, Gateway listener `certificate_refs`, workload volume mounts, and CA-backed issuers (`ca_secret_name`).

The spec covers the complete cert-manager.io/v1 request surface:

- **Names**: DNS (wildcards included), IP, URI (SPIFFE), email, otherName (Microsoft UPN) SANs; common name; subject attributes; or a `literal_subject` LDAP DN when attribute ORDER matters
- **Issuer selection**: a Planton-managed ClusterIssuer or Issuer by reference (composable in charts), or an `external` third-party issuer kind (e.g. AWS Private CA)
- **Lifetime**: `duration`, and renewal as an absolute window (`renew_before`) or a percentage of actual lifetime (`renew_before_percentage` — scales when the CA overrides the requested duration)
- **Key material**: RSA/ECDSA/Ed25519 with per-algorithm size validation, PKCS#1/PKCS#8 encoding, rotation policy
- **CA issuance**: `is_ca` plus X.509 `name_constraints` — the delegated-internal-CA guardrail
- **Outputs beyond PEM**: JKS/PKCS#12 keystores (inline sensitive passwords), DER and combined-PEM additional formats, labels/annotations on the Secret via `secret_template`

Contradictions are rejected at validation with messages that say what to fix: `literal_subject` vs `subject`/`common_name`, `renew_before` vs `renew_before_percentage`, key sizes that don't match the algorithm family, usages outside the x509 vocabulary.

## Deploys never block on issuance

Issuance time belongs to the issuer — an ACME order can take minutes, an unreachable CA would block forever. Neither engine waits for Ready; consumers express the dependency through composition, and `kubectl wait certificate/<name> --for condition=Ready` is the operational check.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: where the Certificate and its Secret live
- **`spec.secret_name`**: the TLS Secret name consumers reference — exported as `status.outputs.secret_name`
- **`spec.issuer_ref`**: `cluster_issuer` / `issuer` / `external`
- At least one requested name (any SAN type, common name, or literal subject)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Certificate (and Secret) namespace |
| `certificate_name` | The Certificate resource name |
| `secret_name` | The TLS Secret handle — Ingress TLS, Gateway certificate refs, CA Issuer input |

## Composing in Infra Charts

Reference `status.outputs.secret_name` from anything that consumes TLS Secrets. The root-CA bootstrap composes this kind twice: once with `is_ca: true` against a self-signed issuer (producing the CA Secret a `ca` Issuer signs with), then as leaf certificates against that CA issuer.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
