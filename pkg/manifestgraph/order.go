package manifestgraph

import (
	"fmt"
	"strings"
)

// TopoOrder returns a deployment order over the set's nodes: every producer
// before its consumers (Kahn's algorithm — the same discipline the platform's
// orchestrator schedules by). Among ready nodes the set's authored order
// breaks ties, so the result is deterministic for a given input. When the
// graph has a cycle, the returned finding names it as a chain and the order
// is nil.
func (g *Graph) TopoOrder() ([]int, *Finding) {
	edges := g.Producers()
	n := len(edges)
	inDegree := make([]int, n)
	consumersOf := make([][]int, n)
	for consumer, producers := range edges {
		inDegree[consumer] = len(producers)
		for _, producer := range producers {
			consumersOf[producer] = append(consumersOf[producer], consumer)
		}
	}

	// A sorted ready-queue (indexes ascend) keeps the order deterministic:
	// ties between independent nodes fall back to authored order.
	var ready []int
	for i := 0; i < n; i++ {
		if inDegree[i] == 0 {
			ready = append(ready, i)
		}
	}

	order := make([]int, 0, n)
	for len(ready) > 0 {
		next := ready[0]
		ready = ready[1:]
		order = append(order, next)
		for _, consumer := range consumersOf[next] {
			inDegree[consumer]--
			if inDegree[consumer] == 0 {
				ready = insertSorted(ready, consumer)
			}
		}
	}

	if len(order) != n {
		cycle := FindCycle(edges)
		parts := make([]string, 0, len(cycle))
		for _, idx := range cycle {
			parts = append(parts, g.Set.Nodes[idx].Identity.String())
		}
		return nil, &Finding{
			Class:   FindingCycle,
			Message: "dependency cycle: " + strings.Join(parts, " -> "),
		}
	}
	return order, nil
}

func insertSorted(ready []int, v int) []int {
	for i, existing := range ready {
		if v < existing {
			return append(ready[:i], append([]int{v}, ready[i:]...)...)
		}
	}
	return append(ready, v)
}

// FindCycle detects a dependency cycle in an adjacency list (edges point from
// a node to the nodes it depends on). It returns the members of one cycle in
// reference order, or nil when the graph is a DAG. One cycle is enough for a
// gate — the author fixes it and re-runs.
func FindCycle(edges [][]int) []int {
	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := make([]int, len(edges))
	var stack []int

	var visit func(node int) []int
	visit = func(node int) []int {
		state[node] = inStack
		stack = append(stack, node)
		for _, next := range edges[node] {
			switch state[next] {
			case inStack:
				// Found a back edge — slice the current stack from the first
				// occurrence of next to get the cycle members in order.
				for i, n := range stack {
					if n == next {
						cycle := make([]int, len(stack)-i)
						copy(cycle, stack[i:])
						return cycle
					}
				}
			case unvisited:
				if cycle := visit(next); cycle != nil {
					return cycle
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[node] = done
		return nil
	}

	for node := range edges {
		if state[node] == unvisited {
			if cycle := visit(node); cycle != nil {
				return cycle
			}
		}
	}
	return nil
}

// String renders one node's place in the graph for debugging and reports.
func (g *Graph) String() string {
	var b strings.Builder
	for i, node := range g.Set.Nodes {
		deps := make([]string, 0, len(g.DependsOn[i]))
		for _, dep := range g.DependsOn[i] {
			deps = append(deps, fmt.Sprintf("%s (%s)", g.Set.Nodes[dep.Producer].Identity, dep.Source))
		}
		fmt.Fprintf(&b, "%s <- [%s]\n", node.Identity, strings.Join(deps, ", "))
	}
	return b.String()
}
