---
title: Serving Domains — a Stable Hostname per Service, per Environment
description: An environment declares its serving domain once, and every service deployed there serves at {service}.{env-domain} — computed by the platform, injected into declared carrier resources, surfaced on deployment records, and probed as the domain_serving rollout check. Read when someone wants their service on their own domain, asks why a hostname is or isn't on a deployment, wants a different URL name than the service's, or a domain_serving check needs walking.
---

# Serving Domains — a Stable Hostname per Service, per Environment

Native endpoints (the run.app URL, the load balancer's DNS name) are infrastructure artifacts: generated, target-leaking, replaced when infrastructure is recreated. What a team publishes is a name on their own domain. That name is an ENVIRONMENT's fact, declared once — never per service, never per deploy.

## The declaration, and the two rules that never bend

An environment declares `spec.servingDomain.domain` — either a literal domain (`staging.acmecorp.com`; DNS can be managed anywhere, including outside Planton) or a `valueFrom` reference to a DNS zone resource managed through Planton. Two rules to hold in every answer:

1. **The declared value is the suffix, verbatim.** The platform never inserts the environment's name into a hostname and there is no prefixing convention to configure. Staging serves env-prefixed names because its declared domain SAYS `staging.…`; production serves clean names because its domain says so. When someone asks "how do I make prod URLs not say prod" — the answer is what they type on prod's declaration.
2. **The label defaults to the service's slug.** `checkout-api` in an environment declaring `acme.dev` serves at `checkout-api.acme.dev`, automatically. `spec.deploy.hostname` on the service overrides the label (`api` → `api.acme.dev`) — one label for ALL environments, deliberately. Slugs are unique per organization, so defaults never collide; two services resolving to one hostname in a shared environment refuse at apply, naming both.

What is deliberately NOT a service field: per-environment naming divergence, apex serving (`acmecorp.com` with no label), multiple hostnames, and CDN fronting. Those are edge infrastructure composed from the infra catalog — real resources the user owns, not switches on the service.

## How the hostname reaches the infrastructure

The platform computes the hostname at deploy collection and INJECTS it into declared carrier manifests — field mutation of resources the author declared, exactly like the built image filling a blank container-image field. It never synthesizes a resource and never writes DNS records itself. The authoring contract: a blank host field is the injection slot. Carriers exist for every compute shape — the Cloudflare Worker itself, the Kubernetes ingress and HTTPRoute, the Cloud Run domain mapping, the ALB listener rule and the ALB's DNS block for ECS — each with its own fill-vs-append law and its own DNS/TLS completion story; the per-target truths live in `service.serving-domains-targets.md`. Whether the hostname actually LANDED in a manifest is a fact the deploy stamps: a carrier kind declared without being wired counts as nothing, honestly.

Promote and rollback stay exact and correct: each environment's captured manifests were rendered with that environment's OWN hostname, so what moves between environments never carries another environment's name.

## Reading the record, and walking a domain_serving check

A deployment in a serving-domain environment carries the computed hostname on its URLs under the `serving_domain` tier as one `https://` URL, beside the `native_endpoint` URLs. Rollout verification probes it from the organization's own runner as the `domain_serving` staged check ("Serving domain serves the service"). The probe keeps asking for up to five minutes while the name is still arriving (the record not resolving yet, the certificate not trusted yet, a 5xx, a refused or timed-out connection), then reports a typed cause, and the check's detail opens with one sentence built from it. Dispatch on its status:

- **verified** — the hostname answered HTTPS 2xx or 3xx over a trusted certificate. Report it as THE address; the native endpoint is the debugging address.
- **failed** — the name answers but not with the service: a 5xx (the carrier holds the name with no healthy backend, or the service fails the request), or the connection is refused or times out (nothing listens on 443, or a firewall drops the runner). The deployment itself still applied. Never relay it as "the deploy failed".
- **unverifiable, with a cause** — the probe could not prove the service serves the name, and says which of these it met:
  - the name doesn't resolve yet (a record not created or not propagated);
  - the certificate isn't trusted yet (a managed certificate still being issued);
  - the name answers 4xx at `/`. That is healthy for an API with no route at `/`, and otherwise means the carrier doesn't route the name to the service.

  A carried domain in this state makes the whole deployment read "Not Verifiable", with that sentence as the reason: the address the environment promised is not proven yet, even when the workload is healthy. The per-carrier ladder is in `service.serving-domains-targets.md`.
- **unverifiable, naming that the hostname was not carried** — nothing in the applied manifests was wired to serve the name, so the platform deliberately did not probe it. Three shapes, all with their working paths in `service.serving-domains-targets.md`: no carrier resource declared, a declared carrier with every slot authored (or its DNS block off), or a PROMOTED deployment whose capture predates a domain change (redeploying from a build adopts the current domain). The computed hostname still shows on the URLs so the person knows what name awaits the wiring.

An environment with no declaration produces no serving URLs and no domain check at all — absent is the honest answer, not a gap to explain away.

For everything outside the convention — apex domains, arbitrary FQDNs, multiple hostnames, CDN fronting — the composed-infrastructure recipe is `service.serving-domains-custom.md`.
