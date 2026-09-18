package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	PostgreSQLPort = 5432

	// PostgreSQLSuperuser is the user most platform services connect as.
	// One cluster, one superuser, self-provisioned databases is the
	// platform-wide contract for the control plane, Temporal, and the
	// identity server: the control-plane fat-jar creates its own databases
	// at boot, Temporal's schema job creates its two, and the identity
	// server's init container ensures its one. The bundled vault is the
	// exception and the shape the rest will converge on: it holds the most
	// sensitive data on the platform, so it connects as its own
	// least-privilege role (PostgreSQLVaultRole) to its own database
	// (DBOpenBAO), both declared to CloudNativePG and reconciled by it.
	PostgreSQLSuperuser = "postgres"

	// postgresqlImage pins the exact PostgreSQL build CloudNativePG runs.
	// CloudNativePG treats a floating tag as an operational hazard (upgrades
	// must be deliberate), so the pin is explicit and bumped on purpose. The
	// major tracks the PostgreSQL line the platform is proven on.
	postgresqlImage = "ghcr.io/cloudnative-pg/postgresql:18.4"

	// postgresqlMaxConnections: one cluster serves the whole platform -- the
	// control-plane monolith (a connection pool per logical database),
	// Temporal's four services, and the identity server. PostgreSQL's
	// default max_connections=100 saturates under that (proven live: "sorry,
	// too many clients already" starved the identity server's pool and
	// crash-looped the control plane's boot).
	postgresqlMaxConnections = "300"

	// The database's sizing, chosen here rather than left to CloudNativePG's
	// default (none): one instance serving every platform database, ~520Mi
	// resident live with the control plane's pools open. A request so it
	// schedules honestly, a memory limit so a runaway query cannot take the
	// node, no CPU limit so a checkpoint or a vacuum is never throttled
	// (requests-only, the house pattern).
	postgresqlCPURequest    = "250m"
	postgresqlMemoryRequest = "512Mi"
	postgresqlMemoryLimit   = "2Gi"

	// DBBase is the ONE database the control plane owns (DB_NAME=planton in
	// every deployment shape); every domain is separated at the schema level
	// inside it. This is the exact contract the desktop daemon boots against
	// (its local boot environment file); the operator
	// and the daemon are the two providers of one control-plane boot
	// contract and must stay in sync.
	//
	// The database is created at cluster birth (bootstrap initdb below) and
	// the control plane self-provisions every OTHER database it needs at
	// boot, connecting as the superuser.
	DBBase = "planton"

	// PostgreSQLOwnerRole owns DBBase. Unused by consumers today (everything
	// connects as the superuser), it exists so the database is born with a
	// non-superuser owner -- the seam a least-privilege split will widen.
	PostgreSQLOwnerRole = "planton"

	// DBOpenFGA backs the OpenFGA authorization server. Created at cluster
	// birth (postInitSQL) so enabling authorization at ANY later point finds
	// its database waiting -- OpenFGA's own migrate job cannot create it.
	DBOpenFGA = "openfga"

	// DBOpenBAO is the bundled secrets manager's database: the vault's
	// storage backend keeps every secret, key, and policy in it as
	// barrier-encrypted rows, so the platform's one archive carries the
	// vault with the records. Declared as a CloudNativePG Database object
	// (not postInitSQL) so a restored cluster that already holds it
	// converges instead of failing a CREATE, and owned by the vault's own
	// role.
	DBOpenBAO = "openbao"

	// PostgreSQLVaultRole is the login role the vault connects as -- owner
	// of DBOpenBAO and nothing else. Declared through the Cluster's
	// managed roles with its password in a Secret CloudNativePG reconciles
	// onto the instance, on a fresh cluster and on a restored one alike, so
	// the vault's credential follows the operator-owned Secret through a
	// restore with nothing carried by hand.
	PostgreSQLVaultRole = "openbao"

	// PostgreSQLClusterKind mirrors CloudNativePG's Cluster naming contract:
	// every object it creates derives from the Cluster name -- instance pods
	// ("{name}-1", ...), the traffic Services ("{name}-rw" primary
	// read-write, "{name}-ro", "{name}-r"), and the credential Secrets
	// ("{name}-app", "{name}-superuser").
	postgresqlAPIGroup     = "postgresql.cnpg.io"
	postgresqlAPIVersion   = "v1"
	postgresqlKind         = "Cluster"
	postgresqlDatabaseKind = "Database"
)

// PostgreSQLClusterGVK is the GroupVersionKind for CloudNativePG Cluster CRs.
var PostgreSQLClusterGVK = schema.GroupVersionKind{
	Group:   postgresqlAPIGroup,
	Version: postgresqlAPIVersion,
	Kind:    postgresqlKind,
}

// PostgreSQLDatabaseGVK is the GroupVersionKind for CloudNativePG Database
// CRs -- a declarative database inside a Cluster, reconciled by the
// database operator against the live instance.
var PostgreSQLDatabaseGVK = schema.GroupVersionKind{
	Group:   postgresqlAPIGroup,
	Version: postgresqlAPIVersion,
	Kind:    postgresqlDatabaseKind,
}

// PostgreSQLVaultRoleSecretName returns the vault role's credential Secret,
// named the way CloudNativePG names the credentials it generates itself
// ("{cluster}-superuser", "{cluster}-app"): "{cluster}-openbao". Basic-auth
// shape (username, password); operator-owned; create-once, so the password
// never rotates under a running vault.
func PostgreSQLVaultRoleSecretName(crName string) string {
	return fmt.Sprintf("%s-%s", PostgreSQLClusterName(crName), PostgreSQLVaultRole)
}

// PostgreSQLVaultDatabaseObjectName returns the name of the Database CR that
// declares the vault's database: "{cluster}-openbao", the same name as the
// role's Secret -- one name for the vault's presence on the cluster, in the
// two kinds that carry it.
func PostgreSQLVaultDatabaseObjectName(crName string) string {
	return fmt.Sprintf("%s-%s", PostgreSQLClusterName(crName), DBOpenBAO)
}

// PostgreSQLClusterName returns the CloudNativePG Cluster name for a
// PlantonPlatform install: "{crName}-postgres".
func PostgreSQLClusterName(crName string) string {
	return fmt.Sprintf("%s-postgres", crName)
}

// PostgreSQLSuperuserSecretName returns the superuser credential Secret
// CloudNativePG generates for the platform cluster (basic-auth shape with
// "username" and "password" keys): "{cluster}-superuser". CloudNativePG owns
// this Secret's lifecycle -- it lives and dies with the Cluster, so
// credentials can never outlive (or predate) the volumes they unlock.
func PostgreSQLSuperuserSecretName(crName string) string {
	return fmt.Sprintf("%s-superuser", PostgreSQLClusterName(crName))
}

// PostgreSQLHost returns the in-cluster DNS hostname of the cluster's
// primary read-write Service: "{cluster}-rw.{namespace}.svc.cluster.local".
// Applications always connect through the -rw Service, never a pod: after a
// failover it re-points to the new primary automatically.
func PostgreSQLHost(crName, namespace string) string {
	return fmt.Sprintf("%s-rw.%s.svc.cluster.local", PostgreSQLClusterName(crName), namespace)
}

// PostgreSQLClusterOptions configures the platform's CloudNativePG Cluster.
type PostgreSQLClusterOptions struct {
	// CRName is the PlantonPlatform CR name the cluster derives its name from.
	CRName string

	// Namespace is the Kubernetes namespace for the cluster.
	Namespace string

	// Instances is the number of PostgreSQL instances (a primary plus hot
	// standbys with automated failover -- HA is one knob on the one cluster).
	Instances int32

	// StorageSize is the per-instance PVC size (e.g., "10Gi"). Sizes can only
	// grow -- CloudNativePG rejects shrinks.
	StorageSize string

	// StorageClassName pins the cluster's volumes to a StorageClass.
	// Empty means the key is omitted and the cluster default provisions.
	StorageClassName string

	// OwnerRef ties the Cluster to the PlantonPlatform CR for garbage
	// collection. CloudNativePG in turn owns the Cluster's Secrets and PVCs,
	// so deleting the platform removes the database WITH its credentials --
	// one coherent lifecycle.
	OwnerRef *metav1.OwnerReference

	// Backup wires the cluster to its object store through the Barman Cloud
	// plugin: WAL is archived continuously under ServerName and base backups
	// land in the same store. Nil renders no plugins entry at all -- the
	// caller sets it only once the plugin serves, because a Cluster naming a
	// plugin the operator cannot reach is parked with no instances.
	Backup *PostgreSQLClusterBackup

	// ServiceAccountAnnotations go on the ServiceAccount CloudNativePG creates
	// for the instance pods (named after the Cluster). This is the keyless
	// backup identity seam -- EKS "eks.amazonaws.com/role-arn", GKE
	// "iam.gke.io/gcp-service-account", AKS "azure.workload.identity/client-id"
	// -- the same shape the runner and the control plane use for theirs.
	ServiceAccountAnnotations map[string]string

	// Recovery bootstraps the cluster from a source's archive instead of
	// initdb. It is honoured by CloudNativePG only at Cluster creation; the
	// component keeps a live Cluster's bootstrap as it is and explains that
	// in status rather than re-rendering a decision that cannot change.
	Recovery *PostgreSQLClusterRecovery

	// ManagedRoles are the login roles CloudNativePG declares and reconciles
	// on the instance -- the least-privilege seam for a consumer that must
	// not connect as the superuser. Each role's password lives in a
	// basic-auth Secret CloudNativePG reads; on a restored cluster the role
	// already exists with the SOURCE's password and CloudNativePG resets it
	// to this Secret's, so consumers keep reading the Secret they always
	// read (the superuser Secret's own behaviour, applied to every role).
	ManagedRoles []PostgreSQLManagedRole
}

// PostgreSQLManagedRole is one login role the Cluster carries under
// spec.managed.roles.
type PostgreSQLManagedRole struct {
	Name               string
	PasswordSecretName string
}

// PostgreSQLClusterBackup names the store and server a cluster archives to.
type PostgreSQLClusterBackup struct {
	ObjectStoreName string
	ServerName      string
}

// PostgreSQLClusterRecovery names the store and server a cluster restores
// from, and optionally the point in time to stop at.
type PostgreSQLClusterRecovery struct {
	ObjectStoreName string
	ServerName      string
	// TargetTime is an RFC 3339 timestamp; empty recovers to the end of the
	// archive (the latest consistent point).
	TargetTime string
}

// NewPostgreSQLCluster builds the platform's postgresql.cnpg.io/v1 Cluster as
// an unstructured object.
//
// The resource floor exists because the control-plane's first boot
// self-provisions every database and runs Flyway migrations across all of
// them at once -- enough concurrent connections and shared buffers to
// OOM-kill a tiny default and cascade the application boot into failure.
// Memory limit only (no CPU limit) so migrations are never CPU-throttled.
//
// Absent quantities are OMITTED, never rendered empty: CloudNativePG's
// mutating webhook rejects "" with a quantity-format error.
func NewPostgreSQLCluster(opts PostgreSQLClusterOptions) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(PostgreSQLClusterGVK)
	obj.SetName(PostgreSQLClusterName(opts.CRName))
	obj.SetNamespace(opts.Namespace)

	if opts.OwnerRef != nil {
		obj.SetOwnerReferences([]metav1.OwnerReference{*opts.OwnerRef})
	}

	storage := map[string]any{
		"size": opts.StorageSize,
	}
	if opts.StorageClassName != "" {
		storage["storageClass"] = opts.StorageClassName
	}

	spec := map[string]any{
		"instances": int64(opts.Instances),
		"imageName": postgresqlImage,
		// Superuser access over the network is off by default in
		// CloudNativePG; the platform contract runs on it (see
		// PostgreSQLSuperuser), so it is enabled and the generated
		// {cluster}-superuser Secret is the one credential every consumer
		// references.
		"enableSuperuserAccess": true,
		"storage":               storage,
		"postgresql": map[string]any{
			"parameters": map[string]any{
				"max_connections": postgresqlMaxConnections,
			},
		},
		"resources": helmResourceValues(postgresqlResources()),
		"bootstrap": postgresqlBootstrap(opts.Recovery),
	}

	if opts.Recovery != nil {
		spec["externalClusters"] = []any{
			map[string]any{
				"name":   recoverySourceName,
				"plugin": pluginConfiguration(opts.Recovery.ObjectStoreName, opts.Recovery.ServerName),
			},
		}
	}

	if opts.Backup != nil {
		plugin := pluginConfiguration(opts.Backup.ObjectStoreName, opts.Backup.ServerName)
		plugin["isWALArchiver"] = true
		spec["plugins"] = []any{plugin}
	}

	if len(opts.ServiceAccountAnnotations) > 0 {
		annotations := make(map[string]any, len(opts.ServiceAccountAnnotations))
		for k, v := range opts.ServiceAccountAnnotations {
			annotations[k] = v
		}
		spec["serviceAccountTemplate"] = map[string]any{
			"metadata": map[string]any{"annotations": annotations},
		}
	}

	if len(opts.ManagedRoles) > 0 {
		roles := make([]any, 0, len(opts.ManagedRoles))
		for _, role := range opts.ManagedRoles {
			roles = append(roles, map[string]any{
				"name":   role.Name,
				"ensure": "present",
				// A login role with no cluster-wide privilege: it owns its
				// database (the Database object names it owner) and
				// nothing else. Superuser, createdb, and createrole stay
				// at CloudNativePG's false.
				"login":          true,
				"passwordSecret": map[string]any{"name": role.PasswordSecretName},
			})
		}
		spec["managed"] = map[string]any{"roles": roles}
	}

	obj.Object["spec"] = spec

	return obj
}

// PostgreSQLDatabaseOptions declares one database inside the platform's
// Cluster through CloudNativePG's Database object.
type PostgreSQLDatabaseOptions struct {
	// CRName is the PlantonPlatform CR name the Cluster derives its name from.
	CRName string

	// Namespace is the Kubernetes namespace of the Cluster.
	Namespace string

	// ObjectName is the Database CR's own name; Name is the database's name
	// inside PostgreSQL; Owner is the role that owns it.
	ObjectName string
	Name       string
	Owner      string

	// OwnerRef ties the object to the PlantonPlatform CR for garbage
	// collection.
	OwnerRef *metav1.OwnerReference
}

// NewPostgreSQLDatabase builds a postgresql.cnpg.io/v1 Database as an
// unstructured object: `ensure: present` is idempotent, so a restored
// cluster that already holds the database converges instead of failing a
// CREATE; `databaseReclaimPolicy: retain` means deleting the object never drops the
// data -- the database lives and dies with the Cluster, not with this
// declaration (a platform that later opts the consumer out leaves the
// database standing, deliberately). CloudNativePG waits for the Cluster and
// the owner role to exist and reports its progress in status.applied and
// status.message.
func NewPostgreSQLDatabase(opts PostgreSQLDatabaseOptions) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(PostgreSQLDatabaseGVK)
	obj.SetName(opts.ObjectName)
	obj.SetNamespace(opts.Namespace)
	if opts.OwnerRef != nil {
		obj.SetOwnerReferences([]metav1.OwnerReference{*opts.OwnerRef})
	}
	obj.Object["spec"] = map[string]any{
		"cluster":               map[string]any{"name": PostgreSQLClusterName(opts.CRName)},
		"name":                  opts.Name,
		"owner":                 opts.Owner,
		"ensure":                "present",
		"databaseReclaimPolicy": "retain",
	}
	return obj
}

// postgresqlBootstrap is how the cluster's data comes to exist: initdb for a
// fresh platform, recovery from a source's archive for a restored one. The
// two are exclusive by CloudNativePG's schema, and neither changes after the
// Cluster exists.
func postgresqlBootstrap(recovery *PostgreSQLClusterRecovery) map[string]any {
	if recovery != nil {
		// A restored platform's databases -- the base, OpenFGA's, the
		// identity server's, Temporal's -- all come back from the archive;
		// nothing here creates them. The superuser credential is the one
		// CloudNativePG generates for THIS cluster (the {cluster}-superuser
		// Secret) and reconciles onto the restored instance, so every
		// consumer keeps reading the Secret it always read.
		rec := map[string]any{"source": recoverySourceName}
		if recovery.TargetTime != "" {
			rec["recoveryTarget"] = map[string]any{"targetTime": recovery.TargetTime}
		}
		return map[string]any{"recovery": rec}
	}
	return map[string]any{
		"initdb": map[string]any{
			"database": DBBase,
			"owner":    PostgreSQLOwnerRole,
			// Databases whose consumers cannot create them are born with
			// the cluster: OpenFGA's migrate job expects its database to
			// exist. Runs once, as the superuser, against the postgres
			// maintenance database.
			"postInitSQL": []any{
				fmt.Sprintf("CREATE DATABASE %s", DBOpenFGA),
			},
		},
	}
}

// postgresqlResources is the container sizing every install gets (the constants
// above carry the reasoning).
func postgresqlResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(postgresqlCPURequest),
			corev1.ResourceMemory: resource.MustParse(postgresqlMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(postgresqlMemoryLimit),
		},
	}
}
