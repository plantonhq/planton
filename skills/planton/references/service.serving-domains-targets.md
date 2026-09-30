---
title: Serving Domains Per Target — Carrier Truths and Walking the domain_serving Check
description: The per-deployment-target truths of serving domains — which declared resource carries the hostname on each target (Cloudflare Worker, Kubernetes ingress and HTTPRoute, Cloud Run domain mapping, the ECS/ALB shapes), what each carrier completes on its own (DNS, TLS) and what stays the author's, and the remediation ladder for a failed or unverifiable domain_serving check keyed by evidence class. Read when a domain_serving check needs walking on a specific target, someone asks how their target gets the hostname wired, or DNS/certificate gaps need naming.
---

# Serving Domains Per Target — Carrier Truths and Walking the domain_serving Check

The core rules live in `serving-domains.md`: the environment declares the suffix verbatim, the label defaults to the slug, and injection is field mutation of DECLARED carrier resources — never synthesis, never the platform writing DNS itself. This file is the per-target depth: what each carrier is, what it completes on its own, and what remains the author's. One law spans all of them: injection fills only what the author left blank, and whether the hostname actually LANDED in a manifest is a stamped fact — a carrier kind that was declared but not wired (every slot authored, a DNS block disabled) reads as an unverifiable check naming the gap, never a probe failure.

## Cloudflare Workers — complete on their own

The worker itself is the carrier: a custom-domain entry is appended (or a blank entry filled), and Cloudflare provisions BOTH the DNS record and the TLS certificate inside the zone it infers from the hostname. Nothing else to wire — the zero-edit story. The one failure class: the hostname's zone is not in the connected Cloudflare account, which fails the worker's own apply (the resources_created check, not the domain check).

## Kubernetes — the ingress and the HTTPRoute, composed beside the workload

Workload kinds (Deployment, StatefulSet) never carry hostnames — exposure is composed from first-class kinds beside them. The ingress's blank `rules[].host` and empty `tls[].hosts` are filled; a fully-authored ingress is untouched, and no rule is ever appended (a rule needs paths and a backend only its author knows). The HTTPRoute is the Gateway API sibling: an EMPTY `hostnames` list is the injection slot (blank entries are unauthorable there, so an empty list is unambiguous authoring); an authored list is never widened.

What the cluster completes: external-dns publishes the injected host automatically where it runs; cert-manager issues for the tls hosts where it is annotated; a Gateway's listener certificates cover routes under it. Without those, DNS is a record the author creates pointing at the ingress load balancer's address (on the deployment record's native-endpoint URLs), and the gap surfaces on the domain check's evidence.

## Cloud Run — the declared domain mapping

The carrier is a `GcpCloudRunDomainMapping` declared beside the Cloud Run resource, its `route` referencing the service and its `domain` left blank as the slot. Three provider truths worth relaying: the domain must be VERIFIED for the provisioning identity (Search Console / `gcloud domains verify` — subdomains of a verified domain need no separate verification, and GCP refuses the create otherwise); the default certificate mode is AUTOMATIC (Google provisions the managed certificate once DNS resolves — provisioning can lag DNS by minutes to hours, a real cert-pending window); and the domain IS the mapping's GCP resource name, create-only — a changed hostname (relabeled service, redeclared environment domain) REPLACES the mapping on the next deploy. The mapping's own outputs name the exact DNS records to create (`resource_records`) — when DNS is the gap, those outputs ARE the remediation. Gen2 Cloud Functions are Cloud Run services under the hood, so a domain mapping naming the function's underlying service is their carrier too.

## ECS behind a load balancer — two shapes

The ALB deliberately carries no routing, so the carrier depends on the shape:

- **Shared environment ALB** (several services behind one balancer): the service declares its OWN listener rule and target group. The rule's host-header condition with an EMPTY pattern list is the slot — the hostname is filled, the rule forwards to the target group the ECS service registers into. Never authored patterns touched, never a condition appended (conditions AND together). TLS is the listener's certificate — a wildcard certificate (`*.{env-domain}`) covers every service; DNS is one wildcard record in the zone pointing at the ALB, authored once. Listener-rule priorities are the author's to keep distinct. This same arrangement is what makes ECS previews nearly free — the worked recipe (manifests, the two authoring laws, the one-file previews tree) lives in `preview-environments.md`.
- **The service's own ALB**: enabling the ALB's `dns` block (Route53 zone + alias records) makes the ALB itself a carrier — the hostname is appended to its `dns.hostnames` and the alias record is created through the customer's own declared resource. Zero DNS edits, the Cloudflare-class story. A hostname outside the declared zone fails the ALB's apply at the provider — the same accepted class as a worker hostname outside the account's zones. An ALB whose `dns` block is NOT enabled is untouched and does not count as wired.

## Walking the check — the remediation ladder, keyed by the evidence

Dispatch on the check's status and its evidence rows' own words, never on guesses:

- **unverifiable, "doesn't resolve yet", with a healthy workload and an answering native endpoint** — DNS not created or not propagated. The fix is per carrier: create the records the Cloud Run mapping's outputs name; the alias or wildcard record pointing at the ALB; a record to the ingress load balancer (or install external-dns); for a worker, confirm the hostname's zone is in the account. Propagation lags are real: the probe waits five minutes, so a record created during the deploy usually passes, and a slower one reads unverifiable on this deploy and verified on the next.
- **unverifiable, "the certificate isn't trusted yet"** — the certificate is pending or missing: Google's managed cert still provisioning (it starts only after DNS resolves), an ACM certificate on the listener missing the hostname (wildcard SANs cover labels, not apex), or a cert-manager order still pending. The carrier applied and DNS resolves — the wait or the SAN is the fix.
- **unverifiable, "answers HTTP 4xx at /"** (404, 401, 403, …) — something serves the name, and the probe cannot tell whether it is the service. An API with no route at `/` answers 404 and is healthy; a service behind sign-in answers 401 or 403 and is healthy too. Otherwise the carrier holds the name without routing it to the service: an ingress or route whose rule doesn't match the hostname, an ALB listener rule missing it, a domain mapping to another service.
- **failed, "answers HTTP 503" (or 502/504)** — the carrier routes the name but has no healthy backend: the service's pods not ready, or failing the carrier's health check. "answers HTTP 500" is the service itself failing the request; its logs say why.
- **failed, "refused the connection" / "didn't answer in time"** — nothing accepts connections on port 443 for the name, or a firewall or security group drops the runner's traffic.
- **unverifiable, "was not carried"** — nothing was wired to serve the name, and the platform deliberately did not probe it. Three shapes: no carrier resource declared (add the ingress/route/mapping/rule beside the workload); a carrier declared with every slot authored to other names or its DNS block off (blank the slot that should take the platform's name, or enable the block); or a PROMOTED deployment whose captured manifests predate a domain change — the capture was rendered with the old name, so redeploying from a build adopts the current domain.
- **the check absent entirely** — the environment declares no serving domain; absent is the honest answer.

Never relay any of these as "the deploy failed" — in every class above, the deployment applied and the environment completed.
