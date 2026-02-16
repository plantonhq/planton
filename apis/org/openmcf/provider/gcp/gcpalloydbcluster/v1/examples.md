# GcpAlloydbCluster Examples

Copy-paste ready YAML manifests for deploying AlloyDB clusters via OpenMCF.

---

## Example 1: Minimal (Dev)

**When to use:** Development or testing. Just required fields, 2 CPU, ZONAL availability.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: dev-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: dev-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/default
  primaryInstance:
    instanceId: dev-primary
    cpuCount: 2
    availabilityType: ZONAL
```

---

## Example 2: Production HA

**When to use:** Production with high availability, initial user, and deletion protection.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: prod-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: prod-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/prod-vpc
  databaseVersion: POSTGRES_15
  displayName: Production AlloyDB
  deletionProtection: true
  initialUser:
    user: dbadmin
    password: SecureP@ssw0rd123!
  primaryInstance:
    instanceId: prod-primary
    cpuCount: 4
    availabilityType: REGIONAL
    displayName: Production Primary
```

---

## Example 3: Enterprise CMEK

**When to use:** Compliance requiring customer-managed encryption for cluster data, automated backups, and continuous backups.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: enterprise-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: enterprise-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/prod-vpc
  databaseVersion: POSTGRES_16
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/cluster-key
  automatedBackupPolicy:
    enabled: true
    quantityBasedRetentionCount: 14
    encryptionKmsKeyName:
      value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/backup-key
  continuousBackupConfig:
    enabled: true
    recoveryWindowDays: 14
    encryptionKmsKeyName:
      value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/pitr-key
  primaryInstance:
    instanceId: enterprise-primary
    cpuCount: 8
    availabilityType: REGIONAL
```

---

## Example 4: Custom Backup Policy

**When to use:** Quantity-based retention with a weekly backup schedule (e.g., Mon/Wed/Fri at 2 AM UTC).

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: backup-scheduled-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: backup-scheduled-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/default
  automatedBackupPolicy:
    enabled: true
    backupWindow: "3600s"
    quantityBasedRetentionCount: 7
    weeklySchedule:
      daysOfWeek:
        - MONDAY
        - WEDNESDAY
        - FRIDAY
      startHour: 2
  primaryInstance:
    instanceId: backup-primary
    cpuCount: 4
    availabilityType: ZONAL
```

---

## Example 5: Time-Based Retention with PITR

**When to use:** Continuous backup with a 21-day point-in-time recovery window.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: pitr-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: pitr-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/default
  automatedBackupPolicy:
    enabled: true
    timeBasedRetentionPeriod: "1814400s"
  continuousBackupConfig:
    enabled: true
    recoveryWindowDays: 21
  primaryInstance:
    instanceId: pitr-primary
    cpuCount: 4
    availabilityType: REGIONAL
```

---

## Example 6: Full-Featured

**When to use:** Maximum configuration with all features: CMEK, custom backups, PITR, maintenance window, query insights, SSL, Auth Proxy enforcement.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpAlloydbCluster
metadata:
  name: full-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: full-alloydb
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/prod-vpc
  allocatedIpRange: my-ip-range
  databaseVersion: POSTGRES_16
  displayName: Full Production AlloyDB
  deletionProtection: true
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/cluster-key
  initialUser:
    user: dbadmin
    password: SecureP@ssw0rd123!
  automatedBackupPolicy:
    enabled: true
    backupWindow: "7200s"
    quantityBasedRetentionCount: 14
    weeklySchedule:
      daysOfWeek:
        - MONDAY
        - THURSDAY
      startHour: 3
    encryptionKmsKeyName:
      value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/backup-key
  continuousBackupConfig:
    enabled: true
    recoveryWindowDays: 21
    encryptionKmsKeyName:
      value: projects/my-gcp-project/locations/us-central1/keyRings/alloydb-kr/cryptoKeys/pitr-key
  maintenanceWindow:
    day: SUNDAY
    startHour: 4
  primaryInstance:
    instanceId: full-primary
    cpuCount: 8
    availabilityType: REGIONAL
    displayName: Full Production Primary
    databaseFlags:
      max_connections: "500"
    queryInsightsConfig:
      queryPlansPerMinute: 10
      queryStringLength: 4096
      recordApplicationTags: true
      recordClientAddress: true
    requireConnectors: true
    sslMode: ENCRYPTED_ONLY
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
