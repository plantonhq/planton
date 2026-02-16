# GcpVertexAiNotebook Examples

## 1. Basic CPU Notebook

The simplest configuration -- a CPU-only notebook for data exploration and light ML work.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: data-exploration
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: e2-standard-4
```

## 2. GPU Notebook for ML Training

A notebook with an NVIDIA Tesla T4 GPU for training ML models.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: ml-training
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: n1-standard-8
  acceleratorConfig:
    type: NVIDIA_TESLA_T4
    coreCount: 1
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
  dataDisk:
    diskType: PD_SSD
    diskSizeGb: 500
  vmImage:
    project: deeplearning-platform-release
    family: tf-latest-gpu
```

## 3. Private Notebook with CMEK Encryption

A security-hardened notebook inside a VPC with no public IP and CMEK-encrypted disks.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: secure-notebook
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: e2-standard-4
  disablePublicIp: true
  networkInterface:
    network:
      value: projects/my-gcp-project/global/networks/ml-vpc
    subnet:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/ml-subnet
  serviceAccount:
    value: notebook-sa@my-gcp-project.iam.gserviceaccount.com
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
    kmsKey:
      value: projects/my-gcp-project/locations/us-central1/keyRings/ml-ring/cryptoKeys/disk-key
  dataDisk:
    diskType: PD_BALANCED
    diskSizeGb: 500
    kmsKey:
      value: projects/my-gcp-project/locations/us-central1/keyRings/ml-ring/cryptoKeys/disk-key
  shieldedInstanceConfig:
    enableSecureBoot: true
    enableVtpm: true
    enableIntegrityMonitoring: true
  tags:
    - notebook
    - no-public-ip
```

## 4. Custom Container Notebook

A notebook using a custom Docker image for specialized environments.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: custom-env
spec:
  projectId:
    value: my-gcp-project
  location: us-west1-b
  machineType: e2-standard-8
  containerImage:
    repository: gcr.io/my-gcp-project/custom-notebook
    tag: v2.1.0
  dataDisk:
    diskType: PD_SSD
    diskSizeGb: 1000
```

## 5. Stopped Notebook (Cost Optimization)

Create a notebook in stopped state -- useful for pre-provisioning infrastructure without incurring compute costs.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: paused-notebook
spec:
  projectId:
    value: my-gcp-project
  location: europe-west1-b
  machineType: n1-standard-4
  desiredState: STOPPED
  instanceOwners:
    - researcher@my-gcp-project.iam.gserviceaccount.com
```

## 6. Notebook with Foreign Key References (Infra Chart Composition)

Using `valueFrom` references for integration with other OpenMCF resources.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: composed-notebook
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: ml-project
  location: us-central1-a
  machineType: n1-standard-8
  networkInterface:
    network:
      valueFrom:
        kind: GcpVpc
        name: ml-vpc
    subnet:
      valueFrom:
        kind: GcpSubnetwork
        name: ml-subnet
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: notebook-sa
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
    kmsKey:
      valueFrom:
        kind: GcpKmsKey
        name: disk-encryption-key
  acceleratorConfig:
    type: NVIDIA_TESLA_T4
    coreCount: 1
```

## 7. High-Memory A100 GPU Notebook

A high-performance notebook with an A100 GPU for large model training.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiNotebook
metadata:
  name: a100-training
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: a2-highgpu-1g
  acceleratorConfig:
    type: NVIDIA_TESLA_A100
    coreCount: 1
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
  dataDisk:
    diskType: PD_SSD
    diskSizeGb: 2000
  vmImage:
    project: deeplearning-platform-release
    family: pytorch-latest-gpu
```
