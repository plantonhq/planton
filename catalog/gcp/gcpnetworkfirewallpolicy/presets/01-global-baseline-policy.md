# Global Baseline Policy

The everyday network firewall policy: a GLOBAL policy on the `main` VPC
declared by a `GcpVpcNetwork` in the same chart, admitting only what every
workload needs -- Identity-Aware Proxy SSH, Google's health-check probers,
and the VPC itself -- then denying and logging every other ingress. It
replaces the handful of legacy `GcpFirewallRule`s a network usually starts
with, in one ordered list.

## What it configures

- `region` left empty — a global policy governing every region of the
  attached network.
- Priority `1000` — `allow` tcp:22 from `35.235.240.0/20` (IAP TCP
  forwarding), so SSH works without a public address.
- Priority `1100` — `allow` tcp from `35.191.0.0/16` and `130.211.0.0/22`,
  Google's health-check probers.
- Priority `1200` — `allow` all with `srcNetworkContext: INTRA_VPC`, the
  VPC talking to itself without listing its ranges.
- Priority `65000` — `deny` all from `0.0.0.0/0`, logged, so a blocked
  connection is explainable.
- One association referencing the `GcpVpcNetwork` named `main`.
- `deletionPolicy: PREVENT` — deleting a baseline reopens what it closed.

## Adjust before deploying

- **`associations[].network.valueFrom.name`** — the name of your
  `GcpVpcNetwork`; add an entry per network the policy should govern.
- **Add allows above `65000`** for the ports your services expose (or
  express them by secure tag, as in **Secure Tag Microsegmentation**).

## When to choose something else

For rules that target an internal managed load balancer, start from
**Regional ILB Policy**; for tag-keyed tier isolation, from **Secure Tag
Microsegmentation**.
