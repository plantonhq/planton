// Package capacityestimator evaluates a kind's capacity derivation
// against one typed manifest. It is the execution engine of the
// capacity-derivation standard, the cluster-capacity twin of the cost
// estimator: workload bindings locate the manifest's ContainerResources
// blocks, instance counts, and per-instance volumes, and the engine sums
// them into the capacity footprint the workload reserves from its target
// cluster -- exact Kubernetes-quantity arithmetic, never a float. A
// manifest that states no reservation at all refuses honestly instead of
// emitting an empty footprint.
//
// The result is the same shape an authored estimate model preset carries
// (a capacity_footprint with configuration-scoped exclusions and notes),
// so the estimate generator's emission runs unchanged on evaluated
// output. Condition semantics for the conditional prose live in the cost
// estimator (one home): effective values, placeholder-as-unset, and
// reference presence behave identically across both standards.
package capacityestimator

import (
	"fmt"
	"math/big"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/plantonhq/planton/catalog/kubernetes"
	capacityv1 "github.com/plantonhq/planton/finops/catalogkindcapacityderivation/v1"
	costestimatev1 "github.com/plantonhq/planton/finops/catalogkindcostestimate/v1"
	estimatemodelv1 "github.com/plantonhq/planton/finops/catalogkindcostestimatemodel/v1"
	"github.com/plantonhq/planton/pkg/finops/costestimator"
	"github.com/plantonhq/planton/pkg/specpath"
	sharedoptions "github.com/plantonhq/planton/shared/options"
)

// Evaluate runs a capacity derivation against one manifest and returns
// exactly one of: the evaluated preset-model shape (a capacity footprint
// with configuration-scoped exclusions and notes), a refusal, or an error
// for a malformed derivation (unresolvable path, unparseable quantity --
// the conformance gate catches path problems before CI ever runs the
// engine, so errors here mean the gate was bypassed or a preset carries a
// quantity outside the grammar).
func Evaluate(
	manifest proto.Message,
	spec *capacityv1.CatalogKindCapacityDerivationSpec,
) (*estimatemodelv1.PresetEstimateModel, *costestimator.Refusal, error) {
	specMsg, err := costestimator.ManifestSpec(manifest)
	if err != nil {
		return nil, nil, err
	}

	totals := &footprintTotals{}
	var segments []string
	for _, workload := range spec.GetWorkloads() {
		segment, err := accumulateWorkload(specMsg, workload, totals)
		if err != nil {
			return nil, nil, err
		}
		if segment != "" {
			segments = append(segments, segment)
		}
	}

	if totals.empty() {
		return nil, &costestimator.Refusal{Reason: "the manifest declares no resource requests, limits, or volumes -- the reservation is unknowable from this manifest"}, nil
	}

	exclusions, err := costestimator.MatchingTexts(specMsg, spec.GetExclusions())
	if err != nil {
		return nil, nil, err
	}
	notes, err := costestimator.MatchingTexts(specMsg, spec.GetNotes())
	if err != nil {
		return nil, nil, err
	}

	footprint, err := totals.render()
	if err != nil {
		return nil, nil, err
	}
	footprint.Basis = strings.Join(segments, "; ")

	return &estimatemodelv1.PresetEstimateModel{
		CapacityFootprint: footprint,
		Exclusions:        exclusions,
		Notes:             strings.Join(notes, " "),
	}, nil, nil
}

// footprintTotals accumulates the five capacity buckets exactly: CPU in
// millicores, memory and storage in bytes -- big.Int arithmetic, so no
// quantity ever rounds.
type footprintTotals struct {
	cpuRequests    big.Int
	memoryRequests big.Int
	cpuLimits      big.Int
	memoryLimits   big.Int
	storage        big.Int
}

func (t *footprintTotals) empty() bool {
	return t.cpuRequests.Sign() == 0 && t.memoryRequests.Sign() == 0 &&
		t.cpuLimits.Sign() == 0 && t.memoryLimits.Sign() == 0 && t.storage.Sign() == 0
}

func (t *footprintTotals) render() (*costestimatev1.CapacityFootprint, error) {
	return &costestimatev1.CapacityFootprint{
		CpuRequests:       renderCpu(&t.cpuRequests),
		MemoryRequests:    renderBytes(&t.memoryRequests),
		CpuLimits:         renderCpu(&t.cpuLimits),
		MemoryLimits:      renderBytes(&t.memoryLimits),
		PersistentStorage: renderBytes(&t.storage),
	}, nil
}

// accumulateWorkload adds one workload's reservations into the totals and
// returns its basis segment ("" when the workload contributes nothing --
// zero instances, or neither resources nor volumes resolved).
func accumulateWorkload(specMsg protoreflect.Message, workload *capacityv1.WorkloadBinding, totals *footprintTotals) (string, error) {
	count, err := instanceCount(specMsg, workload)
	if err != nil {
		return "", err
	}
	if count.Sign() == 0 {
		return "", nil
	}

	resources, resourcesBasis, err := workloadResources(specMsg, workload, count, totals)
	if err != nil {
		return "", err
	}

	volumesBasis, contributed, err := workloadVolumes(specMsg, workload, count, totals)
	if err != nil {
		return "", err
	}

	if !resources && !contributed {
		return "", nil
	}

	noun := workload.GetLabel()
	if count.Cmp(big.NewInt(1)) != 0 {
		noun += "s"
	}
	segment := ""
	if resources {
		segment = fmt.Sprintf("%s %s x %s", count.String(), noun, resourcesBasis)
	}
	if contributed {
		if segment != "" {
			segment += " + " + volumesBasis
		} else {
			segment = fmt.Sprintf("%s %s: %s", count.String(), noun, volumesBasis)
		}
	}
	return segment, nil
}

// instanceCount resolves a workload's multiplier: a literal constant, or
// a numeric spec field with its declared default applying to absent AND
// explicitly-zero values (the FieldValue contract the cost derivations
// share).
func instanceCount(specMsg protoreflect.Message, workload *capacityv1.WorkloadBinding) (*big.Int, error) {
	switch count := workload.GetInstances().GetCount().(type) {
	case *capacityv1.InstanceCount_Constant:
		value, ok := new(big.Int).SetString(count.Constant, 10)
		if !ok {
			return nil, fmt.Errorf("workload %q: instance constant %q is not a whole number", workload.GetLabel(), count.Constant)
		}
		return value, nil
	case *capacityv1.InstanceCount_FieldValue:
		resolved, err := specpath.Resolve(specMsg, count.FieldValue.GetFieldPath())
		if err != nil {
			return nil, fmt.Errorf("workload %q instances: %w", workload.GetLabel(), err)
		}
		value := big.NewInt(0)
		if resolved.Value.IsValid() {
			switch resolved.Field.Kind() {
			case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
				protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
				value = big.NewInt(resolved.Value.Int())
			case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
				protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
				value = new(big.Int).SetUint64(resolved.Value.Uint())
			default:
				return nil, fmt.Errorf("workload %q instances: field %q is not an integer", workload.GetLabel(), count.FieldValue.GetFieldPath())
			}
		}
		if value.Sign() == 0 && count.FieldValue.GetDefaultWhenUnset() != "" {
			defaulted, ok := new(big.Int).SetString(count.FieldValue.GetDefaultWhenUnset(), 10)
			if !ok {
				return nil, fmt.Errorf("workload %q: instance default %q is not a whole number", workload.GetLabel(), count.FieldValue.GetDefaultWhenUnset())
			}
			value = defaulted
		}
		return value, nil
	default:
		return nil, fmt.Errorf("workload %q declares no instance count", workload.GetLabel())
	}
}

// workloadResources reads the workload's ContainerResources block, adds
// count x each quantity into the totals, and renders the per-instance
// basis fragment, e.g. "(1 CPU / 2Gi memory requests, 2 CPU / 4Gi
// limits)". When the manifest omits the block, the resolution falls back
// to the spec field's own (dev.planton.kubernetes.default_container_resources)
// annotation -- the defaults the modules apply at deploy time, so a
// manifest relying on them reserves exactly what the footprint states --
// with the basis naming the defaults' origin. A field carrying neither a
// manifest value nor the annotation contributes nothing.
func workloadResources(specMsg protoreflect.Message, workload *capacityv1.WorkloadBinding, count *big.Int, totals *footprintTotals) (bool, string, error) {
	path := workload.GetResourcesPath()
	if path == "" {
		return false, "", nil
	}
	resolved, err := specpath.Resolve(specMsg, path)
	if err != nil {
		return false, "", fmt.Errorf("workload %q resources: %w", workload.GetLabel(), err)
	}
	defaulted := false
	var resources protoreflect.Message
	if resolved.Present && resolved.Value.IsValid() {
		resources = resolved.Value.Message()
	} else {
		declared := declaredDefaultResources(resolved.Field)
		if declared == nil {
			return false, "", nil
		}
		resources = declared
		defaulted = true
	}

	requestsCpu, requestsMemory := cpuMemory(resources, "requests")
	limitsCpu, limitsMemory := cpuMemory(resources, "limits")

	contributed := false
	add := func(total *big.Int, quantity string, milli bool) error {
		if quantity == "" {
			return nil
		}
		value, err := parseQuantity(quantity, milli)
		if err != nil {
			return fmt.Errorf("workload %q resources: %w", workload.GetLabel(), err)
		}
		total.Add(total, value.Mul(value, count))
		contributed = true
		return nil
	}
	if err := add(&totals.cpuRequests, requestsCpu, true); err != nil {
		return false, "", err
	}
	if err := add(&totals.memoryRequests, requestsMemory, false); err != nil {
		return false, "", err
	}
	if err := add(&totals.cpuLimits, limitsCpu, true); err != nil {
		return false, "", err
	}
	if err := add(&totals.memoryLimits, limitsMemory, false); err != nil {
		return false, "", err
	}
	if !contributed {
		return false, "", nil
	}

	var parts []string
	if fragment := cpuMemoryFragment(requestsCpu, requestsMemory, "memory "); fragment != "" {
		parts = append(parts, fragment+"requests")
	}
	if fragment := cpuMemoryFragment(limitsCpu, limitsMemory, ""); fragment != "" {
		parts = append(parts, fragment+"limits")
	}
	basis := "(" + strings.Join(parts, ", ") + ")"
	if defaulted {
		basis += " from the spec-declared defaults (the modules apply them when the manifest omits resources)"
	}
	return true, basis, nil
}

// declaredDefaultResources reads the spec field's
// (dev.planton.kubernetes.default_container_resources) annotation -- the
// ContainerResources the modules apply when a manifest omits the field.
// Returns nil when the field carries no annotation.
func declaredDefaultResources(field protoreflect.FieldDescriptor) protoreflect.Message {
	options, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok || options == nil {
		return nil
	}
	if !proto.HasExtension(options, kubernetes.E_DefaultContainerResources) {
		return nil
	}
	declared, ok := proto.GetExtension(options, kubernetes.E_DefaultContainerResources).(*kubernetes.ContainerResources)
	if !ok || declared == nil {
		return nil
	}
	return declared.ProtoReflect()
}

// cpuMemory reads one CpuMemory block ("requests" or "limits") off a
// ContainerResources message, returning empty strings for absent halves.
func cpuMemory(resources protoreflect.Message, block string) (string, string) {
	field := resources.Descriptor().Fields().ByName(protoreflect.Name(block))
	if field == nil || !resources.Has(field) {
		return "", ""
	}
	pair := resources.Get(field).Message()
	read := func(name string) string {
		inner := pair.Descriptor().Fields().ByName(protoreflect.Name(name))
		if inner == nil {
			return ""
		}
		return pair.Get(inner).String()
	}
	return read("cpu"), read("memory")
}

// cpuMemoryFragment renders one block's per-instance basis half, e.g.
// "1 CPU / 2Gi memory " (the caller appends "requests"/"limits").
func cpuMemoryFragment(cpu, memory, memoryLabel string) string {
	var parts []string
	if cpu != "" {
		parts = append(parts, cpu+" CPU")
	}
	if memory != "" {
		parts = append(parts, memory+" "+memoryLabel)
	}
	if len(parts) == 0 {
		return ""
	}
	joined := strings.Join(parts, " / ")
	if !strings.HasSuffix(joined, " ") {
		joined += " "
	}
	return joined
}

// workloadVolumes adds count x each resolved volume size into the storage
// total and renders the volumes basis fragment, e.g.
// "3 x (100Gi data + 20Gi WAL) volumes". A size the manifest omits falls
// back to the field's own (dev.planton.shared.options.default) annotation
// -- the size the modules apply at deploy time -- but ONLY when the
// volume's enclosing configuration block is present: the block is the
// volume's existence switch, and a default applied to an unconfigured
// volume would fabricate a reservation for storage that never binds.
func workloadVolumes(specMsg protoreflect.Message, workload *capacityv1.WorkloadBinding, count *big.Int, totals *footprintTotals) (string, bool, error) {
	type resolvedVolume struct{ size, label string }
	var volumes []resolvedVolume
	for _, volume := range workload.GetVolumes() {
		// applies_when gates the volume's EXISTENCE: a mode that
		// provisions no volume (ephemeral on emptyDir) keeps its size
		// value and its spec default while binding nothing -- a size
		// without a volume must contribute nothing.
		exists, err := costestimator.ConditionsHold(specMsg, volume.GetAppliesWhen())
		if err != nil {
			return "", false, fmt.Errorf("workload %q volume %q: %w", workload.GetLabel(), volume.GetLabel(), err)
		}
		if !exists {
			continue
		}
		resolved, err := specpath.Resolve(specMsg, volume.GetSizePath())
		if err != nil {
			return "", false, fmt.Errorf("workload %q volume %q: %w", workload.GetLabel(), volume.GetLabel(), err)
		}
		size := ""
		defaulted := false
		if resolved.Present && resolved.Value.IsValid() {
			size = resolved.Value.String()
		}
		if size == "" || isPlaceholder(size) {
			enclosing, err := enclosingBlockPresent(specMsg, volume.GetSizePath())
			if err != nil {
				return "", false, fmt.Errorf("workload %q volume %q: %w", workload.GetLabel(), volume.GetLabel(), err)
			}
			declared := declaredDefaultString(resolved.Field)
			if !enclosing || declared == "" {
				continue
			}
			size = declared
			defaulted = true
		}
		bytes, err := parseQuantity(size, false)
		if err != nil {
			return "", false, fmt.Errorf("workload %q volume %q: %w", workload.GetLabel(), volume.GetLabel(), err)
		}
		totals.storage.Add(&totals.storage, bytes.Mul(bytes, count))
		label := volume.GetLabel()
		if defaulted {
			label += " (the spec-declared default size)"
		}
		volumes = append(volumes, resolvedVolume{size: size, label: label})
	}
	if len(volumes) == 0 {
		return "", false, nil
	}

	if len(volumes) == 1 {
		noun := "volume"
		if count.Cmp(big.NewInt(1)) != 0 {
			noun = "volumes"
		}
		return fmt.Sprintf("%s x %s %s %s", count.String(), volumes[0].size, volumes[0].label, noun), true, nil
	}
	var terms []string
	for _, volume := range volumes {
		terms = append(terms, volume.size+" "+volume.label)
	}
	return fmt.Sprintf("%s x (%s) volumes", count.String(), strings.Join(terms, " + ")), true, nil
}

// declaredDefaultString reads the spec field's
// (dev.planton.shared.options.default) annotation -- the value the
// modules apply when a manifest omits the field. Empty when the field
// carries no annotation.
func declaredDefaultString(field protoreflect.FieldDescriptor) string {
	options, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok || options == nil {
		return ""
	}
	if !proto.HasExtension(options, sharedoptions.E_Default) {
		return ""
	}
	declared, _ := proto.GetExtension(options, sharedoptions.E_Default).(string)
	return declared
}

// CheckDeclaredDefaults validates a bound field's default annotations at
// gate time, so a malformed spec annotation fails CI naming the file
// instead of erroring at replay: a ContainerResources field's
// (default_container_resources) quantities must parse, and a size
// field's (options.default) must parse as a Kubernetes quantity.
// Exported for the capacity-derivation conformance gate -- the
// annotation semantics have exactly one home, this package.
func CheckDeclaredDefaults(field protoreflect.FieldDescriptor) error {
	if declared := declaredDefaultResources(field); declared != nil {
		for _, block := range []string{"requests", "limits"} {
			cpu, memory := cpuMemory(declared, block)
			if cpu != "" {
				if _, err := parseQuantity(cpu, true); err != nil {
					return fmt.Errorf("default_container_resources %s cpu: %w", block, err)
				}
			}
			if memory != "" {
				if _, err := parseQuantity(memory, false); err != nil {
					return fmt.Errorf("default_container_resources %s memory: %w", block, err)
				}
			}
		}
	}
	if field.Kind() == protoreflect.StringKind && !field.IsList() && !field.IsMap() {
		if declared := declaredDefaultString(field); declared != "" {
			if _, err := parseQuantity(declared, false); err != nil {
				return fmt.Errorf("(dev.planton.shared.options.default) %q: %w", declared, err)
			}
		}
	}
	return nil
}

// enclosingBlockPresent reports whether every intermediate message on a
// path is present on the manifest -- for a root-level field there is no
// enclosing block and the answer is true. The size-default fallback
// gates on this: the enclosing block is the volume's existence switch.
func enclosingBlockPresent(specMsg protoreflect.Message, dotPath string) (bool, error) {
	lastDot := strings.LastIndex(dotPath, ".")
	if lastDot < 0 {
		return true, nil
	}
	parent, err := specpath.Resolve(specMsg, dotPath[:lastDot])
	if err != nil {
		return false, err
	}
	return parent.Present && parent.Value.IsValid(), nil
}

// isPlaceholder recognizes the catalog's template placeholders ("<size>")
// so they never masquerade as real quantities -- the same rule the cost
// estimator applies.
func isPlaceholder(text string) bool {
	return len(text) > 2 && strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">")
}
