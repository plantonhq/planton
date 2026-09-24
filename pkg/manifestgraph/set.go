package manifestgraph

import (
	"fmt"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

// Identity is a node's graph identity: the triple the platform's dependency
// graph keys on. Two manifests with the same identity are the same node.
type Identity struct {
	Kind cloudresourcekind.CloudResourceKind
	Slug string
	Env  string
}

func (id Identity) String() string {
	s := fmt.Sprintf("%s/%s", id.Kind, id.Slug)
	if id.Env != "" {
		s += "@" + id.Env
	}
	return s
}

// Item is one manifest offered to the set: the loaded message and a source
// label (a file path, a rendered-doc label) used in findings.
type Item struct {
	Msg    proto.Message
	Source string
}

// Node is one member of the set.
type Node struct {
	Identity Identity
	// Name is metadata.name as authored (identity carries the derived slug).
	Name   string
	Msg    proto.Message
	Source string

	refUses     []RefUse
	literalUses []LiteralUse
}

// Metadata returns the node's extracted metadata (nil-safe).
func (n *Node) Metadata() *shared.CloudResourceMetadata {
	return metadatareflect.ExtractMetadata(n.Msg)
}

// RefUses returns the node's collected valueFrom references (cached from set
// construction — the one traversal).
func (n *Node) RefUses() []RefUse {
	return n.refUses
}

// LiteralUses returns the node's collected literal StringValueOrRef arms
// (cached from the same traversal as RefUses).
func (n *Node) LiteralUses() []LiteralUse {
	return n.literalUses
}

// Set is a collection of manifests treated as one deployment set.
type Set struct {
	Nodes []Node

	index map[Identity]int
}

// NewSet builds a set from loaded manifests. Every item becomes a node with
// its identity derived (explicit metadata.slug, else generated from
// metadata.name; env from metadata.env). Duplicate identities are findings —
// the FIRST occurrence stays the node, later ones are reported and skipped so
// downstream passes still run on the healthy remainder.
func NewSet(items []Item) (*Set, []Finding) {
	set := &Set{index: make(map[Identity]int, len(items))}
	var findings []Finding

	for _, item := range items {
		meta := metadatareflect.ExtractMetadata(item.Msg)
		kindName, _ := crkreflect.ExtractKindFromProto(item.Msg)
		identity := Identity{
			Kind: crkreflect.KindFromString(kindName),
			Slug: ResolveSlug(meta),
			Env:  meta.GetEnv(),
		}
		if prev, dup := set.index[identity]; dup {
			id := identity
			findings = append(findings, Finding{
				Class:  FindingDuplicateIdentity,
				Source: item.Source,
				Node:   &id,
				Message: fmt.Sprintf("duplicate resource identity: %s is also defined in %s — a dependency graph cannot tell the two apart",
					identity, set.Nodes[prev].Source),
			})
			continue
		}
		refUses, literalUses := collectUses(item.Msg)
		set.index[identity] = len(set.Nodes)
		set.Nodes = append(set.Nodes, Node{
			Identity:    identity,
			Name:        meta.GetName(),
			Msg:         item.Msg,
			Source:      item.Source,
			refUses:     refUses,
			literalUses: literalUses,
		})
	}
	return set, findings
}

// Lookup returns the node index for an identity and whether it is in the set.
func (s *Set) Lookup(id Identity) (int, bool) {
	i, ok := s.index[id]
	return i, ok
}
