package mappingeval

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	suitev1 "github.com/plantonhq/planton/qa/mappingevalsuite/v1"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// SuiteMember is one validated suite member: the declared kind plus its
// loaded fixture manifest.
type SuiteMember struct {
	KindDir      string
	Kind         catalogkind.CatalogKind
	Name         string
	ManifestPath string
	// Manifest is the fixture as authored -- value_from references
	// unresolved. The deployer resolves a COPY against earlier members'
	// outputs; this original stays the ground truth's answer key.
	Manifest proto.Message
	// Edges are the manifest's value_from references.
	Edges []RefEdge
}

// LoadedSuite is a fully validated MappingEvalSuite.
type LoadedSuite struct {
	Suite   *suitev1.MappingEvalSuite
	Members []SuiteMember
}

// LoadSuite reads a MappingEvalSuite from YAML and validates it against the
// repo: every member's manifest exists, parses as its declared kind's
// kind, and references only members listed BEFORE it. Deploy order is list
// order, so a forward or dangling reference means the suite cannot deploy --
// it fails here, before anything touches the cloud.
func LoadSuite(repoRoot, suitePath string) (*LoadedSuite, error) {
	s := &suitev1.MappingEvalSuite{}
	if err := protobufyaml.Load(suitePath, s); err != nil {
		return nil, errors.Wrapf(err, "loading suite from %s", suitePath)
	}
	if s.GetKind() != "MappingEvalSuite" {
		return nil, errors.Errorf("%s: kind is %q, want MappingEvalSuite", suitePath, s.GetKind())
	}
	if s.GetMetadata().GetName() == "" {
		return nil, errors.Errorf("%s: metadata.name is empty", suitePath)
	}
	if len(s.GetSpec().GetMembers()) == 0 {
		return nil, errors.Errorf("%s: suite declares no members", suitePath)
	}
	if s.GetSpec().GetScanScope().GetRegion() == "" {
		return nil, errors.Errorf("%s: scan_scope.region is empty", suitePath)
	}

	loaded := &LoadedSuite{Suite: s}
	// Names seen so far, keyed by kind -- the backward-reference rule's
	// lookup table.
	earlier := map[catalogkind.CatalogKind]map[string]bool{}
	allEarlierNames := map[string]bool{}

	for i, member := range s.GetSpec().GetMembers() {
		kind := catalogkindreflect.KindFromString(member.GetKindDir())
		if kind == catalogkind.CatalogKind_unspecified {
			return nil, errors.Errorf("members[%d]: kind %q is not a registered kind", i, member.GetKindDir())
		}
		manifestPath := filepath.Join(repoRoot, member.GetManifestPath())
		if _, err := os.Stat(manifestPath); err != nil {
			return nil, errors.Errorf("members[%d] (%s): manifest %s does not exist", i, member.GetKindDir(), member.GetManifestPath())
		}
		m, err := manifest.LoadManifest(manifestPath)
		if err != nil {
			return nil, errors.Wrapf(err, "members[%d] (%s): loading %s", i, member.GetKindDir(), member.GetManifestPath())
		}
		if got := catalogkindreflect.KindFromString(kindNameOf(m)); got != kind {
			return nil, errors.Errorf("members[%d]: manifest %s is a %s, but the member declares kind %q",
				i, member.GetManifestPath(), got, member.GetKindDir())
		}
		name := manifestMetadataName(m)
		if name == "" {
			return nil, errors.Errorf("members[%d] (%s): manifest has no metadata.name", i, member.GetKindDir())
		}
		if earlier[kind][name] {
			return nil, errors.Errorf("members[%d]: duplicate member %s %q", i, kind, name)
		}

		edges, err := ExtractRefEdges(m)
		if err != nil {
			return nil, errors.Wrapf(err, "members[%d] (%s %q)", i, kind, name)
		}
		for _, edge := range edges {
			if !earlierTargetExists(edge, earlier, allEarlierNames) {
				return nil, errors.Errorf("members[%d] (%s %q): value_from at %s targets %s %q, which is not an EARLIER member -- deploy order is list order, so references must point backward",
					i, kind, name, edge.FieldPath, edge.TargetKind, edge.TargetName)
			}
		}

		if earlier[kind] == nil {
			earlier[kind] = map[string]bool{}
		}
		earlier[kind][name] = true
		allEarlierNames[name] = true

		loaded.Members = append(loaded.Members, SuiteMember{
			KindDir:      member.GetKindDir(),
			Kind:         kind,
			Name:         name,
			ManifestPath: manifestPath,
			Manifest:     m,
			Edges:        edges,
		})
	}
	return loaded, nil
}

// earlierTargetExists mirrors edgeTargetExists but against the members
// already listed.
func earlierTargetExists(edge RefEdge, earlier map[catalogkind.CatalogKind]map[string]bool, allNames map[string]bool) bool {
	if edge.TargetKind != catalogkind.CatalogKind_unspecified {
		return earlier[edge.TargetKind][edge.TargetName]
	}
	return allNames[edge.TargetName]
}

// kindNameOf reads the "kind" field off a typed KRM message.
func kindNameOf(m proto.Message) string {
	top := m.ProtoReflect()
	kindField := top.Descriptor().Fields().ByName("kind")
	if kindField == nil {
		return ""
	}
	return top.Get(kindField).String()
}
