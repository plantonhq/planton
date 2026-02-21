# Alibaba Cloud RAM Policy Examples

Below are several examples demonstrating how to define an Alibaba Cloud RAM Policy component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Minimal OSS Read-Only Policy

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: oss-reader
spec:
  region: cn-hangzhou
  policyName: oss-read-only
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "oss:GetObject",
            "oss:ListObjects"
          ],
          "Resource": ["acs:oss:*:*:my-bucket/*"]
        }
      ]
    }
```

This example:
- Creates a custom policy granting read-only access to a specific OSS bucket.
- Uses only required fields with no optional configuration.
- Uses default `rotateStrategy` (`None`) and `force` (`false`).

---

## Scoped Bucket Access with Tags

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: app-data-access
spec:
  region: cn-shanghai
  policyName: app-data-bucket-full-access
  description: Grants full access to the application data bucket and its objects
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": ["oss:*"],
          "Resource": [
            "acs:oss:*:*:app-data-prod",
            "acs:oss:*:*:app-data-prod/*"
          ]
        }
      ]
    }
  rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded
  tags:
    team: platform
    costCenter: infrastructure
```

This policy:
- Grants full OSS access scoped to a single bucket (bucket-level and object-level ARNs).
- Enables automatic version rotation to prevent the 5-version limit from blocking deployments.
- Includes organizational tags for cost attribution and filtering.

---

## Multi-Service CI/CD Pipeline Policy

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: cicd-deploy-policy
spec:
  region: cn-hangzhou
  policyName: cicd-pipeline-deploy-policy
  description: Permissions for CI/CD pipeline to build images, deploy to ACK, and manage logs
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "cr:GetRepository",
            "cr:PushRepository",
            "cr:PullRepository"
          ],
          "Resource": ["acs:cr:*:*:repository/my-org/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "cs:DescribeClusterDetail",
            "cs:GetClusterKubeconfig",
            "cs:DescribeClusterNodes"
          ],
          "Resource": ["acs:cs:*:*:cluster/*"]
        },
        {
          "Effect": "Allow",
          "Action": [
            "log:PostLogStoreLogs",
            "log:GetLogStore"
          ],
          "Resource": ["acs:log:*:*:project/cicd-logs/*"]
        }
      ]
    }
  rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded
  force: true
  tags:
    purpose: cicd
    managedBy: platform-team
```

This policy:
- Uses multiple statements for cross-service permissions (Container Registry, Container Service, Log Service).
- Sets `force: true` for clean teardown even when the policy is attached to roles.
- Enables version rotation for frequent CI/CD pipeline updates.

---

## Full-Featured Policy

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: production-oss-policy
spec:
  region: cn-hangzhou
  policyName: prod-oss-read-write-scoped
  description: Production policy granting scoped read-write access to application data bucket
  policyDocument: |
    {
      "Version": "1",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "oss:GetObject",
            "oss:PutObject",
            "oss:DeleteObject",
            "oss:ListObjects",
            "oss:GetBucket"
          ],
          "Resource": [
            "acs:oss:*:*:prod-app-data",
            "acs:oss:*:*:prod-app-data/*"
          ]
        }
      ]
    }
  rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded
  force: false
  tags:
    team: backend
    environment: production
    costCenter: app-platform
```

A production-ready configuration with:
- Specific OSS actions (no wildcard `oss:*`) for least-privilege access.
- Explicit `force: false` for production safety (deletion fails if policy is still attached).
- Version rotation enabled for IaC-managed updates.
- Comprehensive tags for organizational tracking.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm the policy exists using the Alibaba Cloud CLI:

```shell
aliyun ram GetPolicy --PolicyName <your-policy-name> --PolicyType Custom
```

You should see the policy details including its creation timestamp, default version, and attachment count.
