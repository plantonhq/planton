package runner

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/encoding/protojson"
)

// DependencyOutputs holds the captured outputs of every deployed
// prerequisite, keyed by kind and then by manifest name. The second level exists
// because a prerequisite install profile may deploy several instances of the
// same kind (e.g. the two different-AZ subnets a load balancer requires), and a
// reference must be able to pick the specific instance it means.
type DependencyOutputs map[catalogkind.CatalogKind]map[string]map[string]interface{}

// ResolveManifestRefs implements, for the standalone E2E harness, the foreign-key
// resolution the orchestrated lanes perform: it replaces each value_from
// reference in the kind manifest that matches a deployed prerequisite
// with the literal value read from that prerequisite's outputs. Standalone
// Planton otherwise requires literal values -- the tofu generator errors on an
// unresolved ref and the pulumi modules drop it -- so this is the step that
// makes a composed (e.g. subnet -> vpc) topology testable end to end.
//
// The traversal and the strict reference rules live in pkg/manifestgraph (the
// one home shared with chart validation and multi-manifest deployment); what
// stays HERE is harness policy, quarantined deliberately:
//
//   - the sole-instance fallback: when a reference's name matches no deployed
//     instance but exactly one instance of the kind exists, that instance is
//     used -- scenario manifests name their references after real-world
//     topology, while the shared install profiles have fixed names, and
//     forcing them to agree would couple every scenario to the profile file;
//   - unresolved-target tolerance: a reference to a kind the chain never
//     deployed is left untouched by design (the engine's own wall catches it
//     at deploy when it matters).
//
// Two failure modes remain hard errors, so misdeclared scenarios fail loudly:
// a name that matches none of SEVERAL deployed instances (ambiguous), and a
// matched instance whose outputs lack the referenced field.
//
// The resolved manifest is written to a temp file whose path is returned; the
// original is left untouched. When there is nothing to resolve, the original path
// is returned unchanged.
func ResolveManifestRefs(manifestPath string, depOutputs DependencyOutputs) (string, error) {
	if len(depOutputs) == 0 {
		return manifestPath, nil
	}

	manifestObject, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load manifest for ref resolution from %s", manifestPath)
	}

	// Flatten each prerequisite's outputs to dotted keys (the platform's
	// flattening) and index by the SAME slug derivation graph identity uses,
	// so a reference naming "My VPC" joins the instance named "My VPC".
	type instance struct {
		name string
		outs map[string]string
	}
	bySlug := make(map[catalogkind.CatalogKind]map[string]instance, len(depOutputs))
	perKind := make(map[catalogkind.CatalogKind][]instance, len(depOutputs))
	for kind, instances := range depOutputs {
		bySlug[kind] = make(map[string]instance, len(instances))
		for name, out := range instances {
			inst := instance{name: name, outs: outputs.Flatten(out)}
			bySlug[kind][manifestgraph.GenerateSlug(name)] = inst
			perKind[kind] = append(perKind[kind], inst)
		}
	}

	// The lookup implements the harness policy documented above. Ambiguity
	// is captured through the closure (the lookup contract has no error arm
	// because "not found" is a legitimate answer everywhere else).
	var lookupErr error
	lookup := func(id manifestgraph.Identity) (map[string]string, bool) {
		instances := bySlug[id.Kind]
		if len(instances) == 0 {
			return nil, false
		}
		if inst, ok := instances[id.Slug]; ok {
			return inst.outs, true
		}
		if len(instances) == 1 {
			// The sole-instance fallback -- harness policy, see above.
			for _, inst := range instances {
				return inst.outs, true
			}
		}
		lookupErr = errors.Errorf(
			"reference %q matches none of the %d deployed %s prerequisites by name",
			id.Slug, len(instances), id.Kind)
		return nil, false
	}

	// The consumer's own env is the identity fallback for references that
	// name none (the lookup here ignores env anyway -- a harness chain is
	// single-environment by construction).
	env := metadatareflect.ExtractMetadata(manifestObject).GetEnv()

	resolved, findings := manifestgraph.ResolveRefs(manifestObject, env, lookup)
	if lookupErr != nil {
		return "", lookupErr
	}
	for _, f := range findings {
		// Missing output on a resolved target is a misdeclared scenario --
		// hard error. Unresolved targets and rule findings are left for the
		// engine's own wall, per harness policy.
		if f.Class == manifestgraph.FindingMissingOutput {
			return "", errors.New(f.Message)
		}
	}
	if resolved == 0 {
		return manifestPath, nil
	}

	jsonBytes, err := protojson.Marshal(manifestObject)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal resolved manifest")
	}
	yamlBytes, err := protobufyaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert resolved manifest to yaml")
	}

	// A temp file (not next to the scenario, so discovery never picks it up)
	// that KEEPS the scenario's basename: verifier dispatch keys behavioral
	// variants off the scenario name in the manifest path, and a
	// random-only temp name would silently demote every reference-carrying
	// behavioral scenario to its plain verifier (the same identity contract
	// the token-expansion copy honors).
	tmpDir, err := os.MkdirTemp("", "planton-e2e-resolved-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp dir for resolved manifest")
	}
	tmpPath := filepath.Join(tmpDir, filepath.Base(manifestPath))
	if err := os.WriteFile(tmpPath, yamlBytes, 0o600); err != nil {
		return "", errors.Wrap(err, "failed to write resolved manifest")
	}
	return tmpPath, nil
}
