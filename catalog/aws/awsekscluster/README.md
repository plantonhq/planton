# Overview

The AwsEksCluster API resource provisions an EKS cluster CONTROL PLANE:
the managed Kubernetes API server, etcd, and the cluster-level posture --
networking exposure, authentication mode, secrets encryption,
control-plane logging, upgrade policy, and (optionally) EKS Auto Mode.

## Why We Created This API Resource

The cluster is deliberately only the control plane. Modeling it that way
-- compute attaches as separate `AwsEksNodeGroup` nodes, IAM trust flows
through referenced roles and the exported OIDC issuer -- lets you:

- **Operate the control plane and the fleets independently**: upgrade the
  control plane first, then roll each node pool on its own schedule; add
  and remove pools without ever touching the cluster resource.
- **Keep identity honest**: the cluster role is a referenced `AwsIamRole`
  that carries its own `AmazonEKSClusterPolicy` -- the cluster never
  reaches into a role it merely references. IRSA wires up by pointing an
  `AwsIamOidcProvider` at the exported `oidc_issuer_url`.
- **Choose the compute model explicitly**: explicit node groups, or EKS
  Auto Mode (AWS provisions compute, storage, and load balancing itself)
  -- one honest toggle, not three blocks to keep in lockstep.

## Key Features

### Networking Exposure

- **Endpoint pair**: independent public/private API endpoint toggles,
  with `public_access_cidrs` to scope a public endpoint down.
- **Control-plane egress mode**: route control-plane egress through your
  VPC (inspection/firewall architectures) or isolate it.
- **Address families**: `ipv4` or `ipv6` pod/service networking with a
  custom service CIDR.
- **Hybrid nodes**: declare the on-premises/edge node and pod CIDR
  ranges allowed to join over VPN/Direct Connect -- one cluster, cloud
  and on-prem capacity; ranges update in place.

### Security Posture

- **Access entries** (`authentication_mode: API`) -- the modern access
  model; IAM principals granted access as first-class EKS resources.
- **Envelope encryption** of Kubernetes secrets with a referenced
  `AwsKmsKey` (a one-way door, documented as such).
- **Granular control-plane logs**: choose exactly which of the five log
  types stream to CloudWatch (audit and authenticator carry the most
  signal per dollar).
- **Deletion protection** for shared/production control planes.

### Lifecycle

- **Upgrade support tier**: standard schedule or extended support (with
  its surcharge) -- an explicit, reviewable choice.
- **Provisioned control-plane scaling**: pre-provision API-server
  capacity (`tier-xl` through `tier-8xl`, billed hourly) for very large
  or bursty clusters; `standard` keeps EKS's free reactive scaling.
- **Zonal shift** for automatic traffic movement away from an impaired
  availability zone.
- **EKS Auto Mode**: AWS-managed compute, block storage, and load
  balancing -- the all-or-nothing trio expressed as one toggle.

## Benefits

- **Composability**: subnets, roles, security groups, and the KMS key
  attach through `valueFrom` references, and downstream nodes
  (`AwsEksNodeGroup`, `AwsIamOidcProvider`) compose onto this cluster's
  outputs.
- **Impossible states are unrepresentable**: Auto Mode's three AWS
  settings move together by construction; immutable fields and one-way
  doors are called out on the fields themselves.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `endpoint`: Kubernetes API server URL
- `cluster_ca_certificate`: base64 cluster CA (kubeconfig)
- `cluster_security_group_id`: the EKS-managed control-plane security group
- `oidc_issuer_url`: the IRSA trust anchor (feed an `AwsIamOidcProvider`)
- `cluster_arn`: cluster ARN (IAM policies, access entries)
- `name`: cluster name (what node groups reference)
- `platform_version`: EKS platform revision (e.g. `eks.12`)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
