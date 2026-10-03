package manifest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/internal/manifest/protodefaults"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"github.com/plantonhq/planton/pkg/yamldiag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	goyaml "gopkg.in/yaml.v3"
)

// ManifestLoadError represents an error when loading a manifest fails due to proto issues.
type ManifestLoadError struct {
	ManifestPath string
	Err          error
	// Kind is the manifest's resolved kind name, for schema-reference hints.
	Kind string
	// Mismatches is the YAML-aware diagnosis of the failure: real line
	// numbers, field paths, expected shapes. Empty when the failure is a
	// value-format problem only the parser understands.
	Mismatches []yamldiag.Mismatch
}

// Error renders the diagnosis when one exists, the parser error otherwise.
// The diagnosis must live HERE and not only in the styled UI renderer:
// surfaces that print the error value directly (chart validation reports,
// wrapped errors in scripts and agent transcripts) would otherwise regress
// to protojson's offsets into a JSON translation the author never wrote.
func (e *ManifestLoadError) Error() string {
	if len(e.Mismatches) == 0 {
		return e.Err.Error()
	}
	return fmt.Sprintf("manifest does not fit the %s schema:\n%s",
		e.Kind, yamldiag.FormatAll(e.Mismatches, e.Kind))
}

// IsManifestLoadError checks if an error is a ManifestLoadError.
func IsManifestLoadError(err error) bool {
	_, ok := err.(*ManifestLoadError)
	return ok
}

// HandleManifestLoadError displays the error beautifully if it's a ManifestLoadError.
// Returns true if it was handled, false otherwise.
func HandleManifestLoadError(err error) bool {
	if mle, ok := err.(*ManifestLoadError); ok {
		if len(mle.Mismatches) > 0 {
			ui.ManifestDiagnosis(mle.ManifestPath, mle.Kind, mle.Mismatches)
			return true
		}
		// No certain diagnosis -- fall back to the legacy display, which
		// salvages what it can from the parser error text.
		ui.ManifestLoadError(mle.ManifestPath, mle.Err)
		return true
	}
	return false
}

func LoadManifest(manifestPath string) (proto.Message, error) {
	isUrl, err := isManifestPathUrl(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine if manifest path is url")
	}

	if isUrl {
		manifestPath, err = downloadManifest(manifestPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to download manifest using %s", manifestPath)
		}
	}

	manifestYamlBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read manifest file %s", manifestPath)
	}

	return LoadManifestBytes(manifestYamlBytes, manifestPath)
}

// LoadManifestBytes unmarshals an in-memory manifest into its kind's typed proto message
// and applies proto-declared defaults. The manifest never needs to exist on disk, which is
// what rendered-template validation (e.g. infra-chart templates) relies on. sourceName is
// used only in error messages (a file path, or a "chart/template.yaml[docN]" style label).
//
// The input must be exactly ONE YAML document. This is the loader's contract,
// enforced here at the one funnel every load passes through: the YAML-to-JSON
// conversion below reads only the first document of a stream, so accepting
// multi-document input would silently drop every document after the first —
// a multi-resource kustomize overlay would deploy one resource and report
// success. Callers that legitimately handle multi-document YAML (chart
// validation, catalog tooling) split the stream BEFORE loading each document.
func LoadManifestBytes(manifestYamlBytes []byte, sourceName string) (proto.Message, error) {
	if err := refuseMultiDocument(manifestYamlBytes, sourceName); err != nil {
		return nil, err
	}

	jsonBytes, err := protobufyaml.YAMLToJSON(manifestYamlBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load yaml to json")
	}

	kindName, err := extractKindName(manifestYamlBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract catalogKind from %s", sourceName)
	}

	catalogKind := catalogkindreflect.KindFromString(kindName)

	manifest := catalogkindreflect.ToMessageMap[catalogKind]

	if manifest == nil {
		return nil, formatUnsupportedResourceError(kindName)
	}
	// ToMessageMap holds shared instances; clone so concurrent/repeated loads never
	// unmarshal into the same message.
	manifest = proto.Clone(manifest)

	if err := protojson.Unmarshal(jsonBytes, manifest); err != nil {
		// protojson stays the authority on what loads; its error, however,
		// carries offsets into the single-line JSON translation above --
		// meaningless to the author. Diagnose against the ORIGINAL bytes,
		// where the positions still exist. When the diagnoser is not certain
		// it stays empty and the parser error remains the message.
		return nil, &ManifestLoadError{
			ManifestPath: sourceName,
			Err:          err,
			Kind:         kindName,
			Mismatches:   yamldiag.Diagnose(manifestYamlBytes, manifest.ProtoReflect().Descriptor()),
		}
	}

	// Apply defaults from proto field options
	if err := protodefaults.ApplyDefaults(manifest); err != nil {
		return nil, errors.Wrap(err, "failed to apply default values")
	}

	return manifest, nil
}

// extractKindName reads the top-level 'kind' key from raw manifest YAML. It is the
// in-memory equivalent of catalogkindreflect.ExtractKindFromTargetManifest so byte- and
// path-based loading resolve kinds identically.
func extractKindName(manifestYamlBytes []byte) (string, error) {
	var yamlData map[string]interface{}
	if err := goyaml.Unmarshal(manifestYamlBytes, &yamlData); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal manifest YAML")
	}
	kind, ok := yamlData["kind"]
	if !ok {
		return "", errors.New("key 'kind' not found in manifest YAML")
	}
	kindStr, ok := kind.(string)
	if !ok {
		return "", errors.New("value of 'kind' key is not a string")
	}
	return kindStr, nil
}

func downloadManifest(manifestUrl string) (string, error) {
	// Get the directory to save the downloaded file
	dir, err := workspace.GetManifestDownloadDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get manifest download directory")
	}

	// Generate a new ulid for the file name
	fileName := ulidgen.NewGenerator().Generate().String() + ".yaml"

	filePath := filepath.Join(dir, fileName)

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create file")
	}
	defer out.Close()

	// Download the file
	resp, err := http.Get(manifestUrl)
	if err != nil {
		return "", errors.Wrapf(err, "failed to download manifest from %s", manifestUrl)
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to write manifest to file")
	}

	// Return the absolute path of the downloaded file
	return filePath, nil
}

func isManifestPathUrl(manifestPath string) (bool, error) {
	// Attempt to parse the manifestPath as a URL
	parsedUrl, err := url.Parse(manifestPath)
	if err != nil {
		return false, errors.Wrap(err, "failed to parse manifest path as URL")
	}

	// Check if the URL has a scheme and host
	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return false, nil
	}

	return true, nil
}

// formatUnsupportedResourceError creates a helpful error message when a catalog kind is not supported
func formatUnsupportedResourceError(kindName string) error {
	// Create colored output functions
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	var msg strings.Builder

	msg.WriteString("\n")
	msg.WriteString(red("╔═══════════════════════════════════════════════════════════════════════════════╗") + "\n")
	msg.WriteString(red("║") + bold("                ⚠️  UNSUPPORTED CATALOG KIND                           ") + red("║") + "\n")
	msg.WriteString(red("╚═══════════════════════════════════════════════════════════════════════════════╝") + "\n\n")

	msg.WriteString(yellow("Resource Kind:") + " " + bold(kindName) + "\n\n")

	msg.WriteString(red("❌ This catalog kind is not recognized.\n\n"))

	msg.WriteString(cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	msg.WriteString(bold("                           🔧 HOW TO FIX\n"))
	msg.WriteString(cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))

	msg.WriteString(yellow("1. Check your manifest for typos in the 'kind' field\n\n"))
	msg.WriteString("   Common mistakes:\n")
	msg.WriteString("   • Extra characters (e.g., 'AwsEksCluster" + bold("s") + "')\n")
	msg.WriteString("   • Wrong capitalization (e.g., 'Aws" + bold("EKS") + "Cluster')\n")
	msg.WriteString("   • Misspelled resource name (e.g., 'AwsEks" + bold("Clster") + "')\n\n")

	msg.WriteString(yellow("2. If the kind is correct, update your CLI to the latest version:\n\n"))
	msg.WriteString("   " + green("planton upgrade") + "\n\n")
	msg.WriteString("   Then verify:\n\n")
	msg.WriteString("   " + green("planton version") + "\n\n")

	msg.WriteString(yellow("3. Retry your command\n\n"))

	msg.WriteString(cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))

	msg.WriteString(bold("💡 TIP: ") + "If you're developing a new infra component, ensure the proto files\n")
	msg.WriteString("   are compiled and the CLI binary is rebuilt.\n\n")

	return errors.New(msg.String())
}
