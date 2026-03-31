# Examples

## Minimal Configuration

A RAM role with a trust policy allowing ECS service to assume it. No policies attached yet.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: my-ecs-role
spec:
  region: cn-hangzhou
  roleName: my-ecs-service-role
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["ecs.aliyuncs.com"]}
      }],
      "Version": "1"
    }
```

## ECS Service Role with System Policies

A role for ECS instances that need access to OSS and CloudMonitor. Uses Alibaba Cloud managed (system) policies.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: ecs-worker-role
  org: my-org
  env: production
spec:
  region: cn-shanghai
  roleName: ecs-worker-role
  description: Role for ECS worker instances accessing OSS and monitoring
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["ecs.aliyuncs.com"]}
      }],
      "Version": "1"
    }
  maxSessionDuration: 7200
  tags:
    team: platform
    costCenter: infrastructure
  policyAttachments:
    - policyName: AliyunOSSFullAccess
    - policyName: AliyunCloudMonitorFullAccess
    - policyName: AliyunLogFullAccess
```

## Cross-Account Role with Custom Policy

A role that another Alibaba Cloud account can assume, with a custom policy for fine-grained access. Uses `force: true` to allow clean deletion even if policies are attached.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: cross-account-audit-role
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  roleName: cross-account-audit-role
  description: Allows the security audit account to read billing and logs
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"RAM": ["acs:ram::1234567890123456:root"]}
      }],
      "Version": "1"
    }
  maxSessionDuration: 43200
  force: true
  tags:
    purpose: security-audit
  policyAttachments:
    - policyName: AliyunBSSReadOnlyAccess
      policyType: System
    - policyName: audit-log-reader-policy
      policyType: Custom
```
