package manifestgraph

import (
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// The connection-placement edge source reads a two-ended contract written in
// annotations, not in spec fields:
//
//   - a Kubernetes workload names the connection it deploys THROUGH with
//     planton.dev/connection (the consumer end);
//   - a cluster names the connection the platform PUBLISHES when it deploys
//     with planton.dev/connection-name, or takes the default formula
//     <env>-<name> (the producer end).
//
// Both ends slug through the ONE slug function, so a workload whose
// annotation equals the slug a sibling cluster will publish runs on that
// cluster — an ordering fact of the manifest set alone, with no relationship
// authored and no backend consulted. Which kinds publish a connection is the
// kind's own kind_meta (publishes_kubernetes_connection), read here and by
// the platform. The platform's orchestrator derives the same edge (its
// DeclaredConnectionClusterLookup); the corpus scenario
// `connection-placement` pins the two lanes to agree.
const (
	// AnnotationConnection names the provider connection a resource deploys through.
	AnnotationConnection = "planton.dev/connection"
	// AnnotationConnectionName overrides the name of the connection a cluster publishes.
	AnnotationConnectionName = "planton.dev/connection-name"
)

// publishesKubernetesConnection reads the kind's compile-time kind_meta flag:
// the one place, shared with the platform, that says a cluster kind's deploy
// publishes a Kubernetes provider connection. A kind without the flag (or
// unknown to the registry) publishes none.
func publishesKubernetesConnection(kind cloudresourcekind.CloudResourceKind) bool {
	meta, err := crkreflect.KindMeta(kind)
	return err == nil && meta.GetPublishesKubernetesConnection()
}

// PublishedConnectionSlug returns the slug of the Kubernetes provider
// connection the platform publishes when this node's cluster deploys, or ""
// when the node's kind publishes none. The name is the
// planton.dev/connection-name annotation taken literally, else
// <env>-<name>; the slug derives through GenerateSlug — byte-for-byte the
// materializer's frozen formula, so a chart can template one value on both
// ends and never drift.
func PublishedConnectionSlug(node *Node) string {
	if !publishesKubernetesConnection(node.Identity.Kind) {
		return ""
	}
	meta := node.Metadata()
	name := meta.GetAnnotations()[AnnotationConnectionName]
	if name == "" {
		name = meta.GetEnv() + "-" + meta.GetName()
	}
	return GenerateSlug(name)
}

// ConsumedConnectionSlug returns the connection slug a Kubernetes-provider
// node deploys through, as its planton.dev/connection annotation names it, or
// "" when the node is not a Kubernetes kind or names no connection. The
// annotation value is taken as the slug it is: that is exactly what the
// platform resolves the connection by, so a value that is not slug-shaped
// would fail there too, and this lane must not pretend otherwise. A node
// with no annotation may still deploy through an environment or organization
// default connection — a backend fact this lane cannot see, and does not
// guess.
func ConsumedConnectionSlug(node *Node) string {
	if crkreflect.GetProvider(node.Identity.Kind) != cloudresourcekind.CloudResourceProvider_kubernetes {
		return ""
	}
	return node.Metadata().GetAnnotations()[AnnotationConnection]
}

// publishedConnectionIndex maps each connection slug published by a cluster in
// the set to that cluster's node index. Two clusters publishing one slug is
// the platform's foreign-record refusal at deploy; here the first authored
// wins, matching how duplicate identities are kept.
func publishedConnectionIndex(set *Set) map[string]int {
	index := map[string]int{}
	for i := range set.Nodes {
		slug := PublishedConnectionSlug(&set.Nodes[i])
		if slug == "" {
			continue
		}
		if _, dup := index[slug]; !dup {
			index[slug] = i
		}
	}
	return index
}
