# DigitalOcean DNS Record -- Operational Guide

Judgment calls that matter when you run individual DNS records on DigitalOcean.

## Standalone records vs. zone-inline records

Two components can create the same record: this kind (one record, one resource) and the `DigitalOceanDnsZone` kind's inline `records` list. Use zone-inline records when one team owns the whole zone and its records ship together. Use this kind when records have different owners or lifecycles than their zone — an application chart adding its own hostname to a shared company zone is the canonical case. Mixing both against the same name works (DigitalOcean allows duplicate-name records of most types) but splits ownership; pick one home per record.

## Many records on one zone at once: apply them one after another

DigitalOcean's DNS API deadlocks when several record writes to the same domain are in flight together — one of the batch fails with `422 Error 1213 (40001): Deadlock found when trying to get lock; try restarting transaction` (measured: 1 in 8 parallel creates, on a fresh and on a settled zone alike), and the provider does not retry it. A single standalone record is never affected. A chart that deploys many `DigitalOceanDnsRecord` resources into one zone concurrently can be: the failed record simply re-applies cleanly, but the run is not. When a zone's records ship together under one owner, prefer the zone kind's inline `records`, whose modules write them one at a time; when they must stay standalone, sequence their deployment.

## Author hostname targets fully qualified

DigitalOcean reports CNAME/MX/NS/SRV/CAA (`issue`/`issuewild`) targets back fully qualified with a trailing dot, and the provider forgives exactly two manifest spellings: the same dotted form (`mail.example.com.`, `letsencrypt.org.`) or a name relative to the zone (`mail`). Author either and the manifest matches the stored record forever. Author a bare fully-qualified name without the dot (`letsencrypt.org`) and it is not cosmetic: every apply rewrites the record, DigitalOcean hands it back with the dot again, and the diff never settles. IP-valued records (A, AAAA), TXT, and CAA `iodef` (`mailto:` values) are unaffected.

## The explicit-zero trap

The provider validates that MX/SRV records carry `priority` (and SRV its `weight`/`port`), but silently drops a value of exactly 0 from the create request, letting the API default win. DNS semantics make 0 a legal, highest-preference MX priority — but the provider cannot actually deliver it. Use 1 as the top preference and the manifest will always match the zone. CAA `flags: 0` is exempt from the worry: the dropped 0 and the API default are the same value.

## TTL strategy, and why the live TTL can drift

Omit `ttlSeconds` to take DigitalOcean's default (1800). Lower it to 60–300 ahead of planned moves so caches empty quickly; raise it to 3600+ on stable records to cut resolver load. Note DigitalOcean harmonizes TTLs across all records sharing one fully-qualified name (RFC 2181 requires it): changing the TTL on one of two same-name A records changes both, and the provider surfaces the drift as a warning, not an error. Keep same-name records' TTLs identical in manifests to avoid churn.

## The apex is special

`name: "@"` addresses the zone apex. CNAME cannot exist at the apex (DNS forbids it); use an A/AAAA record there and CNAME only on subdomains. NS records at the apex belong to the zone's delegation — leave them to DigitalOcean unless you know exactly why you're overriding them, and treat SOA the same way (the API accepts SOA writes; the zone already has one).

## Deleting a zone deletes its records

Records live inside their zone: destroying the `DigitalOceanDnsZone` removes every record in it, including records this kind created. The verifier treats a vanished domain as a valid record-absent signal for exactly this reason. When a record must outlive infrastructure churn, its zone must too.
