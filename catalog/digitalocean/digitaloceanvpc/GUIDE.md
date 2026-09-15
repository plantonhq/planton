# DigitalOcean VPC -- Operational Guide

Judgment calls that matter when you run private networks on DigitalOcean.

## Plan the range once, or let DigitalOcean plan it forever

`ipRangeCidr` is immutable: the only way to change it is replacing the VPC, which means evacuating every Droplet, cluster, load balancer, and database inside it first. Decide up front which world you are in. If nothing will ever peer with this network or VPN into it, omit the range — DigitalOcean picks a non-conflicting block and the `ip_range` output tells you what it chose. The moment corporate ranges, site-to-site VPNs, or VPC peering enter the picture, choose the range yourself and record the allocation, because DigitalOcean's auto-assigned 10.x blocks will eventually collide with somebody's office network.

## Sizing inside /16–/24

DigitalOcean accepts prefixes from /16 (65,536 addresses) down to /24 (256). A /24 sounds roomy for a dozen Droplets, but managed resources quietly consume addresses too — every DOKS node, load balancer, and database cluster member takes one. Kubernetes clusters are the heavy consumer: autoscaling to twenty nodes eats twenty addresses. A /20 (4,096) is a comfortable default for an environment; reserve /24s for genuinely small, fixed-size networks.

## The region's default VPC: never yours by choice, sometimes yours by accident

Each region has a default VPC; resources created without an explicit network land there. The flag is not settable through this kind or either provisioner. Treat the default VPC as the untyped landing zone and this kind's VPCs as the deliberate ones: always wire the `vpc` reference on Droplets, clusters, load balancers, and databases explicitly, and membership never depends on which VPC happens to be the regional default.

The trap is a region that has no VPC yet — a fresh account, or a region you have never used. DigitalOcean does not pre-create defaults; it makes the FIRST VPC created in that region the default, and default VPCs cannot be deleted (the API answers `403 Can not delete default VPCs`). If that first VPC was declared through this kind, its destroy fails and the VPC stays behind, still the default. Recovery is an account-level action outside infrastructure-as-code: create a plain replacement in that region (`default-<region>` is the name DigitalOcean itself would have used), promote it with the API (`PATCH /v2/vpcs/{id}` with `"default": true` — the one write the flag does accept), then delete the stranded one. Better: before the first deliberate VPC in a new region, let DigitalOcean create its default by creating any resource there without a `vpc`, or seed `default-<region>` yourself the same way.

## One region, no bridges — plan for it

A VPC exists in exactly one region, and members can only join from that region. Cross-region private connectivity does not come from this kind: it comes from VPC peering (a separate resource DigitalOcean offers) or from routing over the public network with TLS. If a workload spans regions, design the split now — one VPC per region with non-overlapping ranges, so future peering stays possible.

## Destroy members first; the VPC goes last

DigitalOcean refuses to delete a VPC that still contains resources, and the module's delete retries only paper over short races (a Droplet mid-destroy), not real membership. Tear environments down in dependency order — workloads, then load balancers and databases, then the VPC. The same applies in reverse for creation, which is why other kinds' E2E lanes install this VPC as their first fixture.
