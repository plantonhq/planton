# GcpComputeMig

Deploys a Compute Engine **Managed Instance Group** — a self-healing,
optionally auto-scaling fleet of identical VMs behind one declarative
surface. One resource manages the whole stack: the instance TEMPLATE
(what each VM looks like), the GROUP MANAGER (how many run, how changes
roll out, how failed VMs are repaired), an optional AUTOSCALER (when the
fleet grows and shrinks), stateful PER-INSTANCE CONFIGS (name/disk/IP
overrides for individual instances), and queued RESIZE REQUESTS
(one-shot capacity asks for scarce shapes).

The group is ZONAL (all VMs in one zone) or REGIONAL (VMs spread across
a region's zones for higher availability) — exactly one of `zone` or
`region` selects the scope, and the module picks the matching provider
resource family kind-wide.

## The template rotates; the group rolls

Instance templates are immutable in GCP (labels excepted). Every change
to the `template` block creates a NEW template, repoints the group, and
deletes the old one — managed natively by both IaC modules through
name-prefix rotation, so the group is never left on a deleted template.
How running VMs pick the change up is `updatePolicy`: `PROACTIVE` rolls
the fleet automatically within the surge/unavailability budget;
`OPPORTUNISTIC` waits for repairs and manual refreshes.

## Composition

The `instance_group` stack output is the load-balancer backend handle: a
`GcpBackendService` backend's `group` takes exactly that value, which
plugs this kind into the modeled HTTPS-LB family (URL map, target
proxies, forwarding rule). Auto-healing references a `GcpHealthCheck`;
the template's NICs reference `GcpVpcNetwork`/`GcpSubnetwork`; disks and
CMEK reference `GcpComputeDisk`/`GcpKmsKey`; the VM identity references
`GcpServiceAccount`.

## Coverage

Built for 100% Terraform parity against the pinned `google` provider:
ten provider resources (`google_compute_instance_template` +
`_region_instance_template`, `google_compute_instance_group_manager` +
`_region_...`, `google_compute_autoscaler` + `_region_...`,
`google_compute_per_instance_config` + `_region_...`,
`google_compute_resize_request` + `_region_resize_request`) at total
accounting — every provider argument matched, mapped, or excluded with a
recorded reason in `iac/provider-parity.yaml`. Notable recorded
exclusions: customer-supplied raw encryption keys (CSEK — use CMEK) and
the template `name`/`name_prefix` pair (module-internal rotation
machinery). The template's managed workload identity
(`workloadIdentityConfig`: a SPIFFE ID per VM, optional X.509
certificates) and the host-error detection timeout
(`scheduling.hostErrorTimeoutSeconds`) are modeled on both engines.

## Docs

- `catalog.md` — the catalog page (what gets created, how to deploy)
- `GUIDE.md` — operational judgment (rollout strategy, stateful one-way
  doors, standby economics, quota)
- `v1alpha1/reference.md` — the generated spec reference
- `iac/pulumi/README.md`, `iac/tf/README.md` — module internals

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
