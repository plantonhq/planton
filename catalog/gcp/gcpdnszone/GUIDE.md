# GcpDnsZone Guide

The judgment this guide protects: a DNS zone is delegated authority —
registrars and parent zones point at it, resolvers cache it, and every
mistake propagates at TTL speed. The zone shell should be boring; the
interesting decisions are visibility and destroy behavior.

## The zone owns the shell, records live elsewhere

This kind deliberately creates the zone WITHOUT records — GcpDnsRecord
owns those, one node per record set. Zones that embed their records
turn every application change into a zone change; the split keeps the
blast radius of "edit one A record" away from "the thing the registrar
delegates to".

## Delegation is the real go-live

Creating a public zone does nothing until the registrar points at the
zone's `nameservers` output — and once it does, those four NS hosts are
load-bearing. Recreating the zone hands out a DIFFERENT name-server set
in the general case, which silently breaks delegation until the
registrar is updated. Treat zone recreation like an IP change: planned,
announced, verified.

## Private visibility is a topology decision

`visibility: private` scopes the zone to the VPC networks listed in
`privateVisibilityConfig`; forwarding and peering configs are
private-only and mutually exclusive. Forwarding targets carry ONE
resolver address family each — IPv4 or IPv6, never both (the spec
enforces the API's rule pre-deploy) — plus `forwarding_path` to choose
whether queries travel via the VPC or the internet.

## DNSSEC: turn on deliberately, never toggle casually

Enabling DNSSEC without completing the DS handshake at the registrar
leaves you signed-but-unvalidated (harmless); DISABLING it while the DS
record still exists at the registrar breaks resolution for validating
resolvers — the dangerous direction. The key-spec `kind` fields the API
carries are vestigial identity markers; they are recorded exclusions,
not configuration.

## Teardown discipline: two levers, in order

`forceDestroy` empties the record sets; `deletionPolicy` decides the
zone shell. GCP refuses to delete a zone with non-default records, so
`DELETE` without forceDestroy fails safely on a still-populated zone.
`PREVENT` suits any delegated production zone. `ABANDON` keeps the zone
answering while dropping management — remember the registrar still
delegates to it, so an abandoned zone is live infrastructure someone
must still own.
