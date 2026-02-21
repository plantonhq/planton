# Alibaba Cloud Security Group

Deploys an Alibaba Cloud Security Group with bundled security rules in a VPC. The component provisions the security group and its ingress/egress rules as a single atomic unit.

## What Gets Created

When you deploy an AliCloudSecurityGroup resource, OpenMCF provisions:

- **Security Group** -- an `ecs.SecurityGroup` resource bound to the specified VPC
- **Security Group Rules** -- one `ecs.SecurityGroupRule` per entry in `rules`, with `nic_type` hardcoded to `intranet`

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An Alibaba Cloud VPC** -- the security group must belong to a VPC

## Quick Start

Create a file `security-group.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudSecurityGroup
metadata:
  name: my-web-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudSecurityGroup.my-web-sg
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  securityGroupName: web-sg
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "443/443"
      cidrIp: "0.0.0.0/0"
```

Deploy:

```shell
openmcf apply -f security-group.yaml
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
