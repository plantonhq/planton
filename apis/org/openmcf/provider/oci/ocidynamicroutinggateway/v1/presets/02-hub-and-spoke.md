# Hub-and-Spoke

This preset creates a Dynamic Routing Gateway configured as a hub for multi-VCN networking. Two spoke VCNs are attached to a shared custom route table that imports routes from all VCN attachments via an import distribution, enabling spoke-to-spoke communication through the DRG. This is the standard OCI pattern for enterprise network segmentation where workload VCNs need controlled connectivity.

## When to Use

- Multi-VCN architectures where spoke VCNs need to communicate through a central hub
- Enterprise network segmentation with separate VCNs for production, staging, and shared services
- Organizations migrating from flat VCN designs to hub-and-spoke for better traffic control
- Environments that will later add on-premises connectivity (IPSec/FastConnect) alongside VCN peering

## Key Configuration Choices

- **Two spoke VCN attachments** sharing a single custom route table -- both spokes use the same routing policy, ensuring symmetric reachability. Add more spoke attachments as needed; all pointing to the same route table.
- **Custom route table** (`spoke-route-table`) with import distribution -- instead of relying on OCI's auto-generated default tables, a custom table imports routes from all VCN attachments, giving explicit control over which routes propagate between spokes.
- **Import route distribution** (`import-all-vcn-routes`) matching all VCN attachment types -- a single statement at priority 10 imports routes from every VCN attachment. This means each spoke automatically learns the CIDR blocks of every other spoke.
- **ECMP disabled** (`isEcmpEnabled: false`) -- Equal-Cost Multi-Path is only useful when multiple IPSec tunnels or virtual circuits provide redundant paths to the same destination. For VCN-to-VCN routing, ECMP is not applicable.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the DRG will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<spoke-vcn-a-ocid>` | OCID of the first spoke VCN | OCI Console > Networking > Virtual Cloud Networks, or `OciVcn` outputs |
| `<spoke-vcn-b-ocid>` | OCID of the second spoke VCN | OCI Console > Networking > Virtual Cloud Networks, or `OciVcn` outputs |

## Related Presets

- **01-single-vcn-attachment** -- Use instead when only one VCN needs DRG connectivity and hub-and-spoke routing is not required
