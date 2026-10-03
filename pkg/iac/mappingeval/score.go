package mappingeval

import (
	"sort"
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// structFullName marks google.protobuf.Struct spec fields (dynamic JSON
// documents like IAM policies). They are compared as single leaves via
// proto.Equal -- recursing into their internal representation would grade
// encoding details, not mapping quality.
const structFullName = "google.protobuf.Struct"

// ScoreOptions carries the declared knowledge the scorer needs beyond the
// ground truth itself. Both members derive from the provider import catalog
// and the kinds' import maps (see NameDerivedIdentityChecks and
// ConfigOnlySpecFieldExclusions) -- nothing here is authored per suite.
type ScoreOptions struct {
	// ExcludedSpecFields are spec field names (proto snake_case) excluded
	// from the spec axis at any depth: values that exist only in IaC
	// configuration, never on the infra component (the catalog's
	// config_only_attributes), so no scan-driven proposer could ever
	// reconstruct them. Expecting them would penalize physics, not mapping
	// quality. This leans on the deliberate convention that spec fields
	// carry the same snake_case names as the module attributes they drive
	// (force_destroy, revoke_rules_on_delete, ...).
	ExcludedSpecFields map[string]bool

	// NameDerivedIdentity maps a kind to the Cloud Control type name whose
	// claimed identifier MUST equal the proposed metadata.name, because the
	// kind's import recipe derives its import id from_metadata_name (S3:
	// the manifest name IS the bucket name). Names are otherwise never
	// scored; this is the one declared exception -- breaking the derivation
	// breaks the downstream zero-typing import.
	NameDerivedIdentity map[catalogkind.CatalogKind]string

	// GradeEnvironmentPartition enables the partition axis. Unlike the
	// members above it derives from the SUITE (grade_environment_partition),
	// not the catalog: whether member environments are recoverable answer
	// keys is an exam-design fact, declared per suite.
	GradeEnvironmentPartition bool
}

// Score grades a proposal against the ground truth. See the package doc for
// the axes and the honesty rules.
func Score(gt *GroundTruth, proposal *LoadedProposal, opts ScoreOptions) *Report {
	report := &Report{}
	universe := gt.Universe()
	matches := matchInstances(gt, proposal)

	matchedGT := map[string]*InstanceMatch{}
	matchedProp := map[*ProposedInstance]*InstanceMatch{}
	for i := range matches {
		m := &matches[i]
		matchedGT[m.GroundTruth.Name] = m
		matchedProp[m.Proposed] = m
	}

	scoreGrouping(gt, proposal, universe, matchedGT, report)
	scoreSpecs(matches, gt, proposal, opts, report)
	scoreRefs(matches, gt, matchedGT, report)
	scorePartition(gt, matchedGT, opts, report)
	scoreCoverage(gt, proposal, universe, report)
	checkNameDerivedIdentity(matches, opts, report)

	for _, instance := range gt.Instances {
		entry := InstanceReport{
			Kind:                   instance.Kind.String(),
			GroundTruthName:        instance.Name,
			InvisibleResourceTypes: instance.InvisibleResourceTypes,
		}
		if m, ok := matchedGT[instance.Name]; ok {
			entry.ProposedName = m.Proposed.Name
			entry.ClaimOverlap = m.Overlap
		}
		report.Instances = append(report.Instances, entry)
	}
	return report
}

// scoreGrouping grades resource-to-instance assignment over the scored
// universe. For every universe resource: it is correctly grouped when
// exactly one proposed instance claims it AND that instance is the matched
// counterpart of the resource's ground-truth owner. Precision counts every
// in-universe claim (so duplicate claims cost), recall counts the universe.
func scoreGrouping(gt *GroundTruth, proposal *LoadedProposal, universe map[AccountResourceRef]string, matchedGT map[string]*InstanceMatch, report *Report) {
	claimers := map[AccountResourceRef][]*ProposedInstance{}
	inUniverseClaims := 0
	for i := range proposal.Instances {
		instance := &proposal.Instances[i]
		for _, claim := range instance.Claims {
			if _, in := universe[claim]; !in {
				continue
			}
			claimers[claim] = append(claimers[claim], instance)
			inUniverseClaims++
		}
	}

	correct := 0
	for ref, ownerName := range universe {
		claiming := claimers[ref]
		if len(claiming) > 1 {
			report.Grouping.DuplicateClaims = append(report.Grouping.DuplicateClaims, ref)
			continue
		}
		if len(claiming) == 0 {
			report.Grouping.Unclaimed = append(report.Grouping.Unclaimed, ref)
			continue
		}
		match, matched := matchedGT[ownerName]
		if matched && claiming[0] == match.Proposed {
			correct++
			continue
		}
		report.Grouping.Misassigned = append(report.Grouping.Misassigned, MisassignedResource{
			Resource:        ref,
			GroundTruthName: ownerName,
			ClaimedBy:       claiming[0].Name,
		})
	}
	sortRefs(report.Grouping.Unclaimed)
	sortRefs(report.Grouping.DuplicateClaims)
	report.Grouping.Correct = correct
	report.Grouping.UniverseSize = len(universe)
	report.Grouping.InUniverseClaims = inUniverseClaims
}

// scoreSpecs grades manifest reconstruction per matched pair: recall over
// the leaves the ground-truth spec sets. Proposer-extra leaves (set in the
// proposal, absent in the ground truth) are reported, never penalized --
// materializing a cloud-observed default explicitly is honest, not wrong.
func scoreSpecs(matches []InstanceMatch, gt *GroundTruth, proposal *LoadedProposal, opts ScoreOptions, report *Report) {
	for _, m := range matches {
		gtSpec := specMessage(m.GroundTruth.Manifest.ProtoReflect())
		propSpec := specMessage(m.Proposed.Manifest.ProtoReflect())
		if gtSpec == nil || propSpec == nil {
			continue
		}
		diffSpecMessages(gtSpec, propSpec, "", m.GroundTruth.Name, opts.ExcludedSpecFields, &report.Spec)
		collectExtraLeaves(gtSpec, propSpec, "", m.GroundTruth.Name, opts.ExcludedSpecFields, &report.Spec)
	}
	// Ground-truth instances with no match at all: every set leaf is
	// missing -- the manifests were simply never proposed.
	matchedNames := map[string]bool{}
	for _, m := range matches {
		matchedNames[m.GroundTruth.Name] = true
	}
	for _, instance := range gt.Instances {
		if matchedNames[instance.Name] {
			continue
		}
		gtSpec := specMessage(instance.Manifest.ProtoReflect())
		if gtSpec == nil {
			continue
		}
		countMissingLeaves(gtSpec, "", instance.Name, opts.ExcludedSpecFields, &report.Spec)
	}
}

// diffSpecMessages compares the leaves the ground-truth message sets against
// the proposal message. StringValueOrRef fields participate only through
// their literal arm (the ref arm is the refs axis's business).
func diffSpecMessages(gtMsg, propMsg protoreflect.Message, path, instanceName string, excluded map[string]bool, spec *SpecScore) {
	fields := gtMsg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !gtMsg.Has(fd) {
			continue
		}
		if excluded[string(fd.Name())] {
			spec.ExcludedConfigOnly++
			continue
		}
		fieldPath := joinPath(path, string(fd.Name()))

		if isRefField(fd) {
			gtRef := gtMsg.Get(fd).Message().Interface().(*foreignkeyv1.StringValueOrRef)
			if gtRef.GetValueFrom() != nil {
				continue // an edge -- scored by the refs axis
			}
			var propValue string
			propSet := propMsg.Has(fd)
			if propSet {
				propRef := propMsg.Get(fd).Message().Interface().(*foreignkeyv1.StringValueOrRef)
				propValue = propRef.GetValue()
			}
			recordLeaf(spec, fieldPath, instanceName, gtRef.GetValue(), propValue, propSet)
			continue
		}

		switch {
		case fd.IsMap():
			gtMap := gtMsg.Get(fd).Map()
			propSet := propMsg.Has(fd)
			equal := propSet && mapsEqual(gtMap, propMsg.Get(fd).Map())
			recordLeaf(spec, fieldPath, instanceName, "<map>", conditional(equal, "<map>", "<different map>"), propSet)
		case fd.IsList():
			gtList := gtMsg.Get(fd).List()
			if isRefListOfScalars(fd) {
				// A repeated StringValueOrRef participates exactly like its
				// singular form: literal arms are ONE whole-list spec leaf
				// (mirroring scalar lists, so element count never moves the
				// denominator); value_from arms are edges, the refs axis's
				// business. A list carrying only edges contributes nothing
				// here.
				gtLiterals := refListLiterals(gtList)
				if len(gtLiterals) == 0 {
					continue
				}
				propSet := propMsg.Has(fd)
				equal := propSet && stringSlicesEqual(gtLiterals, refListLiterals(propMsg.Get(fd).List()))
				recordLeaf(spec, fieldPath, instanceName, "<list>", conditional(equal, "<list>", "<different list>"), propSet)
				continue
			}
			if fd.Kind() == protoreflect.MessageKind {
				propSet := propMsg.Has(fd)
				var propList protoreflect.List
				if propSet {
					propList = propMsg.Get(fd).List()
				}
				for j := 0; j < gtList.Len(); j++ {
					elementPath := fieldPath + "[" + strconv.Itoa(j) + "]"
					if !propSet || j >= propList.Len() {
						countMissingLeavesInMessage(gtList.Get(j).Message(), elementPath, instanceName, excluded, spec)
						continue
					}
					diffSpecMessages(gtList.Get(j).Message(), propList.Get(j).Message(), elementPath, instanceName, excluded, spec)
				}
				continue
			}
			// Scalar list: one leaf, compared whole.
			propSet := propMsg.Has(fd)
			equal := propSet && scalarListsEqual(gtList, propMsg.Get(fd).List())
			recordLeaf(spec, fieldPath, instanceName, "<list>", conditional(equal, "<list>", "<different list>"), propSet)
		case fd.Kind() == protoreflect.MessageKind && string(fd.Message().FullName()) == structFullName:
			propSet := propMsg.Has(fd)
			equal := propSet && proto.Equal(gtMsg.Get(fd).Message().Interface(), propMsg.Get(fd).Message().Interface())
			recordLeaf(spec, fieldPath, instanceName, "<document>", conditional(equal, "<document>", "<different document>"), propSet)
		case fd.Kind() == protoreflect.MessageKind:
			if !propMsg.Has(fd) {
				countMissingLeavesInMessage(gtMsg.Get(fd).Message(), fieldPath, instanceName, excluded, spec)
				continue
			}
			diffSpecMessages(gtMsg.Get(fd).Message(), propMsg.Get(fd).Message(), fieldPath, instanceName, excluded, spec)
		default:
			propSet := propMsg.Has(fd)
			var propValue string
			if propSet {
				propValue = propMsg.Get(fd).String()
			}
			recordLeaf(spec, fieldPath, instanceName, gtMsg.Get(fd).String(), propValue, propSet)
		}
	}
}

// collectExtraLeaves reports scalar leaves the proposal sets where the
// ground truth sets nothing -- informational only.
func collectExtraLeaves(gtMsg, propMsg protoreflect.Message, path, instanceName string, excluded map[string]bool, spec *SpecScore) {
	fields := propMsg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !propMsg.Has(fd) || excluded[string(fd.Name())] || isRefField(fd) {
			continue
		}
		// A repeated ref carrying only value_from arms is pure edge
		// territory (the refs axis reports unexpected edges); only its
		// literal arms are spec material.
		if isRefListOfScalars(fd) && len(refListLiterals(propMsg.Get(fd).List())) == 0 {
			continue
		}
		fieldPath := joinPath(path, string(fd.Name()))
		if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() && !isRefMessage(fd.Message()) {
			if gtMsg.Has(fd) {
				collectExtraLeaves(gtMsg.Get(fd).Message(), propMsg.Get(fd).Message(), fieldPath, instanceName, excluded, spec)
			} else {
				spec.ProposerExtra = append(spec.ProposerExtra, FieldFinding{Instance: instanceName, FieldPath: fieldPath})
			}
			continue
		}
		if !gtMsg.Has(fd) {
			spec.ProposerExtra = append(spec.ProposerExtra, FieldFinding{Instance: instanceName, FieldPath: fieldPath})
		}
	}
}

// countMissingLeaves marks every set leaf of an unmatched instance's spec
// missing.
func countMissingLeaves(gtSpec protoreflect.Message, path, instanceName string, excluded map[string]bool, spec *SpecScore) {
	countMissingLeavesInMessage(gtSpec, path, instanceName, excluded, spec)
}

func countMissingLeavesInMessage(gtMsg protoreflect.Message, path, instanceName string, excluded map[string]bool, spec *SpecScore) {
	fields := gtMsg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !gtMsg.Has(fd) {
			continue
		}
		if excluded[string(fd.Name())] {
			spec.ExcludedConfigOnly++
			continue
		}
		fieldPath := joinPath(path, string(fd.Name()))
		if isRefField(fd) {
			gtRef := gtMsg.Get(fd).Message().Interface().(*foreignkeyv1.StringValueOrRef)
			if gtRef.GetValueFrom() != nil {
				continue
			}
			recordLeaf(spec, fieldPath, instanceName, gtRef.GetValue(), "", false)
			continue
		}
		switch {
		case fd.IsMap(), fd.IsList() && fd.Kind() != protoreflect.MessageKind:
			recordLeaf(spec, fieldPath, instanceName, "<collection>", "", false)
		case fd.IsList():
			if isRefListOfScalars(fd) {
				// Same rule as the matched path: literal arms are one
				// whole-list leaf (owed here), value_from arms are edges
				// (owed by the refs axis's unmatched-instance accounting).
				if len(refListLiterals(gtMsg.Get(fd).List())) > 0 {
					recordLeaf(spec, fieldPath, instanceName, "<list>", "", false)
				}
				continue
			}
			list := gtMsg.Get(fd).List()
			for j := 0; j < list.Len(); j++ {
				countMissingLeavesInMessage(list.Get(j).Message(), fieldPath+"["+strconv.Itoa(j)+"]", instanceName, excluded, spec)
			}
		case fd.Kind() == protoreflect.MessageKind && string(fd.Message().FullName()) == structFullName:
			recordLeaf(spec, fieldPath, instanceName, "<document>", "", false)
		case fd.Kind() == protoreflect.MessageKind:
			countMissingLeavesInMessage(gtMsg.Get(fd).Message(), fieldPath, instanceName, excluded, spec)
		default:
			recordLeaf(spec, fieldPath, instanceName, gtMsg.Get(fd).String(), "", false)
		}
	}
}

// scoreRefs grades value_from wiring per matched pair. A ground-truth edge
// is reproduced when the proposal has an edge at the SAME spec location
// whose target resolves to the matched counterpart of the ground-truth
// edge's target. A literal (or nothing) where the ground truth has an edge
// is a miss; a proposed edge with no ground-truth counterpart costs
// precision.
func scoreRefs(matches []InstanceMatch, gt *GroundTruth, matchedGT map[string]*InstanceMatch, report *Report) {
	for _, m := range matches {
		gtEdges, err := ExtractRefEdges(m.GroundTruth.Manifest)
		if err != nil {
			continue // the manifest loaded once already; unreachable in practice
		}
		propEdgesByPath := map[string]RefEdge{}
		for _, edge := range m.Proposed.Edges {
			propEdgesByPath[edge.FieldPath] = edge
		}

		for _, gtEdge := range gtEdges {
			report.Refs.GroundTruthEdges++
			propEdge, present := propEdgesByPath[gtEdge.FieldPath]
			if !present {
				report.Refs.MissingEdges = append(report.Refs.MissingEdges, EdgeFinding{
					Instance:  m.GroundTruth.Name,
					FieldPath: gtEdge.FieldPath,
					Detail:    "ground truth references " + gtEdge.TargetName + "; proposal has a literal or nothing",
				})
				continue
			}
			delete(propEdgesByPath, gtEdge.FieldPath)

			// Translate the ground-truth target through the instance
			// matching: the proposal cannot know internal names, so the
			// edge is correct when it points at the target's matched
			// counterpart.
			targetMatch, targetMatched := matchedGT[gtEdge.TargetName]
			if !targetMatched {
				report.Refs.MissingEdges = append(report.Refs.MissingEdges, EdgeFinding{
					Instance:  m.GroundTruth.Name,
					FieldPath: gtEdge.FieldPath,
					Detail:    "target " + gtEdge.TargetName + " was never matched to a proposed instance",
				})
				continue
			}
			if propEdge.TargetName == targetMatch.Proposed.Name &&
				(propEdge.TargetKind == catalogkind.CatalogKind_unspecified ||
					targetMatch.Proposed.Kind == propEdge.TargetKind) {
				report.Refs.CorrectEdges++
				continue
			}
			report.Refs.WrongTargetEdges = append(report.Refs.WrongTargetEdges, EdgeFinding{
				Instance:  m.GroundTruth.Name,
				FieldPath: gtEdge.FieldPath,
				Detail:    "proposal references " + propEdge.TargetName + ", want the counterpart of " + gtEdge.TargetName + " (" + targetMatch.Proposed.Name + ")",
			})
		}

		// Whatever proposal edges remain have no ground-truth counterpart.
		for _, propEdge := range propEdgesByPath {
			report.Refs.UnexpectedEdges = append(report.Refs.UnexpectedEdges, EdgeFinding{
				Instance:  m.Proposed.Name,
				FieldPath: propEdge.FieldPath,
				Detail:    "proposal references " + propEdge.TargetName + " where the ground truth has a literal or nothing",
			})
		}
	}

	// Ground-truth instances with no match at all still owe their edges:
	// an unproposed instance's wiring is missing, not exempt. Without this
	// (mirroring the spec axis's countMissingLeaves), skipping a resource
	// would silently shrink the denominator and a proposer could escape
	// its ref debt by proposing less.
	for _, instance := range gt.Instances {
		if _, matched := matchedGT[instance.Name]; matched {
			continue
		}
		gtEdges, err := ExtractRefEdges(instance.Manifest)
		if err != nil {
			continue // the manifest loaded once already; unreachable in practice
		}
		for _, gtEdge := range gtEdges {
			report.Refs.GroundTruthEdges++
			report.Refs.MissingEdges = append(report.Refs.MissingEdges, EdgeFinding{
				Instance:  instance.Name,
				FieldPath: gtEdge.FieldPath,
				Detail:    "ground truth references " + gtEdge.TargetName + "; the instance was never proposed",
			})
		}
	}
}

// scorePartition grades environment assignment on suites that declare it:
// each ground-truth instance that sets metadata.env owes that environment
// on its proposed counterpart. Instances are walked in ground-truth order
// (deterministic reports). Wrong beats missing in severity -- a wrong
// environment poisons env-scoped references, while honest unassignment is
// exactly what the review gate exists to resolve. A proposed env where the
// ground truth has none is informational, mirroring spec proposer-extra.
func scorePartition(gt *GroundTruth, matchedGT map[string]*InstanceMatch, opts ScoreOptions, report *Report) {
	if !opts.GradeEnvironmentPartition {
		return
	}
	report.Partition.Graded = true
	for _, instance := range gt.Instances {
		gtEnv := manifestMetadataEnv(instance.Manifest)
		match, matched := matchedGT[instance.Name]
		var propEnv string
		if matched {
			propEnv = manifestMetadataEnv(match.Proposed.Manifest)
		}

		if gtEnv == "" {
			if propEnv != "" {
				report.Partition.ExtraEnv = append(report.Partition.ExtraEnv, EnvFinding{
					Instance: instance.Name, Actual: propEnv,
				})
			}
			continue
		}

		report.Partition.GroundTruthInstances++
		switch {
		case propEnv == gtEnv:
			report.Partition.Correct++
		case propEnv == "":
			report.Partition.MissingEnv = append(report.Partition.MissingEnv, EnvFinding{
				Instance: instance.Name, Expected: gtEnv,
			})
		default:
			report.Partition.WrongEnv = append(report.Partition.WrongEnv, EnvFinding{
				Instance: instance.Name, Expected: gtEnv, Actual: propEnv,
			})
		}
	}
}

// scoreCoverage reports the honesty ledger: how much of the scored universe
// the proposal accounts for (claimed or explicitly unmapped), what it left
// unaccounted, and what it talked about outside the universe.
func scoreCoverage(gt *GroundTruth, proposal *LoadedProposal, universe map[AccountResourceRef]string, report *Report) {
	claimed := map[AccountResourceRef]bool{}
	for _, instance := range proposal.Instances {
		for _, claim := range instance.Claims {
			if _, in := universe[claim]; in {
				claimed[claim] = true
			} else {
				report.Coverage.OutOfUniverseClaims = append(report.Coverage.OutOfUniverseClaims, claim)
			}
		}
	}
	unmapped := map[AccountResourceRef]bool{}
	for _, u := range proposal.Proposal.GetSpec().GetUnmapped() {
		ref := AccountResourceRef{TypeName: u.GetTypeName(), Identifier: u.GetIdentifier()}
		if _, in := universe[ref]; in {
			unmapped[ref] = true
			report.Coverage.UnmappedInUniverse = append(report.Coverage.UnmappedInUniverse, ref)
		}
	}
	for ref := range universe {
		if !claimed[ref] && !unmapped[ref] {
			report.Coverage.Unaccounted = append(report.Coverage.Unaccounted, ref)
		}
	}
	sortRefs(report.Coverage.OutOfUniverseClaims)
	sortRefs(report.Coverage.UnmappedInUniverse)
	sortRefs(report.Coverage.Unaccounted)
	report.Coverage.UniverseSize = len(universe)
	report.Coverage.ClaimedInUniverse = len(claimed)
}

// checkNameDerivedIdentity enforces the one declared name rule: where a
// kind's import recipe derives the import id from metadata.name, the
// proposed name must equal the claimed identifier of the kind's primary
// resource, or the downstream zero-typing import breaks.
func checkNameDerivedIdentity(matches []InstanceMatch, opts ScoreOptions, report *Report) {
	for _, m := range matches {
		ccType, ruled := opts.NameDerivedIdentity[m.Proposed.Kind]
		if !ruled {
			continue
		}
		for _, claim := range m.Proposed.Claims {
			if claim.TypeName != ccType {
				continue
			}
			if claim.Identifier != m.Proposed.Name {
				report.NameDerivability = append(report.NameDerivability, NameDerivabilityFinding{
					Instance:   m.Proposed.Name,
					Kind:       m.Proposed.Kind.String(),
					Identifier: claim.Identifier,
				})
			}
		}
	}
}

func recordLeaf(spec *SpecScore, fieldPath, instanceName, gtValue, propValue string, propSet bool) {
	spec.GroundTruthLeaves++
	if !propSet {
		spec.Missing = append(spec.Missing, FieldFinding{Instance: instanceName, FieldPath: fieldPath, Expected: gtValue})
		return
	}
	if gtValue == propValue {
		spec.Matched++
		return
	}
	spec.Mismatched = append(spec.Mismatched, FieldFinding{Instance: instanceName, FieldPath: fieldPath, Expected: gtValue, Actual: propValue})
}

func specMessage(top protoreflect.Message) protoreflect.Message {
	specField := top.Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind || !top.Has(specField) {
		return nil
	}
	return top.Get(specField).Message()
}

// isRefField reports a SINGULAR StringValueOrRef field. Repeated refs are
// deliberately excluded: they are graded elementwise (each element's
// literal arm is a spec leaf, each value_from arm an edge), so callers
// must reach their list handling -- treating the whole list as one ref
// here would panic on the list value.
func isRefField(fd protoreflect.FieldDescriptor) bool {
	return fd.Kind() == protoreflect.MessageKind && !fd.IsMap() && !fd.IsList() && string(fd.Message().FullName()) == stringValueOrRefFullName
}

func isRefMessage(md protoreflect.MessageDescriptor) bool {
	return string(md.FullName()) == stringValueOrRefFullName
}

func isRefListOfScalars(fd protoreflect.FieldDescriptor) bool {
	return fd.IsList() && fd.Kind() == protoreflect.MessageKind && isRefMessage(fd.Message())
}

// refListLiterals extracts the literal arms of a repeated StringValueOrRef,
// in order. value_from arms are deliberately absent -- they are edges.
func refListLiterals(list protoreflect.List) []string {
	var literals []string
	for i := 0; i < list.Len(); i++ {
		ref, ok := list.Get(i).Message().Interface().(*foreignkeyv1.StringValueOrRef)
		if !ok || ref.GetValueFrom() != nil {
			continue
		}
		literals = append(literals, ref.GetValue())
	}
	return literals
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func scalarListsEqual(a, b protoreflect.List) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if a.Get(i).String() != b.Get(i).String() {
			return false
		}
	}
	return true
}

func mapsEqual(a, b protoreflect.Map) bool {
	if a.Len() != b.Len() {
		return false
	}
	equal := true
	a.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		other := b.Get(k)
		if !other.IsValid() || other.String() != v.String() {
			equal = false
			return false
		}
		return true
	})
	return equal
}

func sortRefs(refs []AccountResourceRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].TypeName != refs[j].TypeName {
			return refs[i].TypeName < refs[j].TypeName
		}
		return refs[i].Identifier < refs[j].Identifier
	})
}

func conditional(cond bool, whenTrue, whenFalse string) string {
	if cond {
		return whenTrue
	}
	return whenFalse
}
