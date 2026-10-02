# Secure Tag Microsegmentation

Tier isolation without IP ranges: rules keyed on secure tags -- Resource
Manager tag values bound to the VMs -- so a VM's tier, not its address,
decides what it may reach. Frontend reaches backend on 8080, backend reaches
the database on 5432, nothing else reaches the database. Every tag is a
`GcpTagValue` in the same chart, referenced by its `name` output.

## What it configures

- Priority `1000` — `allow` tcp:8080 from VMs tagged `tier-frontend` to
  VMs tagged `tier-backend` (`srcSecureTags` and `targetSecureTags`).
- Priority `1100` — `allow` tcp:5432 from `tier-backend` to
  `tier-database`.
- Priority `2000` — `deny` all to `tier-database` from anywhere, logged.
- One association referencing the `GcpVpcNetwork` named `main`.

## Adjust before deploying

- **The three `GcpTagValue` names** — your tag values, under a `GcpTagKey`
  with `purpose: GCE_FIREWALL` scoped to the network; bind them to the VMs
  with `GcpTagBinding` or at create time.
- **Ports** — your tiers' ports.
- **`associations[].network.valueFrom.name`** — your VPC.

## When to choose something else

For a network's ordinary ingress baseline, start from **Global Baseline
Policy**; for rules that target an internal load balancer, from **Regional
ILB Policy**.
