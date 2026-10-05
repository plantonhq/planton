# Overview

The AwsEksAddon API resource installs an EKS MANAGED add-on on an
`AwsEksCluster`: cluster software (vpc-cni, coredns, kube-proxy, the
EBS/EFS CSI drivers, Pod Identity agent, marketplace add-ons, ...)
whose install, version upgrades, configuration, and removal AWS manages
through the EKS control plane instead of Helm charts you operate
yourself.

## Why We Created This API Resource

Add-ons are per-cluster software with their own version lifecycle --
exactly the shape that deserves a first-class, composable node:

- **One add-on, one node**: a cluster runs many add-ons, each upgraded
  on its own schedule; modeling them individually makes the
  architecture graph show exactly what runs on which cluster, at which
  version.
- **Attach identity by reference**: the IAM role an add-on needs (the
  EBS CSI driver's, for example) is a referenced `AwsIamRole` -- wired
  through IRSA or EKS Pod Identity -- never an inline side effect.
- **Honest conflict handling**: clusters that bootstrap self-managed
  copies of the core add-ons need `OVERWRITE` at install time to adopt
  them; the spec models AWS's asymmetric create/update conflict rules
  exactly, so misconfiguration fails at validation, not mid-deploy.

## Key Features

### Version Lifecycle

- **AWS-default or pinned**: empty `addonVersion` installs the AWS
  default for the cluster's Kubernetes version (never goes stale);
  pinning gives byte-identical clusters and controlled upgrades.
- **Resolved version surfaced**: the `addon_version` output reports
  what is actually running either way.

### Identity Wiring

- **IRSA**: `serviceAccountRoleArn` binds the add-on's service account
  to a referenced role through the cluster's OIDC provider.
- **EKS Pod Identity**: `podIdentityAssociations` binds service
  accounts to roles with no per-cluster OIDC provider -- the modern
  path for new clusters.

### Operational Controls

- **Conflict resolution**: create-time `NONE`/`OVERWRITE` (adopting
  bootstrap self-managed installs) and update-time
  `NONE`/`OVERWRITE`/`PRESERVE` (drift handling), modeled as AWS
  defines them.
- **Configuration values**: each add-on's own JSON schema
  (`aws eks describe-addon-configuration`), passed through unmodified.
- **Preserve on delete**: hand the software back to cluster operators
  without an outage.
- **Custom namespace**: install into a non-default namespace
  (create-time immutable, as in AWS).

## Benefits

- **Composability**: cluster and IAM roles attach through `valueFrom`
  references; the add-on is a referenceable node in the environment
  graph.
- **Honest constraints**: the create/update conflict asymmetry and the
  version format are CEL-enforced at validation time.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `addon_arn`: the add-on's ARN
- `addon_name`: the EKS catalog name it was installed under
- `addon_version`: the version actually running (the resolved AWS
  default when the spec pinned nothing)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
