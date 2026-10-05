# pulumiannotationkeys

This package defines the manifest **annotation** keys that carry Pulumi stack-location
configuration in Planton resource manifests.

## Why annotations, not labels

`metadata.labels` are derived into cloud-provider tags by planton IaC modules, so a
platform key there would leak internal configuration onto the user's real provider
resources. Platform-behavior signals — including the Pulumi stack location — therefore
live in `metadata.annotations`, which never touch the cloud.

## Keys

| Constant | Key |
|----------|-----|
| `StackFqdnAnnotationKey` | `pulumi.planton.dev/stack.fqdn` |
| `OrganizationAnnotationKey` | `pulumi.planton.dev/organization` |
| `ProjectAnnotationKey` | `pulumi.planton.dev/project` |
| `StackNameAnnotationKey` | `pulumi.planton.dev/stack.name` |

`stack.fqdn` (format `organization/project/stack`) takes precedence; when absent, all
three individual keys are required.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: my-vpc
  annotations:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/stack.fqdn: my-org/aws-examples/prod
spec:
  region: us-west-2
```

The consumer is `pkg/iac/pulumi/backendconfig`, which parses these annotations into a
`PulumiBackendConfig`.
