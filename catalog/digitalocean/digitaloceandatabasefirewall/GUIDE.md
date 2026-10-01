# DigitalOcean Database Firewall -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Destroying the firewall OPENS the database

The loudest fact about this kind: "delete" is not a deletion. The rule set is a property of the cluster, and destroying this resource PUTs an EMPTY rule list -- after which the cluster accepts connections from anywhere, exactly as it did before the firewall existed. Never remove this resource as cleanup while the cluster lives; remove it only when the cluster goes with it (teardown order handles that), or replace it with the successor rule set in the same change.

## Exactly one rule set per cluster

DigitalOcean holds ONE trusted-sources list per cluster. Two of these resources pointing at the same cluster do not merge -- each apply overwrites the other's rules, and they will flap forever. Declare every trusted source in one resource per cluster, owned by one manifest.

## A read replica has its own firewall -- the primary's rules do not reach it

A replica is its own cluster to DigitalOcean: it has its own UUID (the replica kind's `replica_id` output), answers `GET /v2/databases/{replica_id}` on its own, and carries its own trusted-sources list, which starts EMPTY (measured 2026-09-17: a primary with one `ip_addr` rule, its fresh replica reading `rules: []`; a rule PUT on the replica's UUID landed on the replica alone and left the primary untouched). Firewalling the primary therefore leaves every replica open to the public endpoint with the shared credentials. Declare a second `DigitalOceanDatabaseFirewall` per replica, with `cluster` pointing at the replica -- `valueFrom: {kind: DigitalOceanDatabaseReplica, name: <replica>, fieldPath: status.outputs.replica_id}`, or the replica's UUID as a literal value -- and keep its rule list in step with the primary's.

## Prefer tags and references over IPs

A `tags` rule tracks every Droplet carrying the tag -- membership updates itself as fleets scale. References (`dropletIds`, `kubernetesClusterIds`, `appIds`) resolve live resource ids at deploy time. Literal IPs are for fixed points (bastions, offices); CIDRs for private ranges. If you find yourself editing IP lists weekly, the fleet wants a tag.

## The whole set replaces on every change

There is no per-rule add/remove -- each apply PUTs the full list. That makes review easy (the manifest IS the allowlist) and partial edits impossible. Rule rows carry server-assigned uuid/created_at noise that never causes diffs (set semantics).

## Broad CIDRs defeat the purpose

`0.0.0.0/0` as an ip_rule is accepted by DigitalOcean and re-admits the whole internet deliberately. Nothing stops you; everything should. The validation floor requires at least one rule -- it cannot require the rule be narrow.

## IPv4 only -- there is no IPv6 rule

DigitalOcean's database firewall refuses every IPv6 shape at apply -- an address and a prefix alike answer `422 invalid rule with type IP_ADDR because: invalid ip format` (measured 2026-09-17). The spec therefore rejects IPv6 in `ipRules` at validation, in seconds, instead of minutes into a deploy. A bare IPv4 address and its `/32` form are both accepted and read back exactly as written, so pick one shape and keep it: switching between them is a rule change.

## What is deliberately NOT here

A spec-side rule `type` field (the five typed lists make type/value mismatches unrepresentable -- the modules derive the type from which list a value came from), and per-rule lifecycle (the API has none).
