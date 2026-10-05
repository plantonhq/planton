package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

func TestPrerequisites_PostgresRequiresCloudNativePgOperator(t *testing.T) {
	prereqs := Prerequisites(catalogkind.CatalogKind_KubernetesPostgres)
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 prerequisite for KubernetesPostgres, got %d", len(prereqs))
	}
	if prereqs[0] != catalogkind.CatalogKind_KubernetesCloudNativePgOperator {
		t.Fatalf("expected KubernetesCloudNativePgOperator, got %s", prereqs[0])
	}
}

func TestPrerequisites_SolrRequiresSolrOperator(t *testing.T) {
	prereqs := Prerequisites(catalogkind.CatalogKind_KubernetesSolr)
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 prerequisite for KubernetesSolr, got %d", len(prereqs))
	}
	if prereqs[0] != catalogkind.CatalogKind_KubernetesSolrOperator {
		t.Fatalf("expected KubernetesSolrOperator, got %s", prereqs[0])
	}
}

func TestPrerequisites_NamespaceHasNone(t *testing.T) {
	prereqs := Prerequisites(catalogkind.CatalogKind_KubernetesNamespace)
	if len(prereqs) != 0 {
		t.Fatalf("expected 0 prerequisites for KubernetesNamespace, got %d", len(prereqs))
	}
}

func TestHasPrerequisites(t *testing.T) {
	if !HasPrerequisites(catalogkind.CatalogKind_KubernetesPostgres) {
		t.Fatal("expected KubernetesPostgres to have prerequisites")
	}
	if HasPrerequisites(catalogkind.CatalogKind_KubernetesValkey) {
		t.Fatal("expected KubernetesValkey to have no prerequisites")
	}
}

func TestTransitivePrerequisites_DirectDep(t *testing.T) {
	prereqs, err := TransitivePrerequisites(catalogkind.CatalogKind_KubernetesKafka)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 transitive prerequisite for KubernetesKafka, got %d", len(prereqs))
	}
	if prereqs[0] != catalogkind.CatalogKind_KubernetesStrimziKafkaOperator {
		t.Fatalf("expected KubernetesStrimziKafkaOperator, got %s", prereqs[0])
	}
}

func TestTransitivePrerequisites_NoDeps(t *testing.T) {
	prereqs, err := TransitivePrerequisites(catalogkind.CatalogKind_KubernetesDeployment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prereqs) != 0 {
		t.Fatalf("expected 0 transitive prerequisites for KubernetesDeployment, got %d", len(prereqs))
	}
}

func TestAllSevenOperatorDependentKinds(t *testing.T) {
	cases := []struct {
		kind   catalogkind.CatalogKind
		expect catalogkind.CatalogKind
	}{
		{catalogkind.CatalogKind_KubernetesPostgres, catalogkind.CatalogKind_KubernetesCloudNativePgOperator},
		{catalogkind.CatalogKind_KubernetesKafka, catalogkind.CatalogKind_KubernetesStrimziKafkaOperator},
		{catalogkind.CatalogKind_KubernetesOpenSearch, catalogkind.CatalogKind_KubernetesOpenSearchOperator},
		{catalogkind.CatalogKind_KubernetesMongodb, catalogkind.CatalogKind_KubernetesPerconaMongoOperator},
		{catalogkind.CatalogKind_KubernetesMysql, catalogkind.CatalogKind_KubernetesPerconaMysqlOperator},
		{catalogkind.CatalogKind_KubernetesSolr, catalogkind.CatalogKind_KubernetesSolrOperator},
		{catalogkind.CatalogKind_KubernetesClickHouse, catalogkind.CatalogKind_KubernetesAltinityOperator},
	}

	for _, tc := range cases {
		t.Run(tc.kind.String(), func(t *testing.T) {
			prereqs := Prerequisites(tc.kind)
			if len(prereqs) == 0 {
				t.Fatalf("expected prerequisites for %s, got none", tc.kind)
			}
			if prereqs[0] != tc.expect {
				t.Fatalf("expected %s, got %s", tc.expect, prereqs[0])
			}
		})
	}
}
