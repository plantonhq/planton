package catalogschema

import (
	"fmt"
	"strings"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// extractFieldSemantics extracts structured semantic annotations and reconstructed
// rawOptions text from a field's resolved options via type-safe proto.GetExtension().
func extractFieldSemantics(fd protoreflect.FieldDescriptor) (*ProtoFieldSemantic, string) {
	opts, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil {
		return nil, ""
	}

	sem := &ProtoFieldSemantic{}
	var parts []string
	hasAnySemantic := false

	// 1. buf.validate.field -- required flag and validation constraints
	if proto.HasExtension(opts, validatepb.E_Field) {
		rules := proto.GetExtension(opts, validatepb.E_Field).(*validatepb.FieldRules)
		if rules != nil {
			valParts := reconstructValidateOptions(rules)
			parts = append(parts, valParts...)
			if rules.GetRequired() {
				sem.IsRequired = true
				hasAnySemantic = true
			}
		}
	}

	// 2. dev.planton.shared.options.default
	if proto.HasExtension(opts, options.E_Default) {
		val := proto.GetExtension(opts, options.E_Default).(string)
		sem.DefaultValue = &val
		hasAnySemantic = true
		parts = append(parts, fmt.Sprintf(`(dev.planton.shared.options.default) = "%s"`, val))
	}

	// 3. dev.planton.shared.options.recommended_default
	if proto.HasExtension(opts, options.E_RecommendedDefault) {
		val := proto.GetExtension(opts, options.E_RecommendedDefault).(string)
		if sem.DefaultValue == nil {
			sem.DefaultValue = &val
			hasAnySemantic = true
		}
		parts = append(parts, fmt.Sprintf(`(dev.planton.shared.options.recommended_default) = "%s"`, val))
	}

	// 4. dev.planton.shared.foreignkey.v1.default_kind
	if proto.HasExtension(opts, foreignkeyv1.E_DefaultKind) {
		kind := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(catalogkind.CatalogKind)
		kindStr := kind.String()
		sem.ForeignKeyKind = &kindStr
		hasAnySemantic = true
		parts = append(parts, fmt.Sprintf("(dev.planton.shared.foreignkey.v1.default_kind) = %s", kindStr))
	}

	// 5. dev.planton.shared.foreignkey.v1.default_kind_field_path
	if proto.HasExtension(opts, foreignkeyv1.E_DefaultKindFieldPath) {
		path := proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
		parts = append(parts, fmt.Sprintf(`(dev.planton.shared.foreignkey.v1.default_kind_field_path) = "%s"`, path))
	}

	// 6. dev.planton.shared.foreignkey.v1.candidate (repeated)
	candidates, _ := proto.GetExtension(opts, foreignkeyv1.E_Candidate).([]*foreignkeyv1.ReferenceCandidate)
	for _, c := range candidates {
		parts = append(parts, fmt.Sprintf(`(dev.planton.shared.foreignkey.v1.candidate) = {kind: %s, field_path: "%s"}`,
			c.GetKind().String(), c.GetFieldPath()))
	}
	if len(candidates) > 0 {
		for _, key := range refannotations.Of(fd).Keys() {
			sem.ForeignKeyCandidates = append(sem.ForeignKeyCandidates,
				ForeignKeyCandidate{Kind: key.Kind.String(), FieldPath: key.FieldPath})
		}
		hasAnySemantic = true
	}

	rawOpts := strings.Join(parts, ", ")

	if !hasAnySemantic {
		if rawOpts == "" {
			return nil, ""
		}
		return nil, rawOpts
	}

	return sem, rawOpts
}

// reconstructValidateOptions builds rawOptions fragments from buf.validate.FieldRules.
// Deliberately partial (a declared fidelity boundary of the contract): the common
// rules render structurally; everything else is readable in rawContent.
func reconstructValidateOptions(rules *validatepb.FieldRules) []string {
	var parts []string

	if rules.GetRequired() {
		parts = append(parts, "(buf.validate.field).required = true")
	}

	switch rule := rules.GetType().(type) {
	case *validatepb.FieldRules_String_:
		if sr := rule.String_; sr != nil {
			if sr.GetMinLen() > 0 {
				parts = append(parts, fmt.Sprintf("(buf.validate.field).string.min_len = %d", sr.GetMinLen()))
			}
			if sr.GetMaxLen() > 0 {
				parts = append(parts, fmt.Sprintf("(buf.validate.field).string.max_len = %d", sr.GetMaxLen()))
			}
		}
	case *validatepb.FieldRules_Int32:
		if ir := rule.Int32; ir != nil {
			if ir.HasGt() {
				parts = append(parts, fmt.Sprintf("(buf.validate.field).int32.gt = %d", ir.GetGt()))
			}
		}
	case *validatepb.FieldRules_Int64:
		if ir := rule.Int64; ir != nil {
			if ir.HasGt() {
				parts = append(parts, fmt.Sprintf("(buf.validate.field).int64.gt = %d", ir.GetGt()))
			}
		}
	}

	if r := rules.GetRepeated(); r != nil {
		if r.GetMinItems() > 0 {
			parts = append(parts, fmt.Sprintf("(buf.validate.field).repeated.min_items = %d", r.GetMinItems()))
		}
	}

	return parts
}
