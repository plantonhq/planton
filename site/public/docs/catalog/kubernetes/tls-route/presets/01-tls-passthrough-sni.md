---
title: "TLS Passthrough by SNI"
description: "The most common TLSRoute: match a TLS connection by its SNI hostname and forward it, unmodified (passthrough), to a single backend Service. The backend terminates TLS itself -- the Gateway never sees..."
type: "preset"
rank: "01"
presetSlug: "01-tls-passthrough-sni"
componentSlug: "tls-route"
componentTitle: "TLS Route"
provider: "kubernetes"
icon: "package"
order: 1
---

# TLS Passthrough by SNI

The most common TLSRoute: match a TLS connection by its SNI hostname and forward
it, unmodified (passthrough), to a single backend Service. The backend terminates
TLS itself -- the Gateway never sees the plaintext. This is the standard pattern
for exposing a service that does its own TLS termination (databases, mTLS
services, or apps that must hold their own certificate).

## When to Use

- You want end-to-end TLS where the backend, not the Gateway, terminates the
  connection.
- You route by SNI hostname only (TLS routes have no path/header matching).
- The parent Gateway has a TLS listener in `Passthrough` mode.

## Key Configuration Choices

- **`parentRefs`** -- attaches the route to the Gateway by name; add `sectionName` to target the TLS listener.
- **`hostnames`** -- the SNI hostnames that select this route; a leading `*.` is a suffix match. At least one is required, and IP addresses are not allowed (RFC 6066).
- **`backendRefs[].port`** -- the backend port that accepts the passthrough TLS connection.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) and has a
  listener of protocol `TLS` with `tls.mode: Passthrough`.
- The target namespace exists (`KubernetesNamespace`).
- The backend Service exists in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<sni-hostname>` | SNI hostname this route serves, e.g. `secure.example.com`. |
| `<service-name>` | Name of the backend Kubernetes Service. |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.
