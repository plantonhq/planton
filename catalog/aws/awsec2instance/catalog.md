# AWS EC2 Instance

Deploys a single EC2 virtual machine -- the pet of EC2 compute: a bastion, a license server, a singleton stateful workload. The kind covers the full instance surface: the launch source (an AMI + instance type, or a launch template with inline overrides), network placement, IAM identity, storage reshaping, IMDS hardening, purchase options (On-Demand or Spot), capacity reservations, placement, and lifecycle protections. For fleets, compose AwsLaunchTemplate + AwsAutoScalingGroup instead -- this kind deliberately shares the launch template's vocabulary so a pet can graduate into a templated fleet without relearning field names.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EC2 Instance** -- launched from the inline AMI + instance type or the referenced launch template, with the configured network identity, storage, posture, and protections
- **Block Device Mappings** -- the root volume reshaped from the AMI's mapping plus any launch-time EBS data volumes and instance-store mappings
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

The instance ATTACHES referenced first-class nodes -- a subnet, security groups, an IAM instance profile, a launch template, KMS keys -- and creates none of them.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A VPC with at least one private subnet** where the instance will be placed. Provide the subnet ID directly or reference an AwsSubnet Infra Component via ValueFromRef. (Unset launches into the account's default VPC -- fine for experiments, not a production posture.)
- **At least one security group** controlling inbound/outbound traffic. Provide group IDs directly or reference AwsSecurityGroup Infra Components.
- **An AMI ID** for the desired operating system (e.g., Amazon Linux 2023, Ubuntu) -- unless a launch template supplies one.
- **An IAM instance profile** whose role carries `AmazonSSMManagedInstanceCore` for keyless SSM Session Manager access (the modern posture). Provide the profile NAME (the instance API takes the name, not the ARN) or reference an AwsIamInstanceProfile Infra Component.
- **An existing EC2 key pair** only if you want SSH access -- leave `keyName` unset for keyless instances.

## Deploy

### Console

Open the deployment store, find **AWS EC2 Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **SSM-Managed Hardened Instance** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEc2Instance
metadata:
  name: backend-server
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  ami: ami-0abcdef1234567890
  instanceType: t4g.small
  subnetId:
    value: "subnet-0a1b2c3d4e5f00001"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
  instanceProfile:
    value: "ssm-managed-profile"
  metadataOptions:
    httpTokens: required
  rootBlockDevice:
    volumeType: gp3
    encrypted: true
  disableApiTermination: true
```

```shell
planton apply -f ec2-instance.yaml
```

This creates a hardened instance in a private subnet with keyless SSM access, IMDSv2 enforced, an encrypted gp3 root volume, and termination protection. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a subnet, security group, and instance profile deployed in the same InfraPipeline:

```yaml
spec:
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: private-a
      fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: app-sg
        fieldPath: status.outputs.security_group_id
  instanceProfile:
    valueFrom:
      kind: AwsIamInstanceProfile
      name: ssm-profile
      fieldPath: status.outputs.instance_profile_name
```

The InfraPipeline resolves the dependency graph, deploys the subnet, security group, and instance profile first, then provisions the EC2 instance with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EC2 instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Launch source** -- Configure `ami` + `instanceType` directly, or launch from an AwsLaunchTemplate (`launchTemplate.id` by reference, or `launchTemplate.name` for templates managed outside the graph). Every inline field OVERRIDES the template's value -- the template is the org's golden baseline, the inline fields are this pet's deviations. Pin `launchTemplate.version` to a number for reproducibility; `$Latest` restarts the instance on every template publish.

**Access posture** -- Attach an `instanceProfile` whose role carries `AmazonSSMManagedInstanceCore` for audited, keyless shell access through SSM Session Manager -- no bastion, no port 22, no key distribution. Set `keyName` only when SSH is genuinely needed; the key pair must already exist and changing it replaces the instance.

**IMDS hardening** -- Set `metadataOptions.httpTokens: required` to enforce IMDSv2, the single most effective hardening against credential-stealing SSRF attacks. Set `httpPutResponseHopLimit: 2` when containers on the instance need metadata access.

**Network identity** -- Place the instance inline (`subnetId`, `securityGroupIds`, optional pinned `privateIp`, IPv6) or attach a pre-provisioned ENI as eth0 (`primaryNetworkInterfaceId`) to inherit a fixed IP, MAC, and security groups that survive instance replacement. The two are mutually exclusive -- the ENI carries the whole network identity, including `sourceDestCheck`. Leave `sourceDestCheck` unset to keep AWS's default (checking ON); an explicit `false` is the router/NAT posture for instances that forward traffic.

**Storage** -- `rootBlockDevice` reshapes the boot disk the AMI defines (unset fields inherit): grow it, switch it to `gp3` (up to 2000 MiB/s throughput), encrypt it with a customer-managed AwsKmsKey. `ebsBlockDevices` attaches launch-time data volumes (create-time mappings -- prefer standalone volumes when data should outlive the instance); `ephemeralBlockDevices` maps instance-store disks or suppresses AMI-baked devices. Tag volumes one of two ways: `volumeTags` applies uniform tags to EVERY volume at creation (including AMI-inherited ones -- the shape ABAC/SCP tag-at-creation policies require), or per-device `tags` on each mapping for per-volume values (applied after creation). The two are mutually exclusive.

**Purchase and placement** -- Leave `marketType` and `spotOptions` unset for On-Demand (the pet posture). For interruption-tolerant standalone workloads, a persistent Spot request with `instanceInterruptionBehavior: stop` survives reclaims. `marketType: capacity-block` launches into a pre-purchased ML Capacity Block (GPU training windows); `interruptible-capacity-reservation` launches into an interruptible reservation -- both require targeting the reservation via `capacityReservation`. `capacityReservation.preference: capacity-reservations-only` guarantees the launch lands in reserved capacity or fails. `placement` pins the AZ, placement group, or tenancy. All fixed at launch.

**Lifecycle protections** -- Set `disableApiTermination: true` for production pets (the wizard preselects it). Pair it with `forceDestroy: true` to keep teardown declarative: destroy lifts the protection itself instead of failing. `userDataReplaceOnChange` decides whether a user-data edit replaces the instance (fresh boot) or stop-updates it (the changed script does NOT re-run -- cloud-init runs once per instance).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetId`, per-row `secondaryNetworkInterfaces[].subnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsIamInstanceProfile** | `instanceProfile` | `status.outputs.instance_profile_name` |
| **AwsLaunchTemplate** | `launchTemplate.id` | `status.outputs.launch_template_id` |
| **AwsKmsKey** | `rootBlockDevice.kmsKeyId`, per-row `ebsBlockDevices[].kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_id` | EC2 instance identifier (e.g., `i-0123456789abcdef0`) | AwsLbTargetGroup instance targets, monitoring, alarm dimensions |
| `arn` | The instance ARN | IAM policies and EventBridge rules scoped to this instance |
| `instance_state` | Lifecycle state as of the last deploy (`running`, `stopped`, ...) | Operational dashboards |
| `availability_zone` | AZ the instance runs in (e.g., `us-west-2a`) | Placement constraints for related resources |
| `private_ip` | Primary private IPv4 address | Service discovery, application configuration |
| `private_dns` | Internal DNS hostname within the VPC | DNS-based service references |
| `public_ip` | Public IPv4 address, when one is associated (changes across stop/start -- compose an AwsElasticIp for stability) | External access for public-facing instances |
| `public_dns` | Public DNS hostname, when a public address is associated | External DNS references |
| `primary_network_interface_id` | The eth0 ENI ID | Elastic IP associations, flow-log scoping |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**SSM-managed hardened pet** -- keyless SSM access through an instance profile, IMDSv2 enforced, encrypted gp3 root, termination protection ON. Connections are brokered through the AWS control plane with full audit logging. Start from the **SSM-Managed Hardened Instance** preset.

**Template-backed instance** -- launch from the org's AwsLaunchTemplate golden baseline and override only what differs for this machine. The template carries the hardening; the instance carries its identity. Start from the **Launch-Template Instance** preset.

**Spot worker** -- a persistent Spot request with `stop` interruption behavior for interruption-tolerant standalone workloads (build agents, batch boxes): AWS reclaims with two minutes' notice, then restarts the machine when capacity returns. Start from the **Spot Worker Instance** preset.

**Static-identity appliance** -- attach a pre-provisioned ENI (`primaryNetworkInterfaceId`) so the firewall/NAT/license-server keeps its fixed private IP and MAC across instance replacements. Start from the **Static Network Identity (Pre-Provisioned ENI)** preset.

**Capacity-block training node** -- launch a GPU instance into a pre-purchased ML Capacity Block (`marketType: capacity-block` targeting the block's reservation) for scheduled training runs on scarce accelerators. Start from the **Capacity Block ML Training Node** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnet where the instance (and any secondary interfaces) are placed
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls inbound and outbound traffic for the instance
- [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) -- the instance's IAM identity for SSM access and AWS API calls
- [**AWS Launch Template**](/infra-catalog/aws-launch-template) -- the golden baseline the instance can launch from
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed encryption for the root and data volumes
- [**AWS Elastic IP**](/infra-catalog/aws-elastic-ip) -- a stable public address associated with the instance's primary interface
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- registers the instance (by `instance_id`) behind a load balancer
