# AliCloudVpc Terraform Examples

Create a YAML manifest using one of the examples below, then deploy with the OpenMCF CLI:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

---

## Basic VPC

A VPC with only the required fields. Suitable for development or testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVpc
metadata:
  name: dev-vpc
spec:
  region: cn-hangzhou
  vpcName: dev-vpc
  cidrBlock: "192.168.0.0/16"
```

This creates a VPC with a `/16` CIDR block in `cn-hangzhou`. No description, IPv6, resource group, or custom tags are configured.

---

## Production VPC with Tags and Resource Group

A production VPC with a large CIDR block, resource group assignment, and organizational tags.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVpc
metadata:
  name: prod-vpc
  org: my-org
  env: production
spec:
  region: cn-shanghai
  vpcName: prod-platform-vpc
  cidrBlock: "10.0.0.0/8"
  description: Production VPC for platform workloads
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

- Large CIDR block accommodates many VSwitches across multiple AZs
- Resource group enables access control and cost attribution
- Tags supplement the system tags that OpenMCF adds automatically

---

## IPv6-Enabled VPC

A dual-stack VPC. Alibaba Cloud allocates a `/56` IPv6 CIDR block automatically when IPv6 is enabled.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVpc
metadata:
  name: ipv6-vpc
  env: staging
spec:
  region: us-west-1
  vpcName: ipv6-enabled-vpc
  cidrBlock: "172.16.0.0/12"
  description: Dual-stack VPC with IPv6 support
  enableIpv6: true
  tags:
    networkType: dual-stack
```

- IPv6 is best enabled at VPC creation time; retrofitting later is disruptive
- VSwitches within this VPC can then be assigned IPv6 CIDR blocks

---

## After Deploying

Verify the VPC was created using the Alibaba Cloud CLI:

```shell
aliyun vpc DescribeVpcs --RegionId cn-hangzhou --VpcName dev-vpc
```

To inspect the VPC's route table and VRouter:

```shell
aliyun vpc DescribeVRouters --RegionId cn-hangzhou --VpcId <your-vpc-id>
aliyun vpc DescribeRouteTables --VpcId <your-vpc-id>
```

To tear down:

```shell
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```
