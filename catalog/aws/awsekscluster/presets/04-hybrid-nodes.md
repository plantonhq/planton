# EKS Hybrid Nodes Cluster

This preset creates a control plane that on-premises or edge machines
join as workers over your VPN or Direct Connect: one Kubernetes API,
cloud and on-prem capacity underneath it. AWS bills hybrid nodes per
vCPU only when they register -- declaring the networks costs nothing.

## When to Use

- Data-gravity workloads: pods must run next to on-prem data or
  hardware (factory floors, hospitals, GPUs you already own) under the
  same cluster and GitOps flow as cloud workloads
- Gradual cloud migrations that keep one scheduling domain while
  capacity moves
- Edge fleets that need a managed control plane without running one
  on-site

## Key Configuration Choices

- **`remoteNetworks.nodeCidrs`** -- the on-prem address blocks kubelets
  register from; a machine outside these ranges cannot join. Must be
  private-range (RFC1918 or CGNAT 100.64/10), routable from the VPC,
  and non-overlapping with it
- **`remoteNetworks.podCidrs`** -- where the on-prem CNI assigns pod
  addresses; set it when cluster components must reach those pods
  directly (admission webhooks, east-west traffic). Node-only setups
  can omit it
- **`endpointPrivateAccess: true`** -- hybrid nodes reach the API
  server over the private connection; the private endpoint is the
  load-bearing one
- **Ranges update in place** on a live cluster -- adding a site later
  is a plain update, not a replacement

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<cluster-name>` | Name for the cluster | Your environment naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<cluster-role-resource-name>` | Name of the AwsIamRole with `AmazonEKSClusterPolicy` | Your role manifest's `metadata.name` |
| `10.80.0.0/16` (example) | Private CIDR your on-prem machines' addresses live in — a real CIDR shape; the field's pattern rejects placeholders | Your network team / site subnet plan |
| `10.85.0.0/16` (example) | Private CIDR the on-prem CNI assigns pods from — same real-CIDR requirement | Your CNI configuration |

## Common Additions

- `controlPlaneScalingTier` (`tier-xl` ... `tier-8xl`) when a large
  hybrid fleet drives sustained API-server load (billed hourly)
- `kmsKeyArn` for customer-managed secrets encryption
- `AwsEksNodeGroup` resources for the cloud side of the fleet --
  hybrid and managed node groups compose on one cluster

## Related Presets

- **01-standard** -- cloud-only cluster with managed node groups
- **02-private-endpoint** -- the fully private control plane posture
