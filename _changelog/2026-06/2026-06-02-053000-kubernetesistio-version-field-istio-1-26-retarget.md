# KubernetesIstio version field + Istio 1.26 retarget

**Date**: June 2, 2026
**Type**: Feature + Behavior Change
**Components**: API Definitions, Kubernetes Provider, crd2pulumi typed SDK

## Summary

Added a user-facing `version` field to `KubernetesIstio` (kind 825) so the deployed Istio
Helm chart version (istio/base, istiod, ingress-gateway) is driven by the spec instead of a
hardcoded module constant. As part of retargeting the Istio API family to **Istio 1.26**, the
crd2pulumi-generated typed SDK under `pkg/kubernetes/kubernetestypes/istio` was regenerated from
`release-1.26` (was `release-1.22`), and the `KubernetesIstio` default version was bumped from
`1.22.3` to **`1.26.8`**.

## What's new

- `KubernetesIstio.spec.version` (`optional string`, default `1.26.8`, full-patch `X.Y.Z`
  pattern). Wired through both the Pulumi and Terraform modules; falls back to the module
  default when unset.
- crd2pulumi Istio typed SDK regenerated at `release-1.26` (CRD source path also moved from
  `manifests/charts/base/crds/` to `manifests/charts/base/files/` upstream; the Makefile target
  was updated accordingly).
- Removed the dead, misleading `DefaultLatestVersion` (`1.23.0`) var from the Istio Pulumi
  module (unused, and lower than the new stable pin).

## ⚠️ Behavior change (read before upgrading existing meshes)

The default Istio version changed `1.22.3` -> `1.26.8`. Because `protodefaults.ApplyDefaults`
runs on every manifest load, **a `KubernetesIstio` manifest that does not set `version` will
resolve to `1.26.8` on its next deploy.** Istio supports only **sequential, single-minor**
upgrades (1.22 -> 1.23 -> ... -> 1.26); a direct 1.22.3 -> 1.26.8 in-place jump is **not
supported** and can break the mesh.

**Action required for existing meshes:** pin the currently-running version explicitly before
redeploying, then upgrade one minor at a time.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstio
metadata:
  name: my-istio
spec:
  namespace:
    value: istio-system
  version: "1.22.3"   # pin current version; bump one minor at a time to reach 1.26.x
  container:
    resources:
      requests: { cpu: 50m, memory: 100Mi }
      limits: { cpu: 1000m, memory: 1Gi }
```

Fresh installs (and the in-repo presets / e2e scenarios, which are fresh-install templates)
correctly ride the new `1.26.8` default and need no change.

## Version model

The Istio *typed-schema* version (the crd2pulumi SDK, the forthcoming typed components, and the
CRDs installed by `KubernetesIstioBaseCrds`) is a property of this OpenMCF release, pinned in one
place (`pkg/kubernetes/kubernetestypes/Makefile` `istio_release`). The only user-facing version
knob is `KubernetesIstio.version` (an untyped Helm mesh install). Coherence rule: to use the
typed Istio components, the cluster's mesh + CRDs must be >= the release's Istio version.
