# GcpCertManagerCert Guide

The judgment this guide protects: Certificate Manager is the certificate
stack you choose when the classic ones run out of road — wildcards,
zero-downtime migration, regional and cross-region serving, private PKI,
backend mTLS. The price is one more decision: how domain control is
proven.

## Choosing the authorization mode (managed arm)

- **DNS authorizations** (first-class `GcpCertManagerDnsAuthorization`
  references): the strongest mode. Validates BEFORE serving — the
  zero-downtime migration path — and the only mode that can issue
  wildcards. One authorization covers a domain and its wildcard.
- **Load-balancer authorization** (omit both `dnsAuthorizations` and
  `issuanceConfig`): simplest — no DNS records to manage — but validates
  only once traffic reaches the load balancer, so it shares the classic
  managed certificate's serving-gap problem and cannot do wildcards.
- **Issuance config** (a `GcpCertManagerIssuanceConfig` reference): your
  private Certificate Authority Service pool issues and renews the
  certificate instead of a public CA -- the private-PKI path for internal
  load balancers, wildcards included. Mutually exclusive with DNS
  authorizations.

The spec validates the coherence rules before deploy: exactly one of
managed/self_managed, wildcards require DNS auth or an issuance config,
DNS auth XOR issuance config.

## The self-managed arm rotates IN PLACE — use that

Unlike `GcpSslCertificate` (fully ForceNew), this resource PATCHes
`pemCertificate`/`pemPrivateKey` updates: rotation is a spec update, no
replacement, no proxy repointing. If you operate self-managed
certificates and rotation pain is real, migrating them into this kind
deletes the whole create-before-destroy dance. The private key is
sensitive: encrypted in Pulumi state, encrypted with the whole OpenTofu
state file when it has an encryption key, and never in outputs.

## Scope and location are create-time decisions

`scope` and `location` are ForceNew, so serving geometry is baked in:
`DEFAULT` for core load balancing (choose this if unsure), `ALL_REGIONS`
for cross-region internal ALBs (global certificates only — validated
pre-deploy), `EDGE_CACHE` for Media CDN, `CLIENT_AUTH` for the
certificate a load balancer PRESENTS to an mTLS backend. Regional
certificates pair with regional load balancers and regional DNS
authorizations.

## Conventions and gotchas

- `managed_state` reaching ACTIVE is the serving signal; the e2e bar for
  a freshly-issued DNS-auth certificate is PROVISIONING (issuance takes
  minutes after the validation record propagates).
- Renewal reuses the same authorization mode — keep DNS authorizations
  and their CNAME records alive for the certificate's whole life, not
  just issuance.
- `deletionPolicy: PREVENT` guards a certificate whose replacement is
  not yet serving; `ABANDON` hands it off mid-migration.

## Pairs well with

- `GcpCertManagerDnsAuthorization` — the pre-issue validation node this
  kind references.
- `GcpDnsRecord` — carries the authorization's CNAME into the zone.
- `GcpTargetHttpsProxy` — consumes Certificate Manager certificates for
  serving.
