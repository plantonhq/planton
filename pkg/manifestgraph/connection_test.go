package manifestgraph

import (
	"testing"

	gcpgkeclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkecluster/v1alpha1"
	gcpvpcnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcnetwork/v1alpha1"
	kubernetesnamespacev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnamespace/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// The two ends of the connection-placement contract, each read on its own.
// The corpus scenarios pin the ORDER the edge produces; these pin the
// naming rule the edge is derived from, so a drift in the formula is named
// at the formula, not discovered as a reordered golden.

func clusterNode(name, env string, annotations map[string]string) *Node {
	msg := &gcpgkeclusterv1alpha1.GcpGkeCluster{
		Kind:     "GcpGkeCluster",
		Metadata: &shared.CloudResourceMetadata{Name: name, Env: env, Annotations: annotations},
	}
	return nodeOf(msg)
}

func nodeOf(msg proto.Message) *Node {
	set, _ := NewSet([]Item{{Msg: msg, Source: "test"}})
	return &set.Nodes[0]
}

func TestPublishedConnectionSlug_TakesTheAnnotationLiterally(t *testing.T) {
	node := clusterNode("platform-cluster", "prod", map[string]string{
		AnnotationConnectionName: "Platform Cluster",
	})
	assert.Equal(t, "platform-cluster", PublishedConnectionSlug(node),
		"the annotation is the name; the slug derives through the one slug function")
}

func TestPublishedConnectionSlug_DefaultsToEnvDashName(t *testing.T) {
	node := clusterNode("platform", "prod", nil)
	assert.Equal(t, "prod-platform", PublishedConnectionSlug(node),
		"absent the annotation the materializer's frozen formula applies")
}

func TestPublishedConnectionSlug_IsEmptyForAKindThatPublishesNothing(t *testing.T) {
	node := nodeOf(&gcpvpcnetworkv1alpha1.GcpVpcNetwork{
		Kind:     "GcpVpcNetwork",
		Metadata: &shared.CloudResourceMetadata{Name: "vpc", Env: "prod", Annotations: map[string]string{AnnotationConnectionName: "x"}},
	})
	assert.Equal(t, "", PublishedConnectionSlug(node))
}

func TestConsumedConnectionSlug_ReadsAKubernetesKindsAnnotationVerbatim(t *testing.T) {
	node := nodeOf(&kubernetesnamespacev1alpha1.KubernetesNamespace{
		Kind:     "KubernetesNamespace",
		Metadata: &shared.CloudResourceMetadata{Name: "apps", Env: "prod", Annotations: map[string]string{AnnotationConnection: "prod-platform"}},
	})
	assert.Equal(t, "prod-platform", ConsumedConnectionSlug(node))
}

func TestConsumedConnectionSlug_IgnoresNonKubernetesKinds(t *testing.T) {
	node := clusterNode("platform", "prod", map[string]string{AnnotationConnection: "some-gcp-connection"})
	assert.Equal(t, "", ConsumedConnectionSlug(node),
		"a GCP kind deploys through a GCP connection, which opens no cluster")
}

func TestBuildGraph_ClusterNeverPlacedInsideItself(t *testing.T) {
	// A cluster that both publishes a connection and (nonsensically) names it
	// as its own consumer must not depend on itself; addEdge refuses self-edges
	// and the consumer check refuses non-Kubernetes kinds — belt and braces.
	msg := &gcpgkeclusterv1alpha1.GcpGkeCluster{
		Kind: "GcpGkeCluster",
		Metadata: &shared.CloudResourceMetadata{Name: "platform", Env: "prod", Annotations: map[string]string{
			AnnotationConnectionName: "prod-platform",
			AnnotationConnection:     "prod-platform",
		}},
	}
	set, _ := NewSet([]Item{{Msg: msg, Source: "test"}})
	g := BuildGraph(set)
	assert.Empty(t, g.DependsOn[0])
}
