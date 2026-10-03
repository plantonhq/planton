# Kubernetes Issuer

## When NOT to Use This

**When one signing authority should serve the whole cluster, use KubernetesClusterIssuer instead.** An Issuer is namespace-scoped: only Certificates in the same namespace can use it, and every Secret it needs must live in that namespace. That scope is the feature — but for the platform-wide "letsencrypt-production" that every team requests certificates from, a ClusterIssuer is the right grain.

## Overview

**KubernetesIssuer** creates one cert-manager Issuer — a namespace-scoped certificate authority front-end, named after the resource (`metadata.name`). The signing configuration is IDENTICAL to KubernetesClusterIssuer (the config message is shared; upstream defines the two kinds with the same spec): ACME at full depth, CA, self-signed, and Vault backends, with the same declared-credential model (values in the spec, Secrets materialized by the modules — here in the Issuer's own namespace).

**The namespace scope is the point:**

- A team's CA keypair and DNS credentials stay readable only inside the team's namespace, instead of being trusted cluster-wide
- The standard internal-PKI bootstrap lives entirely in one namespace: a `self_signed` Issuer → a root CA KubernetesCertificate (`is_ca: true`) → a `ca` Issuer signing with that root's Secret (reference the certificate's `status.outputs.secret_name` — the FK default)
- A team-owned ACME issuer keeps its DNS token in the team namespace

## Deploys never block on readiness

Same posture as KubernetesClusterIssuer: readiness depends on external reachability that is not part of applying the resource; neither engine waits.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: the Issuer's namespace — also where its credential Secrets and its Certificates must live
- **`spec.config`**: exactly one backend (`acme` / `ca` / `self_signed` / `vault`)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | The Issuer's namespace |
| `issuer_name` | The handle same-namespace Certificates reference |
| `acme_account_key_secret_name` | ACME account key Secret (empty for non-ACME backends) |

## Composing in Infra Charts

The CA-chain composition is the signature pattern: self-signed Issuer, root CA Certificate referencing it, CA Issuer referencing the root's Secret output, leaf Certificates referencing the CA Issuer — all four resources in one chart, all wiring through exported outputs.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
