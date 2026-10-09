# A GKE or AKS Cluster Takes a Pre-Minted Token

**Date**: October 9, 2026
**Type**: Feature
**Components**: Kubernetes provider config (`catalog/kubernetes/provider.proto`), kubeconfig builder, kube exec credential command, GKE and AKS token minters, Kubernetes connections docs page

## Summary

**The Kubernetes provider config can now carry a token minted elsewhere, so a keyless cloud connection can reach the clusters it builds.**
- **Before:** a GKE cluster's config could carry only a service-account key, and an AKS cluster's only a client secret. A Planton runner holding a keyless (OIDC) Google or Azure connection had no way to hand the cluster its credential, so every Kubernetes resource on such a cluster failed (platform issue #1928).
- **Now:** `KubernetesProviderConfigGcpGke.access_token` and `KubernetesProviderConfigAzureAks.access_token` carry a short-lived token the caller minted. The rendered kubeconfig hands it to the exec credential command, which returns it as the cluster's bearer token.

## What Changed

- **`catalog/kubernetes/provider.proto`:**
  - `access_token` on the GKE config (field 4) and on the AKS config (field 6);
  - the rules `kubernetes.gcp_gke.one_credential` and `kubernetes.azure_aks.one_credential`: at most one credential per cluster config;
  - the comments say when the ambient chain applies.
- **`pkg/kubernetes/kubeconfig`:**
  - A GKE token rides the exec entry as `GOOGLE_OAUTH_ACCESS_TOKEN`, the name the GKE minter's chain already reads first. The exec entry's environment overrides the engine's, so the cluster's identity can never be confused with another Google identity the engine holds, such as the state bucket's sign-in.
  - An AKS token rides `PLANTON_AKS_ACCESS_TOKEN`.
- **`pkg/kubernetes/kubetoken`:**
  - `AksTokenOptions.AccessToken` is returned as is, and is refused together with a client secret.
  - The expiry reported for a supplied token, ten minutes, is one constant shared by GKE and AKS. `GoogleOAuthAccessTokenEnvVar` is exported so the builder and the minter share the name.
- **Docs:** the Kubernetes clusters page describes GKE, EKS and AKS connections as they work: each borrows a cloud connection of any sign-in method, keyless included. It no longer calls EKS and AKS "not yet implemented".
- **Bazel:** gazelle sorted two unrelated BUILD files (`pkg/iac/iacinput/providerenvvars`, `pkg/skills/hostprobe`) while the stubs regenerated.

## How It Was Proven

- New `catalog/kubernetes/provider_test.go`: both rules accept each credential alone or none, and refuse two together.
- `kubeconfig` tests: a GKE or AKS token lands on the exec entry, never in argv, and ambient mode emits no token entry.
- `execcredential` tests: the command returns a GKE or AKS supplied token as the bearer, offline.
- `kubetoken` tests: a supplied AKS token sends no Entra request; a token beside a client secret is refused.
