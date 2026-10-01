# DigitalOcean Load Balancer -- Operational Guide

Judgment calls that matter when you run DigitalOcean load balancers.

## Size is the cost knob; pick one way to express it

`size` (`lb-small` / `lb-medium` / `lb-large`) and `sizeUnit` (1–200) describe the same capacity. `lb-small` is 1 unit, `lb-medium` is 3, `lb-large` is 6. Past `lb-large`, only `sizeUnit` applies. Carrying both is rejected before any provisioner runs. Unset means DigitalOcean provisions `lb-small`. Scale units, not Droplet count, are what you pay for — the bill scales with units. Verified per-preset figures live in the generated estimate at `catalog/_pricing/estimates/digitaloceanloadbalancer.yaml`, computed from the pinned, source-dated price book.

## Certificates are names, not UUIDs

DigitalOcean Let's Encrypt certificates rotate their UUID on every auto-renewal. The provider's `certificate_id` argument is deprecated for that reason; this component only models `certificateName`. A `DigitalOceanCertificate` reference resolves to `status.outputs.certificate_id`, which at the pinned provider is the certificate NAME — the stable handle. Never paste a certificate UUID into `certificateName`.

## Type decides the rest of the spec

- **REGIONAL** (the default when `type` is unset) and **REGIONAL_NETWORK** require a `region` and `forwardingRules`. They may take a `vpc`.
- **GLOBAL** forbids a `region` and `forwardingRules`. It routes through `glbSettings` (required), plus `domains` and `targetLoadBalancerIds` pointing at regional balancers. A GLOBAL balancer has no VPC.

The provider's own check allows region-without-type (it implies REGIONAL). This component mirrors that.

## Droplet IDs versus a tag

`dropletIds` is a fixed list. `dropletTag` is a living membership: every Droplet carrying the tag is attached, and membership follows the tag as Droplets come and go. They are mutually exclusive. Tag-based targeting is the right default for anything that scales; the explicit list is for a known, small set.

A balancer with neither is valid — it just has no backends. Useful for proving the balancer itself, not for serving traffic.

DigitalOcean accepts a `dropletTag` that no Droplet carries yet (measured: the create succeeds, the tag is stored as a selector, no tag object is created, and the member list stays empty) — the opposite of a DigitalOcean firewall, which rejects an unknown tag with a 422. So a tag-targeted balancer can be created before its Droplets exist; it simply serves nothing until a Droplet carrying the tag appears in the same region. Tagging the Droplets first is still the sensible order: it lets you confirm membership on the first apply instead of wondering why traffic goes nowhere.

## A balancer stuck in `new`

A new balancer normally reaches `active` in one to two minutes; the provider waits up to ten. Once in three identical creates on one day, a balancer never left `new` in that window (no incident on DigitalOcean's status page), the apply failed with `timeout while waiting for state to become 'active'`, and the same manifest succeeded on the next apply. Treat a single stall as DigitalOcean's, not the manifest's: destroy and re-apply. One after-effect to know about: deleting a balancer that never left `new` can leave it listed as a member of its VPC for a while after the balancer itself is gone, and the VPC refuses deletion (`409 Can not delete VPC with members`) until DigitalOcean clears the ghost — observed at about 45 minutes.

## Firewall rules use the provider's own prefix format

`firewall.allow` and `firewall.deny` are strings like `ip:203.0.113.5` or `cidr:10.0.0.0/8`. A bare address is rejected before apply. Deny a specific range and allow the rest, or allow a specific range and deny the rest — DigitalOcean evaluates both lists.

## Write-only knobs: network, network stack, TLS policy

`network` (`EXTERNAL` / `INTERNAL`), `networkStack` (`IPV4` / `DUALSTACK`), and `tlsCipherPolicy` (`DEFAULT` / `STRONG`) are never reported back by the API. Import leaves them empty. Drift on them is invisible. Set them at create time and treat them as create-intent, not as a source of truth you can re-read.

`network` and `networkStack` are also create-only (ForceNew). Changing them replaces the balancer.

## BYOIP and subnet placement

`ip` assigns an unassigned BYOIP address on the account at create time. When unset, DigitalOcean allocates one. The assigned address is always the `ip` stack output.

`subnetUuid` places the balancer in a DigitalOcean-managed VPC subnet and requires `vpc`. Both are create-only, and both deploy on either provisioner.

## Sticky sessions: cookies or none

`stickySessions.type: cookies` requires `cookieName` (2–40 characters) and `cookieTtlSeconds`. `type: none` forbids both and is the API default — set it only to assert "no affinity" explicitly. Leaving `stickySessions` unset also means no affinity.

## Importing an existing balancer

Import uses the bare balancer UUID (`load_balancer_id`). Expect `network`, `networkStack`, and `tlsCipherPolicy` to stay at their configured values after import: the API never reports them back. `size` / `sizeUnit` read back as whichever the API currently reports (the other stays at its prior state). A GLOBAL balancer's `failoverThreshold` only reads back when `regionPriorities` is also set.

## What is deliberately NOT here

The deprecated `algorithm` argument (DigitalOcean no longer lets you pick one) and the deprecated `certificate_id` leaf (use `certificateName`). Droplet firewalls and reserved IPs are separate resources.
