package manifestgraph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// graphOf builds a Graph with synthetic authored edges over n anonymous
// nodes — TopoOrder and FindCycle are pure over the adjacency structure.
func graphOf(dependsOn [][]int) *Graph {
	n := len(dependsOn)
	set := &Set{Nodes: make([]Node, n), index: map[Identity]int{}}
	for i := range set.Nodes {
		set.Nodes[i].Identity = Identity{Slug: string(rune('a' + i))}
	}
	g := &Graph{Set: set, DependsOn: make([][]Dependency, n)}
	for consumer, producers := range dependsOn {
		for _, producer := range producers {
			g.DependsOn[consumer] = append(g.DependsOn[consumer], Dependency{Producer: producer, Source: EdgeSourceValueFrom})
		}
	}
	return g
}

// TestTopoOrder_EveryEdgeRespected pins the ordering property: every
// producer lands before every consumer, and independent nodes fall back to
// authored order (determinism).
func TestTopoOrder_EveryEdgeRespected(t *testing.T) {
	// d depends on b and c; b and c depend on a; e independent.
	g := graphOf([][]int{
		0: {},
		1: {0},
		2: {0},
		3: {1, 2},
		4: {},
	})

	order, cycle := g.TopoOrder()
	assert.Nil(t, cycle)
	assert.Len(t, order, 5)

	position := map[int]int{}
	for pos, node := range order {
		position[node] = pos
	}
	for consumer, producers := range g.Producers() {
		for _, producer := range producers {
			assert.Less(t, position[producer], position[consumer],
				"producer %d must precede consumer %d", producer, consumer)
		}
	}
	// Determinism: among simultaneously-ready nodes the lower authored index
	// deploys first.
	assert.Equal(t, []int{0, 1, 2, 3, 4}, order)
}

func TestTopoOrder_CycleNamesTheChain(t *testing.T) {
	g := graphOf([][]int{
		0: {1},
		1: {0},
	})
	order, cycle := g.TopoOrder()
	assert.Nil(t, order)
	assert.NotNil(t, cycle)
	assert.Equal(t, FindingCycle, cycle.Class)
	assert.Contains(t, cycle.Message, "->")
}

func TestFindCycle_DagReturnsNil(t *testing.T) {
	assert.Nil(t, FindCycle([][]int{{}, {0}, {1}}))
	assert.NotNil(t, FindCycle([][]int{{1}, {2}, {0}}))
}
