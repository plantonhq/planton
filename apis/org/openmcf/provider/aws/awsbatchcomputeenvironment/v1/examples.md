# AwsBatchComputeEnvironment Examples

## 1. Minimal Fargate

The simplest setup: a serverless Fargate compute environment with a single job queue. AWS manages all infrastructure.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: dev-batch
spec:
  region: us-west-2
  computeResources:
    type: FARGATE
    maxVcpus: 256
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
  jobQueues:
    - name: default
      priority: 1
```

## 2. Production EC2 with Multi-Queue Isolation

EC2 compute with two priority-separated queues. The `critical` queue (priority 10) is dispatched before `background` (priority 1). Background jobs stuck in RUNNABLE for over 2 hours are automatically cancelled.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: prod-batch
spec:
  region: us-west-2
  computeResources:
    type: EC2
    maxVcpus: 512
    minVcpus: 0
    instanceTypes:
      - m5.xlarge
      - c5.xlarge
      - r5.xlarge
      - optimal
    allocationStrategy: BEST_FIT_PROGRESSIVE
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
    instanceRole:
      value: arn:aws:iam::123456789012:instance-profile/ecsInstanceRole
  updatePolicy:
    terminateJobsOnUpdate: false
    jobExecutionTimeoutMinutes: 60
  jobQueues:
    - name: critical
      priority: 10
    - name: background
      priority: 1
      jobStateTimeLimitActions:
        - action: CANCEL
          maxTimeSeconds: 7200
          reason: "Job stuck in RUNNABLE for over 2 hours"
          state: RUNNABLE
```

## 3. Spot with Fair-Share Scheduling

Cost-optimized Spot compute with fair-share scheduling across two teams. The `team-data` share has twice the weight of `team-ml`, so data engineering jobs receive a proportionally larger share of compute capacity.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: ml-batch
spec:
  region: us-west-2
  computeResources:
    type: SPOT
    maxVcpus: 1024
    minVcpus: 0
    instanceTypes:
      - m5.2xlarge
      - c5.2xlarge
      - r5.2xlarge
      - p3.2xlarge
    allocationStrategy: SPOT_CAPACITY_OPTIMIZED
    bidPercentage: 70
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
      - value: subnet-0a1b2c3d4e5f00003
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
    instanceRole:
      value: arn:aws:iam::123456789012:instance-profile/ecsInstanceRole
    spotIamFleetRole:
      value: arn:aws:iam::123456789012:role/aws-ec2-spot-fleet-role
  schedulingPolicy:
    computeReservation: 10
    shareDecaySeconds: 3600
    shareDistributions:
      - shareIdentifier: team-data
        weightFactor: 1.0
      - shareIdentifier: team-ml
        weightFactor: 0.5
  jobQueues:
    - name: ml-training
      priority: 1

```

## 4. Cross-Resource Reference (valueFrom)

Using `valueFrom` to reference outputs from other OpenMCF resources instead of hardcoding values.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: connected-batch
spec:
  region: us-west-2
  computeResources:
    type: FARGATE
    maxVcpus: 256
    subnetIds:
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          fieldPath: status.outputs.private_subnets.0.id
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          fieldPath: status.outputs.private_subnets.1.id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: batch-sg
          fieldPath: status.outputs.security_group_id
  jobQueues:
    - name: default
      priority: 1
```
