package mappingeval

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	proposalv1 "github.com/plantonhq/planton/iac/importmappingproposal/v1"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProposedInstance is one proposal resource in validated, typed form: the
// manifest parsed into its kind's api message plus the claim accounting.
type ProposedInstance struct {
	Kind      catalogkind.CatalogKind
	Name      string
	Manifest  proto.Message
	Claims    []AccountResourceRef
	Rationale string
	// Edges are the manifest's value_from references, extracted once at
	// load so every consumer sees the same edge set.
	Edges []RefEdge
}

// LoadedProposal is a fully validated ImportMappingProposal.
type LoadedProposal struct {
	Proposal  *proposalv1.ImportMappingProposal
	Instances []ProposedInstance
}

// LoadProposalFile reads an ImportMappingProposal from a YAML file and
// validates it. See ParseProposal for the contract enforced.
func LoadProposalFile(path string) (*LoadedProposal, error) {
	p := &proposalv1.ImportMappingProposal{}
	if err := protobufyaml.Load(path, p); err != nil {
		return nil, errors.Wrapf(err, "loading proposal from %s", path)
	}
	return ParseProposal(p)
}

// ParseProposal validates an ImportMappingProposal against its contract and
// returns the typed form. The contract is enforced HERE, at the seam every
// proposer's output crosses, so a malformed proposal fails immediately with
// a contract-level error instead of surfacing later as a confusing scoring
// or creation failure:
//
//   - every proposed manifest parses into its kind's typed api message with
//     unknown fields rejected (the exact strictness real manifests get);
//   - metadata.name is present and unique per kind;
//   - every proposed instance claims at least one account resource (a
//     manifest that accounts for nothing is not a mapping);
//   - every value_from reference targets another instance IN the proposal
//     -- a dangling reference would fail resolution at the first infra-job
//     build, so it is rejected at the contract. (References to resources a
//     platform already manages are legal in production; an eval proposal is
//     self-contained by construction.)
func ParseProposal(p *proposalv1.ImportMappingProposal) (*LoadedProposal, error) {
	if p.GetKind() != "ImportMappingProposal" {
		return nil, errors.Errorf("kind is %q, want ImportMappingProposal", p.GetKind())
	}

	loaded := &LoadedProposal{Proposal: p}
	namesByKind := map[catalogkind.CatalogKind]map[string]bool{}

	for i, resource := range p.GetSpec().GetResources() {
		if resource.GetManifest() == nil {
			return nil, errors.Errorf("resources[%d]: no manifest", i)
		}
		// The Struct round-trips through JSON into the kind's typed message
		// via the same loader every real manifest passes through --
		// protojson strictness (unknown fields rejected), kind resolution,
		// and proto-declared defaults included.
		manifestMap := resource.GetManifest().AsMap()
		manifestJSON, err := json.Marshal(manifestMap)
		if err != nil {
			return nil, errors.Wrapf(err, "resources[%d]: serializing manifest", i)
		}
		typed, err := manifest.LoadManifestBytes(manifestJSON, "proposal resources["+strconv.Itoa(i)+"]")
		if err != nil {
			return nil, errors.Wrapf(err, "resources[%d]: manifest does not parse as its kind", i)
		}

		kindName, _ := manifestMap["kind"].(string)
		kind := catalogkindreflect.KindFromString(kindName)
		if kind == catalogkind.CatalogKind_unspecified {
			return nil, errors.Errorf("resources[%d]: manifest kind %q is not a registered CatalogKind", i, kindName)
		}
		name := manifestMetadataName(typed)
		if name == "" {
			return nil, errors.Errorf("resources[%d]: manifest has no metadata.name", i)
		}
		if len(resource.GetClaims()) == 0 {
			return nil, errors.Errorf("resources[%d] (%s %q): claims no account resources -- a manifest that accounts for nothing is not a mapping", i, kind, name)
		}
		for j, claim := range resource.GetClaims() {
			if claim.GetTypeName() == "" || claim.GetIdentifier() == "" {
				return nil, errors.Errorf("resources[%d] (%s %q): claims[%d] missing type_name or identifier", i, kind, name, j)
			}
		}
		if namesByKind[kind] == nil {
			namesByKind[kind] = map[string]bool{}
		}
		if namesByKind[kind][name] {
			return nil, errors.Errorf("resources[%d]: duplicate instance %s %q", i, kind, name)
		}
		namesByKind[kind][name] = true

		edges, err := ExtractRefEdges(typed)
		if err != nil {
			return nil, errors.Wrapf(err, "resources[%d] (%s %q)", i, kind, name)
		}

		claims := make([]AccountResourceRef, 0, len(resource.GetClaims()))
		for _, claim := range resource.GetClaims() {
			claims = append(claims, AccountResourceRef{TypeName: claim.GetTypeName(), Identifier: claim.GetIdentifier()})
		}
		loaded.Instances = append(loaded.Instances, ProposedInstance{
			Kind:      kind,
			Name:      name,
			Manifest:  typed,
			Claims:    claims,
			Rationale: resource.GetRationale(),
			Edges:     edges,
		})
	}

	// Dangling-reference check runs after all instances are known, so
	// ordering inside the proposal carries no meaning.
	for _, instance := range loaded.Instances {
		for _, edge := range instance.Edges {
			if !edgeTargetExists(edge, namesByKind) {
				return nil, errors.Errorf("%s %q: value_from at %s targets %s %q, which this proposal does not propose -- a dangling reference fails resolution at the first infra-job build",
					instance.Kind, instance.Name, edge.FieldPath, edge.TargetKind, edge.TargetName)
			}
		}
	}

	for i, unmapped := range p.GetSpec().GetUnmapped() {
		if unmapped.GetTypeName() == "" || unmapped.GetIdentifier() == "" {
			return nil, errors.Errorf("unmapped[%d]: missing type_name or identifier", i)
		}
		if unmapped.GetReason() == "" {
			return nil, errors.Errorf("unmapped[%d] (%s %s): no reason -- the honest remainder must say WHY it maps to nothing",
				i, unmapped.GetTypeName(), unmapped.GetIdentifier())
		}
	}

	return loaded, nil
}

// edgeTargetExists checks an edge against the proposal's own instances. A
// reference that states its kind must match kind+name; a kind-less
// reference (a polymorphic field with no default_kind, pointing wherever
// its target_type says) matches by name alone.
func edgeTargetExists(edge RefEdge, namesByKind map[catalogkind.CatalogKind]map[string]bool) bool {
	if edge.TargetKind != catalogkind.CatalogKind_unspecified {
		return namesByKind[edge.TargetKind][edge.TargetName]
	}
	for _, names := range namesByKind {
		if names[edge.TargetName] {
			return true
		}
	}
	return false
}

// manifestMetadataName reads metadata.name off a typed KRM message.
func manifestMetadataName(m proto.Message) string {
	return manifestMetadataField(m, "name")
}

// manifestMetadataEnv reads metadata.env off a typed KRM message -- the
// surface the partition axis grades.
func manifestMetadataEnv(m proto.Message) string {
	return manifestMetadataField(m, "env")
}

func manifestMetadataField(m proto.Message, field protoreflect.Name) string {
	top := m.ProtoReflect()
	metadataField := top.Descriptor().Fields().ByName("metadata")
	if metadataField == nil || metadataField.Kind() != protoreflect.MessageKind {
		return ""
	}
	valueField := metadataField.Message().Fields().ByName(field)
	if valueField == nil {
		return ""
	}
	return top.Get(metadataField).Message().Get(valueField).String()
}
