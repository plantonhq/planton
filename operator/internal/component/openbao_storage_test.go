package component

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/plantonhq/planton/operator/internal/resources"
)

// The vault's wait for its database reads two objects CloudNativePG writes
// and relays their words: a role it cannot reconcile names its cause, a
// database not yet applied names its message, and only a reconciled role on
// an applied database lets the server render.

func vaultCluster(managedRolesStatus map[string]any) *unstructured.Unstructured {
	cluster := resources.NewPostgreSQLCluster(resources.PostgreSQLClusterOptions{CRName: "planton", Namespace: "planton", Instances: 1, StorageSize: "1Gi"})
	if managedRolesStatus != nil {
		_ = unstructured.SetNestedMap(cluster.Object, managedRolesStatus, "status", "managedRolesStatus")
	}
	return cluster
}

func vaultDatabase(applied bool, message string) *unstructured.Unstructured {
	db := resources.NewPostgreSQLDatabase(resources.PostgreSQLDatabaseOptions{
		CRName: "planton", Namespace: "planton",
		ObjectName: resources.PostgreSQLVaultDatabaseObjectName("planton"),
		Name:       resources.DBOpenBAO, Owner: resources.PostgreSQLVaultRole,
	})
	status := map[string]any{"applied": applied}
	if message != "" {
		status["message"] = message
	}
	_ = unstructured.SetNestedMap(db.Object, status, "status")
	return db
}

func TestClassifyVaultDatabase(t *testing.T) {
	reconciled := map[string]any{"byStatus": map[string]any{"reconciled": []any{"openbao"}}}

	t.Run("a reconciled role on an applied database is ready", func(t *testing.T) {
		ready, msg, err := classifyVaultDatabase(vaultCluster(reconciled), vaultDatabase(true, ""), "planton-postgres")
		if err != nil || !ready || msg != "" {
			t.Errorf("ready=%v msg=%q err=%v", ready, msg, err)
		}
	})
	t.Run("a role not yet reconciled waits and names the role", func(t *testing.T) {
		pending := map[string]any{"byStatus": map[string]any{"pending-reconciliation": []any{"openbao"}}}
		ready, msg, _ := classifyVaultDatabase(vaultCluster(pending), vaultDatabase(true, ""), "planton-postgres")
		if ready || !strings.Contains(msg, "create the vault's role openbao on planton-postgres") {
			t.Errorf("ready=%v msg=%q", ready, msg)
		}
	})
	t.Run("no managed-roles status at all is a wait, not a crash", func(t *testing.T) {
		ready, msg, _ := classifyVaultDatabase(vaultCluster(nil), vaultDatabase(false, ""), "planton-postgres")
		if ready || !strings.Contains(msg, "vault's role") {
			t.Errorf("ready=%v msg=%q", ready, msg)
		}
	})
	t.Run("a role CloudNativePG cannot reconcile relays its causes, sorted", func(t *testing.T) {
		cannot := map[string]any{
			"byStatus":        map[string]any{"pending-reconciliation": []any{"openbao"}},
			"cannotReconcile": map[string]any{"openbao": []any{"secret planton-postgres-openbao not found", "role owns objects"}},
		}
		ready, msg, _ := classifyVaultDatabase(vaultCluster(cannot), vaultDatabase(true, ""), "planton-postgres")
		if ready {
			t.Fatal("a role that cannot be reconciled must not be ready")
		}
		want := "The database cannot reconcile the vault's role openbao on planton-postgres: role owns objects; secret planton-postgres-openbao not found"
		if msg != want {
			t.Errorf("msg\n got: %s\nwant: %s", msg, want)
		}
	})
	t.Run("a database not yet applied waits and relays CloudNativePG's message", func(t *testing.T) {
		ready, msg, _ := classifyVaultDatabase(vaultCluster(reconciled), vaultDatabase(false, "cluster planton-postgres is not ready"), "planton-postgres")
		if ready {
			t.Fatal("an unapplied database must not be ready")
		}
		want := "Waiting for the database to create the vault's database openbao on planton-postgres: cluster planton-postgres is not ready"
		if msg != want {
			t.Errorf("msg\n got: %s\nwant: %s", msg, want)
		}
	})
	t.Run("a database not yet applied with nothing to say still waits in words", func(t *testing.T) {
		ready, msg, _ := classifyVaultDatabase(vaultCluster(reconciled), vaultDatabase(false, ""), "planton-postgres")
		if ready || msg != "Waiting for the database to create the vault's database openbao on planton-postgres" {
			t.Errorf("ready=%v msg=%q", ready, msg)
		}
	})
}
