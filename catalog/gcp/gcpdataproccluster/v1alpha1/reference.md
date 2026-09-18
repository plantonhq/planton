# GcpDataprocCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDataprocClusterSpec defines the configuration for a Google Cloud
Dataproc cluster running Apache Spark, Hadoop, and related
open-source data processing frameworks.

Dataproc offers two mutually exclusive deployment arms, mirroring the
API's own shape:

  - cluster_config: the standard arm — dedicated Compute Engine VMs
    (master, workers, optional spot secondaries) fully managed by
    Dataproc. When both arms are omitted, GCP creates a default
    GCE-based cluster (2 workers, default machine types).

  - virtual_cluster_config: the Dataproc-on-GKE arm — Spark workloads
    run as pods on an existing GKE cluster, composing against
    GcpGkeCluster and GcpGkeNodePool by reference.

Important behavioral notes:

  - The cluster_name and region fields are immutable after creation.
    Changing them requires recreating the cluster.

  - Most configuration is immutable after creation (recreate on
    change). The in-place exceptions on the GCE arm: primary and
    secondary worker counts (manual scaling), min_num_instances, the
    autoscaling policy attachment, the lifecycle TTLs, and labels.
    The virtual arm is fully immutable.

  - User labels are applied to the cluster and propagate to its VMs.
    The Dataproc API does not support user labels on virtual
    (GKE-based) clusters — leave labels empty for the virtual arm.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocCluster
metadata:
  name: test-dataproc
spec:
  projectId:
    value: my-gcp-project-123 # replace with your project ID
  region: us-central1
  clusterName: test-spark-cluster
  gracefulDecommissionTimeout: "600s"
  deletionPolicy: DELETE
  labels:
    team: data-eng
    workload: spark-etl
  clusterConfig:
    gceConfig:
      subnetwork:
        value: projects/my-gcp-project-123/regions/us-central1/subnetworks/dataproc-subnet # replace with your subnetwork self-link
      internalIpOnly: true
      tags:
        - dataproc
        - spark
      shieldedInstanceConfig:
        enableSecureBoot: true
        enableVtpm: true
        enableIntegrityMonitoring: true
    masterConfig:
      numInstances: 1
      machineType: n2-standard-4
      diskConfig:
        bootDiskSizeGb: 100
        bootDiskType: pd-ssd
    workerConfig:
      numInstances: 3
      minNumInstances: 2
      # No machineType: the flexibility policy below replaces it — the
      # API provisions from the ranked selections and drops a paired
      # machineTypeUri (mutually exclusive, CEL-enforced).
      diskConfig:
        bootDiskSizeGb: 200
        bootDiskType: hyperdisk-balanced
        bootDiskProvisionedIops: 3060
        bootDiskProvisionedThroughput: 155
        numLocalSsds: 1
        localSsdInterface: nvme
      instanceFlexibilityPolicy:
        instanceSelectionList:
          - machineTypes:
              - n2-standard-4
              - n2d-standard-4
            rank: 1
    secondaryWorkerConfig:
      numInstances: 4
      preemptibility: SPOT
      diskConfig:
        bootDiskSizeGb: 100
      instanceFlexibilityPolicy:
        instanceSelectionList:
          - machineTypes:
              - n2-standard-4
              - n2d-standard-4
            rank: 0
          - machineTypes:
              - e2-standard-4
            rank: 1
        provisioningModelMix:
          standardCapacityBase: 1
          standardCapacityPercentAboveBase: 25
    softwareConfig:
      imageVersion: "2.2-debian12"
      optionalComponents:
        - JUPYTER
      properties:
        "spark:spark.executor.memory": "4g"
        "spark:spark.dynamicAllocation.enabled": "true"
        "yarn:yarn.nodemanager.resource.memory-mb": "12288"
    autoscalingPolicyUri:
      value: projects/my-gcp-project-123/locations/us-central1/autoscalingPolicies/etl-autoscaling # replace with your policy resource name
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "1800s"
      idleStopTtl: "900s"
    dataprocMetricConfig:
      metrics:
        - metricSource: SPARK
        - metricSource: YARN
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.clusterName` | `string` | yes |  |  |
| `spec.clusterConfig` | `GcpDataprocClusterConfig` |  |  |  |
| `spec.clusterConfig.stagingBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.clusterConfig.tempBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.clusterConfig.clusterTier` | `string` |  |  |  |
| `spec.clusterConfig.gceConfig` | `GcpDataprocClusterGceConfig` |  |  |  |
| `spec.clusterConfig.gceConfig.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.clusterConfig.gceConfig.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.clusterConfig.gceConfig.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.clusterConfig.gceConfig.serviceAccountScopes` | `[]string` |  |  |  |
| `spec.clusterConfig.gceConfig.zone` | `string` |  |  |  |
| `spec.clusterConfig.gceConfig.internalIpOnly` | `bool` |  |  |  |
| `spec.clusterConfig.gceConfig.tags` | `[]string` |  |  |  |
| `spec.clusterConfig.gceConfig.metadata` | `map<string, string>` |  |  |  |
| `spec.clusterConfig.gceConfig.shieldedInstanceConfig` | `GcpDataprocClusterShieldedInstanceConfig` |  |  |  |
| `spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableVtpm` | `bool` |  |  |  |
| `spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  |  |  |
| `spec.clusterConfig.gceConfig.reservationAffinity` | `GcpDataprocClusterReservationAffinity` |  |  |  |
| `spec.clusterConfig.gceConfig.reservationAffinity.consumeReservationType` | `string` |  |  |  |
| `spec.clusterConfig.gceConfig.reservationAffinity.key` | `string` |  |  |  |
| `spec.clusterConfig.gceConfig.reservationAffinity.values` | `[]string` |  |  |  |
| `spec.clusterConfig.gceConfig.nodeGroupAffinity` | `GcpDataprocClusterNodeGroupAffinity` |  |  |  |
| `spec.clusterConfig.gceConfig.nodeGroupAffinity.nodeGroupUri` | `string` | yes |  |  |
| `spec.clusterConfig.gceConfig.confidentialInstanceConfig` | `GcpDataprocClusterConfidentialInstanceConfig` |  |  |  |
| `spec.clusterConfig.gceConfig.confidentialInstanceConfig.enableConfidentialCompute` | `bool` |  |  |  |
| `spec.clusterConfig.gceConfig.confidentialInstanceConfig.confidentialInstanceType` | `string` |  |  |  |
| `spec.clusterConfig.gceConfig.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.clusterConfig.masterConfig` | `GcpDataprocClusterMasterConfig` |  |  |  |
| `spec.clusterConfig.masterConfig.numInstances` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.machineType` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.accelerators` | `[]GcpDataprocClusterAccelerator` |  |  |  |
| `spec.clusterConfig.masterConfig.accelerators[].acceleratorType` | `string` | yes |  |  |
| `spec.clusterConfig.masterConfig.accelerators[].acceleratorCount` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.minCpuPlatform` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.imageUri` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy` | `GcpDataprocClusterInstanceFlexibilityPolicy` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList` | `[]GcpDataprocClusterInstanceSelection` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes` | `[]string` | yes |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix` | `GcpDataprocClusterProvisioningModelMix` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase` | `int32` |  |  |  |
| `spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig` | `GcpDataprocClusterWorkerConfig` |  |  |  |
| `spec.clusterConfig.workerConfig.numInstances` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.machineType` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.accelerators` | `[]GcpDataprocClusterAccelerator` |  |  |  |
| `spec.clusterConfig.workerConfig.accelerators[].acceleratorType` | `string` | yes |  |  |
| `spec.clusterConfig.workerConfig.accelerators[].acceleratorCount` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.minCpuPlatform` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.imageUri` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.minNumInstances` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy` | `GcpDataprocClusterInstanceFlexibilityPolicy` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList` | `[]GcpDataprocClusterInstanceSelection` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes` | `[]string` | yes |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix` | `GcpDataprocClusterProvisioningModelMix` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase` | `int32` |  |  |  |
| `spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig` | `GcpDataprocClusterSecondaryWorkerConfig` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.numInstances` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.preemptibility` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy` | `GcpDataprocClusterInstanceFlexibilityPolicy` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList` | `[]GcpDataprocClusterInstanceSelection` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes` | `[]string` | yes |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix` | `GcpDataprocClusterProvisioningModelMix` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase` | `int32` |  |  |  |
| `spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase` | `int32` |  |  |  |
| `spec.clusterConfig.softwareConfig` | `GcpDataprocClusterSoftwareConfig` |  |  |  |
| `spec.clusterConfig.softwareConfig.imageVersion` | `string` |  |  |  |
| `spec.clusterConfig.softwareConfig.optionalComponents` | `[]string` |  |  |  |
| `spec.clusterConfig.softwareConfig.properties` | `map<string, string>` |  |  |  |
| `spec.clusterConfig.initializationActions` | `[]GcpDataprocClusterInitAction` |  |  |  |
| `spec.clusterConfig.initializationActions[].script` | `string` | yes |  |  |
| `spec.clusterConfig.initializationActions[].timeoutSec` | `int32` |  |  |  |
| `spec.clusterConfig.autoscalingPolicyUri` | `string \| valueFrom` |  |  | GcpDataprocAutoscalingPolicy (`status.outputs.name`) |
| `spec.clusterConfig.encryptionKmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.clusterConfig.securityConfig` | `GcpDataprocClusterSecurityConfig` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig` | `GcpDataprocClusterKerberosConfig` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.enableKerberos` | `bool` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.rootPrincipalPasswordUri` | `string` | yes |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.kmsKeyUri` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.clusterConfig.securityConfig.kerberosConfig.realm` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.tgtLifetimeHours` | `int32` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.kdcDbKeyUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.keystoreUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.keystorePasswordUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.keyPasswordUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.truststoreUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.truststorePasswordUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustRealm` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustKdc` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustAdminServer` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustSharedPasswordUri` | `string` |  |  |  |
| `spec.clusterConfig.securityConfig.identityConfig` | `GcpDataprocClusterIdentityConfig` |  |  |  |
| `spec.clusterConfig.securityConfig.identityConfig.userServiceAccountMapping` | `map<string, string>` | yes |  |  |
| `spec.clusterConfig.endpointConfig` | `GcpDataprocClusterEndpointConfig` |  |  |  |
| `spec.clusterConfig.endpointConfig.enableHttpPortAccess` | `bool` |  |  |  |
| `spec.clusterConfig.lifecycleConfig` | `GcpDataprocClusterLifecycleConfig` |  |  |  |
| `spec.clusterConfig.lifecycleConfig.idleDeleteTtl` | `string` |  |  |  |
| `spec.clusterConfig.lifecycleConfig.autoDeleteTime` | `string` |  |  |  |
| `spec.clusterConfig.lifecycleConfig.idleStopTtl` | `string` |  |  |  |
| `spec.clusterConfig.lifecycleConfig.autoStopTime` | `string` |  |  |  |
| `spec.clusterConfig.metastoreConfig` | `GcpDataprocClusterMetastoreConfig` |  |  |  |
| `spec.clusterConfig.metastoreConfig.dataprocMetastoreService` | `string \| valueFrom` | yes |  |  |
| `spec.clusterConfig.dataprocMetricConfig` | `GcpDataprocClusterMetricConfig` |  |  |  |
| `spec.clusterConfig.dataprocMetricConfig.metrics` | `[]GcpDataprocClusterMetric` | yes |  |  |
| `spec.clusterConfig.dataprocMetricConfig.metrics[].metricSource` | `string` | yes |  |  |
| `spec.clusterConfig.dataprocMetricConfig.metrics[].metricOverrides` | `[]string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups` | `[]GcpDataprocClusterAuxiliaryNodeGroup` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].roles` | `[]string` | yes |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig` | `GcpDataprocClusterAuxiliaryNodeGroupConfig` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.numInstances` | `int32` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.machineType` | `string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.minCpuPlatform` | `string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig` | `GcpDataprocClusterDiskConfig` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskType` | `string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.numLocalSsds` | `int32` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.localSsdInterface` | `string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskProvisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskProvisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks` | `[]GcpDataprocClusterAttachedDisk` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].diskSizeGb` | `int32` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].diskType` | `string` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].provisionedIops` | `int64` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].provisionedThroughput` | `int64` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators` | `[]GcpDataprocClusterAccelerator` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators[].acceleratorType` | `string` | yes |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators[].acceleratorCount` | `int32` |  |  |  |
| `spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupId` | `string` |  |  |  |
| `spec.clusterConfig.clusterType` | `string` |  |  |  |
| `spec.clusterConfig.engine` | `string` |  |  |  |
| `spec.virtualClusterConfig` | `GcpDataprocClusterVirtualClusterConfig` |  |  |  |
| `spec.virtualClusterConfig.stagingBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.virtualClusterConfig.kubernetesClusterConfig` | `GcpDataprocClusterKubernetesClusterConfig` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesNamespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig` | `GcpDataprocClusterGkeClusterConfig` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget` | `string \| valueFrom` | yes |  | GcpGkeCluster (`status.outputs.cluster_id`) |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget` | `[]GcpDataprocClusterNodePoolTarget` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePool` | `string \| valueFrom` | yes |  | GcpGkeNodePool (`status.outputs.node_pool_id`) |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].roles` | `[]string` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig` | `GcpDataprocClusterNodePoolConfig` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.locations` | `[]string` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling` | `GcpDataprocClusterNodePoolAutoscaling` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling.minNodeCount` | `int32` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling.maxNodeCount` | `int32` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.machineType` | `string` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.localSsdCount` | `int32` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.minCpuPlatform` | `string` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.preemptible` | `bool` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.spot` | `bool` |  |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig` | `GcpDataprocClusterKubernetesSoftwareConfig` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig.componentVersion` | `map<string, string>` | yes |  |  |
| `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig.properties` | `map<string, string>` |  |  |  |
| `spec.virtualClusterConfig.auxiliaryServicesConfig` | `GcpDataprocClusterAuxiliaryServicesConfig` |  |  |  |
| `spec.virtualClusterConfig.auxiliaryServicesConfig.metastoreConfig` | `GcpDataprocClusterMetastoreConfig` |  |  |  |
| `spec.virtualClusterConfig.auxiliaryServicesConfig.metastoreConfig.dataprocMetastoreService` | `string \| valueFrom` | yes |  |  |
| `spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig` | `GcpDataprocClusterSparkHistoryServerConfig` |  |  |  |
| `spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig.dataprocCluster` | `string \| valueFrom` |  |  | GcpDataprocCluster (`status.outputs.cluster_id`) |
| `spec.gracefulDecommissionTimeout` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the Dataproc cluster will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

GCP region for the cluster (e.g., "us-central1", "europe-west12").
All cluster nodes will be placed in this region.
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.clusterName

`string` · required

Name of the Dataproc cluster. Must start with a lowercase letter,
can contain lowercase letters, numbers, and hyphens, and must end
with a lowercase letter or number. Maximum 55 characters.
Immutable after creation.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"55","pattern":"^[a-z][a-z0-9-]{0,53}[a-z0-9]$"}}

### spec.clusterConfig

`GcpDataprocClusterConfig`

Standard (Compute Engine) cluster arm: nodes, software, networking,
encryption, security, and lifecycle. Mutually exclusive with
virtual_cluster_config.

### spec.clusterConfig.stagingBucket

`string | valueFrom`

Cloud Storage bucket for staging job dependencies, jar files, and
other temporary data. If not specified, GCP auto-creates a staging
bucket in the cluster's project and region.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.clusterConfig.tempBucket

`string | valueFrom`

Cloud Storage bucket for ephemeral cluster data (shuffle, spill).
If not specified, GCP auto-creates a temp bucket.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.clusterConfig.clusterTier

`string`

Cluster tier controlling the Dataproc feature set and SLA.
CLUSTER_TIER_STANDARD: the default tier.
CLUSTER_TIER_PREMIUM: premium features (e.g. advanced repair and
faster scaling). Immutable after creation.

- rule: cluster_tier must be CLUSTER_TIER_STANDARD or CLUSTER_TIER_PREMIUM

### spec.clusterConfig.gceConfig

`GcpDataprocClusterGceConfig`

Compute Engine configuration for the cluster's nodes (networking,
service account, zone, tags, metadata, hardening, placement).

- rule: only one of network or subnetwork may be set

### spec.clusterConfig.gceConfig.network

`string | valueFrom`

VPC network for the cluster's nodes. Mutually exclusive with subnetwork.
Expects a network self-link or resource reference.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.clusterConfig.gceConfig.subnetwork

`string | valueFrom`

VPC subnetwork for the cluster's nodes. Mutually exclusive with network.
Using a subnetwork is recommended for production clusters with
controlled IP ranges.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.clusterConfig.gceConfig.serviceAccount

`string | valueFrom`

Service account for the cluster's VMs. If not specified, the default
Compute Engine service account is used. A custom service account with
minimal permissions is recommended for production.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.clusterConfig.gceConfig.serviceAccountScopes

`[]string`

OAuth 2.0 scopes for the service account. If not specified, GCP uses
a default set of scopes. Override only when you need to restrict or
expand API access.

### spec.clusterConfig.gceConfig.zone

`string`

GCP zone within the region for node placement. If not specified,
GCP auto-selects a zone within the cluster's region.

### spec.clusterConfig.gceConfig.internalIpOnly

`bool`

Whether to use only internal IP addresses for cluster nodes.
When true, nodes have no external IP and require Cloud NAT or
Private Google Access for internet connectivity.
Recommended for production to reduce attack surface.

### spec.clusterConfig.gceConfig.tags

`[]string`

GCE network tags applied to all cluster nodes. Useful for
firewall rule targeting.

### spec.clusterConfig.gceConfig.metadata

`map<string, string>`

Compute Engine metadata key-value pairs applied to all cluster nodes.
Common use: startup scripts, environment variables for init actions.

### spec.clusterConfig.gceConfig.shieldedInstanceConfig

`GcpDataprocClusterShieldedInstanceConfig`

Shielded VM hardening (secure boot, vTPM, integrity monitoring)
for all cluster nodes.

### spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableSecureBoot

`bool`

Verify the boot integrity of the VM using a secure boot chain.

### spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableVtpm

`bool`

Enable the virtual Trusted Platform Module (vTPM).

### spec.clusterConfig.gceConfig.shieldedInstanceConfig.enableIntegrityMonitoring

`bool`

Enable integrity monitoring comparing boot measurements against a
known-good baseline.

### spec.clusterConfig.gceConfig.reservationAffinity

`GcpDataprocClusterReservationAffinity`

Reservation affinity for guaranteed compute capacity.

- rule: key and values are required when consume_reservation_type is SPECIFIC_RESERVATION

### spec.clusterConfig.gceConfig.reservationAffinity.consumeReservationType

`string`

How the cluster consumes reservations.
NO_RESERVATION: never consume reservations.
ANY_RESERVATION: consume any matching open reservation (default).
SPECIFIC_RESERVATION: consume only the reservation named by key/values.

- rule: consume_reservation_type must be NO_RESERVATION, ANY_RESERVATION, or SPECIFIC_RESERVATION

### spec.clusterConfig.gceConfig.reservationAffinity.key

`string`

Reservation label key. For SPECIFIC_RESERVATION use
"compute.googleapis.com/reservation-name".

### spec.clusterConfig.gceConfig.reservationAffinity.values

`[]string`

Reservation label values (the reservation name for
SPECIFIC_RESERVATION).

### spec.clusterConfig.gceConfig.nodeGroupAffinity

`GcpDataprocClusterNodeGroupAffinity`

Sole-tenant node group placement for compliance or licensing
isolation.

### spec.clusterConfig.gceConfig.nodeGroupAffinity.nodeGroupUri

`string` · required

URI of the sole-tenant node group the cluster will be created on.
Format: projects/{project}/zones/{zone}/nodeGroups/{node_group}
(shorter forms accepted by the API also work).

- rule: {"required":true}

### spec.clusterConfig.gceConfig.confidentialInstanceConfig

`GcpDataprocClusterConfidentialInstanceConfig`

Confidential VM (data-in-use encryption) for all cluster nodes.
Requires an N2D machine type.

### spec.clusterConfig.gceConfig.confidentialInstanceConfig.enableConfidentialCompute

`bool`

Enable Confidential Compute for all cluster nodes with the AMD SEV
technology. The provider marks this boolean deprecated in favor of
confidential_instance_type, which names the technology explicitly;
prefer that field for new clusters. Both may be set for a cluster
authored before the type field existed.

### spec.clusterConfig.gceConfig.confidentialInstanceConfig.confidentialInstanceType

`string`

Confidential Compute technology for all cluster nodes:
  "SEV"     -- AMD Secure Encrypted Virtualization (N2D, C2D, C3D)
  "SEV_SNP" -- AMD SEV with Secure Nested Paging: adds integrity
               protection and attestation (N2D)
  "TDX"     -- Intel Trust Domain Extensions (C3)
The machine type must support the chosen technology. Immutable.

- rule: confidential_instance_type must be SEV, SEV_SNP, or TDX

### spec.clusterConfig.gceConfig.resourceManagerTags

`map<string, string>`

Resource manager (secure) tags applied to all cluster instances,
keyed "tagKeys/{tag_key_id}" with values "tagValues/{tag_value_id}".
Unlike network tags, these are IAM-governed and usable in org
policies and firewall rules.

### spec.clusterConfig.masterConfig

`GcpDataprocClusterMasterConfig`

Master node configuration. If not specified, GCP defaults to
1 master with a default machine type and 500 GB pd-standard disk.

- rule: provisioning_model_mix applies to secondary workers only
- rule: machine_type and instance_flexibility_policy are mutually exclusive; rank machine types in the flexibility policy's instance_selection_list instead

### spec.clusterConfig.masterConfig.numInstances

`int32`

Number of master instances. Valid values: 1 (standard) or 3 (HA).
If not specified, GCP defaults to 1.

### spec.clusterConfig.masterConfig.machineType

`string`

Compute Engine machine type (e.g., "n2-standard-4", "e2-standard-8").
If not specified, GCP selects a default machine type.
Mutually exclusive with instance_flexibility_policy — a flexibility
policy replaces the single machine type (see that field).

### spec.clusterConfig.masterConfig.diskConfig

`GcpDataprocClusterDiskConfig`

Boot disk and local SSD configuration.

### spec.clusterConfig.masterConfig.diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.masterConfig.diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.masterConfig.diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.masterConfig.diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.masterConfig.diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.accelerators

`[]GcpDataprocClusterAccelerator`

GPU accelerators attached to master nodes.
Typically not needed for masters unless running single-node ML workloads.

### spec.clusterConfig.masterConfig.accelerators[].acceleratorType

`string` · required

Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
See https://cloud.google.com/compute/docs/gpus for available types.

- rule: {"required":true}

### spec.clusterConfig.masterConfig.accelerators[].acceleratorCount

`int32`

Number of accelerators to attach. Must be at least 1.

- rule: {"int32":{"gte":1}}

### spec.clusterConfig.masterConfig.minCpuPlatform

`string`

Minimum CPU platform for the instances (e.g., "Intel Cascade Lake").
Forces nodes onto a specific or newer CPU generation.

### spec.clusterConfig.masterConfig.imageUri

`string`

Custom Dataproc image URI. If not specified, GCP uses the image
determined by software_config.image_version.

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy

`GcpDataprocClusterInstanceFlexibilityPolicy`

Ranked machine-type preferences for master provisioning — keeps the
cluster creatable when the preferred type's zonal capacity dries up.
REPLACES machine_type: with a flexibility policy present the API
provisions solely from the ranked selections and drops a paired
machineTypeUri from the stored config. Masters are on-demand
capacity: provisioning_model_mix does not apply here (secondary
workers only).

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList

`[]GcpDataprocClusterInstanceSelection`

Ranked machine-type preferences. Dataproc provisions from the
lowest-rank entry with available capacity.

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes

`[]string` · required

Machine types to try for this selection entry (e.g.
["n2-standard-8", "n2d-standard-8"]).

- rule: {"repeated":{"minItems":"1"}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank

`int32`

Preference rank. Lower rank is preferred; Dataproc falls back to
higher ranks when capacity for the preferred types is unavailable.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig

`GcpDataprocClusterDiskConfig`

Disk shape for nodes provisioned from THIS selection entry —
overrides the role's disk_config so, for example, the fallback
machine types can carry a different boot disk or local SSD count
than the preferred ones. Leave unset to inherit the role's
disk_config. Immutable.

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix

`GcpDataprocClusterProvisioningModelMix`

Standard/spot capacity mix for the group.

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase

`int32`

Number of secondary workers that must be standard (on-demand)
capacity before any spot VMs are used.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.masterConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase

`int32`

Percentage of capacity above the base that should also be standard
(0-100). The remainder is provisioned as spot.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.clusterConfig.workerConfig

`GcpDataprocClusterWorkerConfig`

Primary worker node configuration. If not specified, GCP defaults
to 2 workers. For a single-node cluster set the software property
"dataproc:dataproc.allow.zero.workers" = "true" instead of a
worker count of zero.

- rule: provisioning_model_mix applies to secondary workers only
- rule: machine_type and instance_flexibility_policy are mutually exclusive; rank machine types in the flexibility policy's instance_selection_list instead

### spec.clusterConfig.workerConfig.numInstances

`int32`

Number of primary worker instances.
If not specified, GCP defaults to 2. This is one of the few fields
that can be changed in place after creation (manual scaling).

### spec.clusterConfig.workerConfig.machineType

`string`

Compute Engine machine type (e.g., "n2-standard-4", "e2-standard-8").
If not specified, GCP selects a default machine type.
Mutually exclusive with instance_flexibility_policy — a flexibility
policy replaces the single machine type (see that field).

### spec.clusterConfig.workerConfig.diskConfig

`GcpDataprocClusterDiskConfig`

Boot disk and local SSD configuration.

### spec.clusterConfig.workerConfig.diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.workerConfig.diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.workerConfig.diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.workerConfig.diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.workerConfig.diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.accelerators

`[]GcpDataprocClusterAccelerator`

GPU accelerators attached to worker nodes.
Common for Spark ML workloads using GPU-accelerated libraries.

### spec.clusterConfig.workerConfig.accelerators[].acceleratorType

`string` · required

Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
See https://cloud.google.com/compute/docs/gpus for available types.

- rule: {"required":true}

### spec.clusterConfig.workerConfig.accelerators[].acceleratorCount

`int32`

Number of accelerators to attach. Must be at least 1.

- rule: {"int32":{"gte":1}}

### spec.clusterConfig.workerConfig.minCpuPlatform

`string`

Minimum CPU platform for the instances.

### spec.clusterConfig.workerConfig.imageUri

`string`

Custom Dataproc image URI.

### spec.clusterConfig.workerConfig.minNumInstances

`int32`

Minimum number of primary worker instances when autoscaling is active.
The autoscaler will not scale below this threshold. Updatable in place.

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy

`GcpDataprocClusterInstanceFlexibilityPolicy`

Ranked machine-type preferences for primary-worker provisioning —
keeps scale-ups schedulable when the preferred type's zonal capacity
dries up. REPLACES machine_type: with a flexibility policy present
the API provisions solely from the ranked selections and drops a
paired machineTypeUri from the stored config. Primary workers are
on-demand capacity: provisioning_model_mix does not apply here
(secondary workers only).

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList

`[]GcpDataprocClusterInstanceSelection`

Ranked machine-type preferences. Dataproc provisions from the
lowest-rank entry with available capacity.

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes

`[]string` · required

Machine types to try for this selection entry (e.g.
["n2-standard-8", "n2d-standard-8"]).

- rule: {"repeated":{"minItems":"1"}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank

`int32`

Preference rank. Lower rank is preferred; Dataproc falls back to
higher ranks when capacity for the preferred types is unavailable.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig

`GcpDataprocClusterDiskConfig`

Disk shape for nodes provisioned from THIS selection entry —
overrides the role's disk_config so, for example, the fallback
machine types can carry a different boot disk or local SSD count
than the preferred ones. Leave unset to inherit the role's
disk_config. Immutable.

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix

`GcpDataprocClusterProvisioningModelMix`

Standard/spot capacity mix for the group.

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase

`int32`

Number of secondary workers that must be standard (on-demand)
capacity before any spot VMs are used.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.workerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase

`int32`

Percentage of capacity above the base that should also be standard
(0-100). The remainder is provisioned as spot.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.clusterConfig.secondaryWorkerConfig

`GcpDataprocClusterSecondaryWorkerConfig`

Secondary (preemptible/spot) worker node configuration.
If not specified, no secondary workers are created.

### spec.clusterConfig.secondaryWorkerConfig.numInstances

`int32`

Number of secondary worker instances. Default: 0.
Updatable in place (manual scaling).

### spec.clusterConfig.secondaryWorkerConfig.preemptibility

`string`

Preemptibility of the secondary workers.
SPOT: modern spot VMs with dynamic pricing (recommended).
PREEMPTIBLE: legacy preemptible VMs with fixed pricing.
NON_PREEMPTIBLE: standard on-demand pricing (unusual for secondary workers).
If not specified, GCP defaults to PREEMPTIBLE.
Immutable after creation.

- rule: preemptibility must be PREEMPTIBLE, SPOT, or NON_PREEMPTIBLE

### spec.clusterConfig.secondaryWorkerConfig.diskConfig

`GcpDataprocClusterDiskConfig`

Boot disk and local SSD configuration for secondary workers.
Secondary workers do not support custom machine types or accelerators
directly; they inherit machine configuration from the primary worker
config unless an instance_flexibility_policy overrides it.

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy

`GcpDataprocClusterInstanceFlexibilityPolicy`

Machine-type flexibility and standard/spot capacity mix for the
group — only secondary workers support flexible provisioning.

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList

`[]GcpDataprocClusterInstanceSelection`

Ranked machine-type preferences. Dataproc provisions from the
lowest-rank entry with available capacity.

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].machineTypes

`[]string` · required

Machine types to try for this selection entry (e.g.
["n2-standard-8", "n2d-standard-8"]).

- rule: {"repeated":{"minItems":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].rank

`int32`

Preference rank. Lower rank is preferred; Dataproc falls back to
higher ranks when capacity for the preferred types is unavailable.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig

`GcpDataprocClusterDiskConfig`

Disk shape for nodes provisioned from THIS selection entry —
overrides the role's disk_config so, for example, the fallback
machine types can carry a different boot disk or local SSD count
than the preferred ones. Leave unset to inherit the role's
disk_config. Immutable.

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.instanceSelectionList[].diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix

`GcpDataprocClusterProvisioningModelMix`

Standard/spot capacity mix for the group.

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityBase

`int32`

Number of secondary workers that must be standard (on-demand)
capacity before any spot VMs are used.

- rule: {"int32":{"gte":0}}

### spec.clusterConfig.secondaryWorkerConfig.instanceFlexibilityPolicy.provisioningModelMix.standardCapacityPercentAboveBase

`int32`

Percentage of capacity above the base that should also be standard
(0-100). The remainder is provisioned as spot.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.clusterConfig.softwareConfig

`GcpDataprocClusterSoftwareConfig`

Software configuration including Dataproc image version, optional
components, and framework property overrides.

### spec.clusterConfig.softwareConfig.imageVersion

`string`

Dataproc image version (e.g., "2.2-debian12", "2.1-ubuntu20").
See https://cloud.google.com/dataproc/docs/concepts/versioning/dataproc-versions
If not specified, GCP uses the latest stable version.

### spec.clusterConfig.softwareConfig.optionalComponents

`[]string`

Optional components to install on the cluster. Common values:
JUPYTER, DOCKER, PRESTO, ZEPPELIN, HIVE_WEBHCAT, FLINK, TRINO.
See https://cloud.google.com/dataproc/docs/concepts/components/overview

### spec.clusterConfig.softwareConfig.properties

`map<string, string>`

Key-value pairs to override or set Hadoop, Spark, YARN, and other
framework properties. Keys use the format "prefix:property", e.g.:
  "spark:spark.executor.memory" = "4g"
  "hdfs:dfs.replication" = "2"
  "dataproc:dataproc.allow.zero.workers" = "true"  (single-node cluster)
See https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/cluster-properties

### spec.clusterConfig.initializationActions

`[]GcpDataprocClusterInitAction`

Initialization actions (startup scripts) that run on all nodes
when the cluster is created.

### spec.clusterConfig.initializationActions[].script

`string` · required

GCS URI of the initialization script (must start with "gs://").

- rule: {"required":true}

### spec.clusterConfig.initializationActions[].timeoutSec

`int32`

Maximum time (in seconds) the script is allowed to run before
being forcefully terminated. Default: 300 (5 minutes).

### spec.clusterConfig.autoscalingPolicyUri

`string | valueFrom`

Autoscaling policy that governs worker scaling for this cluster.
Resolves to the policy's full resource name
(projects/{project}/locations/{location}/autoscalingPolicies/{policy}).
Attaching, swapping, or detaching the policy updates in place.

- references: GcpDataprocAutoscalingPolicy (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDataprocAutoscalingPolicy, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.clusterConfig.encryptionKmsKeyName

`string | valueFrom`

Cloud KMS key for encrypting persistent disks attached to cluster
nodes (CMEK). Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
If not specified, disks are encrypted with Google-managed keys.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.clusterConfig.securityConfig

`GcpDataprocClusterSecurityConfig`

In-cluster authentication hardening: Kerberos or personal-cluster
identity mapping.

- rule: exactly one of kerberos_config or identity_config must be set

### spec.clusterConfig.securityConfig.kerberosConfig

`GcpDataprocClusterKerberosConfig`

Kerberos (Hadoop Secure Mode) configuration.

### spec.clusterConfig.securityConfig.kerberosConfig.enableKerberos

`bool`

Flag to indicate whether to Kerberize the cluster.

### spec.clusterConfig.securityConfig.kerberosConfig.rootPrincipalPasswordUri

`string` · required

Cloud Storage URI of a KMS-encrypted file containing the root
principal password. Required to enable Kerberos.

- rule: {"required":true}

### spec.clusterConfig.securityConfig.kerberosConfig.kmsKeyUri

`string | valueFrom` · required

URI of the KMS key used to decrypt the root password and the other
encrypted files below.
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.clusterConfig.securityConfig.kerberosConfig.realm

`string`

Custom Kerberos realm. If unset, Dataproc derives a realm from the
cluster's hostname.

### spec.clusterConfig.securityConfig.kerberosConfig.tgtLifetimeHours

`int32`

Lifetime of ticket-granting tickets, in hours. Default: 10.

### spec.clusterConfig.securityConfig.kerberosConfig.kdcDbKeyUri

`string`

Cloud Storage URI of a KMS-encrypted file containing the master key
of the KDC database.

### spec.clusterConfig.securityConfig.kerberosConfig.keystoreUri

`string`

Cloud Storage URI of the keystore file (TLS). If unset, Dataproc
generates a self-signed certificate.

### spec.clusterConfig.securityConfig.kerberosConfig.keystorePasswordUri

`string`

Cloud Storage URI of a KMS-encrypted file with the keystore password.

### spec.clusterConfig.securityConfig.kerberosConfig.keyPasswordUri

`string`

Cloud Storage URI of a KMS-encrypted file with the key password.

### spec.clusterConfig.securityConfig.kerberosConfig.truststoreUri

`string`

Cloud Storage URI of the truststore file (TLS).

### spec.clusterConfig.securityConfig.kerberosConfig.truststorePasswordUri

`string`

Cloud Storage URI of a KMS-encrypted file with the truststore password.

### spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustRealm

`string`

Remote realm for cross-realm trust.

### spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustKdc

`string`

KDC host of the remote trusted realm.

### spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustAdminServer

`string`

Admin server host of the remote trusted realm.

### spec.clusterConfig.securityConfig.kerberosConfig.crossRealmTrustSharedPasswordUri

`string`

Cloud Storage URI of a KMS-encrypted file with the shared password
between the on-cluster KDC and the remote trusted realm.

### spec.clusterConfig.securityConfig.identityConfig

`GcpDataprocClusterIdentityConfig`

Personal cluster authentication (user-to-service-account mapping).

### spec.clusterConfig.securityConfig.identityConfig.userServiceAccountMapping

`map<string, string>` · required

Map of user accounts to the service accounts they operate as
(e.g. "bob@example.com" -> "bob-sa@project.iam.gserviceaccount.com").

- rule: {"map":{"minPairs":"1"}}

### spec.clusterConfig.endpointConfig

`GcpDataprocClusterEndpointConfig`

Component Gateway configuration for authenticated web UI access.

### spec.clusterConfig.endpointConfig.enableHttpPortAccess

`bool`

Whether to enable the Dataproc Component Gateway for web UI access.
When enabled, GCP creates authenticated HTTPS endpoints for each
cluster component. Requires the cluster to have external IP access
or appropriate Private Google Access configuration.

### spec.clusterConfig.lifecycleConfig

`GcpDataprocClusterLifecycleConfig`

Lifecycle configuration for automatic cluster deletion.
Critical for cost management of ephemeral batch processing clusters.

### spec.clusterConfig.lifecycleConfig.idleDeleteTtl

`string`

Duration of inactivity after which the cluster is automatically
deleted. Format: duration in seconds with 's' suffix (e.g., "1800s"
for 30 minutes). Valid range: 10 minutes to 14 days.

A cluster is considered idle when no jobs are running and no
interactive sessions are active.

- rule: idle_delete_ttl must be a duration in seconds (e.g., '1800s')

### spec.clusterConfig.lifecycleConfig.autoDeleteTime

`string`

RFC3339 timestamp at which the cluster is automatically deleted,
regardless of activity (e.g., "2026-03-01T00:00:00Z").
Useful for time-boxed clusters with a known end date.

### spec.clusterConfig.lifecycleConfig.idleStopTtl

`string`

Duration of inactivity after which the cluster is automatically
STOPPED (not deleted) — VMs shut down, storage and configuration
retained, restartable later. Same format and range as
idle_delete_ttl (e.g., "1800s"; 10 minutes to 14 days).

- rule: idle_stop_ttl must be a duration in seconds (e.g., '1800s')

### spec.clusterConfig.lifecycleConfig.autoStopTime

`string`

RFC3339 timestamp at which the cluster is automatically STOPPED
(not deleted), regardless of activity.

### spec.clusterConfig.metastoreConfig

`GcpDataprocClusterMetastoreConfig`

Attach the cluster to a persistent Dataproc Metastore service.

### spec.clusterConfig.metastoreConfig.dataprocMetastoreService

`string | valueFrom` · required

Resource name of an existing Dataproc Metastore service.
Format: projects/{project}/locations/{location}/services/{service}
Accepts a literal resource name today; references attach when a
metastore-service kind lands in the catalog.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.clusterConfig.dataprocMetricConfig

`GcpDataprocClusterMetricConfig`

OSS metric collection into Cloud Monitoring.

### spec.clusterConfig.dataprocMetricConfig.metrics

`[]GcpDataprocClusterMetric` · required

Metric sources to collect.

- rule: {"repeated":{"minItems":"1"}}

### spec.clusterConfig.dataprocMetricConfig.metrics[].metricSource

`string` · required

Source to collect metrics from.

- rule: {"required":true,"string":{"in":["MONITORING_AGENT_DEFAULTS","HDFS","SPARK","YARN","SPARK_HISTORY_SERVER","HIVESERVER2"]}}

### spec.clusterConfig.dataprocMetricConfig.metrics[].metricOverrides

`[]string`

Specific metrics to collect, overriding the source's default set
(e.g. "yarn:ResourceManager:QueueMetrics:AppsCompleted").

### spec.clusterConfig.auxiliaryNodeGroups

`[]GcpDataprocClusterAuxiliaryNodeGroup`

Dedicated DRIVER node groups, separating Spark driver capacity
from the master's control-plane duties.

### spec.clusterConfig.auxiliaryNodeGroups[].roles

`[]string` · required

Roles assigned to the group. The API currently supports DRIVER.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"auxiliary_node_group_role_valid_value","message":"each role must be DRIVER","expression":"this in ['DRIVER']"}]}}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig

`GcpDataprocClusterAuxiliaryNodeGroupConfig`

VM sizing for the group.

- rule: attached_disks apply to master, worker, and secondary worker nodes only -- Dataproc does not attach additional disks to auxiliary (driver) node groups

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.numInstances

`int32`

Number of instances in the group.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.machineType

`string`

Compute Engine machine type for the group's instances.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.minCpuPlatform

`string`

Minimum CPU platform for the group's instances.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig

`GcpDataprocClusterDiskConfig`

Boot disk and local SSD configuration.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskSizeGb

`int32`

Size of the boot disk in GB. Minimum 10 GB.
If not specified, GCP defaults to 500 GB for master and worker nodes.

- rule: boot_disk_size_gb must be at least 10

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskType

`string`

Boot disk type: "pd-standard" (GCP default), "pd-ssd", "pd-balanced",
or "hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). The API validates availability per image version and
machine family at deploy time.

- rule: boot_disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.numLocalSsds

`int32`

Number of local SSDs to attach. Each local SSD is 375 GB.
Default: 0 (no local SSDs).

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.localSsdInterface

`string`

Interface used to attach local SSDs: "scsi" (default) or "nvme".
NVMe offers higher throughput for shuffle-heavy Spark workloads but
requires an image that ships NVMe drivers (all current Dataproc
images do).

- rule: local_ssd_interface must be scsi or nvme

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskProvisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second for the boot disk — the IOPS
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.bootDiskProvisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for the boot disk — the bandwidth
dial decoupled from disk size, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks

`[]GcpDataprocClusterAttachedDisk`

Additional persistent disks attached to every node of this role,
beyond the boot disk — for HDFS data or shuffle spill that should not
share the boot volume, or for Hyperdisk performance tiers the boot
disk cannot use. Each entry is one disk on each node. Immutable: the
whole set is fixed at cluster creation (ForceNew). Not available on
auxiliary (driver) node groups.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].diskSizeGb

`int32`

Size of the disk in GB. Leave unset for the API's default size.

- rule: disk_size_gb must be at least 10

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].diskType

`string`

Disk type: "pd-standard", "pd-ssd", "pd-balanced", or
"hyperdisk-balanced" (the class whose provisioned IOPS/throughput
dials apply). Leave empty for the API's default.

- rule: disk_type must be pd-standard, pd-ssd, pd-balanced, or hyperdisk-balanced

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].provisionedIops

`int64` · optional (explicit presence)

Provisioned I/O operations per second, honored by disk types that
support provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.diskConfig.attachedDisks[].provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s, honored by disk types that support
provisioned performance (hyperdisks).

- rule: {"int64":{"gte":"1"}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators

`[]GcpDataprocClusterAccelerator`

GPU accelerators attached to the group's instances.

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators[].acceleratorType

`string` · required

Full accelerator type name (e.g., "nvidia-tesla-t4", "nvidia-tesla-v100").
See https://cloud.google.com/compute/docs/gpus for available types.

- rule: {"required":true}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupConfig.accelerators[].acceleratorCount

`int32`

Number of accelerators to attach. Must be at least 1.

- rule: {"int32":{"gte":1}}

### spec.clusterConfig.auxiliaryNodeGroups[].nodeGroupId

`string`

Optional stable identifier for the group (3-33 characters).
If unset, Dataproc generates one.

- rule: node_group_id must be 3-33 characters

### spec.clusterConfig.clusterType

`string`

The cluster's structural type. STANDARD (default) has masters and
workers; SINGLE_NODE runs everything on one VM (the modern
alternative to the "dataproc:dataproc.allow.zero.workers" property);
ZERO_SCALE keeps only the control plane warm and provisions workers
on demand. Immutable after creation.

- rule: cluster_type must be STANDARD, SINGLE_NODE, or ZERO_SCALE

### spec.clusterConfig.engine

`string`

The execution engine. DEFAULT runs open-source Spark as-is;
LIGHTNING enables the Lightning Engine (Google's accelerated Spark
runtime, premium tier). Immutable after creation.

- rule: engine must be DEFAULT or LIGHTNING

### spec.virtualClusterConfig

`GcpDataprocClusterVirtualClusterConfig`

Dataproc-on-GKE arm: run Spark as pods on an existing GKE cluster.
Mutually exclusive with cluster_config.

### spec.virtualClusterConfig.stagingBucket

`string | valueFrom`

Cloud Storage bucket for staging job dependencies. If not
specified, GCP auto-creates a staging bucket.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.virtualClusterConfig.kubernetesClusterConfig

`GcpDataprocClusterKubernetesClusterConfig` · required

The Kubernetes-side configuration: target GKE cluster, node-pool
role mapping, namespace, and component versions.

- rule: {"required":true}

### spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesNamespace

`string | valueFrom`

Kubernetes namespace Dataproc deploys workloads into. If unset,
Dataproc uses a namespace named after the virtual cluster.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig

`GcpDataprocClusterGkeClusterConfig` · required

The target GKE cluster and node-pool role mapping.

- rule: {"required":true}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget

`string | valueFrom` · required

The target GKE cluster. Resolves to the cluster's fully qualified
resource name (projects/{project}/locations/{location}/clusters/{cluster}).

- references: GcpGkeCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget

`[]GcpDataprocClusterNodePoolTarget`

Node pools Dataproc schedules roles onto. When omitted, Dataproc
creates and manages a default pool on the target cluster.

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePool

`string | valueFrom` · required

The target GKE node pool. Resolves to the pool's fully qualified
resource name
(projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{pool}).
The pool may pre-exist (composed via GcpGkeNodePool) or be created
by Dataproc using node_pool_config.

- references: GcpGkeNodePool (`status.outputs.node_pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeNodePool, name: <that resource's name>, fieldPath: status.outputs.node_pool_id}} -- a bare string does not parse

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].roles

`[]string` · required

Dataproc roles scheduled onto this pool.
DEFAULT: catch-all for roles without a dedicated pool.
CONTROLLER: the Dataproc control plane pods.
SPARK_DRIVER: Spark driver pods.
SPARK_EXECUTOR: Spark executor pods.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"node_pool_role_valid_value","message":"each role must be one of DEFAULT, CONTROLLER, SPARK_DRIVER, SPARK_EXECUTOR","expression":"this in ['DEFAULT', 'CONTROLLER', 'SPARK_DRIVER', 'SPARK_EXECUTOR']"}]}}}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig

`GcpDataprocClusterNodePoolConfig`

Sizing for the pool when Dataproc creates it. Ignored for
pre-existing pools.

- rule: preemptible and spot are mutually exclusive

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.locations

`[]string` · required

Compute Engine zones where the pool's nodes are placed
(e.g. ["us-central1-a"]). Required when Dataproc creates the pool.

- rule: {"repeated":{"minItems":"1"}}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling

`GcpDataprocClusterNodePoolAutoscaling`

GKE autoscaling bounds for the pool.

- rule: max_node_count must be greater than or equal to min_node_count

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling.minNodeCount

`int32`

Minimum number of nodes per zone. May be 0 for scale-to-zero pools.

- rule: {"int32":{"gte":0}}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.autoscaling.maxNodeCount

`int32`

Maximum number of nodes per zone.

- rule: {"int32":{"gte":0}}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.machineType

`string`

Compute Engine machine type for the pool's nodes.

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.localSsdCount

`int32`

Number of local SSDs attached to each node.

- rule: {"int32":{"gte":0}}

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.minCpuPlatform

`string`

Minimum CPU platform for the pool's nodes.

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.preemptible

`bool`

Use legacy preemptible VMs for the pool. Mutually exclusive with spot.

### spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePoolConfig.spot

`bool`

Use modern spot VMs for the pool (recommended over preemptible).

### spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig

`GcpDataprocClusterKubernetesSoftwareConfig` · required

Component versions (SPARK is required) and framework properties.

- rule: {"required":true}

### spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig.componentVersion

`map<string, string>` · required

Component-to-version map. The SPARK component is required, e.g.
{"SPARK": "3.5-dataproc-17"}.
See https://cloud.google.com/dataproc/docs/guides/dpgke/dataproc-gke-version-compatibility

- rule: {"map":{"minPairs":"1"}}

### spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesSoftwareConfig.properties

`map<string, string>`

Framework properties (e.g. "spark:spark.eventLog.enabled" = "true").

### spec.virtualClusterConfig.auxiliaryServicesConfig

`GcpDataprocClusterAuxiliaryServicesConfig`

Shared persistent services (metastore, Spark History Server).

### spec.virtualClusterConfig.auxiliaryServicesConfig.metastoreConfig

`GcpDataprocClusterMetastoreConfig`

Persistent Hive metastore for the virtual cluster's jobs.

### spec.virtualClusterConfig.auxiliaryServicesConfig.metastoreConfig.dataprocMetastoreService

`string | valueFrom` · required

Resource name of an existing Dataproc Metastore service.
Format: projects/{project}/locations/{location}/services/{service}
Accepts a literal resource name today; references attach when a
metastore-service kind lands in the catalog.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig

`GcpDataprocClusterSparkHistoryServerConfig`

Persistent Spark History Server hosted on another Dataproc cluster.

### spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig.dataprocCluster

`string | valueFrom`

The Dataproc cluster hosting the Spark History Server. Resolves to
the cluster's fully qualified resource name
(projects/{project}/regions/{region}/clusters/{cluster}).

- references: GcpDataprocCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDataprocCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.gracefulDecommissionTimeout

`string`

Timeout for graceful YARN decommissioning when reducing the number
of workers. During this period, YARN waits for running tasks to
complete before shutting down nodes. Without this, scaling down
can terminate running jobs. Only meaningful for the GCE arm.
Format: duration in seconds with 's' suffix (e.g., "3600s").
Default: "0s" (immediate decommission).

- rule: graceful_decommission_timeout must be a duration in seconds (e.g., '3600s')

### spec.labels

`map<string, string>`

User-defined labels applied to the cluster (and propagated to its
VMs). Merged beneath Planton's platform attribution labels
(platform keys win on conflict). Not supported by the API for the
virtual (GKE-based) arm.

### spec.deletionPolicy

`string`

Engine-side teardown behavior. "DELETE" (default) destroys the
cluster; "PREVENT" fails any plan that would destroy it; "ABANDON"
removes it from IaC management while leaving it running in GCP.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `at_most_one_deployment_arm`: cluster_config and virtual_cluster_config are mutually exclusive
- `labels_unsupported_on_virtual_clusters`: user labels are not supported on virtual (GKE-based) clusters

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDataprocCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | Fully qualified cluster resource name. Format: projects/{project}/regions/{region}/clusters/{cluster} The composition handle downstream resources reference — including another cluster's spark_history_server_config, which consumes this exact format. |
| `status.outputs.cluster_name` | `string` | Short name of the cluster (same as the spec's cluster_name input). Useful for display, logging, and human-readable references. |
| `status.outputs.staging_bucket` | `string` | Cloud Storage bucket used for staging job dependencies. This is either the user-supplied staging_bucket or the auto-created bucket chosen by GCP. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.clusterConfig.stagingBucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.clusterConfig.tempBucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.clusterConfig.gceConfig.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.clusterConfig.gceConfig.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.clusterConfig.gceConfig.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.clusterConfig.autoscalingPolicyUri` | GcpDataprocAutoscalingPolicy | `status.outputs.name` |
| `spec.clusterConfig.encryptionKmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.clusterConfig.securityConfig.kerberosConfig.kmsKeyUri` | GcpKmsKey | `status.outputs.key_id` |
| `spec.virtualClusterConfig.stagingBucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesNamespace` | KubernetesNamespace | `spec.name` |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget` | GcpGkeCluster | `status.outputs.cluster_id` |
| `spec.virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.nodePoolTarget[].nodePool` | GcpGkeNodePool | `status.outputs.node_pool_id` |
| `spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig.dataprocCluster` | GcpDataprocCluster | `status.outputs.cluster_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDataprocCluster | `spec.virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig.dataprocCluster` | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
