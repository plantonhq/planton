# AwsEc2Instance

A single Amazon EC2 virtual machine: the launch source (an AMI + instance type, a launch template, or a template with inline overrides), network placement, IAM identity, storage, instance posture (IMDSv2, monitoring, CPU topology), purchase options (On-Demand or Spot), and lifecycle protections.

A standalone instance is the pet of EC2 compute -- a bastion, a license server, a singleton stateful workload -- as opposed to the cattle an auto-scaling group manages. The spec deliberately shares `AwsLaunchTemplate`'s vocabulary (metadata options, CPU options, spot options, block devices), so a pet can graduate into a templated fleet without relearning field names.

Everything the instance composes with attaches by reference and none of it is created here: the subnet (`AwsSubnet`), security groups (`AwsSecurityGroup`), the IAM instance profile (`AwsIamInstanceProfile` -- referenced by NAME, which is what the EC2 API takes), the launch template (`AwsLaunchTemplate`), and the KMS keys that encrypt volumes (`AwsKmsKey`). The instance's display identity is the `Name` tag, carried from `metadata.name`.

## Spec highlights

- **Launch sources, three-shaped**: inline `ami` + `instanceType`; a referenced `launchTemplate` (by ID or name, pinned to a version or tracking `$Default`); or both -- inline fields override the template per instance. Validation enforces AWS's own rule that an image and a type must come from somewhere.
- **Networking**: subnet and security-group references, a pre-provisioned primary ENI (`primaryNetworkInterfaceId` -- static IP/MAC identity), static private IPs, secondary private IPs, public-IP tri-state, IPv6 (count or explicit addresses, stable primary IPv6), private DNS hostname options, and secondary network interfaces on multi-card instance types.
- **Storage**: full root-volume reshaping (size, gp3 tuning up to 2000 MiB/s, encryption with a KMS reference), additional EBS data volumes keyed by device name, instance-store mappings, EBS optimization, and volume tagging both ways -- uniform at-creation `volumeTags` (the ABAC-compliant arm, covering AMI-inherited volumes) XOR per-device `tags`.
- **Posture**: IMDSv2 enforcement (`metadataOptions.httpTokens: required` -- the recommended default), detailed monitoring, CPU topology (core count, SMT, AMD SEV-SNP, nested virtualization), burstable credit mode, Nitro Enclaves XOR hibernation, auto-recovery, shutdown behavior, stop/termination API protection, and `forceDestroy` -- the declarative escape hatch that lifts those protections at destroy instead of failing it.
- **Purchasing**: three markets -- Spot with persistent/stop-resume semantics, pre-purchased ML Capacity Blocks (`marketType: capacity-block`), and interruptible Capacity Reservations -- plus Capacity Reservation targeting (open/none/capacity-reservations-only, or a specific reservation/group).
- **Placement**: AZ pinning, placement groups (by name or ID, with partition numbers), tenancy, Dedicated Hosts, and host resource groups.
- **User data**: plain-text or base64 (mutually exclusive), with an explicit replace-on-change switch; `${...}` content passes through literally on both engines.

## Outputs

`instance_id` (the join key -- `AwsLbTargetGroup` instance targets reference it), `arn`, `instance_state`, `availability_zone`, `private_ip`, `private_dns`, `public_ip`, `public_dns`, `primary_network_interface_id`.

## How it works

Both the Terraform/OpenTofu and Pulumi modules implement the same contract: exactly one `aws_instance` whose `Name` tag carries `metadata.name`. Launch identity (AMI, type, subnet, key pair, placement, CPU topology, purchase option) is create-time; the operational posture (security groups, IAM profile, metadata options, protections) edits in place. Optional fields pass through only when set, so an omitted field keeps the AWS or launch-template default.

## References

- [Amazon EC2 User Guide](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/concepts.html)
- [Instance metadata service (IMDSv2)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html)
- [Spot Instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
- [Capacity Reservations](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-reservations-using.html)
- [Launch an instance from a launch template](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-launch-templates.html)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
