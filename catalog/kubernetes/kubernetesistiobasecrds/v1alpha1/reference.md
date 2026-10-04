# KubernetesIstioBaseCrds

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesIstioBaseCrdsSpec defines configuration for installing the Istio CRDs
(the `istio/base` Custom Resource Definitions) on a target Kubernetes cluster,
WITHOUT installing istiod or any controller.

This is the lightweight prerequisite for the typed Istio API kinds
(KubernetesDestinationRule, KubernetesServiceEntry, KubernetesPeerAuthentication,
KubernetesRequestAuthentication, KubernetesAuthorizationPolicy, KubernetesTelemetry,
KubernetesEnvoyFilter). It places the networking/security/telemetry CRDs on the
cluster so those resources can be applied and server-side validated -- analogous to
KubernetesGatewayApiCrds for the Gateway API family.

Deliberately has NO version field. The installed CRD schema is pinned to the Istio
version this Planton release's typed SDK was generated against (see
pkg/kubernetes/kubernetestypes/Makefile `istio_release` and the IaC module's
IstioRelease constant). A user-selectable CRD version would be incoherent: the typed
custom resources are frozen to the SDK version, so a mismatched CRD set would silently
prune or reject fields. (This intentionally avoids the gateway family's footgun, where
KubernetesGatewayApiCrds.version defaults below its typed SDK version.)

## Example

```yaml
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstioBaseCrds
metadata:
  name: istio-base-crds-test
spec: {}
```

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesIstioBaseCrds, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.installed_release` | `string` | Istio release the CRDs were installed from (the pinned SDK tag, e.g. "1.30.3"). |
| `status.outputs.installed_manifest_url` | `string` | Full URL of the istio/base CRD bundle that was applied -- the authoritative record of exactly what landed on the cluster. |

## See Also

- [Overview](../README.md)
