package catalogkindreflect

import (
	"sort"
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// publishes_kubernetes_connection is an ORDERING fact every lane reads: a
// workload naming the connection a flagged kind will publish orders after
// that kind. The rules below keep the flag coherent with what the platform
// can honor -- it materializes a connection only for a cluster it created,
// and only a cluster is a place workloads run on.

func flaggedKinds() []catalogkind.CatalogKind {
	var flagged []catalogkind.CatalogKind
	for _, kind := range KindsList() {
		meta, err := KindMeta(kind)
		if err == nil && meta.GetPublishesKubernetesConnection() {
			flagged = append(flagged, kind)
		}
	}
	sort.Slice(flagged, func(i, j int) bool { return flagged[i].Number() < flagged[j].Number() })
	return flagged
}

func TestPublishesKubernetesConnectionKindsAreContainerKindsOnACloudProvider(t *testing.T) {
	for _, kind := range flaggedKinds() {
		meta, _ := KindMeta(kind)
		if !meta.GetContainerKind() {
			t.Errorf("%s publishes a kubernetes connection but is not a container kind: workloads run INSIDE a cluster", kind)
		}
		if meta.GetProvider() == catalogkind.CatalogProvider_kubernetes {
			t.Errorf("%s is a kubernetes-provider kind: it deploys THROUGH a connection and cannot publish the one it needs", kind)
		}
	}
}

// The flagged set is pinned: the platform pairs every flagged kind with a
// connection materializer, so adding a kind here is a two-repo change made
// deliberately (flag the kind, ship its materializer), never a drive-by.
func TestPublishesKubernetesConnectionKindsArePinned(t *testing.T) {
	want := []catalogkind.CatalogKind{
		catalogkind.CatalogKind_AwsEksCluster,
		catalogkind.CatalogKind_AzureAksCluster,
		catalogkind.CatalogKind_GcpGkeCluster,
	}
	got := flaggedKinds()
	if len(got) != len(want) {
		t.Fatalf("publishes_kubernetes_connection kinds changed:\n want %v\n got  %v\n(a new cluster kind needs its platform connection materializer in the same release train)", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("publishes_kubernetes_connection kinds changed at %d: want %s, got %s", i, want[i], got[i])
		}
	}
}
