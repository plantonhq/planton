package crkreflect

import (
	"testing"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

func TestPrerequisites_PostgresRequiresZalandoOperator(t *testing.T) {
	prereqs := Prerequisites(cloudresourcekind.CloudResourceKind_KubernetesPostgres)
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 prerequisite for KubernetesPostgres, got %d", len(prereqs))
	}
	if prereqs[0] != cloudresourcekind.CloudResourceKind_KubernetesZalandoPostgresOperator {
		t.Fatalf("expected KubernetesZalandoPostgresOperator, got %s", prereqs[0])
	}
}

func TestPrerequisites_SolrRequiresSolrOperator(t *testing.T) {
	prereqs := Prerequisites(cloudresourcekind.CloudResourceKind_KubernetesSolr)
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 prerequisite for KubernetesSolr, got %d", len(prereqs))
	}
	if prereqs[0] != cloudresourcekind.CloudResourceKind_KubernetesSolrOperator {
		t.Fatalf("expected KubernetesSolrOperator, got %s", prereqs[0])
	}
}

func TestPrerequisites_NamespaceHasNone(t *testing.T) {
	prereqs := Prerequisites(cloudresourcekind.CloudResourceKind_KubernetesNamespace)
	if len(prereqs) != 0 {
		t.Fatalf("expected 0 prerequisites for KubernetesNamespace, got %d", len(prereqs))
	}
}

func TestHasPrerequisites(t *testing.T) {
	if !HasPrerequisites(cloudresourcekind.CloudResourceKind_KubernetesPostgres) {
		t.Fatal("expected KubernetesPostgres to have prerequisites")
	}
	if HasPrerequisites(cloudresourcekind.CloudResourceKind_KubernetesRedis) {
		t.Fatal("expected KubernetesRedis to have no prerequisites")
	}
}

func TestTransitivePrerequisites_DirectDep(t *testing.T) {
	prereqs, err := TransitivePrerequisites(cloudresourcekind.CloudResourceKind_KubernetesKafka)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prereqs) != 1 {
		t.Fatalf("expected 1 transitive prerequisite for KubernetesKafka, got %d", len(prereqs))
	}
	if prereqs[0] != cloudresourcekind.CloudResourceKind_KubernetesStrimziKafkaOperator {
		t.Fatalf("expected KubernetesStrimziKafkaOperator, got %s", prereqs[0])
	}
}

func TestTransitivePrerequisites_NoDeps(t *testing.T) {
	prereqs, err := TransitivePrerequisites(cloudresourcekind.CloudResourceKind_KubernetesDeployment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prereqs) != 0 {
		t.Fatalf("expected 0 transitive prerequisites for KubernetesDeployment, got %d", len(prereqs))
	}
}

func TestAllSixOperatorDependentComponents(t *testing.T) {
	cases := []struct {
		kind   cloudresourcekind.CloudResourceKind
		expect cloudresourcekind.CloudResourceKind
	}{
		{cloudresourcekind.CloudResourceKind_KubernetesPostgres, cloudresourcekind.CloudResourceKind_KubernetesZalandoPostgresOperator},
		{cloudresourcekind.CloudResourceKind_KubernetesKafka, cloudresourcekind.CloudResourceKind_KubernetesStrimziKafkaOperator},
		{cloudresourcekind.CloudResourceKind_KubernetesElasticsearch, cloudresourcekind.CloudResourceKind_KubernetesElasticOperator},
		{cloudresourcekind.CloudResourceKind_KubernetesMongodb, cloudresourcekind.CloudResourceKind_KubernetesPerconaMongoOperator},
		{cloudresourcekind.CloudResourceKind_KubernetesSolr, cloudresourcekind.CloudResourceKind_KubernetesSolrOperator},
		{cloudresourcekind.CloudResourceKind_KubernetesClickHouse, cloudresourcekind.CloudResourceKind_KubernetesAltinityOperator},
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
