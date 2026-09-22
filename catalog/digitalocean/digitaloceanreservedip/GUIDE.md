# DigitalOcean Reserved IP -- Operational Guide

What experience with this component teaches that the field reference cannot.

## The idle state is the expensive state

Most orphaned cloud resources cost nothing; a forgotten UNASSIGNED reserved IPv4 accrues a monthly charge forever. Assignment suspends the charge entirely. If a reservation stops being useful, destroy it -- "keep it around just in case" is the one posture this kind punishes.

## The address is the asset -- and destroy releases it permanently

The whole point of reserving is that the address survives droplet churn: DNS points at the reservation, droplets come and go behind it. The flip side: destroying the reservation releases the address to DigitalOcean's pool, and recreating gets a DIFFERENT one -- every DNS record naming the old address goes stale. Destroy is a DNS event; plan it as one.

## Same region or no assignment

A reserved IP assigns only to droplets in ITS region. Cross-region failover is not what this kind does (that is DNS or a global load balancer); this is intra-region droplet swapping. Pin the reservation's region to where its droplets actually live.

## Re-pointing is the failover move -- but not through the six-phase runner

Assign, re-point, and unassign all apply in place on IPv4 (the address never changes). In practice: run a standby droplet, and failover is a one-field manifest change. Note the v6 flavor re-points by replacing its assignment object -- same one-field edit in the manifest, one extra resource turn under the hood.

## An IPv6 create can fail with "inconsistent result" -- and leave the address behind

The provider reserves the IPv6 address and reads it straight back; when DigitalOcean's index has not caught up yet, that read answers 404 and the provider treats it as "gone", so Terraform reports `Provider produced inconsistent result after apply: root object was present, but now absent` -- while the reservation EXISTS, unrecorded in state (measured 2026-09-17: three consecutive applies hit it about seven seconds apart, then the fourth passed; by API alone it did not reproduce, so it is an intermittent lag, not the rule). IPv6 reservations are free, but an orphan still holds an address. When you see that error: `doctl compute reserved-ipv6 list`, then either `tofu import` the address into `digitalocean_reserved_ipv6.ipv6[0]` (the module's address) and apply again, or delete it and apply again. Pulumi previews without refresh and its create path did not hit the lag in proof. The upstream fix is a readability wait inside the provider's create, the same shape DigitalOcean's provider already gave Kafka topics; tracked as [digitalocean/terraform-provider-digitalocean#1607](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1607) (upstream PR #1600 adds the wait to the IPv4 resource only).

## IPv6 delete errors are swallowed upstream -- and an out-of-band unassign trips the destroy

The provider's v6 delete ignores every error except 404 (an inverted error check, verified in source at the pin) -- a failed release can look like a success in IaC output. The proof lane's live absence check is the trustworthy signal; operators cleaning up by hand should verify with `doctl compute reserved-ipv6 list`. One more edge: the v6 assignment's destroy reads the reservation and dereferences its attached droplet without checking for none, so an address someone unassigned in the control panel first makes the destroy crash instead of clean up. Recover by unassigning nothing further and removing the assignment from state (`tofu state rm`) or re-assigning the address before destroying.

## Adopting an existing reservation works on both families

Both reservations import by the bare address, and the v6 assignment imports by `{address},{droplet_id}` -- the provider parses the pair and re-derives its own id, so the timestamp in that id is no obstacle (live-proven 2026-09-17: every scenario re-imported blind with a clean plan, the v6 lane bringing both resources back). An assigned v4 address reads its droplet back on import, so the manifest's `droplet` matches without any tolerance.

## What is deliberately NOT here

`ip_address` as an input (the provider declares it but never sends it -- the address is an OUTPUT of reservation); the standalone v4 assignment resource (all-ForceNew with a timestamped id that can never round-trip an import -- the reservation's own mutable argument is strictly better); and project placement (the DigitalOceanProject kind's membership list carries the reservation's `urn`).
