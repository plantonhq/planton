# GCP Firestore Database Examples

This document provides YAML examples for deploying Cloud Firestore databases via OpenMCF. Each example includes a use-case description and the manifest.

---

## Example 1: Default Database (Minimal)

**When to use:** Simplest starting point. Creates the project's default Firestore Native database. Client libraries connect to this database by default.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: default-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpFirestoreDatabase.default-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: nam5
  databaseName: "(default)"
  type: FIRESTORE_NATIVE
```

---

## Example 2: Named Database with PITR

**When to use:** Create a named database separate from the default, with point-in-time recovery enabled for disaster recovery (7-day retention).

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: orders-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirestoreDatabase.orders-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: us-east1
  databaseName: orders-db
  type: FIRESTORE_NATIVE
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
```

---

## Example 3: Datastore Mode Database

**When to use:** Legacy Datastore applications that need the Datastore client library API with server-side entity-group transactions.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: legacy-ds
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirestoreDatabase.legacy-ds
spec:
  projectId:
    value: my-gcp-project-123
  locationId: us-central1
  databaseName: legacy-datastore
  type: DATASTORE_MODE
  concurrencyMode: PESSIMISTIC
```

---

## Example 4: Enterprise Edition with CMEK

**When to use:** Compliance requirements demand customer-managed encryption keys and the enhanced Enterprise SLA. The KMS key must be in the same location as the database.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: secure-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirestoreDatabase.secure-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: nam5
  databaseName: secure-db
  type: FIRESTORE_NATIVE
  databaseEdition: ENTERPRISE
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
  kmsKeyName:
    value: projects/my-gcp-project-123/locations/us/keyRings/firestore-ring/cryptoKeys/firestore-key
```

---

## Example 5: Infra Chart Composition (valueFrom References)

**When to use:** When deploying as part of an infra chart where the project and KMS key are created by other resources in the same chart.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: composed-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirestoreDatabase.composed-db
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  locationId: nam5
  databaseName: composed-db
  type: FIRESTORE_NATIVE
  databaseEdition: ENTERPRISE
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: firestore-key
      fieldPath: status.outputs.key_id
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
```

---

## Example 6: European Multi-Region Database

**When to use:** Data residency requirements demand European data storage. Uses the eur3 multi-region location for high availability across Europe.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFirestoreDatabase
metadata:
  name: eu-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpFirestoreDatabase.eu-db
spec:
  projectId:
    value: my-gcp-project-123
  locationId: eur3
  databaseName: eu-customers
  type: FIRESTORE_NATIVE
  concurrencyMode: OPTIMISTIC
  pointInTimeRecoveryEnablement: POINT_IN_TIME_RECOVERY_ENABLED
  deleteProtectionState: DELETE_PROTECTION_ENABLED
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
