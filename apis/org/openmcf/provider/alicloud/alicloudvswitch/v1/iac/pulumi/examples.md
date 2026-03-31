# AliCloudVswitch Pulumi Examples

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

## Minimal Single-Zone VSwitch

A VSwitch with only the required fields. Suitable for development or quick testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: dev-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-abc123def456
  zoneId: cn-hangzhou-a
  cidrBlock: "192.168.0.0/24"
  vswitchName: dev-vswitch
```

This creates a VSwitch with a `/24` CIDR block in the `cn-hangzhou-a` Availability Zone. No description, IPv6, or custom tags are configured.

---

## Production VSwitch with Tags

A production VSwitch using a large CIDR block for Kubernetes node pools, with organizational tags.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
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
- Tags supplement the system tags that OpenMCF adds automatically (resource name, kind, org, env)

---

## VSwitch with Cross-Resource Reference

Uses a `valueFrom` reference to resolve the VPC ID from an existing AliCloudVpc component.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: db-vswitch
  env: staging
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  zoneId: cn-hangzhou-c
  cidrBlock: "10.2.0.0/24"
  vswitchName: staging-db-vswitch
  description: Database tier VSwitch for RDS and Redis instances
```

- The `valueFrom` reference is resolved by the platform before Pulumi executes
- At runtime, `vpc_id.GetValue()` returns the literal VPC ID from the referenced AliCloudVpc's outputs

---

## IPv6-Enabled VSwitch

A dual-stack VSwitch with IPv6 support. The parent VPC must also have IPv6 enabled for this to take effect.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
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
