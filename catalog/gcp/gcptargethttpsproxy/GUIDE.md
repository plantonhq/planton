# GcpTargetHttpsProxy Guide

The judgment this guide protects: the HTTPS proxy owns the client-facing
TLS handshake. The certificate mechanism, the TLS floor, and the rotation
choreography are all decisions made HERE, not on the certificates
themselves.

## One kind, two scopes

`region` empty builds the GLOBAL proxy — the TLS front door of the global
external ALB, the cross-region internal ALB, and Traffic Director. `region`
set builds the REGIONAL proxy — the TLS front door of the regional external
ALB and the regional internal ALB. Scope is a chain-wide decision: a
regional proxy's URL map must be a regional `GcpUrlMap`, its certificates
regional (`GcpSslCertificate` with `region`, attached with an explicit
`valueFrom.kind`, or regional Certificate Manager certificates), its
`sslPolicy` a regional `GcpSslPolicy`, and the forwarding rule in front a
regional `GcpGlobalForwardingRule`. Four levers exist only on the global
proxy and are rejected when `region` is set: `certificateMap` (SNI-scale
maps are a global external ALB feature), `quicOverride` (regional ALBs do
not negotiate QUIC), `tlsEarlyData`, and `proxyBind`. A proxy never moves
between scopes — `region` is immutable.

## Pick the certificate mechanism by scale and scheme

Exactly one of three (GCP rejects combinations):

- **`sslCertificates`** (1–15): classic compute certificates —
  `GcpManagedSslCertificate` for Google-managed issuance,
  `GcpSslCertificate` for bring-your-own. The default for external ALBs
  with a handful of domains.
- **`certificateMap`**: a Certificate Manager map that selects the served
  certificate by SNI hostname — the SaaS custom-domain mechanism, and
  mandatory past ~15 certificates. External ALBs only.
- **`certificateManagerCertificates`**: Certificate Manager certificates
  attached directly — only honored by the cross-region internal ALB
  (`INTERNAL_MANAGED`).

Traffic Director (`INTERNAL_SELF_MANAGED`) ignores client certificates
entirely — `serverTlsPolicy` is its only TLS lever, and also the mTLS
mechanism (demand and validate client certificates) for external ALBs.

## Zero-downtime certificate rotation

`sslCertificates` swaps in place (setSslCertificates). Rotation is:
attach the replacement alongside the old certificate, wait for it to
serve, then detach the old one. Never a destroy-and-recreate of the
proxy, never a moment with zero valid certificates.

## Set an sslPolicy — the default is permissive

An unset `sslPolicy` leaves GCP's default in charge: minimum TLS 1.0,
COMPATIBLE profile. Any compliance regime wants an explicit
`GcpSslPolicy` reference; it swaps in place, so hardening a live frontend
is non-disruptive.

## QUIC and early data are latency/replay trades

`quicOverride: NONE` lets Google manage HTTP/3 — the right default.
`tlsEarlyData` accepts TLS 1.3 0-RTT requests, which are replayable by
design: `STRICT` (safe methods, no query params) is the sane opt-in;
`PERMISSIVE`/`UNRESTRICTED` belong only in front of idempotent services.
Early data is immutable — decide it at create time.

## The mutation map

`urlMap`, all three certificate mechanisms, `sslPolicy`,
`serverTlsPolicy`, and `quicOverride` update in place — routing cutovers,
cert rotation, and TLS hardening are all zero-downtime edits. Name,
description, keep-alive, `tlsEarlyData`, and `proxyBind` are ForceNew,
and replacement must be create-before-destroy: GCP refuses to delete a
proxy a forwarding rule still references.

## Teardown discipline

`deletionPolicy: PREVENT` for any production TLS frontend; `ABANDON`
leaves the proxy terminating TLS unmanaged — a certificate that keeps
serving with nobody watching its renewal is a slow-motion outage.
