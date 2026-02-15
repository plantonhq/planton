# GcpDataprocCluster Examples

## Minimal Cluster

The simplest possible Dataproc cluster. GCP provides defaults for master/worker configuration.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: my-spark-cluster
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: my-spark-cluster
```

## Development Cluster with Jupyter

Interactive development cluster with Jupyter notebooks and auto-delete after 30 minutes idle.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: dev-jupyter
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: dev-jupyter
  clusterConfig:
    masterConfig:
      numInstances: 1
      machineType: e2-standard-4
    workerConfig:
      numInstances: 2
      machineType: e2-standard-4
    softwareConfig:
      imageVersion: "2.2-debian12"
      optionalComponents:
        - JUPYTER
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "1800s"
```

## HA Production Cluster

High-availability cluster with 3 masters, SSD storage, CMEK encryption, and private networking.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: prod-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: prod-spark
  gracefulDecommissionTimeout: "3600s"
  clusterConfig:
    gceConfig:
      subnetwork:
        value: "projects/my-project/regions/us-central1/subnetworks/dataproc"
      serviceAccount:
        value: "dataproc-sa@my-project.iam.gserviceaccount.com"
      internalIpOnly: true
      tags:
        - dataproc
    masterConfig:
      numInstances: 3
      machineType: n2-standard-8
      diskConfig:
        bootDiskSizeGb: 200
        bootDiskType: pd-ssd
    workerConfig:
      numInstances: 5
      machineType: n2-standard-8
      diskConfig:
        bootDiskSizeGb: 500
        bootDiskType: pd-ssd
        numLocalSsds: 2
    softwareConfig:
      imageVersion: "2.2-debian12"
      properties:
        "spark:spark.executor.memory": "12g"
        "spark:spark.driver.memory": "8g"
    encryptionKmsKeyName:
      value: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key"
    endpointConfig:
      enableHttpPortAccess: true
```

## Cost-Optimized Batch Cluster with Spot Workers

Ephemeral cluster for batch jobs using Spot VMs for secondary workers.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: batch-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: batch-spark
  clusterConfig:
    masterConfig:
      numInstances: 1
      machineType: n2-standard-4
    workerConfig:
      numInstances: 2
      machineType: n2-standard-4
    secondaryWorkerConfig:
      numInstances: 10
      preemptibility: SPOT
    softwareConfig:
      imageVersion: "2.2-debian12"
      properties:
        "spark:spark.dynamicAllocation.enabled": "true"
        "spark:spark.shuffle.service.enabled": "true"
    lifecycleConfig:
      idleDeleteTtl: "900s"
```

## ML Training Cluster with GPU Accelerators

Cluster with GPU-enabled workers for distributed ML training with Spark MLlib.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: ml-training
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: ml-training
  clusterConfig:
    masterConfig:
      numInstances: 1
      machineType: n1-standard-8
    workerConfig:
      numInstances: 4
      machineType: n1-standard-8
      accelerators:
        - acceleratorType: nvidia-tesla-t4
          acceleratorCount: 1
      diskConfig:
        bootDiskSizeGb: 200
        bootDiskType: pd-ssd
    softwareConfig:
      imageVersion: "2.2-debian12"
      optionalComponents:
        - JUPYTER
      properties:
        "spark:spark.rapids.sql.enabled": "true"
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "3600s"
```

## Cluster with Custom Initialization Actions

Cluster with startup scripts for installing additional software.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: custom-init
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: custom-init
  clusterConfig:
    masterConfig:
      numInstances: 1
      machineType: n2-standard-4
    workerConfig:
      numInstances: 3
      machineType: n2-standard-4
    softwareConfig:
      imageVersion: "2.2-debian12"
    initializationActions:
      - script: "gs://my-bucket/scripts/install-conda.sh"
        timeoutSec: 600
      - script: "gs://my-bucket/scripts/setup-monitoring.sh"
        timeoutSec: 300
```

## Foreign Key References (Infra Chart Composition)

Using `valueFrom` references for infra-chart composability.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocCluster
metadata:
  name: composed-spark
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
  region: us-central1
  clusterName: composed-spark
  clusterConfig:
    stagingBucket:
      valueFrom:
        kind: GcpGcsBucket
        name: staging-bucket
    gceConfig:
      subnetwork:
        valueFrom:
          kind: GcpSubnetwork
          name: dataproc-subnet
      serviceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: dataproc-sa
    encryptionKmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: dataproc-key
    masterConfig:
      numInstances: 1
      machineType: n2-standard-4
    workerConfig:
      numInstances: 4
      machineType: n2-standard-8
```
