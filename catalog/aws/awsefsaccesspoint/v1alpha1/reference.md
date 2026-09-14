# AwsEfsAccessPoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEfsAccessPointSpec defines the desired configuration for an EFS access
point — an application-specific entry point into an Elastic File System that
enforces a POSIX user/group identity and optionally restricts the visible
root directory.

Access points are deliberately their own resource rather than a field on the
file system: one file system serves many applications, each behind its own
access point (up to 1,000 per file system), and the access point — not the
file system — is what Lambda file-system configs and ECS task-definition EFS
volumes reference. Splitting them lets each application team own its access
point without touching the shared file-system node.

Access points are the recommended way to give Lambda functions and ECS tasks
file system access: the NFS client's identity is overridden with the enforced
POSIX user, and the visible tree is pinned to the root directory, so
applications cannot wander the file system regardless of what they run as.

Key design notes:
- The ENTIRE access point is create-time immutable — every field below
  (file system, POSIX user, root directory) replaces the access point when
  changed. Only tags are mutable. Plan the identity and root path upfront.
- When `root_directory.path` does not exist on the file system yet, provide
  `root_directory.creation_info` — EFS then creates the directory with that
  ownership and permissions on first mount. Without creation_info, mounting
  an access point whose path does not exist fails.

Credentials, region, and deployment workflow live outside this spec in
stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEfsAccessPoint
metadata:
  org: example-org
  env: dev
  name: test-efs-ap
  id: test-efs-ap-id
spec:
  region: us-east-1
  fileSystemId:
    value: fs-0123456789abcdef0
  posixUser:
    uid: 1000
    gid: 1000
  rootDirectory:
    path: /app/data
    creationInfo:
      ownerUid: 1000
      ownerGid: 1000
      permissions: "0755"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.posixUser` | `AwsEfsAccessPointPosixUser` |  |  |  |
| `spec.posixUser.uid` | `int64` | yes |  |  |
| `spec.posixUser.gid` | `int64` | yes |  |  |
| `spec.posixUser.secondaryGids` | `[]int64` |  |  |  |
| `spec.rootDirectory` | `AwsEfsAccessPointRootDirectory` |  |  |  |
| `spec.rootDirectory.path` | `string` | yes |  |  |
| `spec.rootDirectory.creationInfo` | `AwsEfsAccessPointCreationInfo` |  |  |  |
| `spec.rootDirectory.creationInfo.ownerUid` | `int64` |  |  |  |
| `spec.rootDirectory.creationInfo.ownerGid` | `int64` |  |  |  |
| `spec.rootDirectory.creationInfo.permissions` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.fileSystemId

`string | valueFrom` · required

The Elastic File System this access point enters. ForceNew — an access
point cannot be moved between file systems.

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.posixUser

`AwsEfsAccessPointPosixUser`

POSIX user and group identity enforced for ALL file operations through
this access point. When set, the NFS client's identity is overridden —
regardless of what UID/GID the client claims, all operations use these
values. ForceNew.

Omit to let clients keep their own POSIX identity (rare — enforcing the
identity is most of the point of an access point).

### spec.posixUser.uid

`int64` · required · optional (explicit presence)

POSIX user ID (0–4294967295). All file system operations through this
access point use this UID as the file owner. Required inside posix_user;
0 is root.

- rule: {"required":true,"int64":{"lte":"4294967295","gte":"0"}}

### spec.posixUser.gid

`int64` · required · optional (explicit presence)

POSIX primary group ID (0–4294967295). All file system operations through
this access point use this GID as the file group. Required inside
posix_user; 0 is root.

- rule: {"required":true,"int64":{"lte":"4294967295","gte":"0"}}

### spec.posixUser.secondaryGids

`[]int64`

Secondary POSIX group IDs supplementing the primary GID for group
permission checks. Maximum 16 (the NFS AUTH_SYS group limit).

- rule: {"repeated":{"maxItems":"16","items":{"int64":{"lte":"4294967295","gte":"0"}}}}

### spec.rootDirectory

`AwsEfsAccessPointRootDirectory`

Root directory exposed as "/" when mounting through this access point.
ForceNew. Omit to expose the entire file system.

### spec.rootDirectory.path

`string` · required

Path on the file system to expose as the root directory. Must be an
absolute path (starts with "/"), up to 4 subdirectories deep and at most
100 characters (AWS limits, enforced server-side).

If this path does not exist yet, provide `creation_info` — EFS creates the
directory with that ownership and permissions on first mount. Without it,
mounting an access point whose path does not exist fails.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^/"}}

### spec.rootDirectory.creationInfo

`AwsEfsAccessPointCreationInfo`

POSIX ownership and permissions applied when EFS creates the root
directory. Required in practice whenever `path` does not already exist on
the file system (AWS validates existence at mount time, not create time).

### spec.rootDirectory.creationInfo.ownerUid

`int64`

POSIX user ID for the directory owner (0–4294967295).

- rule: {"int64":{"lte":"4294967295","gte":"0"}}

### spec.rootDirectory.creationInfo.ownerGid

`int64`

POSIX group ID for the directory owner (0–4294967295).

- rule: {"int64":{"lte":"4294967295","gte":"0"}}

### spec.rootDirectory.creationInfo.permissions

`string` · required

POSIX permissions for the directory, as 3–4 octal digits (e.g., "755",
"0755", "0750").

- rule: {"required":true,"string":{"pattern":"^[0-7]{3,4}$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEfsAccessPoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.access_point_id` | `string` | The ID of the access point (e.g., "fsap-0123456789abcdef0"). ECS task definition EFS volume authorization references this. |
| `status.outputs.access_point_arn` | `string` | The Amazon Resource Name of the access point. Lambda file system configurations and IAM policy conditions reference this. |
| `status.outputs.file_system_id` | `string` | The ID of the file system this access point enters (e.g., "fs-0123456789abcdef0"). Exported so a consumer composing on the access point (an ECS EFS volume needs both) can wire everything from this one node. |
| `status.outputs.file_system_arn` | `string` | The ARN of the file system this access point enters. Used in IAM policies that scope permissions to the file system while mounting through the access point. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBatchJobDefinition | `spec.container.volumes[].efs.accessPointId` | `status.outputs.access_point_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].runtimeEnvironment.filesystems[].efsAccessPointArn` | `status.outputs.access_point_arn` |
| AwsBedrockAgentCoreRuntime | `spec.filesystems[].efsAccessPointArn` | `status.outputs.access_point_arn` |
| AwsEcsTaskDefinition | `spec.volumes[].efs.accessPointId` | `status.outputs.access_point_id` |
| AwsLambda | `spec.fileSystemConfig.accessPointArn` | `status.outputs.access_point_arn` |

## See Also

- [Overview](../README.md)
