# Alibaba Cloud RAM Role — Pulumi Examples

## CLI Usage

Preview changes before applying:

```shell
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Apply the manifest:

```shell
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

Destroy all resources:

```shell
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal ECS Service Role

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

Creates a RAM role with an ECS service trust policy. No policies attached — the role can be assumed but has no permissions.

---

## Service Role with System Policies

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: ecs-worker-role
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
  policyAttachments:
    - policyName: AliyunOSSFullAccess
    - policyName: AliyunCloudMonitorFullAccess
    - policyName: AliyunLogFullAccess
```

Attaches three system-managed policies to the role. The 2-hour session duration suits CI/CD pipelines and batch workloads.

---

## Cross-Account Role with Custom Policy

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: cross-account-audit
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
    - policyName: AliyunLogReadOnlyAccess
    - policyName: audit-log-reader-policy
      policyType: Custom
```

Uses an account-trust policy instead of a service-trust policy. Mixes system and custom policies. The `force: true` setting allows clean deletion even when policies are attached.
