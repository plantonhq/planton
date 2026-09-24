// containment_decisions_test.go — The containment-decision registry gate.
//
// WHY THIS EXISTS:
// A kind marked `container_kind: true` (see CloudResourceKindMeta.container_kind)
// is drawn as a BOUNDARY on architecture diagrams, and every resource that
// references it nests visually inside it — unless the reference field carries
// `(dev.planton.shared.foreignkey.v1.containment_exempt) = true`, which declares
// the reference expresses ACCESS ("admit traffic from / let me talk to") rather
// than PLACEMENT ("deploy me into this"). Getting this wrong draws a FALSE
// diagram: a storage account nested inside a subnet it merely admits traffic
// from, a web app nested inside the storage account it mounts.
//
// This test walks every compiled-in provider spec via proto reflection, finds
// every reference field whose default_kind targets a container kind, and diffs
// the resulting verdict list (contained vs. exempt) against the committed
// golden file. Any NEW reference into a container kind — a new field, a new
// kind, or a newly marked container — changes the list and FAILS this test
// until a human-legible verdict is recorded by regenerating the golden file:
//
//	go test ./shared/cloudresourcekind/... -run TestContainmentDecisions -update
//
// Review the diff before committing: every added "contained" line is a claim
// that the referencing resource physically lives inside the referenced
// container. If the reference is access-style, author `containment_exempt` on
// the field instead and regenerate.
//
// WHY AN EXTERNAL TEST PACKAGE (`package cloudresourcekind_test`):
// The test imports pkg/crkreflect (to register every kind's proto files in the
// global registry), and crkreflect imports cloudresourcekind — the external
// test package breaks the cycle, same convention as
// shared/foreignkey/v1/foreign_key_consumer_test.go.
package cloudresourcekind_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/plantonhq/planton/shared/options"

	// Registers every cloud-resource kind's proto files in the global
	// registry so the reflection walk below sees the full compiled-in surface.
	_ "github.com/plantonhq/planton/pkg/crkreflect"
)

var update = flag.Bool("update", false, "regenerate testdata/containment_decisions.txt from the compiled-in protos")

const goldenPath = "testdata/containment_decisions.txt"

// stringValueOrRefFQN / valueFromRefFQN identify the only field types on which
// reference annotations are meaningful. Annotations authored anywhere else
// (e.g. on a plain string, or on a wrapper's parent message field) are never
// read by the platform's edge selector and would be silently dead.
var (
	stringValueOrRefFQN = protoreflect.FullName("dev.planton.shared.foreignkey.v1.StringValueOrRef")
	valueFromRefFQN     = protoreflect.FullName("dev.planton.shared.foreignkey.v1.ValueFromRef")
)

// TestExtensionNumbersArePinned guards the wire-level contract: downstream
// platforms read these extensions by number from descriptors, so the numbers
// are load-bearing API surface.
func TestExtensionNumbersArePinned(t *testing.T) {
	if n := options.E_DiagramLabel.TypeDescriptor().Number(); n != 60006 {
		t.Fatalf("diagram_label extension number changed: got %d, want 60006", n)
	}
	if n := foreignkeyv1.E_ContainmentExempt.TypeDescriptor().Number(); n != 200003 {
		t.Fatalf("containment_exempt extension number changed: got %d, want 200003", n)
	}
}

// TestReferenceAnnotationsSitOnReferenceFields guards against authoring the
// options at the wrong nesting level. The platform's edge selector resolves
// annotations from the StringValueOrRef/ValueFromRef-typed consumer field; an
// annotation on any other field type is unreachable and therefore a bug at
// authoring time, not a style choice.
func TestReferenceAnnotationsSitOnReferenceFields(t *testing.T) {
	var violations []string
	walkProviderFields(func(fd protoreflect.FieldDescriptor, opts *descriptorOptions) {
		if !opts.exempt && opts.label == "" {
			return
		}
		if !isReferenceField(fd) {
			violations = append(violations, string(fd.FullName()))
		}
	})
	if len(violations) > 0 {
		t.Fatalf("containment_exempt/diagram_label authored on non-reference fields (the platform never reads them there — move the annotation onto the StringValueOrRef/ValueFromRef field):\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// TestContainmentExemptTargetsContainerKinds: an exemption on a reference whose
// default_kind is a NON-container kind is inert and can only mislead the next
// author — reject it.
//
// A reference with NO default_kind (a kind-less reference: an event target, a
// forwarding address, a monitored resource — fields whose kind varies per
// manifest and is stated in valueFrom) is a different case. The platform
// resolves containment by the kind of the node the manifest actually names,
// so a kind-less reference into a container kind nests exactly like a typed
// one — and the exemption is the only way its author can say "access, not
// placement". Kind-less exemptions are therefore legal here and are listed in
// the registry under their own form (`-> *`).
func TestContainmentExemptTargetsContainerKinds(t *testing.T) {
	var violations []string
	walkProviderFields(func(fd protoreflect.FieldDescriptor, opts *descriptorOptions) {
		if !opts.exempt || opts.kind == cloudresourcekind.CloudResourceKind_unspecified {
			return
		}
		if !isContainerKind(opts.kind) {
			violations = append(violations, fmt.Sprintf("%s (default_kind=%s)", fd.FullName(), opts.kind))
		}
	})
	if len(violations) > 0 {
		t.Fatalf("containment_exempt authored on references to non-container kinds (inert — remove it or mark the kind as a container):\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// TestContainmentDecisions is the registry gate described in the file header.
func TestContainmentDecisions(t *testing.T) {
	var lines []string
	walkProviderFields(func(fd protoreflect.FieldDescriptor, opts *descriptorOptions) {
		if opts.kind == cloudresourcekind.CloudResourceKind_unspecified {
			// A kind-less reference is listed only when its author declared it
			// access: without the exemption it may or may not reach a container
			// (that depends on the manifest), and the registry records
			// authored verdicts, never possibilities.
			if opts.exempt && isReferenceField(fd) {
				lines = append(lines, fmt.Sprintf("exempt    %s -> *", fd.FullName()))
			}
			return
		}
		if !isContainerKind(opts.kind) {
			return
		}
		verdict := "contained"
		if opts.exempt {
			verdict = "exempt   "
		}
		lines = append(lines, fmt.Sprintf("%s %s -> %s", verdict, fd.FullName(), opts.kind))
	})
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(header+got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("regenerated %s with %d decisions", goldenPath, len(lines))
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden file missing — run with -update to create it: %v", err)
	}
	want := stripHeader(string(wantBytes))
	if got == want {
		return
	}
	t.Fatalf("containment decisions changed.\n%s\nEvery reference into a container kind needs a conscious verdict:\n"+
		"  contained — the referencing resource physically lives inside the referenced container\n"+
		"  exempt    — the reference is access-style; author (dev.planton.shared.foreignkey.v1.containment_exempt) = true on the field\n"+
		"  exempt … -> * — a kind-less reference (no default_kind) whose author declared it access, whatever kind a manifest names\n"+
		"Review the diff below, fix annotations if needed, then regenerate the golden file with -update.\n\n%s",
		goldenPath, diffLines(want, got))
}

const header = `# Containment-decision registry — generated by containment_decisions_test.go (-update); do not edit by hand.
#
# Every line is a reference field whose default_kind targets a container kind:
#   contained — the referencing resource nests inside the referenced container on diagrams
#   exempt    — the reference is access-style (containment_exempt authored on the field); no nesting
# A line ending in "-> *" is a kind-less reference (no default_kind; the kind is stated per manifest in valueFrom)
# whose author declared it access: it never nests, whichever container kind a manifest names.
#
`

func stripHeader(s string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(s, "\n") {
		if strings.HasPrefix(line, "#") || line == "\n" && b.Len() == 0 {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

// diffLines renders a minimal set diff (ordering-insensitive content diff is
// enough here — both sides are sorted).
func diffLines(want, got string) string {
	wantSet := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(want), "\n") {
		wantSet[l] = true
	}
	gotSet := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(got), "\n") {
		gotSet[l] = true
	}
	var b strings.Builder
	for l := range gotSet {
		if !wantSet[l] {
			b.WriteString("+ " + l + "\n")
		}
	}
	for l := range wantSet {
		if !gotSet[l] {
			b.WriteString("- " + l + "\n")
		}
	}
	return b.String()
}

// descriptorOptions carries the resolved reference annotations for one field.
type descriptorOptions struct {
	kind   cloudresourcekind.CloudResourceKind
	exempt bool
	label  string
}

// walkProviderFields visits every field of every message (recursively,
// including nested messages) declared in catalog/ proto files.
func walkProviderFields(visit func(fd protoreflect.FieldDescriptor, opts *descriptorOptions)) {
	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(file.Path(), "catalog/") {
			return true
		}
		msgs := file.Messages()
		for i := 0; i < msgs.Len(); i++ {
			walkMessage(msgs.Get(i), visit)
		}
		return true
	})
}

func walkMessage(md protoreflect.MessageDescriptor, visit func(fd protoreflect.FieldDescriptor, opts *descriptorOptions)) {
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		visit(fd, fieldOptions(fd))
	}
	nested := md.Messages()
	for i := 0; i < nested.Len(); i++ {
		if nested.Get(i).IsMapEntry() {
			continue
		}
		walkMessage(nested.Get(i), visit)
	}
}

func fieldOptions(fd protoreflect.FieldDescriptor) *descriptorOptions {
	out := &descriptorOptions{kind: cloudresourcekind.CloudResourceKind_unspecified}
	opts, ok := fd.Options().(interface{ ProtoReflect() protoreflect.Message })
	if !ok || opts == nil {
		return out
	}
	msg, ok := opts.(proto.Message)
	if !ok {
		return out
	}
	if k, ok := proto.GetExtension(msg, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind); ok {
		out.kind = k
	}
	if e, ok := proto.GetExtension(msg, foreignkeyv1.E_ContainmentExempt).(bool); ok {
		out.exempt = e
	}
	if l, ok := proto.GetExtension(msg, options.E_DiagramLabel).(string); ok {
		out.label = l
	}
	return out
}

func isReferenceField(fd protoreflect.FieldDescriptor) bool {
	if fd.Kind() != protoreflect.MessageKind {
		return false
	}
	name := fd.Message().FullName()
	return name == stringValueOrRefFQN || name == valueFromRefFQN
}

func isContainerKind(kind cloudresourcekind.CloudResourceKind) bool {
	vd := kind.Descriptor().Values().ByNumber(kind.Number())
	if vd == nil {
		return false
	}
	meta, ok := proto.GetExtension(vd.Options(), cloudresourcekind.E_KindMeta).(*cloudresourcekind.CloudResourceKindMeta)
	if !ok || meta == nil {
		return false
	}
	return meta.GetContainerKind()
}
