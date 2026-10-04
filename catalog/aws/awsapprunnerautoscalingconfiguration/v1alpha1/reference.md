# AwsAppRunnerAutoScalingConfiguration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsAppRunnerAutoScalingConfigurationSpec defines the desired configuration
for an AWS App Runner auto scaling configuration -- the reusable scaling
policy that controls how App Runner scales a service's instance count in
response to incoming request concurrency.

An auto scaling configuration is deliberately its own resource rather than
a field on the service: one configuration is shared by any number of App
Runner services (each service references it by ARN), so a fleet of
services can adopt a common scaling posture that is tuned in one place.

How App Runner scales: when the number of concurrent requests routed to a
single instance exceeds max_concurrency, App Runner launches additional
instances (up to max_size). When traffic drops, instances above min_size
are stopped; App Runner keeps min_size instances warm at all times --
warm instances are billed for memory only, not CPU.

Revision semantics (important): AWS versions these configurations. Every
scaling value is create-time immutable -- changing any of them registers a
NEW revision under the same configuration name and points referencing
services at it on their next deployment. The exported ARN carries the
revision, so a change here rolls referencing services naturally through
the resource graph. The set_as_account_default designation is the one
exception: it is a pointer AWS moves in place, not part of the revision.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerAutoScalingConfiguration
metadata:
  name: test-asc
  org: test-org
  env: dev
  id: test-asc-dev
spec:
  region: us-west-2
  maxConcurrency: 80
  maxSize: 10
  minSize: 2
  setAsAccountDefault: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.maxConcurrency` | `int32` |  | `100` |  |
| `spec.maxSize` | `int32` |  | `25` |  |
| `spec.minSize` | `int32` |  | `1` |  |
| `spec.setAsAccountDefault` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the auto scaling configuration will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.maxConcurrency

`int32` · optional (explicit presence)

Maximum number of concurrent requests routed to a single instance
before App Runner launches additional instances. Lower values give each
instance more headroom (better tail latency) at higher cost -- more
instances serve the same traffic. 1 effectively dedicates an instance
per request, the serverless-function posture.

- default: `100`
- rule: {"int32":{"lte":200,"gte":1}}

### spec.maxSize

`int32` · optional (explicit presence)

Maximum number of instances App Runner scales out to during traffic
spikes -- the cost ceiling for the services using this configuration.
AWS caps this at 25 by default (a service-quota increase can raise it).

- default: `25`
- rule: {"int32":{"gte":1}}

### spec.minSize

`int32` · optional (explicit presence)

Minimum number of instances App Runner keeps provisioned at all times.
These warm instances serve traffic without cold-start latency and are
billed for memory only (CPU is billed only while actively serving).
Raise above 1 for latency-sensitive services worth the standing memory
charge.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.setAsAccountDefault

`bool`

Claim this configuration as the ACCOUNT-WIDE default for its region:
App Runner services created WITHOUT an explicit auto scaling
configuration use the account default. One default exists per account
per region -- claiming it silently displaces whichever configuration
held the designation (AWS marks it non-default), and the claim only
affects services created afterwards; existing services keep their
current associations. One-way at AWS: un-setting this flag (or
deleting this resource) does NOT restore the previous default -- AWS
has no restore API, so the designation stays with this configuration
until another one claims it. Coordinate so at most one configuration
per account/region sets this flag.

## Validation Rules

- `max_size_gte_min_size`: max_size must be greater than or equal to min_size -- the scale-out ceiling cannot sit below the warm floor

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAppRunnerAutoScalingConfiguration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.configuration_arn` | `string` | The ARN of this configuration revision (e.g. "arn:aws:apprunner: us-west-2:123456789012:autoscalingconfiguration/my-asc/3/abc123"). The ARN carries the revision number, so registering a new revision changes this output and rolls referencing services on their next deployment. |
| `status.outputs.configuration_revision` | `int64` | The revision this deployment registered (e.g. 3). Revisions are immutable; every value change registers the next number under the same configuration name. |
| `status.outputs.latest` | `bool` | Whether this revision is the latest one registered under the configuration name. |
| `status.outputs.is_default` | `bool` | Whether this configuration holds the account/region default designation at the end of the deployment (true when set_as_account_default claimed it). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppRunnerService | `spec.autoScalingConfigurationArn` | `status.outputs.configuration_arn` |

## See Also

- [Overview](../README.md)
