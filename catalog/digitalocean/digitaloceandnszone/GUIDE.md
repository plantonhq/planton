# DigitalOcean DNS Zone -- Operational Guide

Judgment calls that matter when you run DNS zones on DigitalOcean.

## The zone works before the delegation does

Adding a domain to DigitalOcean hosts it instantly on ns1/ns2/ns3.digitalocean.com — queries against those servers answer immediately, which is what the E2E lanes verify. The public internet, though, resolves through whatever nameservers the registrar advertises, so nothing changes for real users until the registrar's NS delegation is updated to DigitalOcean's set (the `name_servers` output) and the old delegation's TTL expires. Plan cutovers in that order: create the zone, populate the records, verify against DigitalOcean's nameservers directly (`dig @ns1.digitalocean.com`), then flip the registrar.

## One zone name per all of DigitalOcean

Domain names are unique across every DigitalOcean account, not just yours. An "already exists" error on create usually means another account (a colleague's personal account, a previous company account) holds the domain — DigitalOcean support, not a retry, is the fix. The reverse also matters: deleting a zone releases the name for anyone to claim.

## Inline records vs. the standalone record kind

This kind's `records` list is for zones whose records ship as one unit under one owner — the company website, a product domain. The standalone `DigitalOceanDnsRecord` kind exists for records with different owners or lifecycles — an application chart adding its hostname to a shared zone. Both write the same DigitalOcean records; pick one home per record so ownership is never split.

## Records are written one at a time, on purpose

DigitalOcean's DNS API deadlocks when several record writes to one domain are in flight together (`422 Error 1213 (40001): Deadlock found when trying to get lock` — measured 1 in 8 parallel creates, fresh or settled zone alike), and the provider does not retry it. Both provisioners therefore serialize this kind's inline records: Pulumi through an explicit dependency chain, Terraform through a one-request-per-second client limit. A zone of N records applies in roughly 2N seconds and destroys the same way — slower than a parallel write, and the only shape that never fails at random. This is also why inline records are the better home for a zone's own record set than many standalone records applied at once.

## Multi-value entries are real round-robin

An entry with several `values` creates one record per value with the same name and type. For A records that is DNS round-robin — resolvers rotate through the addresses. It is not health-checked failover: a dead address keeps being served. Round-robin is fine for stateless redundancy; use a load balancer when you need health-aware traffic steering.

## Leave the apex NS and SOA alone

Every zone comes with DigitalOcean's own NS records at the apex and one SOA. The API lets you write both types, and there is almost never a reason to: rewriting apex NS records breaks delegation inside the zone, and the SOA is DigitalOcean's operational record. The one legitimate NS use is delegating a subdomain (`sub.example.com`) to other nameservers.

## Write hostname values with a trailing dot

DigitalOcean reports every CNAME, MX, NS, SRV, and CAA (`issue`/`issuewild`) value back fully qualified with a trailing dot, and the provider forgives exactly two spellings in a manifest: the same fully-qualified form with the dot (`mail.example.com.`, `letsencrypt.org.`) or a name relative to the zone (`mail`). A bare fully-qualified name without the dot (`letsencrypt.org`) is neither, so every run re-applies it and DigitalOcean hands it back changed again — a diff that never settles. Pick the dotted form for anything outside the zone; it is what DNS itself means by an absolute name.

## One TTL per name

DigitalOcean gives every record that shares a fully-qualified name one TTL (RFC 2181 §5.2) and rewrites stragglers server-side; the provider only warns. So when several records sit on one name — three A values at the apex, an apex MX beside an apex TXT — give them the same `ttlSeconds` or leave them all unset. A lone custom TTL on a shared name is overwritten under you and shows as a change on every run. The record `ipAddress` seeds (below) counts as an apex A record for this rule.

## `ipAddress` seeds a record nobody manages

The `ipAddress` field makes DigitalOcean create an apex A record at zone creation — and that record is invisible to this component forever after: it is not in `records`, later edits do not see it, and deleting the value from the manifest changes nothing live. It exists for migration compatibility. New zones should declare the apex record in `records`, where it is tracked, diffed, and updatable.

## Adopting an existing zone is safe, by design

Import uses the domain name (the `zone_name` output) for the zone and `{domain},{record_id}` for each inline record — the `record_ids` output carries every id, keyed by `<record name>-<record index>-<value index>`. The API never reports `ipAddress` back, so a fresh import leaves it empty in state while the manifest still carries it; the provider's only answer to that difference would be to destroy and recreate the whole zone, every record and the domain's resolution with it. Both provisioners refuse that trade: after creation they ignore changes to `ipAddress`, so the first apply after import plans no replacement, and editing the value later is a no-op rather than an outage.

## Destroying a zone destroys every record in it

Including records created by the standalone record kind and records added by hand in the control panel. Before destroying a shared zone, enumerate what lives there (`doctl compute domain records list <domain>`) — the blast radius is the whole domain's resolution.
