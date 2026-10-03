package mappingeval

import (
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// AccountResourceRef identifies one discovered account resource by the same
// coordinates the read-only inventory scan reports: CloudFormation type name
// plus the type's primary identifier. It is the shared currency between the
// scan, a proposal's claims, and the ground truth's ownership records.
type AccountResourceRef struct {
	TypeName   string `json:"typeName"`
	Identifier string `json:"identifier"`
}

// ScannedResource is one account resource as the blind side sees it: the
// scan coordinates plus the type's property model (Cloud Control's JSON
// property document, optionally completed by declared enrichments).
type ScannedResource struct {
	TypeName   string         `json:"typeName"`
	Identifier string         `json:"identifier"`
	Properties map[string]any `json:"properties"`
}

// Ref returns the resource's scan coordinates.
func (r ScannedResource) Ref() AccountResourceRef {
	return AccountResourceRef{TypeName: r.TypeName, Identifier: r.Identifier}
}

// Scan is the complete blind-side input to a proposer: everything the
// read-only scan surfaced within the declared scope. It is deliberately
// JSON-serializable so live scans can be recorded as offline fixtures.
type Scan struct {
	// Region is the single region the scan covered.
	Region string `json:"region"`
	// Resources are the discovered account resources across every scanned
	// type.
	Resources []ScannedResource `json:"resources"`
}

// GroundTruth is the answer key a proposal is graded against: the kind
// instances that were actually deployed, with their manifests, the account
// resources each one owns, and what the scan structurally cannot see.
type GroundTruth struct {
	Instances []GroundTruthInstance
}

// GroundTruthInstance is one deployed component instance.
type GroundTruthInstance struct {
	// KindDir is the kind directory name (e.g. "awsvpc").
	KindDir string
	// Kind is the kind's resolved CatalogKind.
	Kind catalogkind.CatalogKind
	// Name is the deployed manifest's metadata.name.
	Name string
	// Manifest is the kind's typed api message AS AUTHORED -- value_from
	// references unresolved. It is the answer key for both the spec axis
	// (which fields, what values) and the refs axis (where the edges run).
	Manifest proto.Message
	// Claims are the scan-visible account resources this instance owns:
	// every IaC state resource whose type carries a cloud_control_type_name
	// in the provider import catalog, identified exactly as the scan
	// reports it.
	Claims []AccountResourceRef
	// InvisibleResourceTypes are the instance's IaC resource types with NO
	// scan-side correspondence -- resources a proposer can never claim
	// because the scan cannot show them. Recorded so reports state what was
	// structurally out of reach rather than silently shrinking the exam.
	InvisibleResourceTypes []string
}

// Universe returns every scan-visible account resource the ground truth
// owns, mapped to the name of the owning instance. This is the scored
// universe: grouping precision/recall and coverage are defined over exactly
// these resources.
func (gt *GroundTruth) Universe() map[AccountResourceRef]string {
	universe := make(map[AccountResourceRef]string)
	for _, instance := range gt.Instances {
		for _, claim := range instance.Claims {
			universe[claim] = instance.Name
		}
	}
	return universe
}
