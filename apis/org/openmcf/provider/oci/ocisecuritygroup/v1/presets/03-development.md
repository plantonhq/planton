# Development NSG

This preset creates a fully permissive Network Security Group that allows all inbound and outbound traffic on all protocols and ports. This is the simplest NSG configuration, suitable for development, testing, proof-of-concept work, and ephemeral environments where convenience matters more than network security. Do not use this preset for production workloads.

## When to Use

- Development and testing environments where fine-grained firewall rules add unnecessary friction
- Proof-of-concept deployments that need unrestricted connectivity to validate functionality
- Ephemeral environments spun up and torn down frequently where security posture is not a concern
- Learning and experimentation with OCI networking where you want to rule out NSG rules as a source of connectivity issues
- Debugging connectivity problems by temporarily attaching a permissive NSG to isolate whether the issue is firewall-related

## Key Configuration Choices

- **All inbound** (`protocol: all` from `0.0.0.0/0`) -- Permits all traffic from any source on any protocol and port. This is intentionally wide open for development convenience. In production, use the web-tier or private-backend preset instead.
- **All outbound** (`protocol: all` to `0.0.0.0/0`) -- Permits all outbound traffic. Combined with the permissive ingress rule, this NSG imposes zero firewall restrictions.
- **Stateful rules** (`stateless` not set, defaults to `false`) -- All rules are stateful, keeping the behavior consistent with the production presets so that configurations can be promoted from dev to production by swapping the NSG preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NSG will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this NSG belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |

## Related Presets

- **01-web-tier** -- Use instead for production internet-facing resources with HTTP/HTTPS-only ingress
- **02-private-backend** -- Use instead for production backend resources that should only accept traffic from within the VCN
