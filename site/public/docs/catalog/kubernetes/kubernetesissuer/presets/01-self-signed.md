---
title: "Self-Signed Issuer"
description: "This preset creates a self-signed Issuer, the simplest issuer type. Self-signed Issuers require no external dependencies and are commonly used to bootstrap a CA chain or for development/testing..."
type: "preset"
rank: "01"
presetSlug: "01-self-signed"
componentSlug: "kubernetesissuer"
componentTitle: "KubernetesIssuer"
provider: "kubernetes"
icon: "package"
order: 1
---

# Self-Signed Issuer

This preset creates a self-signed Issuer, the simplest issuer type. Self-signed Issuers require no external dependencies and are commonly used to bootstrap a CA chain or for development/testing environments.

## When to Use

- You need to bootstrap a root CA by issuing a self-signed CA Certificate
- You're setting up a development or testing environment that needs TLS without external CAs
- You want a zero-dependency Issuer for quick iteration

## CA Chain Bootstrap Pattern

The most common production use of a SelfSigned Issuer is bootstrapping a CA chain:

1. Create a SelfSigned Issuer (this preset)
2. Create a KubernetesCertificate with `is_ca=true` referencing this Issuer -- this generates a CA Secret
3. Create a CA Issuer (preset 02-ca) referencing the CA Secret
4. Issue leaf certificates using the CA Issuer

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Namespace where the Issuer will live | Your cluster namespace list |

## Related Presets

- **02-ca** -- Use after bootstrapping a CA Secret from this SelfSigned Issuer
