# GcpDataprocVirtualCluster Examples

## Minimal Virtual Cluster

The simplest Dataproc on GKE setup: one node pool with the DEFAULT role and Spark 3.5.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: minimal-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
```

## Custom Namespace with Staging Bucket

Explicit Kubernetes namespace and a dedicated GCS bucket for staging artifacts.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: spark-with-staging
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  kubernetesNamespace:
    value: "spark-etl"
  stagingBucket:
    value: "spark-etl-staging-bucket"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
```

## Multi-Pool: DEFAULT + SPARK_DRIVER + SPARK_EXECUTOR

Separate node pools for drivers and executors. Drivers run on on-demand nodes for stability; executors run on a dedicated pool that can be scaled independently.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: multi-pool-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
    properties:
      "spark:spark.executor.memory": "8g"
      "spark:spark.driver.memory": "4g"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
    - nodePool:
        value: "spark-drivers"
      roles:
        - SPARK_DRIVER
    - nodePool:
        value: "spark-executors"
      roles:
        - SPARK_EXECUTOR
```

## Node Pool Config with Autoscaling

Executor pool with autoscaling from 0 to 20 Spot nodes for cost-optimized burst capacity.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: autoscaling-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
    properties:
      "spark:spark.dynamicAllocation.enabled": "true"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
        - SPARK_DRIVER
    - nodePool:
        value: "executor-pool"
      roles:
        - SPARK_EXECUTOR
      nodePoolConfig:
        machineType: n2-standard-8
        spot: true
        autoscaling:
          minNodeCount: 0
          maxNodeCount: 20
        locations:
          - us-central1-a
          - us-central1-b
```

## With Metastore Integration

Virtual cluster connected to a Dataproc Metastore service for Hive catalog access. Spark jobs can query Hive tables without configuring a standalone metastore.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: spark-with-metastore
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
    properties:
      "spark:spark.sql.catalogImplementation": "hive"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
  auxiliaryServicesConfig:
    metastoreService: "projects/my-gcp-project/locations/us-central1/services/shared-metastore"
```

## Full-Featured Production Setup

Production virtual cluster with role separation, autoscaling Spot executors, a staging bucket, Metastore integration, and Spark History Server.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: prod-spark-on-gke
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: prod-spark-on-gke
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/prod-gke-cluster"
  kubernetesNamespace:
    value: "spark-production"
  stagingBucket:
    value: "prod-spark-staging"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
    properties:
      "spark:spark.executor.memory": "12g"
      "spark:spark.driver.memory": "8g"
      "spark:spark.dynamicAllocation.enabled": "true"
      "spark:spark.sql.catalogImplementation": "hive"
  nodePoolTargets:
    - nodePool:
        value: "controller-pool"
      roles:
        - DEFAULT
        - CONTROLLER
      nodePoolConfig:
        machineType: e2-standard-4
        autoscaling:
          minNodeCount: 1
          maxNodeCount: 3
    - nodePool:
        value: "driver-pool"
      roles:
        - SPARK_DRIVER
      nodePoolConfig:
        machineType: n2-standard-4
        autoscaling:
          minNodeCount: 1
          maxNodeCount: 5
    - nodePool:
        value: "executor-pool"
      roles:
        - SPARK_EXECUTOR
      nodePoolConfig:
        machineType: n2-standard-8
        localSsdCount: 1
        spot: true
        autoscaling:
          minNodeCount: 0
          maxNodeCount: 50
        locations:
          - us-central1-a
          - us-central1-b
          - us-central1-c
  auxiliaryServicesConfig:
    metastoreService: "projects/my-gcp-project/locations/us-central1/services/prod-metastore"
    sparkHistoryServerCluster: "projects/my-gcp-project/regions/us-central1/clusters/spark-history"
```

## Foreign Key References (Composition)

Using `valueFrom` references to compose the virtual cluster with other OpenMCF resources.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: composed-spark
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
  region: us-central1
  gkeClusterTarget:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-gke-cluster
  kubernetesNamespace:
    valueFrom:
      kind: KubernetesNamespace
      name: spark-ns
  stagingBucket:
    valueFrom:
      kind: GcpGcsBucket
      name: spark-staging
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
  nodePoolTargets:
    - nodePool:
        valueFrom:
          kind: GcpGkeNodePool
          name: default-pool
      roles:
        - DEFAULT
    - nodePool:
        valueFrom:
          kind: GcpGkeNodePool
          name: executor-pool
      roles:
        - SPARK_EXECUTOR
```
