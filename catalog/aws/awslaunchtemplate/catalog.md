# AWS Launch Template

Deploys an EC2 launch template — the reusable blueprint that describes how to launch a machine: AMI, instance type (or attribute-based requirements), storage, networking posture, IAM identity, metadata-service hardening, and purchase options. The template is the composition anchor of EC2 fleet compute: auto-scaling groups, EKS managed node groups, AWS Batch compute environments, and one-off launches all reference it, which is why it is a first-class resource with its own lifecycle.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Launch Template** -- the named, versioned blueprint. Versions are immutable in AWS: every spec change publishes a NEW version, and the module promotes it to the template's default so consumers following `$Default` pick it up on their next launch or instance refresh
- **Block Device Mappings** -- attached only when `blockDeviceMappings` entries exist; each entry reshapes the AMI's root volume, attaches a data volume, maps instance store, or suppresses an AMI-baked device
- **Network Interface Specifications** -- attached only when `networkInterfaces` are declared; explicit interfaces control public-IP association, static addressing, prefix delegation, multiple NICs, EFA, connection-tracking timeouts, ENA Express (SRD), and Wavelength carrier IPs; `secondaryInterfaces` adds launch-time interfaces for multi-homed instances
- **IMDS Configuration** -- attached only when `metadataOptions` is set; `httpTokens: required` enforces IMDSv2 on every launch
- **Purchase Market Configuration** -- attached when `spotOptions` (Spot requests) or `marketType: capacity-block` (pre-purchased ML Capacity Blocks) is set; `capacityReservation` targets On-Demand Capacity Reservations for reserved fleets
- **Launch-Time Tag Specifications** -- always attached; the platform's identity tags render onto the instances and volumes each launch creates, since template tags do not propagate on their own

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An AMI in the template's region** -- AMI IDs are region-specific. Optional at the template level, but auto-scaling groups require the template they reference to carry one; leave it unset only for consumers that inject their own image (EKS node groups, EC2 Fleet).
- **An IAM instance profile** (recommended) -- the launched machines' identity for SSM access, ECR pulls, and every AWS API call. Reference an AwsIamInstanceProfile Infra Component or pass a literal profile ARN.
- **Security groups in the target VPC** -- the firewall posture every launch inherits. With explicit network interfaces, groups attach per interface instead.

## Deploy

### Console

Open the deployment store, find **AWS Launch Template**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Server Fleet** preset in the [Presets](#presets) tab to pre-populate a working golden-template configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLaunchTemplate
metadata:
  name: web-fleet-base
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: amzn2023 + IMDSv2 + gp3
  imageId: ami-0abcdef1234567890
  instanceType: t3.small
  instanceProfile:
    valueFrom:
      kind: AwsIamInstanceProfile
      name: web-fleet-profile
      fieldPath: status.outputs.instance_profile_arn
  metadataOptions:
    httpTokens: required
    httpPutResponseHopLimit: 2
  blockDeviceMappings:
    - deviceName: /dev/xvda
      ebs:
        volumeSizeGb: 30
        volumeType: gp3
        encrypted: true
```

```shell
planton apply -f launch-template.yaml
```

This publishes the golden web-fleet blueprint: IMDSv2 enforced, an encrypted gp3 root volume, and the fleet's IAM identity — every group and node group built on it inherits the posture. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the template to resources deployed in the same InfraPipeline:

```yaml
spec:
  instanceProfile:
    valueFrom:
      kind: AwsIamInstanceProfile
      name: web-fleet-profile
      fieldPath: status.outputs.instance_profile_arn
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: web-fleet-sg
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the role, profile, and security group first, then provisions the template with the resolved values.

## Key Configuration

These are the most important decisions when configuring a launch template. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Everything is a launch-time default** -- A consumer (an ASG override, an EKS node group, a RunInstances call) may override any field, and every UNSET field is decided by the consumer or the AWS account default. That is how a golden template is built: set the opinionated parts (IMDSv2, encrypted volumes, the instance profile, monitoring), leave the workload-specific parts open.

**Exact type XOR attribute-based requirements** -- Name one instance type, or describe the compute you need (`instanceRequirements`: memory and vCPU ranges, CPU manufacturers, accelerators, price protection) and AWS resolves the matching types at launch. Attribute-based selection is the foundation of Spot diversification — a fleet drawing from dozens of matching pools rides out capacity events that would starve a single-type group. The two are mutually exclusive; both may stay empty when the consumer supplies the type.

**IMDSv2 enforcement** -- `metadataOptions.httpTokens: required` is the single most effective EC2 hardening: the metadata endpoint serves the fleet's IAM credentials, and requiring session tokens blocks the classic SSRF credential-theft path. Set the hop limit to 2 when containerized workloads need metadata access.

**Root-volume override** -- One `blockDeviceMappings` entry on the AMI's root device name (`/dev/xvda` for Amazon Linux, `/dev/sda1` for Ubuntu) grows the boot volume, switches it to gp3, and encrypts it — the classic golden-template mapping. Unset fields inherit from the AMI's own mapping.

**Security groups vs explicit interfaces** -- Security groups on `securityGroupIds` attach to every launch's primary interface. Declaring ANY `networkInterfaces` entry moves security groups onto the interfaces (AWS ignores the top-level list once interfaces exist). A subnet set on an interface pins every launch — fleet templates leave placement to the group.

**Termination protection is a fleet trap** -- `disableApiTermination` on a template blocks the auto-scaling group from ever scaling in: a protected fleet only grows. Use the group's own scale-in protection for graceful-drain semantics instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamInstanceProfile** | `instanceProfile` | `status.outputs.instance_profile_arn` |
| **AwsSecurityGroup** | `securityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsSecurityGroup** (per interface) | `networkInterfaces[].securityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsSubnet** (per interface) | `networkInterfaces[].subnetId` | `status.outputs.subnet_id` |
| **AwsKmsKey** (per EBS mapping) | `blockDeviceMappings[].ebs.kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `launch_template_id` | The template ID (lt-…) | AwsAutoScalingGroup `launchTemplate.launchTemplateId`, EKS node groups, Batch compute environments |
| `launch_template_arn` | Amazon Resource Name of the template | IAM policies scoping `ec2:RunInstances` to approved templates |
| `latest_version` | The newest immutable version number | Version pinning and rollout auditing |
| `default_version` | The version `$Default` consumers launch from | Confirming a rollout reached the fleet |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Golden web-server template** -- AMI + type + instance profile + IMDSv2 + encrypted gp3 root: the baseline every web fleet launches from. Start from the **Web Server Fleet** preset.

**Spot-flexible blueprint** -- Attribute-based requirements (memory/vCPU ranges, current generation, no bare metal) instead of a named type — the shape a diversified Spot fleet draws pools from. Start from the **Spot-Flexible Workers** preset.

**Hardened pet** -- Stop/termination protection, a customer-managed KMS key on the root volume, and tags in metadata — for standalone instances that must survive fat fingers. Start from the **Hardened Baseline** preset.

**Capacity Block training fleet** -- `marketType: capacity-block` plus the reservation target, with a dataset volume pre-warmed from its snapshot at a paid initialization rate — GPU training that starts on time and reads at full speed inside the paid window. Start from the **Capacity Block ML Training Fleet** preset.

## Works With

- [**AWS Auto Scaling Group**](/infra-catalog/aws-auto-scaling-group) -- the fleet manager that launches from this template and rolls it out on version changes
- [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) -- the fleet's identity for SSM, ECR, and every AWS API call
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- the firewall posture every launch inherits
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed custody for encrypted EBS volumes
- [**AWS EKS Node Group**](/infra-catalog/aws-eks-node-group) -- launches worker nodes from a template that carries no AMI (EKS injects its own)
