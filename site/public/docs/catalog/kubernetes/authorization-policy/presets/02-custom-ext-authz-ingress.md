---
title: "Delegate to an External Authorizer (CUSTOM) on the Ingress Gateway"
description: "Delegate the authorization decision for sensitive ingress paths to an external authorization service (an OPA sidecar, a custom authz server, an OAuth2 proxy, ...) registered as an extension provider..."
type: "preset"
rank: "02"
presetSlug: "02-custom-ext-authz-ingress"
componentSlug: "authorization-policy"
componentTitle: "Authorization Policy"
provider: "kubernetes"
icon: "package"
order: 2
---

# Delegate to an External Authorizer (CUSTOM) on the Ingress Gateway

Delegate the authorization decision for sensitive ingress paths to an external
authorization service (an OPA sidecar, a custom authz server, an OAuth2 proxy, ...)
registered as an extension provider in the mesh's MeshConfig. Here, requests to the
ingress gateway whose path begins with `/admin/` are sent to the named provider, which
returns allow or deny.

## When to Use

- You already run an external authorization service and want istiod to consult it for
  specific routes (e.g. an admin console) before traffic reaches the backend.
- Your policy logic is too dynamic or org-specific to express as static ALLOW/DENY
  rules.

## How It Works

CUSTOM policies are evaluated before the native ALLOW and DENY actions and can only
further restrict the decision -- they cannot override an ALLOW/DENY denial. The external
provider must be declared in MeshConfig under `extensionProviders` (by the name
referenced in `spec.provider.name`); this policy only references it.

## Key Configuration Choices

- **`spec.namespace: istio-system`** + **`selector.match_labels.app: istio-ingressgateway`**
  -- target the ingress gateway workload. Adjust to your gateway's namespace/labels.
- **`action: CUSTOM`** with **`provider.name`** -- delegate to the named MeshConfig
  extension provider.
- **`rules[].to[].operation.paths`** -- restrict the delegation to the paths that need
  external authorization; other paths are unaffected by this policy.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running, and the ingress gateway is deployed (`KubernetesIstio`).
- The external authorizer is registered in MeshConfig `extensionProviders` under the
  name set in `provider.name`.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<ext-authz-provider>` | The MeshConfig `extensionProviders` entry name to delegate to (e.g. `my-custom-authz`). |
