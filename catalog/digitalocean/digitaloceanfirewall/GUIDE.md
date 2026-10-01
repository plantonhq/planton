# DigitalOcean Firewall -- Operational Guide

Judgment calls that matter when you run DigitalOcean Cloud Firewalls.

## Target by tag; reserve IDs for fixtures and one-offs

Tag targeting (`tags`) follows Droplets automatically as they come and go, has no churn cost, and is how every tier should be wired in production. ID targeting (`dropletIds`) caps at 10 Droplets, must be edited on every membership change, and exists for the cases where a firewall genuinely protects one known machine — a bastion, a fixture, a snowflake. The same logic applies inside rules: `sourceTags` scales, `sourceDropletIds` pins.

One ordering rule comes with tag targeting: the tag must exist before the firewall does. DigitalOcean creates a tag the first time a Droplet (or volume) declares it, but a firewall naming a tag nobody carries yet is rejected outright (`422 tag <name> does not exist`) — firewalls never create tags. In a chart, deploy the tagged Droplets first and the firewall after; for a firewall that must exist before its first Droplet, target a tag some Droplet already carries (every Planton-managed Droplet carries its `planton-ai_*` label tags) and add the intended tag once a Droplet has created it.

## Write "all", never "1-65535"

The DigitalOcean API reports "all ports" as its own value, and the provider reads it back as the literal string `all`. A rule authored as `1-65535` deploys fine and then diffs forever, because the stored rule no longer matches the manifest. The same class applies to icmp: it has no ports, and any `portRange` set on an icmp rule is silently dropped on read — omit it.

## Rules are a set, and normalization moves whole rules

DigitalOcean hashes each rule as a unit. Any leaf the API normalizes (a port range rewritten to `all`, an empty source list dropped) changes the hash, and a plan then shows the entire rule as removed-and-re-added rather than modified. This is cosmetic — apply converges — but it is why this module and its manifests keep rules in the provider's canonical spelling: canonical input produces empty diffs.

## Egress policy is a tier decision

Web tiers need open egress (package mirrors, external APIs, DNS) and get `all` outbound. Data tiers do not: a database host whose egress is limited to DNS and HTTPS cannot stream its contents to an arbitrary endpoint, which turns a class of exfiltration into a firewall log entry. Decide egress per tier, not per account.

## The double-firewall trap

A Cloud Firewall filters at DigitalOcean's edge; a host firewall (ufw, nftables) filters on the Droplet. Images that enable a host firewall by default will block traffic the Cloud Firewall allows, and the symptom — connection timeouts on an explicitly allowed port — looks exactly like a Cloud Firewall bug. When an allowed port does not answer, check the host firewall before editing rules here.

## A firewall with no rules cannot exist, and one rule direction is enough

DigitalOcean rejects an empty rule set, and validation enforces it before any provisioner runs. An inbound-only firewall leaves all egress open at the DigitalOcean layer (the default when no outbound rules exist is deny — so in practice every real firewall carries at least one outbound rule too; the presets show both directions).

## Propagation is asynchronous

Rule changes apply to targets asynchronously: the firewall reports `waiting` while DigitalOcean pushes rules to each member, then `succeeded`. Automation that changes a rule and immediately probes the port races the propagation; give it seconds, not milliseconds. A `failed` status is never transient — investigate the pending-changes list in the control panel.
