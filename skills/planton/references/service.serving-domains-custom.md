---
title: Custom Domains — Apex, Arbitrary FQDNs, Multi-Host, and CDN as Composed Infrastructure
description: The first-class recipe for everything outside the {label}.{env-domain} convention — serving at an apex domain, an arbitrary FQDN, multiple hostnames, or behind a CDN — built as real infra-catalog resources the user owns, composed beside the service with valueFrom references onto its deployed outputs, never as switches on the service record. Read when someone wants their service at acmecorp.com or www, more than one hostname, a name that doesn't follow the convention, or a CDN in front.
---

# Custom Domains — Apex, Arbitrary FQDNs, Multi-Host, and CDN as Composed Infrastructure

The serving-domain convention answers one name per service per environment: `{label}.{env-domain}`. Everything past it — apex serving, arbitrary FQDNs, several hostnames, CDN fronting — is deliberately NOT a field on the service: it is edge infrastructure, and edge infrastructure here is composed from the infra catalog as real resources the user declares and owns. The assistant's job is assembling the right recipe, not looking for a hidden toggle.

## The recipe's shape

Every custom-domain arrangement is one or more catalog resources declared in the service's own environment manifest set (or as standalone environment infrastructure), wired to the service through `valueFrom` references onto its deployed resources' outputs. That bridge is the load-bearing piece: a DNS record resource can reference the ALB's DNS name output, a CDN distribution can reference the run.app URL, a domain mapping references the Cloud Run service's name — the platform resolves the reference at deploy time and the dependency appears on the resource graph.

Authored carriers stay authored: injection only fills BLANK slots, so a carrier whose hostname fields the user filled serves exactly what they wrote. And an authored carrier is still verified — its hostname surfaces on the deployment record's URLs (the native-endpoint tier) and the rollout probe answers whether it serves, so a custom setup is provable, never a leap of faith.

## Worked shapes

- **Apex on Cloud Run** (`acmecorp.com`): declare the `GcpCloudRunDomainMapping` with the domain AUTHORED (`domain: acmecorp.com`) instead of blank. The domain must be verified for the provisioning identity, the domain is the mapping's create-only GCP name, and the mapping's outputs name the DNS records to create — apex records are A/AAAA (no CNAME at apex), which is exactly what the outputs will say.
- **Apex or arbitrary FQDN on AWS**: a listener rule with the AUTHORED host pattern, the certificate (or an added SAN) on the listener, and the Route53 record — an alias record handles apex where CNAME cannot. Each is a catalog resource; nothing on the service changes.
- **Multiple hostnames**: additional authored entries beside the platform's — extra worker custom domains, extra ingress rules with their backends, extra listener rules. The platform's injected name and the authored names coexist; authored entries are never touched.
- **www + apex together**: the apex arrangement above plus one more record (or a redirect rule at the edge) — a deliberate pair of authored resources, so both names are visible, owned infrastructure.
- **CDN in front** (CloudFront, Cloudflare): the distribution is the catalog resource; its origin references the service's native endpoint output through `valueFrom`; the public DNS points at the distribution. The service's own serving-domain name can stay as the origin-facing address or be skipped entirely — the environment declaration is optional.

## What to hold the line on

Never simulate any of this by editing the service record's hostname label into something it is not (the label is one DNS label, not an FQDN valve), and never suggest the platform "just add a field" — the domain authority stays on the environment, and custom edge stays composed, visible, and owned. When a custom arrangement's DNS or certificate is the gap, the `domain_serving` walking ladder in `service.serving-domains-targets.md` applies to authored carriers exactly as to injected ones.
