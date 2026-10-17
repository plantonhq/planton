package manifestprojection_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/pruning"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"sigs.k8s.io/yaml"
)

// The mirror test proves a projection kind's spec has the custom resource's
// keys. This test proves the bytes the engines actually send are accepted:
// every e2e scenario and the full-surface e2e/manifest.yaml of each kind with a
// pinned definition is rendered through the projection both engines apply, and
// the object is run through the API server's own structural-schema code
// (k8s.io/apiextensions-apiserver) -- pruning, which must remove nothing, and
// OpenAPI validation, which must find nothing. A value the projection writes in
// the wrong form (a 64-bit integer as text, a quantity the server's pattern
// refuses, a required key a manifest can omit) fails here, with no cluster.
//
// CEL rules (x-kubernetes-validations) are not evaluated: they need the API
// server's CEL environment, and the e2e lanes apply the same objects to a real
// one. Presets are not rendered, because they carry placeholders for the values
// an adopter fills in; the preset-validity gate holds them to the spec.

// scenarioTokens are the e2e harness's run-time substitutions, replaced here
// with a plain value so the manifest loads.
var scenarioTokens = regexp.MustCompile(`\$\{E2E_[A-Z_]+(:[^}]*)?\}`)

func TestProjection_RenderedExamplesValidateAgainstServedSchema(t *testing.T) {
	root := repositoryRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.manifest-projection lane")
	}
	paths, err := filepath.Glob(filepath.Join(pinnedSchemasDir, "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			var crd apiextensionsv1.CustomResourceDefinition
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal(raw, &crd); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			kind, version, ok := projectionKindFor(crd.Spec.Group, crd.Spec.Names.Kind)
			if !ok {
				t.Fatalf("%s pins %s/%s, which no projection kind applies", path, crd.Spec.Group, crd.Spec.Names.Kind)
			}
			structural, validator := servedSchema(t, &crd, version)

			examples := kindExamples(t, root, strings.ToLower(kind.String()))
			if len(examples) == 0 {
				t.Fatalf("%s has no e2e scenarios or e2e/manifest.yaml to render", kind)
			}
			for _, example := range examples {
				rel, _ := filepath.Rel(root, example)
				obj := renderedObject(t, example)
				pruned := pruning.PruneWithOptions(obj, structural, true, structuralschema.UnknownFieldPathOptions{TrackUnknownFieldPaths: true})
				for _, p := range pruned {
					t.Errorf("%s: the API server would prune %s, so the setting never lands", rel, p)
				}
				for _, e := range apiservervalidation.ValidateCustomResource(nil, obj, validator) {
					t.Errorf("%s: the API server would refuse the object: %v", rel, e)
				}
			}
		})
	}
}

// servedSchema builds the structural schema and the validator the API server
// derives from the version a projection kind applies.
func servedSchema(t *testing.T, crd *apiextensionsv1.CustomResourceDefinition, version string) (*structuralschema.Structural, apiservervalidation.SchemaValidator) {
	t.Helper()
	for _, v := range crd.Spec.Versions {
		if v.Name != version || v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
			continue
		}
		internal := &apiextensions.JSONSchemaProps{}
		if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(v.Schema.OpenAPIV3Schema, internal, nil); err != nil {
			t.Fatalf("convert %s schema: %v", version, err)
		}
		structural, err := structuralschema.NewStructural(internal)
		if err != nil {
			t.Fatalf("structural %s schema: %v", version, err)
		}
		validator, _, err := apiservervalidation.NewSchemaValidator(internal)
		if err != nil {
			t.Fatalf("validator for %s schema: %v", version, err)
		}
		return structural, validator
	}
	t.Fatalf("no served %s schema", version)
	return nil, nil
}

// kindExamples lists the concrete manifests a kind's folder carries: its full
// surface and every e2e scenario.
func kindExamples(t *testing.T, root, folder string) []string {
	t.Helper()
	dirs, _ := filepath.Glob(filepath.Join(root, "catalog", "*", folder))
	var out []string
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "e2e", "manifest.yaml")); err == nil {
			out = append(out, filepath.Join(dir, "e2e", "manifest.yaml"))
		}
		scenarios, _ := filepath.Glob(filepath.Join(dir, "e2e", "scenarios", "*.yaml"))
		out = append(out, scenarios...)
	}
	sort.Strings(out)
	return out
}

// renderedObject renders a manifest into the object both engines apply, as the
// plain JSON value the API server decodes. References resolve to the name of
// the object they target, the value the platform would substitute.
func renderedObject(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := manifest.LoadManifestBytes(scenarioTokens.ReplaceAll(raw, []byte("e2e")), path)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	for _, use := range manifestgraph.CollectRefUses(msg) {
		use.Replace(&foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: use.Ref.GetName()},
		})
	}
	m, ok := msg.(manifestprojection.Manifest)
	if !ok {
		t.Fatalf("%s is not a catalog manifest", path)
	}
	rendered, err := manifestprojection.Render(m)
	if err != nil {
		t.Fatalf("render %s: %v", path, err)
	}
	metadata := map[string]interface{}{"name": rendered.Name, "labels": rendered.Labels}
	if rendered.Namespace != "" {
		metadata["namespace"] = rendered.Namespace
	}
	if len(rendered.Annotations) > 0 {
		metadata["annotations"] = rendered.Annotations
	}
	encoded, err := json.Marshal(map[string]interface{}{
		"apiVersion": rendered.APIVersion,
		"kind":       rendered.Kind,
		"metadata":   metadata,
		"spec":       rendered.Spec,
	})
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(encoded, &obj); err != nil {
		t.Fatal(err)
	}
	return obj
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
}
