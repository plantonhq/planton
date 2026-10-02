# GcpTargetHttpProxy Guide

The judgment this guide protects: the HTTP proxy is deliberately thin —
its production role is almost always the redirect half of a frontend
pair, and its one mutable field is the mechanism behind zero-downtime
routing changes.

## One kind, two scopes

`region` empty builds the GLOBAL proxy — the frontend adapter of the
global external ALB, the cross-region internal ALB, and Traffic Director.
`region` set builds the REGIONAL proxy — the frontend adapter of the
regional external ALB and the regional internal ALB. Scope is a chain-wide
decision: a regional proxy's URL map must be a regional `GcpUrlMap` in the
same region, and the forwarding rule in front of it a regional
`GcpGlobalForwardingRule` (set its `region` too). The regional proxy has no
`proxyBind` (Traffic Director has no regional proxy); the spec rejects it
when `region` is set. A proxy never moves between scopes — `region` is
immutable.

## The standard pattern is a pair

A production frontend runs TWO proxies sharing one story: this HTTP proxy
points at a redirect-only URL map (http→https 301), while the HTTPS proxy
serves the real application. Serving plaintext application traffic
through this proxy is the exception (internal tooling, legacy clients) —
if that is not deliberate, the URL map behind this proxy should redirect.

## url_map is the zero-downtime lever

Everything on this resource is immutable EXCEPT `urlMap`: GCP swaps it in
place with a dedicated setUrlMap call. Repointing a live frontend at a
new routing table — a blue/green cutover of the entire routing brain — is
one spec edit with no downtime. Structure migrations around that: build
the new URL map beside the old one, then flip the reference.

## Everything else recreates the proxy

Name, description, region, keep-alive, and `proxyBind` are ForceNew — and a proxy
recreation briefly breaks every forwarding rule referencing the old
self_link. GCP also refuses to delete a proxy a forwarding rule still
references (`resourceInUseByAnotherResource`), so replacements must be
create-before-destroy: new proxy, repoint the forwarding rule, then
delete the old.

## Keep-alive is scheme-scoped

`httpKeepAliveTimeoutSec` (5–1200s) is honored only by the envoy-based
`EXTERNAL_MANAGED` load balancer, where GCP defaults to 610s; the classic
`EXTERNAL` scheme ignores it. Set it above your clients' own keep-alive
so the load balancer never closes first. On the wrong scheme it is a
silent no-op — not an error.

## proxyBind is Traffic Director only

Leave `proxyBind` false for internet-facing frontends. It binds the proxy
to Traffic Director mesh VIPs and is only meaningful behind
`INTERNAL_SELF_MANAGED` forwarding rules.

## Teardown discipline

`deletionPolicy: PREVENT` suits any proxy fronting production traffic:
the in-use refusal already protects an attached proxy, but PREVENT also
covers the window where the forwarding rule was destroyed first.
`ABANDON` leaves the proxy serving unmanaged.
