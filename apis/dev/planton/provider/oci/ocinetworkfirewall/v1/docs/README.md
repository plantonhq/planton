# OciNetworkFirewall — Design Notes

## Design Rationale

OciNetworkFirewall bundles the firewall appliance, its policy, and all policy sub-resources into a single component. This is the most complex OCI component in Planton, with 7 distinct resource types created from one manifest.

### Why bundle the policy with the firewall?

A network firewall without a policy cannot inspect traffic — it has no rules to evaluate. A firewall policy without a firewall has no appliance to enforce it. They are functionally inseparable. Bundling them ensures that the entire security posture is declared in one manifest, making it auditable and version-controllable as a unit.

### Why bundle policy sub-resources (address lists, services, etc.)?

Policy sub-resources exist only within a policy and reference each other by name. An address list has no meaning outside the policy that uses it; a security rule's condition refers to address lists and services by name. Managing them separately would require explicit policy references on every sub-resource and introduce name-resolution complexity across manifests. Keeping everything inline provides a self-contained security configuration.

### Why derive security rule priority from list order?

The OCI API uses integer priority for rule evaluation order. Explicit priority numbers are error-prone (gaps, conflicts, renumbering) and don't add value when rules are already in a list. Deriving priority from position (first rule = priority 1, second = priority 2, etc.) is the natural mental model — "rules are evaluated top to bottom" — and matches how operators think about firewall policies.

### Why use name-based cross-references instead of indices?

Security rules reference address lists, services, and URL lists by name (e.g., `sourceAddresses: ["internal"]`). Names are human-readable, self-documenting, and stable across re-ordering of the definition lists. Index-based references would be fragile and opaque.

### Why use separate Service and ServiceList?

Services define individual port definitions (e.g., "HTTPS is TCP 443"). Service lists group services for reuse (e.g., "web-traffic" = HTTP + HTTPS). This two-level model avoids duplicating port definitions across multiple rules while keeping individual services composable.

### Why exclude TLS inspection and decryption?

TLS inspection requires decryption profiles, mapped secrets (certificates), and decryption rules — a complex subsystem that most network firewall deployments do not use initially. Including it would significantly increase spec complexity for a minority use case. It can be added in a future version.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Bundle firewall + policy + sub-resources | Single audit unit; self-contained | Large manifests for complex policies |
| Priority from list order | Natural top-to-bottom mental model | Must reorder list to change priority |
| Name-based cross-references | Human-readable; stable | Names must be unique and consistent |
| Separate Service and ServiceList | Composable; DRY port definitions | Two types for service-related config |
| Exclude TLS inspection | Simpler v1 spec | Cannot do TLS termination/inspection |
| Exclude ICMP type/code (Applications) | Simpler v1; ICMP matching is niche | Cannot filter by ICMP message type |

## Resource Graph

```
OciNetworkFirewall
├── oci_network_firewall_policy (always, created first)
│   └── outputs: policy_id
├── oci_network_firewall_policy_address_list (0..N)
│   └── DependsOn: policy
├── oci_network_firewall_policy_service (0..N)
│   └── DependsOn: policy
├── oci_network_firewall_policy_service_list (0..N)
│   └── DependsOn: policy, services
├── oci_network_firewall_policy_url_list (0..N)
│   └── DependsOn: policy
├── oci_network_firewall_policy_security_rule (0..N)
│   └── DependsOn: policy, address_lists, services, service_lists, url_lists
└── oci_network_firewall (always, created last)
    ├── DependsOn: policy, security_rules
    └── outputs: firewall_id, ipv4_address
```

The creation order is enforced via `DependsOn` chains: policy first, then sub-resources, then security rules (which reference sub-resources by name), then the firewall appliance (which references the policy by OCID).

## Deferred from v1

- **Applications / Application Groups** — ICMP type/code matching. Low adoption for managed firewall use cases.
- **Decryption Profiles / Rules / Mapped Secrets** — TLS inspection and termination. Complex subsystem deferred to reduce v1 scope.
- **NAT Rules** — specialized NAT handling beyond the basic `natConfiguration`. Rarely needed.
- **Tunnel Inspection Rules** — VXLAN inspection for overlay networks. Specialized use case.
- **defined_tags / system_tags / freeform_tags** — auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags on both the firewall and the policy:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciNetworkFirewall` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
