# AwsMemcachedElasticache Examples

## 1. Minimal Single-Node (Development)

The simplest possible Memcached cluster — one node, no VPC, no encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: dev-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.t3.micro
  numCacheNodes: 1
```

## 2. Multi-Node with Cross-AZ

Three nodes distributed across Availability Zones for resilience. If one AZ fails, two-thirds of the cache remains available.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: session-cache
spec:
  region: us-east-1
  engineVersion: "1.6.22"
  nodeType: cache.t3.medium
  numCacheNodes: 3
  azMode: cross-az
  preferredAvailabilityZones:
    - us-east-1a
    - us-east-1b
    - us-east-1c
```

## 3. VPC-Integrated with Security Groups

Deploy into a VPC using `valueFrom` references to an AwsVpc and AwsSecurityGroup.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: app-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.r7g.large
  numCacheNodes: 3
  azMode: cross-az
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: cache-sg
        fieldPath: status.outputs.security_group_id
```

## 4. Transit Encryption Enabled

Enable TLS for client-to-node communication. Requires engine version 1.6.12+.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: encrypted-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.t3.medium
  numCacheNodes: 2
  azMode: cross-az
  transitEncryptionEnabled: true
```

## 5. Custom Parameters

Override default Memcached engine parameters for workload-specific tuning.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: tuned-cache
spec:
  region: us-west-2
  engineVersion: "1.6.22"
  nodeType: cache.r7g.large
  numCacheNodes: 3
  azMode: cross-az
  parameterGroupFamily: memcached1.6
  parameters:
    - name: chunk_size
      value: "96"
    - name: max_simultaneous_connections
      value: "65000"
    - name: binding_protocol
      value: auto
```

## 6. Production-Ready (Full Configuration)

A complete production setup with VPC integration, cross-AZ distribution, transit encryption, custom parameters, SNS notifications, and a maintenance window.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: prod-session-cache
spec:
  region: us-east-1
  engineVersion: "1.6.22"
  nodeType: cache.r7g.xlarge
  numCacheNodes: 5
  azMode: cross-az
  port: 11211
  transitEncryptionEnabled: true
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: memcached-sg
        fieldPath: status.outputs.security_group_id
  parameterGroupFamily: memcached1.6
  parameters:
    - name: chunk_size
      value: "96"
  maintenanceWindow: sun:05:00-sun:06:00
  autoMinorVersionUpgrade: true
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
  preferredAvailabilityZones:
    - us-east-1a
    - us-east-1b
    - us-east-1c
    - us-east-1a
    - us-east-1b
```

## 7. Infra Chart Reference Pattern

Shows how AwsMemcachedElasticache composes with other resources in an infra chart template using `valueFrom` references and template variables.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: "{{ values.env }}-session-cache"
spec:
  region: us-west-2
  engineVersion: "{{ values.memcached_version | default: '1.6.22' }}"
  nodeType: "{{ values.cache_node_type | default: 'cache.t3.medium' }}"
  numCacheNodes: "{{ values.cache_node_count | default: 3 }}"
  azMode: cross-az
  transitEncryptionEnabled: true
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: "{{ values.env }}-vpc"
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: "{{ values.env }}-vpc"
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: "{{ values.env }}-cache-sg"
        fieldPath: status.outputs.security_group_id
```
