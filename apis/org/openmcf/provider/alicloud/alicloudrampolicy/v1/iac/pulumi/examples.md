# Alibaba Cloud RAM Policy — Pulumi Examples

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

## Minimal OSS Read-Only Policy

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRamPolicy
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

Creates a custom policy granting read-only access to a specific OSS bucket. Uses only required fields — no description, tags, or version rotation.

---

## Scoped Bucket Access with Version Rotation

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRamPolicy
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

Grants full OSS access scoped to a single bucket. The `rotateStrategy` prevents version exhaustion when the policy document is updated frequently through IaC. Tags provide organizational metadata.

---

## Multi-Service CI/CD Pipeline Policy

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRamPolicy
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

A cross-service policy with multiple statements for Container Registry, Container Service, and Log Service. Uses `force: true` for clean teardown even when the policy is attached to roles. The `rotateStrategy` prevents version exhaustion during frequent CI/CD pipeline updates.
