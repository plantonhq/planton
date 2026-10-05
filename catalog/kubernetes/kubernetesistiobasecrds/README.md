# KubernetesIstioBaseCrds

Installs the **Istio CRDs** (the `istio/base` Custom Resource Definitions) on a target
Kubernetes cluster — **CRDs only, no istiod and no controller**.

## When NOT to Use This

**If you want a running mesh, use KubernetesIstio instead.** That component owns
everything, CRDs included — deploying both on one cluster is unnecessary (though safe:
both apply the CRDs via server-side apply, so they co-own rather than conflict).

## When to Use This

This is the lightweight prerequisite for the typed Istio API kinds on clusters that
use the Istio policy/config APIs **without running a mesh**:

- `KubernetesDestinationRule`
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`

Those kinds only need the CRDs present so their objects can be applied and server-side
validated — they do not need a control plane. Each declares this kind as its
prerequisite. It is the Istio analog of `KubernetesGatewayApiCrds` for the Gateway API
family.

## Upgrade Path to a Full Mesh

When a mesh is later wanted on a CRDs-only cluster, deploy `KubernetesIstio` — a plain
redeploy. Both kinds apply the CRDs via **server-side apply** (never Helm-owned), so the
full-mesh install reconciles the existing CRDs in place instead of fighting over
ownership. From that point `KubernetesIstio` owns everything, CRDs included.

## Why There Is No `version` Field

The installed CRD schema is **pinned** to the Istio version the typed SDK was generated
against. The typed custom resources are frozen to that schema, so a user-selectable CRD
version would be incoherent — a mismatched CRD set would silently prune or reject
fields. The pin is a build-time constant in both IaC modules, moved in lockstep with the
SDK pin.

The pin is always an exact release **tag** (e.g. `1.30.3`), never a release branch: a
branch ref moves as patches land, so the same deployed resource would install different
CRD schemas at different times — tag pinning keeps installs reproducible and exactly
matched to the generated SDK.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstioBaseCrds
metadata:
  name: my-istio-base-crds
spec: {}
```

## What Gets Installed

The upstream `istio/base` `crd-all.gen.yaml` bundle at the pinned tag, containing the
`networking.istio.io`, `security.istio.io`, and `telemetry.istio.io` (and related)
CustomResourceDefinitions.

## IaC

- **Pulumi**: applies the bundle via `yaml.NewConfigFile`.
- **Terraform**: fetches the bundle via `http` and applies it with `kubectl_manifest`
  (server-side apply).

Outputs: `installed_release`, `installed_manifest_url`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
