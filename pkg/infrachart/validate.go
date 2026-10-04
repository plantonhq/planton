package infrachart

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/deferrules"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
	yamlv3 "gopkg.in/yaml.v3"
)

// refTarget keys a chart's resources by (kind, raw name) for the
// cross-document reference pass. Chart validation deliberately keys on the
// AUTHORED name, not the derived identity slug: within one chart the author
// writes both sides verbatim, and reporting in their own spelling is the
// honest granularity. The deployment graph's slug-derived identity lives in
// pkg/manifestgraph, which owns the walker and rules this pass consumes.
type refTarget struct {
	kind catalogkind.CatalogKind
	name string
}

// Severity of a validation issue.
type Severity string

const (
	// SeverityError fails the gate.
	SeverityError Severity = "error"
	// SeverityWarning is surfaced but does not fail the gate.
	SeverityWarning Severity = "warning"
)

// Issue is one problem found while validating a chart, attributed to the
// template file it came from (the honest granularity — rendering merges
// templates, so line numbers do not survive).
type Issue struct {
	Severity     Severity
	File         string
	ResourceKind string
	ResourceName string
	Message      string
}

// Doc is one rendered catalog object.
type Doc struct {
	File string
	Kind string
	Name string
	YAML []byte
	Msg  proto.Message

	refUses []manifestgraph.RefUse
}

// VariantResult is the outcome of validating one render variant of a chart.
type VariantResult struct {
	// Name identifies the variant: "defaults", or "<param>=<value>" for a
	// flipped bool toggle.
	Name string
	// Docs are the successfully loaded manifest documents the variant rendered.
	Docs []Doc
	// Issues are the variant's failures and warnings.
	Issues []Issue

	checks        []DocumentCheck
	deferredToken func(string) bool
}

// Report is the outcome of validating one chart across all its render
// variants: the defaults (plus any --set overrides), then each bool param
// flipped once so every conditional manifest is exercised in both branches —
// one variant per toggle keeps the run linear in the number of toggles while
// still rendering every conditional block at least once in each state.
type Report struct {
	ChartName string
	ChartDir  string
	Variants  []VariantResult
}

// HasErrors reports whether any variant carries an error-severity issue.
func (r *Report) HasErrors() bool {
	for _, v := range r.Variants {
		for _, issue := range v.Issues {
			if issue.Severity == SeverityError {
				return true
			}
		}
	}
	return false
}

// Options configures a validation run.
type Options struct {
	// Org and Env are bound as the reserved org/env template variables (the
	// platform injects the real ones at render time). Sensible synthetic
	// defaults are applied when empty.
	Org string
	Env string

	// Set overrides param values by name (and may override org/env), letting
	// the gate exercise specific parameter combinations beyond the automatic
	// per-toggle flips.
	Set map[string]string

	// DocumentChecks are a host's own rules over each rendered document, run
	// only once the document's schema passed -- the order the platform applies
	// its own checks in. Each finding is an error on that document. The
	// open-source CLI has none; the Planton Platform CLI adds where secrets
	// may go, a rule the backendless deploy (which takes literals, never
	// $secret/ references) must not apply.
	DocumentChecks []DocumentCheck

	// IsDeferredToken names the values a host resolves before anything
	// deploys -- the Planton Platform CLI passes its `$secret/` and `$var/`
	// reference grammar. A schema rule located on such a spec value (a base64
	// pattern, a CIDR format) is set aside rather than failed, because the
	// value it is written about exists only once the token resolves; the host
	// applies it to the resolved value there. The open-source CLI, whose
	// deploy takes literals, passes none, so every rule judges what is
	// written.
	IsDeferredToken func(value string) bool
}

// DocumentCheck is one host rule over a rendered, schema-valid document; it
// returns its findings as plain-language sentences, none when the document
// passes.
type DocumentCheck func(doc proto.Message) []string

// Validate loads the chart at dir and validates it offline across render
// variants: the defaults variant (declared values plus any Set overrides),
// then one variant per bool param with that toggle flipped. Every rendered
// document and reference is checked per variant; reference-target severity is
// variant-aware (see checkReferences). It returns an error only when the
// chart cannot be processed at all; rendering and validation problems are
// returned as Issues in the Report.
func Validate(dir string, opts Options) (*Report, error) {
	chart, err := LoadDir(dir)
	if err != nil {
		return nil, err
	}

	org := opts.Org
	if org == "" {
		org = "acme"
	}
	env := opts.Env
	if env == "" {
		env = "dev"
	}

	report := &Report{ChartName: chart.Name, ChartDir: dir}

	for _, variant := range planVariants(chart.Params, opts.Set) {
		ctx, placeholders, err := buildContext(chart.Params, org, env, variant.set)
		if err != nil {
			return nil, err
		}
		result := VariantResult{Name: variant.name, checks: opts.DocumentChecks, deferredToken: opts.IsDeferredToken}
		for _, tpl := range chart.Templates {
			result.validateTemplate(tpl, ctx, placeholders)
		}
		report.Variants = append(report.Variants, result)
	}

	report.checkReferences()
	return report, nil
}

// variantPlan is one render variant to validate: its display name and the
// full --set-style override map to build its context with.
type variantPlan struct {
	name string
	set  map[string]string
}

// planVariants produces the defaults variant (declared values + explicit Set
// overrides) followed by one variant per bool param with the toggle flipped
// relative to its effective default. Params the caller explicitly Set are not
// auto-flipped — an explicit override is a deliberate arm choice.
func planVariants(params []Param, set map[string]string) []variantPlan {
	variants := []variantPlan{{name: "defaults", set: set}}
	for _, p := range params {
		if p.Type != paramTypeBool {
			continue
		}
		if _, overridden := set[p.Name]; overridden {
			continue
		}
		current, _ := p.Value.(bool)
		flipped := strconv.FormatBool(!current)

		variantSet := make(map[string]string, len(set)+1)
		for k, v := range set {
			variantSet[k] = v
		}
		variantSet[p.Name] = flipped
		variants = append(variants, variantPlan{name: p.Name + "=" + flipped, set: variantSet})
	}
	return variants
}

// validateTemplate renders one template file and validates every document in
// it, appending issues and docs to the variant result.
func (r *VariantResult) validateTemplate(tpl TemplateFile, ctx map[string]any, placeholders []string) {
	addError := func(format string, args ...any) {
		r.Issues = append(r.Issues, Issue{Severity: SeverityError, File: tpl.Name, Message: fmt.Sprintf(format, args...)})
	}

	for _, finding := range scanBannedConstructs(tpl.Content) {
		addError("%s", finding)
	}

	rendered, err := renderTemplate(tpl.Name, tpl.Content, ctx)
	if err != nil {
		addError("template rendering failed: %v", err)
		return
	}

	// A value-less param renders as its "<name>" placeholder — legal for the
	// platform (the value arrives from the InfraStack), but this gate's
	// promise is that DEFAULTS render deployable, so depending on one is
	// worth a warning even before schema validation rejects it.
	for _, placeholder := range placeholders {
		if strings.Contains(rendered, placeholder) {
			r.Issues = append(r.Issues, Issue{
				Severity: SeverityWarning, File: tpl.Name,
				Message: fmt.Sprintf("template uses param %s, which has no default value — supply one in values.yaml or via --set", placeholder),
			})
		}
	}

	docs, err := splitDocs(rendered)
	if err != nil {
		addError("rendered output is not valid YAML: %v", err)
		return
	}

	for _, docYaml := range docs {
		r.validateDoc(tpl.Name, docYaml)
	}
}

// validateDoc validates one rendered document: strict schema load, presence
// of metadata.name, the whole document's protovalidate/CEL rule set, and collects
// its valueFrom references for the cross-document pass.
func (r *VariantResult) validateDoc(file string, docYaml []byte) {
	msg, err := loadRenderedDoc(docYaml)
	if err != nil {
		r.Issues = append(r.Issues, Issue{Severity: SeverityError, File: file, Message: compactError(err)})
		return
	}

	doc := Doc{
		File: file,
		Kind: string(msg.ProtoReflect().Descriptor().Name()),
		YAML: docYaml,
		Msg:  msg,
	}
	if meta := metadatareflect.ExtractMetadata(msg); meta != nil {
		doc.Name = meta.GetName()
	}

	addIssue := func(severity Severity, format string, args ...any) {
		r.Issues = append(r.Issues, Issue{
			Severity: severity, File: file,
			ResourceKind: doc.Kind, ResourceName: doc.Name,
			Message: fmt.Sprintf(format, args...),
		})
	}

	if doc.Name == "" {
		addIssue(SeverityError, "metadata.name is required (it is the resource's identity in the chart's dependency graph)")
	}

	// A rendered document may omit apiVersion/kind consts (the platform stamps
	// the authoritative values on write), but a PRESENT value that conflicts
	// with the document's kind is an authoring error — surfacing it here keeps
	// chart validation in agreement with what the deploy boundary rejects.
	for _, mismatch := range manifest.EnvelopeMismatches(msg) {
		addIssue(SeverityError, "%s", mismatch)
	}

	// The whole document, stamped the way the platform stamps it on write, is
	// what the install validates -- metadata rules (the slug alphabet) as well
	// as the spec's -- so the offline gate validates the same document.
	validationErr := protovalidate.GlobalValidator.Validate(manifest.StampEnvelope(msg))
	// A rule about a value the host resolves later is the host's to apply to
	// the resolved value (see Options.IsDeferredToken).
	validationErr, _ = deferrules.Split(validationErr, r.deferredToken)
	if validationErr != nil {
		// One shared validator for the whole run, not a fresh instance per
		// document. The per-doc `New(WithDisableLazy(), WithMessages(spec))`
		// this replaces looked principled -- eager compile-error surfacing --
		// but recompiled the same kind's full CEL rule set for EVERY document
		// of EVERY variant of EVERY chart: ~253 CPU-seconds of redundant
		// compilation across the charts tree, the charts lane's dominant
		// cost. protovalidate's own docs prescribe the fix ("each Validator
		// instance has its own caches"): GlobalValidator lazily compiles each
		// message type once per process and is safe for concurrent use.
		// Compile failures still get their distinct diagnosis -- Validate
		// returns them typed, branched below.
		var compileErr *protovalidate.CompilationError
		if errors.As(validationErr, &compileErr) {
			addIssue(SeverityError, "failed to initialize validator: %v", validationErr)
		} else {
			addIssue(SeverityError, "schema validation failed: %s", compactError(validationErr))
		}
	} else {
		for _, check := range r.checks {
			for _, finding := range check(msg) {
				addIssue(SeverityError, "%s", finding)
			}
		}
	}

	doc.refUses = manifestgraph.CollectRefUses(msg)
	r.Docs = append(r.Docs, doc)
}

// checkReferences runs the cross-document pass on every variant:
// per-reference integrity (annotations, field-path resolution), reference
// target resolution, and dependency-cycle detection.
//
// Target resolution is variant-aware, and unresolved targets are WARNINGS
// with two distinct diagnoses:
//
//   - target defined by some OTHER variant of the same chart: either a
//     toggle removed the resource while a reference still stands (a real
//     defect), or this is a bring-your-own arm whose parameter redirects
//     the same reference to a resource owned outside the chart (a designed
//     pattern — e.g. one network_name param naming the chart's own VPC in
//     one arm and a landing zone's VPC in the other). The two are
//     offline-indistinguishable, so the warning says exactly what to verify
//     rather than false-failing the designed pattern;
//   - target no variant defines: the plain cross-chart case — charts
//     compose onto resources owned elsewhere by design, resolved in the
//     target environment at deploy time.
//
// A reference that names an env explicitly points outside this chart's
// deployment by design and is never treated as an in-chart edge.
func (r *Report) checkReferences() {
	// Union of every identity any variant renders — the "this chart can
	// define it" set that separates toggle breakage from cross-chart intent.
	anyVariant := map[refTarget]bool{}
	for _, v := range r.Variants {
		for _, doc := range v.Docs {
			if doc.Name == "" {
				continue
			}
			anyVariant[refTarget{kind: catalogkindreflect.KindFromString(doc.Kind), name: doc.Name}] = true
		}
	}

	for vi := range r.Variants {
		v := &r.Variants[vi]

		// Index the variant's resources by (kind, name). A duplicate identity
		// is an error — the platform's dependency graph could not tell the
		// two resources apart.
		index := map[refTarget]int{}
		for i, doc := range v.Docs {
			if doc.Name == "" {
				continue
			}
			key := refTarget{kind: catalogkindreflect.KindFromString(doc.Kind), name: doc.Name}
			if prev, dup := index[key]; dup {
				v.Issues = append(v.Issues, Issue{
					Severity: SeverityError, File: doc.File, ResourceKind: doc.Kind, ResourceName: doc.Name,
					Message: fmt.Sprintf("duplicate resource identity: %s %q is also defined in %s", doc.Kind, doc.Name, v.Docs[prev].File),
				})
				continue
			}
			index[key] = i
		}

		edges := make([][]int, len(v.Docs))
		for i, doc := range v.Docs {
			for _, use := range doc.refUses {
				checked, problems := manifestgraph.CheckRef(use)
				for _, p := range problems {
					v.Issues = append(v.Issues, Issue{
						Severity: SeverityError, File: doc.File, ResourceKind: doc.Kind, ResourceName: doc.Name,
						Message: p,
					})
				}
				target := refTarget{kind: checked.Kind, name: checked.Name}
				if target.kind == catalogkind.CatalogKind_unspecified || target.name == "" {
					continue
				}
				if use.Ref.GetEnv() != "" {
					continue
				}
				if targetIdx, ok := index[target]; ok {
					edges[i] = append(edges[i], targetIdx)
				} else if anyVariant[target] {
					v.Issues = append(v.Issues, Issue{
						Severity: SeverityWarning, File: doc.File, ResourceKind: doc.Kind, ResourceName: doc.Name,
						Message: fmt.Sprintf("%s: references %s %q, which this chart defines in another variant but not in this one — verify this is a bring-your-own arm resolving to an existing resource, not a toggle that removed the target while the reference still stands", use.FieldPath, target.kind, target.name),
					})
				} else {
					v.Issues = append(v.Issues, Issue{
						Severity: SeverityWarning, File: doc.File, ResourceKind: doc.Kind, ResourceName: doc.Name,
						Message: fmt.Sprintf("%s: references %s %q, which this chart does not define — it must already exist in the target environment at deploy time", use.FieldPath, target.kind, target.name),
					})
				}
			}
		}

		if cycle := manifestgraph.FindCycle(edges); cycle != nil {
			var parts []string
			for _, idx := range cycle {
				parts = append(parts, fmt.Sprintf("%s %q", v.Docs[idx].Kind, v.Docs[idx].Name))
			}
			v.Issues = append(v.Issues, Issue{
				Severity: SeverityError, File: v.Docs[cycle[0]].File,
				ResourceKind: v.Docs[cycle[0]].Kind, ResourceName: v.Docs[cycle[0]].Name,
				Message: "dependency cycle: " + strings.Join(parts, " -> "),
			})
		}
	}
}

// loadRenderedDoc writes a rendered document to a temp file and loads it
// through the CLI's canonical manifest loader — the same strict path (kind
// registry, protojson with unknown fields rejected, proto field defaults)
// every other planton command uses.
func loadRenderedDoc(docYaml []byte) (proto.Message, error) {
	tmp, err := os.CreateTemp("", "planton-chart-doc-*.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file")
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(docYaml); err != nil {
		tmp.Close()
		return nil, errors.Wrap(err, "failed to write temp file")
	}
	if err := tmp.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close temp file")
	}
	return manifest.LoadManifest(tmp.Name())
}

// splitDocs splits rendered multi-document YAML into per-document byte
// slices, skipping empty documents (a template whose conditionals all
// evaluate false legitimately renders to nothing).
func splitDocs(rendered string) ([][]byte, error) {
	var out [][]byte
	decoder := yamlv3.NewDecoder(strings.NewReader(rendered))
	for {
		var node yamlv3.Node
		err := decoder.Decode(&node)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if isNullNode(&node) {
			continue
		}
		docBytes, err := yamlv3.Marshal(&node)
		if err != nil {
			return nil, err
		}
		out = append(out, docBytes)
	}
	return out, nil
}

func isNullNode(node *yamlv3.Node) bool {
	if node == nil || node.Kind == 0 {
		return true
	}
	if node.Kind == yamlv3.DocumentNode {
		if len(node.Content) == 0 {
			return true
		}
		return node.Content[0].Tag == "!!null"
	}
	return node.Tag == "!!null"
}

// compactError flattens a (possibly multi-line, CLI-decorated) error into a
// single readable message for the issue report.
func compactError(err error) string {
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "validation error:")
	msg = strings.ReplaceAll(msg, "\n", " ")
	return strings.Join(strings.Fields(msg), " ")
}
