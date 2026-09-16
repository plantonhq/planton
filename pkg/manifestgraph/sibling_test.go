package manifestgraph

import (
	"testing"

	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetescloudnativepgoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescloudnativepgoperator/v1alpha1"
	kubernetesgatewayv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgateway/v1alpha1"
	kuberneteshttproutev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteshttproute/v1alpha1"
	kubernetespostgresv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// The literal-sibling and operator-prerequisite sources, each pinned at its
// rule. The corpus scenarios pin the ORDER each edge produces; these pin the
// matching rule itself, so a drift in the rule is named at the rule.

func routeNamingGateway(name, env, gatewayName string) proto.Message {
	return &kuberneteshttproutev1alpha1.KubernetesHttpRoute{
		Kind:     "KubernetesHttpRoute",
		Metadata: &shared.CloudResourceMetadata{Name: name, Env: env},
		Spec: &kuberneteshttproutev1alpha1.KubernetesHttpRouteSpec{
			Namespace:  literal("shop"),
			ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{{Name: literal(gatewayName)}},
		},
	}
}

func gateway(name, env string) proto.Message {
	return &kubernetesgatewayv1alpha1.KubernetesGateway{
		Kind:     "KubernetesGateway",
		Metadata: &shared.CloudResourceMetadata{Name: name, Env: env},
		Spec:     &kubernetesgatewayv1alpha1.KubernetesGatewaySpec{Namespace: literal("istio-ingress")},
	}
}

func postgres(name, env string) proto.Message {
	return &kubernetespostgresv1alpha1.KubernetesPostgres{
		Kind:     "KubernetesPostgres",
		Metadata: &shared.CloudResourceMetadata{Name: name, Env: env},
		Spec:     &kubernetespostgresv1alpha1.KubernetesPostgresSpec{Namespace: literal("app")},
	}
}

func cnpgOperator(name, env string) proto.Message {
	return &kubernetescloudnativepgoperatorv1alpha1.KubernetesCloudNativePgOperator{
		Kind:     "KubernetesCloudNativePgOperator",
		Metadata: &shared.CloudResourceMetadata{Name: name, Env: env},
		Spec:     &kubernetescloudnativepgoperatorv1alpha1.KubernetesCloudNativePgOperatorSpec{Namespace: literal("cnpg-system")},
	}
}

func graphOver(msgs ...proto.Message) *Graph {
	items := make([]Item, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, Item{Msg: m, Source: "test"})
	}
	set, _ := NewSet(items)
	return BuildGraph(set)
}

func TestCollectLiteralUses_ReadsTheLiteralArmAtDepth(t *testing.T) {
	uses := CollectLiteralUses(routeNamingGateway("shop-web", "dev", "public-gateway"))
	paths := map[string]string{}
	for _, u := range uses {
		paths[u.FieldPath] = u.Value
	}
	assert.Equal(t, "public-gateway", paths["spec.parent_refs[0].name"],
		"a literal inside a repeated message is found on the same traversal that finds references")
	assert.Equal(t, cloudresourcekind.CloudResourceKind_KubernetesGateway,
		annotatedKind(uses[len(uses)-1].Field), "the declaring field carries the default kind the literal is matched against")
}

func TestLiteralSibling_MatchesTheDeclaredKindBySlugInTheSameEnv(t *testing.T) {
	g := graphOver(routeNamingGateway("shop-web", "dev", "Public Gateway"), gateway("public-gateway", "dev"))
	assert.Equal(t, []Dependency{{Producer: 1, Source: EdgeSourceLiteralSibling, FieldPath: "spec.parent_refs[0].name"}}, g.DependsOn[0],
		"the literal slugs through the one slug function and meets the sibling")
}

func TestLiteralSibling_NeverCrossesEnvsOrKinds(t *testing.T) {
	g := graphOver(routeNamingGateway("shop-web", "dev", "public-gateway"), gateway("public-gateway", "prod"))
	assert.Empty(t, g.DependsOn[0], "a sibling in another env is not the one the literal names")
	assert.Empty(t, g.Findings, "an unmatched literal is the common case, never a finding")
}

func TestLiteralSibling_LeavesTheNamespaceToPlacement(t *testing.T) {
	g := graphOver(gateway("public-gateway", "dev"))
	assert.Empty(t, g.DependsOn[0])
	assert.Equal(t, []Identity{{Kind: cloudresourcekind.CloudResourceKind_KubernetesNamespace, Slug: "istio-ingress", Env: "dev"}}, g.Derived,
		"a literal namespace is the placement source's: it alone mints a derived target")
}

func TestOperatorPrerequisites_FilterToTheOperatorGroup(t *testing.T) {
	assert.Equal(t, []cloudresourcekind.CloudResourceKind{cloudresourcekind.CloudResourceKind_KubernetesCloudNativePgOperator},
		operatorPrerequisites(cloudresourcekind.CloudResourceKind_KubernetesPostgres))
	assert.Equal(t, []cloudresourcekind.CloudResourceKind{cloudresourcekind.CloudResourceKind_KubernetesGatewayApiCrds},
		operatorPrerequisites(cloudresourcekind.CloudResourceKind_KubernetesGateway),
		"a CRD bundle is a cluster-singleton prerequisite of the same class as an operator")
	assert.Empty(t, operatorPrerequisites(cloudresourcekind.CloudResourceKind_KubernetesNamespace),
		"a kind with no prerequisite yields nothing")
}

func TestOperatorPrerequisite_EdgeOnlyWhenExactlyOneInstance(t *testing.T) {
	one := graphOver(postgres("app-db", "dev"), cnpgOperator("cnpg", "dev"))
	assert.Equal(t, []Dependency{{Producer: 1, Source: EdgeSourceOperatorPrerequisite}}, one.DependsOn[0])

	two := graphOver(postgres("app-db", "dev"), cnpgOperator("cnpg-a", "dev"), cnpgOperator("cnpg-b", "dev"))
	assert.Empty(t, two.DependsOn[0], "two candidates: the set does not say which reconciles the operand")

	none := graphOver(postgres("app-db", "dev"), cnpgOperator("cnpg", "prod"))
	assert.Empty(t, none.DependsOn[0], "an operator in another env is not this env's")
}

func TestBuildGraph_AnInferredEdgeYieldsToTheAuthorsCycle(t *testing.T) {
	operator := &kubernetescloudnativepgoperatorv1alpha1.KubernetesCloudNativePgOperator{
		Kind:     "KubernetesCloudNativePgOperator",
		Metadata: &shared.CloudResourceMetadata{Name: "cnpg", Env: "dev"},
		Spec: &kubernetescloudnativepgoperatorv1alpha1.KubernetesCloudNativePgOperatorSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind: cloudresourcekind.CloudResourceKind_KubernetesPostgres, Name: "app-db", FieldPath: "status.outputs.namespace",
			}}},
		},
	}
	g := graphOver(postgres("app-db", "dev"), operator)

	assert.Empty(t, g.DependsOn[0], "the inferred operator edge is dropped")
	assert.Equal(t, []Dependency{{Producer: 0, Source: EdgeSourceValueFrom, FieldPath: "spec.namespace"}}, g.DependsOn[1],
		"the authored edge stands")
	order, cycle := g.TopoOrder()
	assert.Nil(t, cycle)
	assert.Equal(t, []int{0, 1}, order)
	if assert.Len(t, g.Findings, 1) {
		assert.Equal(t, FindingDerivedEdgeDropped, g.Findings[0].Class)
		assert.Equal(t, "KubernetesPostgres/app-db@dev", g.Findings[0].Node.String())
	}
}

func TestBuildGraph_AnAuthoredCycleStands(t *testing.T) {
	g := graphOf([][]int{0: {1}, 1: {0}})
	g.dropInferredEdgesOnCycles()
	assert.Len(t, g.DependsOn[0], 1)
	assert.Len(t, g.DependsOn[1], 1)
	assert.Empty(t, g.Findings, "the author wrote it, and the author must break it")
}
