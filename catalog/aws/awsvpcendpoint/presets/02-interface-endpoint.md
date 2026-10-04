# Interface Endpoint for an AWS Service

This preset places an ENI-based PrivateLink endpoint for an AWS service
in two private subnets, with private DNS on -- workloads in the VPC
reach the service through their default SDK endpoints, privately, with
zero code changes.

## When to Use

- Private subnets that must call AWS APIs (STS, ECR, CloudWatch Logs,
  Secrets Manager, SSM, KMS) without an internet path
- Locked-down environments where API traffic must not transit a NAT
- The ECR pair (`ecr.api` + `ecr.dkr`) for pulling images from private
  subnets -- deploy this preset once per service

## Key Configuration Choices

- **Two subnets across AZs** -- one ENI each; an endpoint with a single
  AZ is a single point of failure for every client in the VPC. Each AZ
  is billed separately, so two is the availability/cost sweet spot.
- **A dedicated security group** -- AWS attaches the VPC's DEFAULT
  group when none is given; a purpose-built group allowing 443 from the
  VPC CIDR is the production shape.
- **`privateDnsEnabled: true`** -- the service's public name resolves
  to the endpoint inside the VPC. Requires the VPC to have DNS support
  and DNS hostnames enabled (the `AwsVpc` kind's recommended
  defaults).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<endpoint-resource-name>` | Name for this endpoint resource (e.g. `platform-sts-endpoint`) | Your naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) -- also update the region inside `serviceName` | Your deployment region |
| `<vpc-resource-name>` | Name of the AwsVpc resource | Your VPC manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the AwsSubnet resources (different AZs) | Your subnet manifests' `metadata.name` |
| `<endpoint-sg-resource-name>` | Name of the AwsSecurityGroup allowing 443 from the VPC CIDR | Your security-group manifest's `metadata.name` |

## Common Additions

- `policy` restricting which principals may use the endpoint
- `ipAddressType: dualstack` + `dnsOptions.dnsRecordIpType` for IPv6
  VPCs
- `subnetConfigurations` to pin static ENI addresses for firewall rules

## Related Presets

- **01-s3-gateway** -- the free route-based endpoint for S3
- **03-privatelink-service** -- consuming a third-party PrivateLink
  service
