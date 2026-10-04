# AwsAppRunnerObservabilityConfiguration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsAppRunnerObservabilityConfigurationSpec defines the desired
configuration for an AWS App Runner observability configuration -- the
reusable tracing policy that App Runner services reference to enable
distributed request tracing.

An observability configuration is deliberately its own resource rather
than a field on the service: one configuration is shared by any number of
App Runner services (each service references it by ARN), so a fleet of
services adopts a common tracing posture that is tuned in one place.

Revision semantics (important): AWS versions these configurations. The
trace settings are create-time immutable -- changing them registers a NEW
revision under the same configuration name. The exported ARN carries the
revision, so a change here rolls referencing services naturally through
the resource graph.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerObservabilityConfiguration
metadata:
  name: test-observability
  org: test-org
  env: dev
  id: test-observability-dev
spec:
  region: us-west-2
  traceConfiguration:
    vendor: AWSXRAY
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.traceConfiguration` | `AwsAppRunnerObservabilityConfigurationTraceConfiguration` |  |  |  |
| `spec.traceConfiguration.vendor` | `string` |  | `AWSXRAY` |  |

## Field Details

### spec.region

`string` · required

The AWS region where the observability configuration will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.traceConfiguration

`AwsAppRunnerObservabilityConfigurationTraceConfiguration`

Distributed tracing configuration. When set, services referencing this
configuration send request traces to the configured vendor -- each
instance runs an OpenTelemetry collector sidecar that forwards spans.
When omitted, the configuration is registered without tracing (a valid
but inert configuration; set trace_configuration to get value from this
resource).

### spec.traceConfiguration.vendor

`string` · optional (explicit presence)

The tracing vendor. "AWSXRAY" (AWS X-Ray) is the only vendor App Runner
supports today; the application must be instrumented with the AWS
Distro for OpenTelemetry (ADOT) SDK to emit spans the collector can
forward.

- default: `AWSXRAY`
- rule: {"string":{"in":["AWSXRAY"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAppRunnerObservabilityConfiguration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.configuration_arn` | `string` | The ARN of this configuration revision (e.g. "arn:aws:apprunner: us-west-2:123456789012:observabilityconfiguration/my-oc/2/abc123"). The ARN carries the revision number, so registering a new revision changes this output and rolls referencing services on their next deployment. |
| `status.outputs.configuration_revision` | `int64` | The revision this deployment registered (e.g. 2). Revisions are immutable; every trace-setting change registers the next number under the same configuration name. |
| `status.outputs.latest` | `bool` | Whether this revision is the latest one registered under the configuration name. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppRunnerService | `spec.observabilityConfigurationArn` | `status.outputs.configuration_arn` |

## See Also

- [Overview](../README.md)
