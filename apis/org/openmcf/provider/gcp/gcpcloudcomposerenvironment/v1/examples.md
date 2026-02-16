# GcpCloudComposerEnvironment Examples

Copy-paste ready YAML manifests for deploying Cloud Composer environments via OpenMCF.

---

## Example 1: Minimal Environment

**When to use:** Development or testing. Minimal configuration with default settings.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: dev-composer
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_SMALL
```

---

## Example 2: With Software Configuration

**When to use:** Specify Composer/Airflow version and install custom PyPI packages.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: composer-with-packages
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_SMALL
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
    pypiPackages:
      numpy: ">=1.21.0"
      pandas: ">=1.3.0"
      requests: ""
    airflowConfigOverrides:
      core-dags_are_paused_at_creation: "False"
      webserver-expose_config: "True"
```

---

## Example 3: Production with VPC Networking and Private Endpoint

**When to use:** Production environment with private networking using VPC peering (Composer 2.x).

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: prod-composer-vpc
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_MEDIUM
  nodeConfig:
    network:
      value: projects/my-gcp-project/global/networks/prod-vpc
    subnetwork:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/prod-subnet
  privateEnvironmentConfig:
    enablePrivateEndpoint: true
    connectionType: VPC_PEERING
    masterIpv4CidrBlock: 172.16.0.0/28
    cloudSqlIpv4CidrBlock: 10.0.0.0/24
    cloudComposerNetworkIpv4CidrBlock: 10.1.0.0/24
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
  webServerNetworkAccessControl:
    allowedIpRanges:
      - value: 10.0.0.0/8
        description: "Corporate network"
      - value: 203.0.113.0/24
        description: "VPN range"
```

---

## Example 4: Full-Featured Production Environment

**When to use:** Maximum configuration with workloads config, maintenance window, CMEK, recovery, and access control.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: prod-composer-full
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_LARGE
  resilienceMode: HIGH_RESILIENCE
  nodeConfig:
    network:
      value: projects/my-gcp-project/global/networks/prod-vpc
    subnetwork:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/prod-subnet
    serviceAccount:
      value: composer-sa@my-gcp-project.iam.gserviceaccount.com
    tags:
      - composer-worker
      - airflow
  privateEnvironmentConfig:
    enablePrivateEndpoint: true
    connectionType: PRIVATE_SERVICE_CONNECT
    masterIpv4CidrBlock: 172.16.0.0/28
    cloudSqlIpv4CidrBlock: 10.0.0.0/24
    cloudComposerNetworkIpv4CidrBlock: 10.1.0.0/24
    cloudComposerConnectionSubnetwork: projects/my-gcp-project/regions/us-central1/subnetworks/composer-psc-subnet
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
    pypiPackages:
      apache-airflow-providers-google: ">=8.0.0"
      apache-airflow-providers-postgres: ">=5.0.0"
    airflowConfigOverrides:
      core-parallelism: "32"
      core-dag_concurrency: "16"
      webserver-expose_config: "True"
    envVariables:
      ENVIRONMENT: production
      LOG_LEVEL: INFO
  workloadsConfig:
    scheduler:
      cpu: 2.0
      memoryGb: 7.5
      storageGb: 5.0
      count: 1
    webServer:
      cpu: 2.0
      memoryGb: 4.0
      storageGb: 2.0
    worker:
      cpu: 4.0
      memoryGb: 15.0
      storageGb: 10.0
      minCount: 2
      maxCount: 10
    triggerer:
      cpu: 1.0
      memoryGb: 2.0
      count: 2
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/composer-kr/cryptoKeys/composer-key
  maintenanceWindow:
    startTime: "2026-01-01T02:00:00Z"
    endTime: "2026-01-01T10:00:00Z"
    recurrence: "FREQ=WEEKLY;BYDAY=TU,WE,TH"
  recoveryConfig:
    enabled: true
    snapshotLocation: gs://my-composer-backups/snapshots
    snapshotCreationSchedule: "0 4 * * *"
    timeZone: America/Los_Angeles
  webServerNetworkAccessControl:
    allowedIpRanges:
      - value: 10.0.0.0/8
        description: "Corporate network"
      - value: 172.16.0.0/12
        description: "VPN range"
```

---

## Example 5: Composer 3 with PSC Networking

**When to use:** Composer 3 environment with Private Service Connect networking and private environment.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: composer3-psc
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_MEDIUM
  nodeConfig:
    composerNetworkAttachment: projects/my-gcp-project/regions/us-central1/networkAttachments/composer-psc-attachment
    composerInternalIpv4CidrBlock: 10.2.0.0/20
    serviceAccount:
      value: composer-sa@my-gcp-project.iam.gserviceaccount.com
  enablePrivateEnvironment: true
  enablePrivateBuildsOnly: true
  softwareConfig:
    imageVersion: composer-3.0.0-airflow-2.9.3
    pypiPackages:
      apache-airflow-providers-google: ">=8.0.0"
    webServerPluginsMode: ENABLED
  workloadsConfig:
    scheduler:
      cpu: 2.0
      memoryGb: 7.5
      storageGb: 5.0
      count: 1
    webServer:
      cpu: 2.0
      memoryGb: 4.0
      storageGb: 2.0
    worker:
      cpu: 4.0
      memoryGb: 15.0
      storageGb: 10.0
      minCount: 2
      maxCount: 10
    triggerer:
      cpu: 1.0
      memoryGb: 2.0
      count: 2
    dagProcessor:
      cpu: 1.0
      memoryGb: 2.0
      storageGb: 1.0
      count: 1
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/composer-kr/cryptoKeys/composer-key
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
