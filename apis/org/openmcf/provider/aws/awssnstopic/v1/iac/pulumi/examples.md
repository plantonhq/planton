# AwsSnsTopic Pulumi Examples

## Minimal Standard Topic

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: my-notifications
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSnsTopic.my-notifications
spec:
  region: us-east-1
  signatureVersion: 2
```

## Topic with SQS Subscription

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: order-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSnsTopic.order-events
spec:
  region: us-east-1
  subscriptions:
    - name: order-queue
      protocol: sqs
      endpoint:
        value: arn:aws:sqs:us-east-1:123456789012:order-queue
      rawMessageDelivery: true
```

## Deploy commands

```bash
openmcf pulumi preview --manifest manifest.yaml --module-dir ./apis/org/openmcf/provider/aws/awssnstopic/v1/iac/pulumi
openmcf pulumi up --manifest manifest.yaml --module-dir ./apis/org/openmcf/provider/aws/awssnstopic/v1/iac/pulumi --yes
```
