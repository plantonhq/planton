# Alibaba Cloud RAM Role Examples

Below are several examples demonstrating how to define an Alibaba Cloud RAM Role component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Minimal ECS Service Role

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

This example:
- Creates a RAM role with an ECS service trust policy.
- No policies attached — the role can be assumed but has no permissions.
- Uses default session duration (3600 seconds) and `force: false`.

---

## Service Role with System Policies

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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
    costCenter: infrastructure
  policyAttachments:
    - policyName: AliyunOSSFullAccess
    - policyName: AliyunCloudMonitorFullAccess
    - policyName: AliyunLogFullAccess
```

This role:
- Attaches three system-managed policies for OSS, CloudMonitor, and Log Service access.
- Sets a 2-hour session duration suitable for CI/CD pipelines.
- Includes custom tags for organizational tracking.

---

## Cross-Account Audit Role

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

This role:
- Uses an account-trust policy (RAM principal) instead of a service-trust policy.
- Mixes system and custom policies.
- Sets `force: true` for clean deletion even when policies are attached.
- Uses the maximum session duration (12 hours) for long-running audit sessions.

---

## Full-Featured Role

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamRole
metadata:
  name: fc-execution-role
spec:
  region: cn-hangzhou
  roleName: fc-execution-role
  description: Execution role for Function Compute with VPC and logging access
  assumeRolePolicyDocument: |
    {
      "Statement": [{
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {"Service": ["fc.aliyuncs.com"]}
      }],
      "Version": "1"
    }
  maxSessionDuration: 3600
  force: false
  tags:
    service: function-compute
    environment: staging
  policyAttachments:
    - policyName: AliyunVPCFullAccess
      policyType: System
    - policyName: AliyunECSNetworkInterfaceManagement
      policyType: System
    - policyName: AliyunLogFullAccess
      policyType: System
```

A production-ready configuration with:
- Function Compute service trust policy.
- Explicit `policyType: System` for clarity (matches the default, but stated for documentation purposes).
- Default session duration and safe deletion behavior (`force: false`).
- Tags for environment and service identification.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm the role exists using the Alibaba Cloud CLI:

```shell
aliyun ram GetRole --RoleName <your-role-name>
```

You should see the role details including its ARN, trust policy, and creation timestamp.
