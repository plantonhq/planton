package manifestgraph

import (
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// The operator-prerequisite edge source reads a fact the kind metadata
// already states: a hosted platform names the operator it needs in its
// kind_meta.prerequisites (KubernetesPostgres needs
// KubernetesCloudNativePgOperator because it creates postgresql.cnpg.io
// Cluster objects only that operator admits and reconciles). Nothing in the
// manifest spells the dependency — there is no field to reference an
// operator by — yet on a fresh cluster the operator must exist first.
//
// The rule is deliberately narrow so the graph never guesses:
//
//   - only prerequisites in the Kubernetes operators-and-controllers service
//     group qualify: operators, controllers, and the CRD bundles that admit
//     a kind's objects (a Gateway needs the Gateway API CRDs the same way a
//     Postgres cluster needs CloudNativePG). An ordinary prerequisite (a
//     topic's Kafka, Airflow's Postgres) is authored through a reference
//     field and never drawn from metadata; what admits and reconciles is a
//     cluster-singleton fact and has no field;
//   - only when EXACTLY ONE instance of the prerequisite kind is in the set,
//     in the same env. Two candidate operators mean the set does not say
//     which one reconciles the operand, and the graph stays silent rather
//     than choose (the same silence the connection-placement source keeps
//     when no sibling publishes a connection);
//   - the direct prerequisites list only. Transitivity re-emerges from the
//     set itself when the intermediate kinds are also present.
//
// The platform's DAG factory derives the same edge from the same metadata
// (its OperatorPrerequisiteSelector); the corpus scenarios
// `operator-prerequisite` and `operator-prerequisite-ambiguous` pin the two
// lanes to agree.

// kindInstanceKey indexes the set's nodes by kind within an env.
type kindInstanceKey struct {
	Kind cloudresourcekind.CloudResourceKind
	Env  string
}

// kindInstanceIndex maps every (kind, env) to the node indexes of that kind
// in that env, in authored order.
func kindInstanceIndex(set *Set) map[kindInstanceKey][]int {
	index := map[kindInstanceKey][]int{}
	for i := range set.Nodes {
		key := kindInstanceKey{Kind: set.Nodes[i].Identity.Kind, Env: set.Nodes[i].Identity.Env}
		index[key] = append(index[key], i)
	}
	return index
}

// operatorPrerequisites returns the operator kinds a node's kind declares as
// prerequisites: the kind's direct prerequisites filtered to the Kubernetes
// operators-and-controllers service group.
func operatorPrerequisites(kind cloudresourcekind.CloudResourceKind) []cloudresourcekind.CloudResourceKind {
	var operators []cloudresourcekind.CloudResourceKind
	for _, prerequisite := range crkreflect.Prerequisites(kind) {
		group, err := crkreflect.ServiceGroup(prerequisite)
		if err != nil || group != cloudresourcekind.CloudProviderServiceGroup_kubernetes_operators_controllers {
			continue
		}
		operators = append(operators, prerequisite)
	}
	return operators
}

// soleInstance returns the one node of the given kind in the env, or -1 when
// the set holds none or more than one.
func soleInstance(index map[kindInstanceKey][]int, kind cloudresourcekind.CloudResourceKind, env string) int {
	instances := index[kindInstanceKey{Kind: kind, Env: env}]
	if len(instances) != 1 {
		return -1
	}
	return instances[0]
}
