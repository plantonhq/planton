# AwsSqsQueue Pulumi Examples

## Minimal Standard Queue

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: my-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSqsQueue.my-queue
spec:
  sqsManagedSseEnabled: true
```

## FIFO Queue

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: order-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSqsQueue.order-events
spec:
  fifoQueue: true
  contentBasedDeduplication: true
  sqsManagedSseEnabled: true
  visibilityTimeoutSeconds: 120
```

## Deploy commands

```bash
openmcf pulumi preview --manifest manifest.yaml --module-dir ./apis/org/openmcf/provider/aws/awssqsqueue/v1/iac/pulumi
openmcf pulumi up --manifest manifest.yaml --module-dir ./apis/org/openmcf/provider/aws/awssqsqueue/v1/iac/pulumi --yes
```
