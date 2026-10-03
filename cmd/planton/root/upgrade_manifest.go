package root

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/conversion"
	"github.com/plantonhq/planton/pkg/conversion/embedded"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/spf13/cobra"
)

var UpgradeManifest = &cobra.Command{
	Use:   "upgrade-manifest [manifest-path]",
	Short: "convert a manifest to its kind's served api version, offline",
	Long: `Converts a manifest written at an older api version to the version the
kind currently serves, using the same declarative conversion specs the
platform executes -- entirely offline. Declared losses (values the newer
version cannot express) are reported explicitly; the converted manifest is
validated before it is printed.`,
	Example: `
	# Upgrade a manifest file and print the converted YAML
	planton upgrade-manifest manifest.yaml

	# Write the converted manifest to a file
	planton upgrade-manifest manifest.yaml --output upgraded.yaml
	`,
	Args: cobra.MaximumNArgs(1),
	Run:  upgradeManifestHandler,
}

var upgradeManifestOutput string

func init() {
	iacflags.AddManifestSourceFlags(UpgradeManifest)
	// No -o shorthand: an embedding host (the Planton Platform CLI) owns -o as
	// its global --output-format, and a second -o panics when cobra merges the
	// flag sets.
	UpgradeManifest.Flags().StringVar(&upgradeManifestOutput, "output", "", "write the converted manifest to this file instead of stdout")
}

func upgradeManifestHandler(cmd *cobra.Command, args []string) {
	var manifestPath string
	var isTemp bool
	var err error

	if len(args) > 0 {
		manifestPath = args[0]
	} else {
		manifestPath, isTemp, err = climanifest.ResolveManifestPath(cmd)
		if err != nil {
			if climanifest.HandleClipboardError(err) {
				os.Exit(1)
			}
			cliprint.PrintError(fmt.Sprintf("failed to resolve manifest: %v", err))
			os.Exit(1)
		}
		if isTemp {
			defer os.Remove(manifestPath)
		}
	}

	converted, losses, err := upgradeManifestFile(manifestPath)
	if err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}

	for _, loss := range losses {
		cliprint.PrintWarning(fmt.Sprintf("declared loss at %s: %s", loss.Path, loss.Reason))
	}

	if upgradeManifestOutput != "" {
		if err := os.WriteFile(upgradeManifestOutput, converted, 0o644); err != nil {
			cliprint.PrintError(fmt.Sprintf("writing %s: %v", upgradeManifestOutput, err))
			os.Exit(1)
		}
		cliprint.PrintSuccessMessage(fmt.Sprintf("upgraded manifest written to %s", upgradeManifestOutput))
		return
	}
	fmt.Print(string(converted))
}

// upgradeManifestFile converts the manifest at path to its kind's served
// version and returns the converted YAML plus any declared losses. The
// conversion operates on the raw document (an old-version manifest cannot
// parse into current stubs -- that inability is the whole point), and the
// RESULT is validated through the normal offline validator before returning.
//
// The document must be in MANIFEST spelling (camelCase field names -- the
// form conversion specs address). This binary compiles only served-version
// schemas, so it cannot canonicalize a document authored in storage spelling
// (proto-name keys, the form stored-document exports carry) at its OLD
// version: a storage-spelled document refuses honestly below with the way
// out named, never a silent op no-op. Descriptor-holding lanes (the
// platform's server and CLI) canonicalize any spelling.
func upgradeManifestFile(manifestPath string) ([]byte, []conversion.DeclaredLoss, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading manifest: %w", err)
	}
	docJson, err := protobufyaml.YAMLToJSON(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("the manifest is not valid YAML: %w", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(docJson, &doc); err != nil {
		return nil, nil, fmt.Errorf("the manifest is not valid YAML: %w", err)
	}

	kindName, _ := doc["kind"].(string)
	if kindName == "" {
		return nil, nil, fmt.Errorf("the manifest has no kind field -- nothing to upgrade")
	}
	kind := catalogkindreflect.KindFromString(kindName)
	served, err := catalogkindreflect.KindVersion(kind)
	if err != nil {
		return nil, nil, fmt.Errorf("kind %q is not a registered catalog kind: %w", kindName, err)
	}

	apiVersion, _ := doc["apiVersion"].(string)
	if apiVersion == "" {
		// The envelope read is dual-spelling: a document whose version stamp
		// arrives under the storage key is a stored-document export, and this
		// binary cannot canonicalize its body (see the function doc) -- refuse
		// with the way out, never no-op the conversion ops on storage keys.
		if storageStamp, _ := doc["api_version"].(string); storageStamp != "" {
			return nil, nil, fmt.Errorf(
				"this document carries storage field spelling (api_version) -- a stored-document" +
					" export, not an authored manifest. This offline converter reads manifest" +
					" spelling (camelCase field names) only; re-author the manifest in that" +
					" spelling, or use the Planton platform CLI's upgrade-manifest, which" +
					" canonicalizes any spelling")
		}
	}
	_, current, found := cutLastSlash(apiVersion)
	if !found {
		return nil, nil, fmt.Errorf("the manifest's apiVersion %q carries no version segment", apiVersion)
	}
	if current == served {
		return nil, nil, fmt.Errorf("the manifest is already at %s -- the version this kind serves; nothing to upgrade", served)
	}

	specsFS, err := embedded.SpecsFS()
	if err != nil {
		return nil, nil, err
	}
	specs, err := conversion.SpecsForKind(specsFS, kind)
	if err != nil {
		return nil, nil, err
	}
	steps, err := conversion.Path(specs, current, served)
	if err != nil {
		return nil, nil, err
	}

	var losses []conversion.DeclaredLoss
	converted := doc
	for _, step := range steps {
		var stepLosses []conversion.DeclaredLoss
		converted, stepLosses, err = conversion.Apply(step.Spec, step.Direction, converted)
		if err != nil {
			return nil, nil, err
		}
		losses = append(losses, stepLosses...)
	}

	convertedJson, err := json.Marshal(converted)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing the converted manifest: %w", err)
	}
	out, err := protobufyaml.JSONToYAML(convertedJson)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing the converted manifest: %w", err)
	}

	// The conversion is only done when its OUTPUT passes the same offline
	// validation apply runs -- an upgrade that produces an invalid manifest
	// surfaces here, not at the server. Two causes share this failure: field
	// names the conversion ops could not address (a manifest carrying proto-
	// name spelling this binary cannot translate), or a genuine conversion-
	// spec defect. The wording admits both -- blaming release data for a
	// spelling this tool declares out of scope would be a lie.
	loaded, err := manifest.LoadManifestBytes(out, manifestPath+" (upgraded)")
	if err != nil {
		return nil, nil, fmt.Errorf("the upgraded manifest does not load against %s -- if the input"+
			" manifest carries storage field spelling (proto-name keys), re-author it in manifest"+
			" spelling or use the Planton platform CLI's upgrade-manifest; otherwise this is a"+
			" conversion-spec defect, report it: %w", served, err)
	}
	if err := manifest.ValidateLoaded(loaded); err != nil {
		return nil, nil, fmt.Errorf("the upgraded manifest does not validate against %s -- if the"+
			" input manifest carries storage field spelling (proto-name keys), re-author it in"+
			" manifest spelling or use the Planton platform CLI's upgrade-manifest; otherwise this"+
			" is a conversion-spec defect, report it: %w", served, err)
	}
	return out, losses, nil
}

func cutLastSlash(s string) (before, after string, found bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}
