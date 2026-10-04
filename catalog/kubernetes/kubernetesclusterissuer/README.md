# Kubernetes Cluster Issuer

## When NOT to Use This

**When a namespace should own its signing authority, use KubernetesIssuer instead.** A ClusterIssuer is cluster-wide: every namespace can request certificates from it, and its credentials live in cert-manager's cluster-resource namespace. A namespace-scoped Issuer keeps a team's CA keypair or DNS credentials readable only inside that team's namespace. The signing capabilities are identical — scope is the only difference, and the deciding factor.

## Overview

**KubernetesClusterIssuer** creates one cert-manager ClusterIssuer — a cluster-scoped certificate authority front-end. The issuer is named after the resource (`metadata.name`); Certificates select it by that name, and ingress-shim annotations (`cert-manager.io/cluster-issuer: <name>`) use the same name.

Four signing backends cover the full upstream surface (the config is shared with KubernetesIssuer, so the two kinds can never drift):

| Backend | Signs with | Typical use |
|---|---|---|
| `acme` | A public CA speaking ACME | Public TLS — Let's Encrypt, ZeroSSL, Google CA |
| `ca` | A CA keypair in a Kubernetes Secret | Internal PKI, service mTLS |
| `self_signed` | The certificate's own key | Bootstrapping a root CA; dev/test |
| `vault` | Vault/OpenBao PKI engine | Centralized enterprise PKI |

**ACME at full depth**: multiple solvers with selectors (`dns_zones` / `dns_names` / `match_labels` — most specific wins), HTTP-01 through an Ingress class or Gateway API HTTPRoute, DNS-01 across nine providers (Cloudflare, Route53, Azure DNS, Google Cloud DNS, DigitalOcean, RFC 2136, acme-dns, Akamai, and the `webhook` extension point for everything else), External Account Binding for CAs that require it, certificate profiles, and preferred chains.

**Credentials are declared, not pre-provisioned**: wherever upstream expects a `secretRef` to a hand-created Secret, the spec takes the credential VALUE (marked sensitive). The modules materialize each credential as a Kubernetes Secret (named `<resource-name>-<purpose>`) in cert-manager's cluster-resource namespace and wire the CR's secretRef to it — identically on both engines.

**Keyless where the platform allows it**: Route53, Cloud DNS, and Azure DNS solvers with no static credentials authenticate through the identity configured on the KubernetesCertManager controller (`workload_identity`) — no long-lived DNS credentials on the cluster at all.

## Deploys never block on readiness

Issuer readiness depends on external reachability (the ACME server, Vault, DNS) that is not part of applying the resource. Neither engine waits for Ready — the same posture as Ingress never blocking on a controller. Check `kubectl get clusterissuer` (or compose consumers through references, which is what Infra Charts do).

## Essential Configuration Fields

### Required

- **`spec.cert_manager_namespace`**: where credential Secrets are materialized — reference a KubernetesCertManager's `status.outputs.cluster_resource_namespace` (its FK default) or supply the literal namespace
- **`spec.config`**: exactly one backend

### ACME quick reference

- `config.acme.email` + at least one solver (an ACME issuer without solvers can never satisfy a challenge — rejected at validation instead of hanging at issuance)
- Use the Let's Encrypt **staging** server while testing: production rate limits are strict and exhaustible
- Wildcards need DNS-01; HTTP-01 needs public port-80 reachability

## Outputs

| Output | Purpose |
|---|---|
| `cluster_issuer_name` | The handle Certificates and `cert-manager.io/cluster-issuer` annotations reference |
| `secrets_namespace` | Where this issuer's credential Secrets were materialized |
| `acme_account_key_secret_name` | ACME account key Secret (empty for non-ACME backends) |

## Composing in Infra Charts

`KubernetesCertManager → KubernetesClusterIssuer → KubernetesCertificate` deploys in one chart run: the issuer references the installation's cluster-resource namespace, certificates reference `status.outputs.cluster_issuer_name`. The cross-cloud pattern (cluster in EKS, DNS in Cloudflare) is one solver block with a Cloudflare token — no cloud identity required.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
