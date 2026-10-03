//go:build !codegen
// +build !codegen

package importmap

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	yamlv3 "gopkg.in/yaml.v3"

	kindv1 "github.com/plantonhq/planton/iac/catalogkindimportmap/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
)

// requireStrictYAML fails the test when a file does not survive a strict YAML
// parse -- most importantly, when a mapping carries the same key twice. The
// production loader (pkg/protobufyaml -> sigs.k8s.io/yaml) silently keeps the
// LAST duplicate key, so a duplicated key passes this repo's own loading while
// strict downstream parsers of the SAME files (the platform console's
// catalog-data bundler) refuse the whole document. yaml.v3 rejects duplicate
// keys by default, naming both lines.
func requireStrictYAML(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var doc any
	if err := yamlv3.Unmarshal(data, &doc); err != nil {
		t.Errorf("%s: strict YAML parse failed (duplicate keys break strict downstream parsers even when the lenient loader tolerates them): %v", path, err)
	}
}

var tfResourcePattern = regexp.MustCompile(`(?m)^resource\s+"([a-z0-9_]+)"\s+"([A-Za-z0-9_-]+)"`)

// cloudControlTypeNamePattern is the CloudFormation type-name shape
// (e.g. "AWS::S3::Bucket") that Cloud Control keys its list/get calls on.
var cloudControlTypeNamePattern = regexp.MustCompile(`^[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}$`)

// TestImportMapConformance validates every authored import recipe against the
// module sources it maps, offline:
//
//  1. The provider catalog and each kind map parse against their proto
//     schemas.
//  2. Every resource type the kind's OpenTofu module declares has an
//     id_format in the provider catalog -- an unmapped type would surface as
//     an unimportable enumerated address.
//  3. Every {placeholder} those formats reference is declared by the
//     kind map, and every declared value either derives from somewhere
//     or tells the user where to find it.
//  4. from_spec_field paths resolve to scalar leaves on the kind's spec
//     proto; from_output keys are real fields on the kind's
//     Outputs proto. A typo'd path would silently downgrade a zero-input
//     import to a manual one.
//
// Enrollment is the import-map file itself: every discovered
// `v1/iac/import-map.yaml` is checked, with no allowlist to keep in sync.
// The platform's catalog bundler and the E2E round-trip gate key off the
// same file presence, so an authored map can never ship to the import
// wizard while dodging this guard. (Whether a map is LIVE-proven is a
// separate, human-readable ledger in this package's README; the round-trip
// lane is its only honest enforcement.)
func TestImportMapConformance(t *testing.T) {
	root := repoRoot(t)

	mappedKinds, err := DiscoverCatalogKindImportMaps(root)
	if err != nil {
		t.Fatalf("discovering catalog kind import maps: %v", err)
	}
	if len(mappedKinds) == 0 {
		t.Fatal("no catalog kind import maps discovered -- wrong repo root?")
	}

	for provider, kindDirs := range mappedKinds {
		requireStrictYAML(t, ProviderCatalogPath(root, provider))
		catalog, err := LoadProviderCatalog(root, provider)
		if err != nil {
			t.Fatalf("provider catalog for %s: %v", provider, err)
		}
		if catalog.GetKind() != "ProviderImportCatalog" {
			t.Fatalf("provider catalog for %s: kind is %q, want ProviderImportCatalog", provider, catalog.GetKind())
		}

		formatsByTerraformType := map[string]string{}
		notImportableTypes := map[string]bool{}
		for _, rt := range catalog.GetSpec().GetResourceTypes() {
			if rt.GetTerraformType() == "" {
				t.Errorf("%s catalog: entry with empty terraform_type", provider)
				continue
			}
			if _, dup := formatsByTerraformType[rt.GetTerraformType()]; dup {
				t.Errorf("%s catalog: duplicate entry for %s", provider, rt.GetTerraformType())
			}
			formatsByTerraformType[rt.GetTerraformType()] = rt.GetIdFormat()
			// Exactly one of id_format / not_importable_upstream_reason: a
			// format on an importer-less type would fail at import time, and
			// an entry with neither declares nothing.
			switch {
			case rt.GetNotImportableUpstreamReason() != "" && rt.GetIdFormat() != "":
				t.Errorf("%s catalog: %s declares BOTH id_format and not_importable_upstream_reason -- they are mutually exclusive",
					provider, rt.GetTerraformType())
			case rt.GetNotImportableUpstreamReason() != "":
				notImportableTypes[rt.GetTerraformType()] = true
			case rt.GetIdFormat() == "":
				t.Errorf("%s catalog: %s has an empty id_format (declare not_importable_upstream_reason instead if the upstream resource ships no importer)",
					provider, rt.GetTerraformType())
			}
			// The tolerance lists feed the round-trip proof's plan oracle; an
			// empty entry would tolerate nothing meaningful and signals a
			// YAML authoring slip.
			for _, attr := range rt.GetConfigOnlyAttributes() {
				if strings.TrimSpace(attr) == "" {
					t.Errorf("%s catalog: %s declares an empty config_only_attributes entry", provider, rt.GetTerraformType())
				}
			}
			for _, attr := range rt.GetWriteNormalizedAttributes() {
				if strings.TrimSpace(attr) == "" {
					t.Errorf("%s catalog: %s declares an empty write_normalized_attributes entry", provider, rt.GetTerraformType())
				}
			}
		}

		// The scan-side correspondence must be well-formed and unambiguous:
		// ground-truth building and mapping proposals translate between what
		// a scan reports (Cloud Control type names) and what IaC state holds
		// (terraform types), so two terraform types claiming the same Cloud
		// Control type would make that translation a guess.
		cloudControlClaims := map[string]string{}
		for _, rt := range catalog.GetSpec().GetResourceTypes() {
			ccType := rt.GetCloudControlTypeName()
			if ccType == "" {
				continue
			}
			if !cloudControlTypeNamePattern.MatchString(ccType) {
				t.Errorf("%s catalog: %s declares malformed cloud_control_type_name %q (want e.g. AWS::S3::Bucket)",
					provider, rt.GetTerraformType(), ccType)
			}
			if prior, dup := cloudControlClaims[ccType]; dup {
				t.Errorf("%s catalog: cloud_control_type_name %q claimed by both %s and %s -- the scan-to-state translation must be unambiguous",
					provider, ccType, prior, rt.GetTerraformType())
			}
			cloudControlClaims[ccType] = rt.GetTerraformType()
		}

		for _, kindDir := range kindDirs {
			kindDir := kindDir
			t.Run(provider+"/"+kindDir, func(t *testing.T) {
				m, err := LoadCatalogKindImportMap(root, provider, kindDir)
				if err != nil {
					t.Fatalf("catalog kind import map: %v", err)
				}
				if m.GetMetadata().GetName() != kindDir {
					t.Errorf("metadata.name is %q, want %q", m.GetMetadata().GetName(), kindDir)
				}

				mapRelPath, err := CatalogKindImportMapPath("", provider, kindDir)
				if err != nil {
					t.Fatalf("catalog kind import map path: %v", err)
				}
				requireStrictYAML(t, filepath.Join(root, mapRelPath))

				declaredValues := map[string]*kindv1.ImportValue{}
				for _, v := range m.GetSpec().GetValues() {
					if v.GetName() == "" {
						t.Error("import value with empty name")
						continue
					}
					declaredValues[v.GetName()] = v
					// A value that cannot derive from anywhere MUST carry guidance --
					// the empathy contract: when we must ask, show where to find it.
					if len(v.GetDerivations()) == 0 && v.GetWhereToFind() == "" {
						t.Errorf("value %q has no derivations and no where_to_find", v.GetName())
					}
				}

				// Every module-declared resource type must be importable.
				moduleTypes, moduleNames := terraformResourceTypes(t, filepath.Join(root, "catalog", provider, kindDir, "iac/tf"))
				if len(moduleTypes) == 0 {
					t.Fatal("module declares no resources -- wrong path?")
				}

				// A tofu_resource_name scope must name a REAL logical resource:
				// a typo'd scope silently falls back to the unscoped declaration
				// at resolve time, which imports the wrong resource.
				for _, v := range m.GetSpec().GetValues() {
					if scope := v.GetTofuResourceName(); scope != "" && !moduleNames[scope] {
						t.Errorf("value %q is scoped to tofu_resource_name %q, but the module declares no resource with that logical name",
							v.GetName(), scope)
					}
				}

				// Import-normalized declarations are the narrowest tolerance
				// vocabulary and must stay narrow: a real resource, a dotted
				// sub-path (whole-attribute tolerance belongs in the provider
				// catalog), and a stated reason (an undocumented tolerance is
				// indistinguishable from a silenced defect). A typo'd resource
				// name would silently tolerate nothing.
				for _, nr := range m.GetSpec().GetImportNormalized() {
					if nr.GetTofuResourceName() == "" || !moduleNames[nr.GetTofuResourceName()] {
						t.Errorf("import_normalized entry names tofu_resource_name %q, but the module declares no resource with that logical name",
							nr.GetTofuResourceName())
					}
					if len(nr.GetSubPaths()) == 0 {
						t.Errorf("import_normalized entry for %q declares no sub_paths", nr.GetTofuResourceName())
					}
					for _, sp := range nr.GetSubPaths() {
						if len(SplitAttributePath(sp.GetPath())) < 2 {
							t.Errorf("import_normalized sub-path %q on %q does not reach below a top-level attribute -- whole-attribute tolerance belongs in the provider catalog",
								sp.GetPath(), nr.GetTofuResourceName())
						}
						if strings.TrimSpace(sp.GetReason()) == "" {
							t.Errorf("import_normalized sub-path %q on %q carries no reason",
								sp.GetPath(), nr.GetTofuResourceName())
						}
					}
				}
				for _, resourceType := range moduleTypes {
					// Types the upstream provider ships no importer for are
					// accounted by their recorded reason -- the round-trip
					// skips their import and proves the re-create path instead.
					if notImportableTypes[resourceType] {
						continue
					}
					idFormat, mapped := formatsByTerraformType[resourceType]
					if !mapped {
						t.Errorf("module resource %s has no id_format in the %s catalog -- "+
							"the module gained a resource type its import recipes don't cover; add a %s entry "+
							"(with its provider import-ID format, or not_importable_upstream_reason if the upstream resource ships no importer) to %s, then re-run the live round-trip lane for %s",
							resourceType, provider, resourceType,
							ProviderCatalogPath("", provider), kindDir)
						continue
					}
					for _, placeholder := range Placeholders(idFormat) {
						if _, declared := declaredValues[placeholder]; !declared {
							t.Errorf("placeholder {%s} (id_format of %s) not declared in the kind map -- "+
								"add a value entry for it to %s (prefer a derivable source; where_to_find is mandatory when nothing derives)",
								placeholder, resourceType,
								mapRelPath)
						}
					}
				}

				// Derivation paths must resolve against the kind's protos.
				kind := catalogkindreflect.KindFromString(kindDir)
				apiMessage, err := catalogkindreflect.NewInstance(kind)
				if err != nil {
					t.Fatalf("NewInstance(%s): %v", kindDir, err)
				}
				specDescriptor := specMessageDescriptor(t, apiMessage)
				outputsDescriptor := outputsDescriptor(t, apiMessage)
				for _, v := range m.GetSpec().GetValues() {
					for _, d := range v.GetDerivations() {
						switch source := d.GetSource().(type) {
						case *kindv1.ImportValueDerivation_FromSpecField:
							if err := validateScalarPath(specDescriptor, source.FromSpecField); err != nil {
								t.Errorf("value %q from_spec_field %q: %v", v.GetName(), source.FromSpecField, err)
							}
						case *kindv1.ImportValueDerivation_FromOutput:
							if outputsDescriptor.Fields().ByName(protoreflect.Name(source.FromOutput)) == nil {
								t.Errorf("value %q from_output %q: no such field on %s",
									v.GetName(), source.FromOutput, outputsDescriptor.FullName())
							}
						case *kindv1.ImportValueDerivation_FromOutputKeyedByAddress:
							// The arm names a map output whose entries are keyed by
							// the resource's for_each keys -- a scalar field here
							// means the map author reached for the wrong arm (plain
							// from_output serves scalars).
							field := outputsDescriptor.Fields().ByName(protoreflect.Name(source.FromOutputKeyedByAddress))
							if field == nil {
								t.Errorf("value %q from_output_keyed_by_address %q: no such field on %s",
									v.GetName(), source.FromOutputKeyedByAddress, outputsDescriptor.FullName())
							} else if !field.IsMap() {
								t.Errorf("value %q from_output_keyed_by_address %q: field on %s is not a map -- the arm exists for per-instance entries keyed by the for_each key; use from_output for scalar outputs",
									v.GetName(), source.FromOutputKeyedByAddress, outputsDescriptor.FullName())
							}
						case *kindv1.ImportValueDerivation_FromArnPart:
							switch source.FromArnPart {
							case "resource_id", "resource_name", "account_id", "region", "arn":
							default:
								t.Errorf("value %q from_arn_part %q: not a recognized ARN part", v.GetName(), source.FromArnPart)
							}
						}
					}
				}
			})
		}
	}
}

// terraformResourceTypes scans a module directory's .tf files for top-level
// resource declarations and returns the distinct resource types.
func terraformResourceTypes(t *testing.T, moduleDir string) ([]string, map[string]bool) {
	t.Helper()
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		t.Fatalf("reading module dir %s: %v", moduleDir, err)
	}
	seen := map[string]bool{}
	names := map[string]bool{}
	var types []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(moduleDir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		for _, match := range tfResourcePattern.FindAllStringSubmatch(string(content), -1) {
			if !seen[match[1]] {
				seen[match[1]] = true
				types = append(types, match[1])
			}
			names[match[2]] = true
		}
	}
	return types, names
}

// specMessageDescriptor returns the descriptor of the kind's spec message
// (the api envelope's `spec` field).
func specMessageDescriptor(t *testing.T, apiMessage proto.Message) protoreflect.MessageDescriptor {
	t.Helper()
	specField := apiMessage.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind {
		t.Fatalf("%s has no spec message field", apiMessage.ProtoReflect().Descriptor().FullName())
	}
	return specField.Message()
}

// outputsDescriptor locates the kind's Outputs message from the
// sibling outputs.proto of the kind's api.proto.
func outputsDescriptor(t *testing.T, apiMessage proto.Message) protoreflect.MessageDescriptor {
	t.Helper()
	apiPath := apiMessage.ProtoReflect().Descriptor().ParentFile().Path()
	outputsPath := filepath.Join(filepath.Dir(apiPath), "outputs.proto")
	file, err := protoregistry.GlobalFiles.FindFileByPath(outputsPath)
	if err != nil {
		t.Fatalf("finding %s: %v", outputsPath, err)
	}
	for i := 0; i < file.Messages().Len(); i++ {
		msg := file.Messages().Get(i)
		if strings.HasSuffix(string(msg.Name()), "Outputs") {
			return msg
		}
	}
	t.Fatalf("%s declares no *Outputs message", outputsPath)
	return nil
}

// validateScalarPath walks a dot path of field names and requires a scalar
// leaf with no repeated/map segments (a repeated segment would make the
// derivation ambiguous -- which instance?).
func validateScalarPath(desc protoreflect.MessageDescriptor, dotPath string) error {
	if dotPath == "" {
		return fmt.Errorf("empty path")
	}
	current := desc
	segments := strings.Split(dotPath, ".")
	for i, segment := range segments {
		field := current.Fields().ByName(protoreflect.Name(segment))
		if field == nil {
			return fmt.Errorf("no field %q on %s", segment, current.FullName())
		}
		if field.IsList() || field.IsMap() {
			return fmt.Errorf("segment %q is repeated/map -- ambiguous for derivation", segment)
		}
		if i == len(segments)-1 {
			if field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind {
				return fmt.Errorf("leaf %q is a message, not a scalar", segment)
			}
			return nil
		}
		if field.Kind() != protoreflect.MessageKind {
			return fmt.Errorf("segment %q is a scalar but the path continues", segment)
		}
		current = field.Message()
	}
	return nil
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
