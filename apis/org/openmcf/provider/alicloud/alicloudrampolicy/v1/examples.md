# Examples

## Minimal Configuration

A custom policy granting read-only access to all OSS buckets.

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
            "oss:GetBucket",
            "oss:ListObjects",
            "oss:ListBuckets"
          ],
          "Resource": ["acs:oss:*:*:*"]
        }
      ]
    }
```

## Scoped Bucket Access with Tags

A policy granting full access to a specific OSS bucket and its objects. Includes description, tags, and automatic version rotation for policies that get updated frequently.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: app-data-bucket-access
  org: my-org
  env: production
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

## Multi-Service Policy with Force Delete

A comprehensive policy granting cross-service permissions for a CI/CD pipeline role. Uses `force: true` to allow clean teardown even when the policy is still attached to roles.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRamPolicy
metadata:
  name: cicd-pipeline-policy
  org: my-org
  env: production
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
