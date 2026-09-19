# Regional ILB Policy

A REGIONAL policy whose rules target an internal Application Load Balancer
instead of VMs -- the one firewall shape that exists only at regional scope.
Office ranges may reach the load balancer on 80 and 443; everything else is
denied and logged.

## What it configures

- `region: us-central1` — the regional resource family; rules with
  `targetType: INTERNAL_MANAGED_LB` are regional by Google's design.
- Priority `1000` — `allow` tcp:80/443 from the office ranges, applied to
  the forwarding rule named in `targetForwardingRules`.
- Priority `2000` — `deny` all from `0.0.0.0/0` to the same load balancer,
  logged.
- One association referencing the `GcpVpcNetwork` named `main`.

## Adjust before deploying

- **`targetForwardingRules[].value`** — the self-link of your internal
  load balancer's regional forwarding rule; reference the forwarding-rule
  block by `valueFrom` once your chart declares it regionally.
- **`srcIpRanges`** on priority `1000` — the ranges that may reach the
  load balancer.
- **`region`** and **`associations[].network.valueFrom.name`** — your
  region and VPC.

## When to choose something else

For a network's ordinary ingress baseline, start from **Global Baseline
Policy**; for tag-keyed tier isolation, from **Secure Tag Microsegmentation**.
