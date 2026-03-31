# AliCloudVpc Pulumi Examples

Create a YAML manifest using one of the examples below, then deploy with the OpenMCF CLI:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal VPC

A VPC with only the required fields, suitable for development or quick testing.

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

This creates a VPC with a `/16` CIDR block in the `cn-hangzhou` region. No description, IPv6, resource group, or custom tags are configured.

---

## Production VPC with Tags

A production VPC using a large CIDR block with resource group assignment and tags for organizational tracking.

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
- Tags supplement the system tags that OpenMCF adds automatically (resource name, kind, org, env)

---

## IPv6-Enabled VPC

A dual-stack VPC with IPv6 support. Alibaba Cloud allocates a `/56` IPv6 CIDR block automatically when IPv6 is enabled.

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
- VSwitches within this VPC can be assigned IPv6 CIDR blocks from the allocated pool
