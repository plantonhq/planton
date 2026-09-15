package component

import (
	"context"
	"fmt"
	"sort"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The vault's storage is the platform's PostgreSQL. This file is the vault's
// half of that seam: reading the state the database component and
// CloudNativePG produce (the role reconciled onto the instance, the database
// applied) before the server is rendered. The database component's half --
// declaring the role, its credential Secret, and the database -- lives with
// the Cluster it renders (postgresql.go); the vault consumes that Secret the
// way every consumer of the platform's database does, as a variable
// projected straight from it (openbao_helm.go), so nothing here composes
// or copies a credential.

// vaultDatabaseReady reports whether the vault's role and database exist on
// the live cluster, and when they do not, a sentence that says what is still
// happening in CloudNativePG's own words. Read every pass: cheap, and the one
// place a database-side failure (a role CloudNativePG cannot reconcile, a
// database it cannot create) surfaces as the vault's status instead of as a
// crash-looping pod.
func (o *OpenBAO) vaultDatabaseReady(ctx context.Context, c client.Client, planton *v1.PlantonPlatform) (bool, string, error) {
	clusterName := resources.PostgreSQLClusterName(planton.Name)

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(resources.PostgreSQLClusterGVK)
	if err := c.Get(ctx, types.NamespacedName{Name: clusterName, Namespace: planton.Namespace}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return false, fmt.Sprintf("Waiting for the platform's database %s to exist before the vault's role can be created on it", clusterName), nil
		}
		return false, "", fmt.Errorf("reading PostgreSQL cluster %s for the vault's role: %w", clusterName, err)
	}

	database := &unstructured.Unstructured{}
	database.SetGroupVersionKind(resources.PostgreSQLDatabaseGVK)
	databaseName := resources.PostgreSQLVaultDatabaseObjectName(planton.Name)
	if err := c.Get(ctx, types.NamespacedName{Name: databaseName, Namespace: planton.Namespace}, database); err != nil {
		if apierrors.IsNotFound(err) {
			return false, fmt.Sprintf("Waiting for the vault's database %s to be declared on %s", resources.DBOpenBAO, clusterName), nil
		}
		return false, "", fmt.Errorf("reading the vault's Database %s: %w", databaseName, err)
	}

	return classifyVaultDatabase(cluster, database, clusterName)
}

// classifyVaultDatabase is the pure verdict over the two objects' status:
// the role must be in CloudNativePG's reconciled set and the database must
// report applied. Anything else is a wait, with the reason relayed from the
// object that knows it -- a role CloudNativePG cannot reconcile names its
// cause; a database not yet applied names its message.
func classifyVaultDatabase(cluster, database *unstructured.Unstructured, clusterName string) (bool, string, error) {
	role := resources.PostgreSQLVaultRole

	if causes := managedRoleCauses(cluster, role); len(causes) > 0 {
		return false, fmt.Sprintf("The database cannot reconcile the vault's role %s on %s: %s", role, clusterName, strings.Join(causes, "; ")), nil
	}
	if !managedRoleReconciled(cluster, role) {
		return false, fmt.Sprintf("Waiting for the database to create the vault's role %s on %s", role, clusterName), nil
	}

	applied, _, _ := unstructured.NestedBool(database.Object, "status", "applied")
	if !applied {
		msg := fmt.Sprintf("Waiting for the database to create the vault's database %s on %s", resources.DBOpenBAO, clusterName)
		if reason, _, _ := unstructured.NestedString(database.Object, "status", "message"); reason != "" {
			msg += ": " + reason
		}
		return false, msg, nil
	}
	return true, "", nil
}

// managedRoleReconciled reads CloudNativePG's status.managedRolesStatus.byStatus:
// a map from state ("reconciled", "pending-reconciliation", "not-managed",
// "reserved") to the roles in it.
func managedRoleReconciled(cluster *unstructured.Unstructured, role string) bool {
	byStatus, _, _ := unstructured.NestedMap(cluster.Object, "status", "managedRolesStatus", "byStatus")
	reconciled, _ := byStatus["reconciled"].([]any)
	for _, r := range reconciled {
		if r == role {
			return true
		}
	}
	return false
}

// managedRoleCauses reads CloudNativePG's status.managedRolesStatus.cannotReconcile
// for one role: the causes it recorded (a PostgreSQL refusal, a password
// Secret it cannot fetch), sorted so the sentence is stable across passes.
func managedRoleCauses(cluster *unstructured.Unstructured, role string) []string {
	cannot, _, _ := unstructured.NestedMap(cluster.Object, "status", "managedRolesStatus", "cannotReconcile")
	raw, _ := cannot[role].([]any)
	causes := make([]string, 0, len(raw))
	for _, cause := range raw {
		if s, ok := cause.(string); ok && s != "" {
			causes = append(causes, s)
		}
	}
	sort.Strings(causes)
	return causes
}
