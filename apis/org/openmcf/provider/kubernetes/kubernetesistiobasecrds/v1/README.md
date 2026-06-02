# KubernetesIstioBaseCrds

Installs the **Istio CRDs** (the `istio/base` Custom Resource Definitions) on a target
Kubernetes cluster — **CRDs only, no istiod and no controller**.

This is the lightweight prerequisite for the typed Istio API components:

- `KubernetesDestinationRule`
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`

Each of those kinds declares `prerequisites: [KubernetesIstioBaseCrds]`, so the platform
(and the E2E harness) installs these CRDs before applying any Istio custom resource.

It is the Istio analog of `KubernetesGatewayApiCrds` for the Gateway API family.

## Why there is no `version` field

The installed CRD schema is **pinned** to the Istio version this OpenMCF release's typed
SDK was generated against (`pkg/kubernetes/kubernetestypes/Makefile` `istio_release`, and
the IaC module's `IstioRelease` constant). The typed custom resources are frozen to that
SDK version, so a user-selectable CRD version would be incoherent — a mismatched CRD set
would silently prune or reject fields. To move the Istio version, bump the SDK pin and
this component's constant together (they carry a cross-reference breadcrumb).

> If you want to run the **full Istio mesh** (istiod + ingress gateway), use
> `KubernetesIstio` (kind 825) instead — that component is a Helm install with a
> user-facing `version` field. `KubernetesIstioBaseCrds` installs only the CRDs.

## Example

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstioBaseCrds
metadata:
  name: my-istio-base-crds
spec: {}
```

## What gets installed

The upstream `istio/base` `crd-all.gen.yaml` bundle for the pinned release, which contains
the `networking.istio.io`, `security.istio.io`, and `telemetry.istio.io` (and related)
CustomResourceDefinitions.

## IaC

- **Pulumi**: applies the bundle via `yaml.NewConfigFile`.
- **Terraform**: fetches the bundle via `http` and applies it with `kubectl_manifest`
  (server-side apply).

Outputs: `installed_release`, `installed_manifest_url`.
