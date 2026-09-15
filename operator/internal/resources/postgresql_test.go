package resources

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewPostgreSQLCluster_GVK(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "ns",
		Instances:   1,
		StorageSize: "5Gi",
	})

	gvk := obj.GroupVersionKind()
	if gvk.Group != "postgresql.cnpg.io" {
		t.Errorf("expected group postgresql.cnpg.io, got %s", gvk.Group)
	}
	if gvk.Version != "v1" {
		t.Errorf("expected version v1, got %s", gvk.Version)
	}
	if gvk.Kind != "Cluster" {
		t.Errorf("expected kind Cluster, got %s", gvk.Kind)
	}
}

func TestNewPostgreSQLCluster_NameAndNamespace(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "my-planton",
		Namespace:   "planton-system",
		Instances:   1,
		StorageSize: "10Gi",
	})

	if obj.GetName() != "my-planton-postgres" {
		t.Errorf("expected name my-planton-postgres, got %s", obj.GetName())
	}
	if obj.GetNamespace() != "planton-system" {
		t.Errorf("expected namespace planton-system, got %s", obj.GetNamespace())
	}
}

func TestNewPostgreSQLCluster_SpecFields(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   3,
		StorageSize: "50Gi",
	})

	instances, _, _ := unstructured.NestedInt64(obj.Object, "spec", "instances")
	if instances != 3 {
		t.Errorf("expected 3 instances, got %d", instances)
	}

	// The image must be pinned to an exact tag: CloudNativePG treats floating
	// tags as an operational hazard, and upgrades must be deliberate.
	imageName, _, _ := unstructured.NestedString(obj.Object, "spec", "imageName")
	if imageName == "" || strings.HasSuffix(imageName, ":latest") {
		t.Errorf("imageName must be pinned to an exact tag, got %q", imageName)
	}
	if !strings.Contains(imageName, "ghcr.io/cloudnative-pg/postgresql:") {
		t.Errorf("expected the official CloudNativePG PostgreSQL image, got %q", imageName)
	}

	size, _, _ := unstructured.NestedString(obj.Object, "spec", "storage", "size")
	if size != "50Gi" {
		t.Errorf("expected storage size 50Gi, got %s", size)
	}
}

// Superuser access over the network is OFF by default in CloudNativePG; the
// platform's single-user contract requires it on, and the generated
// {cluster}-superuser Secret is the credential every consumer references.
func TestNewPostgreSQLCluster_SuperuserAccessEnabled(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})

	enabled, found, _ := unstructured.NestedBool(obj.Object, "spec", "enableSuperuserAccess")
	if !found || !enabled {
		t.Error("expected spec.enableSuperuserAccess to be true")
	}
}

// The StorageClass key must be OMITTED when unpinned -- an explicit "" is
// rejected by CloudNativePG's admission webhook, and in the wider convention
// an empty class means "disable dynamic provisioning".
func TestNewPostgreSQLCluster_StorageClass(t *testing.T) {
	pinned := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",

		StorageClassName: "fast-ssd",
	})
	class, found, _ := unstructured.NestedString(pinned.Object, "spec", "storage", "storageClass")
	if !found || class != "fast-ssd" {
		t.Errorf("expected storageClass fast-ssd, got %q (found=%v)", class, found)
	}

	unpinned := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})
	if _, found, _ := unstructured.NestedString(unpinned.Object, "spec", "storage", "storageClass"); found {
		t.Error("storageClass must be omitted entirely when unpinned")
	}
}

// One cluster serves the whole platform, so PostgreSQL's default
// max_connections=100 saturates (proven live: "sorry, too many clients
// already" starved the identity server's pool and crash-looped the control
// plane's boot).
func TestNewPostgreSQLCluster_ConnectionHeadroom(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})

	maxConn, _, _ := unstructured.NestedString(obj.Object, "spec", "postgresql", "parameters", "max_connections")
	if maxConn != "300" {
		t.Errorf("expected max_connections 300, got %q", maxConn)
	}
}

// The resource floor exists because the control-plane's first boot
// self-provisions every database and runs Flyway migrations across all of
// them at once -- a tiny default gets OOM-killed and cascades the application
// boot into failure. Memory limit only: migrations must never be
// CPU-throttled.
func TestNewPostgreSQLCluster_ResourceFloor(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})

	memRequest, _, _ := unstructured.NestedString(obj.Object, "spec", "resources", "requests", "memory")
	if memRequest != "512Mi" {
		t.Errorf("expected memory request 512Mi, got %q", memRequest)
	}
	memLimit, _, _ := unstructured.NestedString(obj.Object, "spec", "resources", "limits", "memory")
	if memLimit != "2Gi" {
		t.Errorf("expected memory limit 2Gi, got %q", memLimit)
	}
	if _, found, _ := unstructured.NestedString(obj.Object, "spec", "resources", "limits", "cpu"); found {
		t.Error("no CPU limit: migrations must never be CPU-throttled")
	}
}

// The database the control plane owns is born with the cluster; every OTHER
// database it needs is self-provisioned at boot as the superuser -- no
// declared list to drift from the app's canonical, migration-derived set.
func TestNewPostgreSQLCluster_BootstrapInitDB(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})

	database, _, _ := unstructured.NestedString(obj.Object, "spec", "bootstrap", "initdb", "database")
	if database != DBBase {
		t.Errorf("expected initdb database %q, got %q", DBBase, database)
	}
	owner, _, _ := unstructured.NestedString(obj.Object, "spec", "bootstrap", "initdb", "owner")
	if owner != PostgreSQLOwnerRole {
		t.Errorf("expected initdb owner %q, got %q", PostgreSQLOwnerRole, owner)
	}
}

// The openfga database is born with the cluster: OpenFGA's migrate job cannot
// create it, and authorization can be enabled at ANY later point -- the
// database must already be waiting.
func TestNewPostgreSQLCluster_OpenFGADatabaseBornWithCluster(t *testing.T) {
	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName:      "test",
		Namespace:   "default",
		Instances:   1,
		StorageSize: "10Gi",
	})

	statements, found, _ := unstructured.NestedSlice(obj.Object, "spec", "bootstrap", "initdb", "postInitSQL")
	if !found {
		t.Fatal("expected postInitSQL in bootstrap initdb")
	}
	for _, stmt := range statements {
		if s, ok := stmt.(string); ok && strings.Contains(s, "CREATE DATABASE "+DBOpenFGA) {
			return
		}
	}
	t.Errorf("expected a CREATE DATABASE %s statement in postInitSQL, got %v", DBOpenFGA, statements)
}

func TestPostgreSQLClusterName(t *testing.T) {
	if got := PostgreSQLClusterName("my-planton"); got != "my-planton-postgres" {
		t.Errorf("expected my-planton-postgres, got %s", got)
	}
}

func TestPostgreSQLSuperuserSecretName(t *testing.T) {
	if got := PostgreSQLSuperuserSecretName("my-planton"); got != "my-planton-postgres-superuser" {
		t.Errorf("expected my-planton-postgres-superuser, got %s", got)
	}
}

// Applications connect through the -rw Service, never a pod: after a failover
// it re-points to the new primary automatically.
func TestPostgreSQLHost(t *testing.T) {
	got := PostgreSQLHost("my-planton", "planton-system")
	want := "my-planton-postgres-rw.planton-system.svc.cluster.local"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

// The vault connects as its own least-privilege role: a login role with no
// cluster-wide privilege, its password in a basic-auth Secret CloudNativePG
// reconciles onto the instance. No role, no managed block at all.
func TestNewPostgreSQLCluster_ManagedRoles(t *testing.T) {
	plain := NewPostgreSQLCluster(PostgreSQLClusterOptions{CRName: "test", Namespace: "default", Instances: 1, StorageSize: "10Gi"})
	if _, found, _ := unstructured.NestedMap(plain.Object, "spec", "managed"); found {
		t.Error("a cluster with no consumer roles must carry no managed block")
	}

	obj := NewPostgreSQLCluster(PostgreSQLClusterOptions{
		CRName: "test", Namespace: "default", Instances: 1, StorageSize: "10Gi",
		ManagedRoles: []PostgreSQLManagedRole{{Name: PostgreSQLVaultRole, PasswordSecretName: "test-postgres-openbao"}},
	})
	roles, found, _ := unstructured.NestedSlice(obj.Object, "spec", "managed", "roles")
	if !found || len(roles) != 1 {
		t.Fatalf("expected one managed role, got %v", roles)
	}
	role := roles[0].(map[string]any)
	if role["name"] != "openbao" || role["login"] != true || role["ensure"] != "present" {
		t.Errorf("role rendered %v; want a present login role named openbao", role)
	}
	if secret := role["passwordSecret"].(map[string]any); secret["name"] != "test-postgres-openbao" {
		t.Errorf("passwordSecret = %v; want the role's basic-auth Secret", secret)
	}
	for _, privilege := range []string{"superuser", "createdb", "createrole"} {
		if _, set := role[privilege]; set {
			t.Errorf("a consumer role must not carry %s", privilege)
		}
	}
}

// The vault's database is a declared object, not a CREATE statement: present
// is idempotent on a restored cluster that already holds it, and retain means
// the data lives and dies with the Cluster, never with the declaration.
func TestNewPostgreSQLDatabase(t *testing.T) {
	obj := NewPostgreSQLDatabase(PostgreSQLDatabaseOptions{
		CRName: "test", Namespace: "default",
		ObjectName: PostgreSQLVaultDatabaseObjectName("test"),
		Name:       DBOpenBAO, Owner: PostgreSQLVaultRole,
	})
	if obj.GetKind() != "Database" || obj.GetAPIVersion() != "postgresql.cnpg.io/v1" {
		t.Errorf("GVK = %s %s", obj.GetAPIVersion(), obj.GetKind())
	}
	if obj.GetName() != "test-postgres-openbao" || obj.GetNamespace() != "default" {
		t.Errorf("name/namespace = %s/%s", obj.GetNamespace(), obj.GetName())
	}
	spec := obj.Object["spec"].(map[string]any)
	if cluster := spec["cluster"].(map[string]any); cluster["name"] != "test-postgres" {
		t.Errorf("cluster = %v; want the platform's cluster", cluster)
	}
	for key, want := range map[string]any{"name": "openbao", "owner": "openbao", "ensure": "present", "databaseReclaimPolicy": "retain"} {
		if spec[key] != want {
			t.Errorf("spec.%s = %v, want %v", key, spec[key], want)
		}
	}
}

func TestPostgreSQLVaultNames(t *testing.T) {
	if got := PostgreSQLVaultRoleSecretName("my-planton"); got != "my-planton-postgres-openbao" {
		t.Errorf("role Secret = %s", got)
	}
	if got := PostgreSQLVaultDatabaseObjectName("my-planton"); got != "my-planton-postgres-openbao" {
		t.Errorf("Database object = %s", got)
	}
}
