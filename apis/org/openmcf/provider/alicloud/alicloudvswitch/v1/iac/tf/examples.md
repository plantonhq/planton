# AlicloudVswitch Terraform Examples

Create a YAML manifest using one of the examples below, then deploy with the OpenMCF CLI:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

---

## Basic Single-Zone VSwitch

A VSwitch with only the required fields. Suitable for development or testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVswitch
metadata:
  name: dev-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-abc123def456
  zoneId: cn-hangzhou-a
  cidrBlock: "192.168.0.0/24"
  vswitchName: dev-vswitch
```

This creates a VSwitch with a `/24` CIDR block in `cn-hangzhou-a`. No description, IPv6, or custom tags are configured.

---

## Production VSwitch with Tags

A production VSwitch using a large CIDR block for Kubernetes node pools, with organizational tags.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVswitch
metadata:
  name: prod-app-vswitch
  org: my-org
  env: production
spec:
  region: cn-shanghai
  vpcId: vpc-prod-001
  zoneId: cn-shanghai-b
  cidrBlock: "10.1.0.0/20"
  vswitchName: prod-app-tier-b
  description: Application tier VSwitch in zone B for Kubernetes workers
  tags:
    team: platform
    costCenter: engineering
    tier: application
```

- Large `/20` CIDR block accommodates Kubernetes pod networking (4,092 usable addresses)
- Tags supplement the system tags that OpenMCF adds automatically

---

## IPv6-Enabled VSwitch

A dual-stack VSwitch. The parent VPC must have IPv6 enabled for the IPv6 CIDR allocation to succeed.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVswitch
metadata:
  name: ipv6-vswitch
spec:
  region: us-west-1
  vpcId: vpc-ipv6-enabled
  zoneId: us-west-1a
  cidrBlock: "172.16.0.0/24"
  vswitchName: ipv6-app-vswitch
  description: Dual-stack VSwitch for IPv6 workloads
  enableIpv6: true
  ipv6CidrBlockMask: 42
  tags:
    networkType: dual-stack
```

- IPv6 is best enabled at VSwitch creation time; retrofitting later requires resource updates
- The `ipv6CidrBlockMask` (0-255) selects a `/64` segment from the parent VPC's `/56` IPv6 allocation

---

## After Deploying

Verify the VSwitch was created using the Alibaba Cloud CLI:

```shell
aliyun vpc DescribeVSwitches --RegionId cn-hangzhou --VSwitchId <your-vswitch-id>
```

To list all VSwitches in a VPC:

```shell
aliyun vpc DescribeVSwitches --RegionId cn-hangzhou --VpcId <your-vpc-id>
```

To tear down:

```shell
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```
