# AwsDynamodb examples

## Minimal manifest (YAML)
A minimal PAY_PER_REQUEST table with a single HASH key.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDynamodb
metadata:
  name: orders
  labels:
    app: shop
spec:
  region: us-east-1
  billingMode: BILLING_MODE_PAY_PER_REQUEST
  attributeDefinitions:
    - name: pk
      type: ATTRIBUTE_TYPE_S
  keySchema:
    - attributeName: pk
      keyType: KEY_TYPE_HASH
```

## PROVISIONED with GSI and TTL
Table using PROVISIONED capacity, a GSI, TTL enabled, and streams enabled.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDynamodb
metadata:
  name: orders
spec:
  region: us-east-1
  billingMode: BILLING_MODE_PROVISIONED
  provisionedThroughput:
    readCapacityUnits: 10
    writeCapacityUnits: 10
  attributeDefinitions:
    - name: pk
      type: ATTRIBUTE_TYPE_S
    - name: sk
      type: ATTRIBUTE_TYPE_S
    - name: status
      type: ATTRIBUTE_TYPE_S
  keySchema:
    - attributeName: pk
      keyType: KEY_TYPE_HASH
    - attributeName: sk
      keyType: KEY_TYPE_RANGE
  globalSecondaryIndexes:
    - name: status-index
      keySchema:
        - attributeName: status
          keyType: KEY_TYPE_HASH
      projection:
        type: PROJECTION_TYPE_KEYS_ONLY
      provisionedThroughput:
        readCapacityUnits: 5
        writeCapacityUnits: 5
  ttl:
    enabled: true
    attributeName: expiresAt
  streamEnabled: true
  streamViewType: STREAM_VIEW_TYPE_NEW_AND_OLD_IMAGES
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./manifest.yaml | cat
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack myorg/infra/dev --module-dir ./apis/project/planton/provider/aws/awsdynamodb/v1/iac/pulumi | cat
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve | cat
```

> Note: Provider credentials are supplied via stack input, not in the spec.


