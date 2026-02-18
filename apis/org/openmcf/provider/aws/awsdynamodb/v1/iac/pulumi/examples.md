# Pulumi examples for AwsDynamodb

## Minimal manifest (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsDynamodb
metadata:
  name: orders
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

Preview:
```bash
openmcf pulumi preview --manifest ../hack/manifest.yaml --stack organization/<project>/<stack> --module-dir . | cat
```

Update (apply):
```bash
openmcf pulumi update --manifest ../hack/manifest.yaml --stack organization/<project>/<stack> --module-dir . --yes | cat
```

Refresh:
```bash
openmcf pulumi refresh --manifest ../hack/manifest.yaml --stack organization/<project>/<stack> --module-dir . | cat
```

Destroy:
```bash
openmcf pulumi destroy --manifest ../hack/manifest.yaml --stack organization/<project>/<stack> --module-dir . | cat
```


