package module

import (
	"strings"
	"testing"

	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// The orphaned-prerequisite refusal reads two things: the definition Stigmer's
// operator detects a prerequisite by, and the Deployments of the Helm release
// that installed it. The lanes prove the reads on a cluster; these pin how the
// answers map onto refusing or trusting.

func cluster(objects ...runtime.Object) *dynamicfake.FakeDynamicClient {
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
		definitions: "CustomResourceDefinitionList",
		deployments: "DeploymentList",
	}, objects...)
}

func definitionObject(name string, labels, annotations map[string]string) *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetAPIVersion("apiextensions.k8s.io/v1")
	o.SetKind("CustomResourceDefinition")
	o.SetName(name)
	o.SetLabels(labels)
	o.SetAnnotations(annotations)
	return o
}

func helmDefinition(name, release, namespace string) *unstructured.Unstructured {
	return definitionObject(name, map[string]string{"app.kubernetes.io/managed-by": "Helm"},
		map[string]string{"meta.helm.sh/release-name": release, "meta.helm.sh/release-namespace": namespace})
}

func deploymentObject(namespace, name, instance string) *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetAPIVersion("apps/v1")
	o.SetKind("Deployment")
	o.SetNamespace(namespace)
	o.SetName(name)
	o.SetLabels(map[string]string{"app.kubernetes.io/instance": instance})
	return o
}

func bundledLocals(backup bool) *Locals {
	spec := paritySpec()
	bundled := &v1.KubernetesStigmerBundledPostgres{}
	if backup {
		bundled.Backup = &v1.KubernetesStigmerPostgresBackup{}
	}
	spec.Database = &v1.KubernetesStigmerDatabase{Postgresql: &v1.KubernetesStigmerDatabase_Bundled{Bundled: bundled}}
	return &Locals{Spec: spec, Namespace: "stigmer", OperatorNamespace: "stigmer-system"}
}

func TestOrphanedCloudNativePgIsRefusedWithTheWayOut(t *testing.T) {
	client := cluster(helmDefinition(vars.CloudNativePgDefinition, "cnpg", "cnpg-system"))
	err := refuseOrphanedPrerequisites(client, bundledLocals(false))
	if err == nil {
		t.Fatal("a CloudNativePG definition whose Helm release has no Deployment left must be refused")
	}
	for _, want := range []string{
		"observed: CloudNativePG's definition clusters.postgresql.cnpg.io is on the cluster, installed by the Helm release cnpg in namespace cnpg-system, and no controller runs it",
		"meaning: Stigmer's operator trusts a CloudNativePG someone else installed",
		`kubectl get crd -o name | grep '\.postgresql\.cnpg\.io$' | xargs kubectl delete`,
		"set spec.prerequisites.postgres_operator to skip",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal is missing %q:\n%s", want, err)
		}
	}
}

func TestRunningOrOperatorOwnedPrerequisitesAreTrusted(t *testing.T) {
	for name, objects := range map[string][]runtime.Object{
		"its release still runs a Deployment": {
			helmDefinition(vars.CloudNativePgDefinition, "cnpg", "cnpg-system"),
			deploymentObject("cnpg-system", "cnpg-cloudnative-pg", "cnpg"),
		},
		"Stigmer's own operator installed it": {
			definitionObject(vars.CloudNativePgDefinition, map[string]string{vars.StigmerManagedByLabel: vars.StigmerManagedByValue}, nil),
		},
		"installed without Helm, so it names no controller": {
			definitionObject(vars.CloudNativePgDefinition, nil, nil),
		},
		"not on the cluster at all": nil,
	} {
		t.Run(name, func(t *testing.T) {
			if err := refuseOrphanedPrerequisites(cluster(objects...), bundledLocals(true)); err != nil {
				t.Fatalf("want trusted, got %v", err)
			}
		})
	}
}

func TestBackupPluginIsJudgedOnlyWhenTheDatabaseBacksUp(t *testing.T) {
	orphan := helmDefinition(vars.BackupPluginDefinition, "plugin-barman-cloud", "cnpg-system")
	if err := refuseOrphanedPrerequisites(cluster(orphan), bundledLocals(false)); err != nil {
		t.Fatalf("without backups the plugin is not used, so its leftover is not this install's concern: %v", err)
	}
	err := refuseOrphanedPrerequisites(cluster(orphan), bundledLocals(true))
	if err == nil || !strings.Contains(err.Error(), "CloudNativePG's backup plugin's definition objectstores.barmancloud.cnpg.io") {
		t.Fatalf("with backups an orphaned plugin must be refused, got %v", err)
	}
}

func TestSkipAndExternalDatabasesAreNeverJudged(t *testing.T) {
	orphan := helmDefinition(vars.CloudNativePgDefinition, "cnpg", "cnpg-system")
	skip := bundledLocals(false)
	skip.Spec.Prerequisites = &v1.KubernetesStigmerPrerequisites{PostgresOperator: "skip"}
	if err := refuseOrphanedPrerequisites(cluster(orphan), skip); err != nil {
		t.Fatalf("skip trusts the declaration completely: %v", err)
	}
	external := bundledLocals(false)
	external.Spec.Database = &v1.KubernetesStigmerDatabase{Postgresql: &v1.KubernetesStigmerDatabase_External{External: &v1.KubernetesStigmerExternalPostgres{}}}
	if err := refuseOrphanedPrerequisites(cluster(orphan), external); err != nil {
		t.Fatalf("an external database uses no CloudNativePG: %v", err)
	}
}
