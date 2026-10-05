package capacityderivation

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	capacityv1 "github.com/plantonhq/planton/finops/catalogkindcapacityderivation/v1"
	derivationv1 "github.com/plantonhq/planton/finops/catalogkindcostderivation/v1"
	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/finops/capacityestimator"
	"github.com/plantonhq/planton/pkg/finops/costderivation"
	"github.com/plantonhq/planton/pkg/finops/costprofile"
	"github.com/plantonhq/planton/pkg/finops/estimatemodel"
	"github.com/plantonhq/planton/pkg/specpath"
)

// containerResourcesMessage is the shared requests/limits message every
// resources_path must terminate at -- capacity is read by TYPE, never by
// guessed field names.
const containerResourcesMessage = "dev.planton.kubernetes.ContainerResources"

var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// TestCapacityDerivationConformance holds every capacity derivation to
// its contract, offline:
//
//  1. The document parses strictly, names its kind (metadata.name
//     equals the filename), and the kind ships a cost profile whose
//     billing model IS cluster_capacity -- priced kinds derive
//     dollars through the cost-derivation standard, never capacity.
//  2. The kind ships NEITHER a hand-authored estimate model NOR a
//     cost derivation beside this document -- a kind's quantities
//     have exactly one home.
//  3. Every workload binds real spec structure: resources_path resolves
//     to the shared ContainerResources message, the instance count is a
//     literal decimal or a numeric spec field, every volume's size_path
//     resolves to a string scalar (the Kubernetes quantity shape), and
//     labels are non-empty and unique (they carry the basis prose).
//  4. Conditions on exclusions and notes name resolvable paths, carry an
//     op, the comparand exactly when the op compares, and never compare
//     a value-or-reference wrapper (a referenced value is unknowable --
//     the same rule the cost-derivation gate enforces).
//
// Whether the rules reproduce the hand-verified footprints is the
// estimate generator's replay check -- this gate keeps each derivation
// internally sound.
func TestCapacityDerivationConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	kindDirs, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering capacity derivations: %v", err)
	}
	if len(kindDirs) == 0 {
		t.Skip("no capacity derivations authored yet")
	}

	for _, kindDir := range kindDirs {
		kindDir := kindDir
		t.Run(kindDir, func(t *testing.T) {
			derivation, err := Load(root, kindDir)
			if err != nil {
				t.Fatalf("capacity derivation: %v", err)
			}
			if derivation.GetKind() != "CatalogKindCapacityDerivation" {
				t.Fatalf("kind is %q, want CatalogKindCapacityDerivation", derivation.GetKind())
			}
			if derivation.GetMetadata().GetName() != kindDir {
				t.Errorf("metadata.name is %q, want %q (the filename is the kind's identity)",
					derivation.GetMetadata().GetName(), kindDir)
			}

			if _, err := os.Stat(estimatemodel.Path(root, kindDir)); err == nil {
				t.Errorf("kind also ships an estimate model (%s) -- quantities have exactly one home; delete the model",
					estimatemodel.Path(root, kindDir))
			}
			if _, err := os.Stat(costderivation.Path(root, kindDir)); err == nil {
				t.Errorf("kind also ships a cost derivation (%s) -- a cluster-capacity kind prices no meters",
					costderivation.Path(root, kindDir))
			}

			provider := kindProvider(t, root, kindDir)
			profile, err := costprofile.Load(root, provider, kindDir)
			if err != nil {
				t.Fatalf("the kind must ship a cost profile: %v", err)
			}
			if profile.GetSpec().GetBillingModel() != costprofilev1.BillingModel_cluster_capacity {
				t.Fatalf("billing model is %s -- capacity derivations exist exactly for cluster_capacity kinds; priced kinds derive dollars",
					profile.GetSpec().GetBillingModel())
			}

			specDescriptor := kindSpecDescriptor(t, kindDir)
			spec := derivation.GetSpec()

			if len(spec.GetWorkloads()) == 0 {
				t.Error("derivation declares no workloads -- what capacity does it derive?")
			}
			labels := map[string]bool{}
			for i, workload := range spec.GetWorkloads() {
				checkWorkload(t, specDescriptor, i, workload)
				label := strings.TrimSpace(workload.GetLabel())
				if labels[label] {
					t.Errorf("workload %d reuses label %q -- labels carry the basis prose and must be unique", i, label)
				}
				labels[label] = true
			}

			for i, text := range spec.GetExclusions() {
				if strings.TrimSpace(text.GetText()) == "" {
					t.Errorf("exclusion %d has no text", i)
				}
				checkConditions(t, specDescriptor, text.GetAppliesWhen())
			}
			for i, note := range spec.GetNotes() {
				if strings.TrimSpace(note.GetText()) == "" {
					t.Errorf("note %d has no text", i)
				}
				checkConditions(t, specDescriptor, note.GetAppliesWhen())
			}
		})
	}
}

// checkWorkload verifies one workload binding's internal soundness.
func checkWorkload(t *testing.T, specDescriptor protoreflect.MessageDescriptor, index int, workload *capacityv1.WorkloadBinding) {
	t.Helper()
	if strings.TrimSpace(workload.GetLabel()) == "" {
		t.Errorf("workload %d has no label -- the basis prose needs its noun", index)
	}

	if path := workload.GetResourcesPath(); path != "" {
		terminal, err := specpath.ResolvableTerminal(specDescriptor, path)
		if err != nil {
			t.Errorf("workload %d resources_path: %v", index, err)
		} else if terminal.Kind() != protoreflect.MessageKind ||
			string(terminal.Message().FullName()) != containerResourcesMessage {
			t.Errorf("workload %d resources_path %q terminates at %s, want the shared %s -- capacity is read by type",
				index, path, describeTerminal(terminal), containerResourcesMessage)
		} else if err := capacityestimator.CheckDeclaredDefaults(terminal); err != nil {
			t.Errorf("workload %d resources_path %q: the spec's own default annotation is malformed -- %v", index, path, err)
		}
	} else if len(workload.GetVolumes()) == 0 {
		t.Errorf("workload %d binds neither a resources block nor volumes -- it reserves nothing", index)
	}

	switch count := workload.GetInstances().GetCount().(type) {
	case *capacityv1.InstanceCount_Constant:
		if !decimalPattern.MatchString(count.Constant) {
			t.Errorf("workload %d instance constant %q is not a plain decimal string", index, count.Constant)
		}
	case *capacityv1.InstanceCount_FieldValue:
		checkNumericField(t, specDescriptor, index, count.FieldValue)
	default:
		t.Errorf("workload %d declares no instance count -- every reservation multiplies by one", index)
	}

	for j, volume := range workload.GetVolumes() {
		if strings.TrimSpace(volume.GetLabel()) == "" {
			t.Errorf("workload %d volume %d has no label", index, j)
		}
		terminal, err := specpath.ResolvableTerminal(specDescriptor, volume.GetSizePath())
		if err != nil {
			t.Errorf("workload %d volume %d size_path: %v", index, j, err)
		} else if terminal.Kind() != protoreflect.StringKind || terminal.IsList() || terminal.IsMap() {
			t.Errorf("workload %d volume %d size_path %q is not a string scalar -- volume sizes are Kubernetes quantity strings",
				index, j, volume.GetSizePath())
		} else if err := capacityestimator.CheckDeclaredDefaults(terminal); err != nil {
			t.Errorf("workload %d volume %d size_path %q: the spec's own default annotation is malformed -- %v", index, j, volume.GetSizePath(), err)
		}
		checkConditions(t, specDescriptor, volume.GetAppliesWhen())
	}
}

// checkNumericField verifies an instance-count field binding: a resolvable
// numeric path and a decimal default.
func checkNumericField(t *testing.T, specDescriptor protoreflect.MessageDescriptor, index int, field *derivationv1.FieldValue) {
	t.Helper()
	terminal, err := specpath.ResolvableTerminal(specDescriptor, field.GetFieldPath())
	if err != nil {
		t.Errorf("workload %d instance field: %v", index, err)
		return
	}
	switch terminal.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		// Numeric -- the count shape.
	default:
		t.Errorf("workload %d instance field %q is %s, want an integer field -- instance counts are whole numbers",
			index, field.GetFieldPath(), terminal.Kind())
	}
	if terminal.IsList() || terminal.IsMap() {
		t.Errorf("workload %d instance field %q is repeated -- which element would the count be?", index, field.GetFieldPath())
	}
	if d := field.GetDefaultWhenUnset(); d != "" && !decimalPattern.MatchString(d) {
		t.Errorf("workload %d instance default %q is not a plain decimal string", index, d)
	}
}

// checkConditions mirrors the cost-derivation gate's condition rules: a
// resolvable path, an op, the comparand exactly when the op compares, and
// never a comparison on a value-or-reference wrapper (the engine compares
// only the literal arm; a referenced value is unknowable).
func checkConditions(t *testing.T, specDescriptor protoreflect.MessageDescriptor, conditions []*derivationv1.Condition) {
	t.Helper()
	for _, condition := range conditions {
		terminal, err := specpath.ResolvableTerminal(specDescriptor, condition.GetFieldPath())
		if err != nil {
			t.Errorf("condition: %v", err)
			continue
		}
		switch condition.GetOp() {
		case derivationv1.Condition_equals, derivationv1.Condition_not_equals:
			if condition.GetValue() == "" {
				t.Errorf("condition on %q compares against an empty value -- use is_unset for absence", condition.GetFieldPath())
			}
			if isReferenceCapable(terminal) {
				t.Errorf("condition on %q compares a value-or-reference field -- a referenced value is unknowable at estimate time; restructure the rule or use a presence op", condition.GetFieldPath())
			}
		case derivationv1.Condition_is_set, derivationv1.Condition_is_unset:
			if condition.GetValue() != "" {
				t.Errorf("condition on %q is a presence check but carries value %q", condition.GetFieldPath(), condition.GetValue())
			}
		default:
			t.Errorf("condition on %q has no op", condition.GetFieldPath())
		}
	}
}

// isReferenceCapable reports whether a field is the value-or-reference
// wrapper shape: a non-repeated message carrying a scalar `value` field
// beside at least one reference arm.
func isReferenceCapable(field protoreflect.FieldDescriptor) bool {
	if field.IsList() || field.IsMap() || field.Kind() != protoreflect.MessageKind {
		return false
	}
	value := field.Message().Fields().ByName("value")
	if value == nil || value.Kind() == protoreflect.MessageKind || value.IsList() || value.IsMap() {
		return false
	}
	return field.Message().Fields().Len() > 1
}

// describeTerminal names a terminal field's shape for error messages.
func describeTerminal(field protoreflect.FieldDescriptor) string {
	if field.Kind() == protoreflect.MessageKind {
		return string(field.Message().FullName())
	}
	return field.Kind().String()
}

// kindProvider locates the provider directory a kind lives under.
func kindProvider(t *testing.T, repoRoot, kindDir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", kindDir))
	if err != nil || len(matches) == 0 {
		t.Fatalf("capacity derivation names kind %q, which exists nowhere under catalog/", kindDir)
	}
	if len(matches) > 1 {
		t.Fatalf("kind %q exists under multiple providers: %v", kindDir, matches)
	}
	return filepath.Base(filepath.Dir(matches[0]))
}

// kindSpecDescriptor resolves a kind directory name to its kind's spec
// message descriptor via the kind registry.
func kindSpecDescriptor(t *testing.T, kindDir string) protoreflect.MessageDescriptor {
	t.Helper()
	kind := catalogkindreflect.KindFromString(kindDir)
	apiMessage, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("NewInstance(%s): %v", kindDir, err)
	}
	specField := apiMessage.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind {
		t.Fatalf("%s has no spec message field", apiMessage.ProtoReflect().Descriptor().FullName())
	}
	return specField.Message()
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}
