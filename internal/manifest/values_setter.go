package manifest

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest/manifestprotobuf"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

func LoadWithOverrides(manifestPath string, valueOverrides map[string]string) (proto.Message, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		// Preserve ManifestLoadError type for beautiful error display
		if IsManifestLoadError(err) {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to load manifest")
	}
	for key, value := range valueOverrides {
		manifest, err = manifestprotobuf.SetProtoField(manifest, key, value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set %s=%s", key, value)
		}
	}
	return manifest, nil
}

// ApplyOverridesToFile loads a manifest, applies overrides, and writes to a new temp file.
// Returns the path to the new temp file and whether it's a temp file that needs cleanup.
// If no overrides are provided, returns the original path unchanged.
func ApplyOverridesToFile(manifestPath string, valueOverrides map[string]string) (string, bool, error) {
	// If no overrides, return original path
	if len(valueOverrides) == 0 {
		return manifestPath, false, nil
	}

	// Load manifest with overrides applied
	manifest, err := LoadWithOverrides(manifestPath, valueOverrides)
	if err != nil {
		return "", false, formatOverrideError(err)
	}

	// Convert to YAML
	jsonBytes, err := protojson.Marshal(manifest)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to marshal manifest to json")
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to convert json to yaml")
	}

	// Write to temp file
	tempFile, err := os.CreateTemp("", "manifest-with-overrides-*.yaml")
	if err != nil {
		return "", false, errors.Wrap(err, "failed to create temp file for overrides")
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(yamlBytes); err != nil {
		os.Remove(tempFile.Name())
		return "", false, errors.Wrap(err, "failed to write manifest with overrides")
	}

	return tempFile.Name(), true, nil
}

func formatOverrideError(err error) error {
	// Create colored output functions
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Build the error message
	var msg strings.Builder

	msg.WriteString("\n")
	msg.WriteString(red("╔═══════════════════════════════════════════════════════════════════════════════╗") + "\n")
	msg.WriteString(red("║") + bold("                    ❌  FIELD OVERRIDE FAILED                                  ") + red("║") + "\n")
	msg.WriteString(red("╚═══════════════════════════════════════════════════════════════════════════════╝") + "\n\n")

	msg.WriteString(yellow("⚠️  Override Error:\n\n"))

	// Display the actual error (clean it up)
	errMsg := err.Error()
	errMsg = strings.TrimPrefix(errMsg, "failed to load manifest with overrides: ")
	errMsg = strings.TrimSpace(errMsg)
	msg.WriteString(cyan("   "+errMsg) + "\n\n")

	// Provide helpful guidance
	msg.WriteString(bold("💡 Common Issues:\n\n"))

	if strings.Contains(errMsg, "unsupported field type") {
		msg.WriteString("   You're trying to override a complex/nested field with a simple value.\n")
		msg.WriteString("   For nested fields, you need to specify the complete path:\n\n")
		msg.WriteString(green("   # Wrong:\n"))
		msg.WriteString(green("   --set spec.container.app.image=nginx\n\n"))
		msg.WriteString(green("   # Correct:\n"))
		msg.WriteString(green("   --set spec.container.app.image.repo=nginx\n"))
		msg.WriteString(green("   --set spec.container.app.image.tag=latest\n"))
	} else if strings.Contains(errMsg, "field not found") {
		msg.WriteString("   The field path you specified doesn't exist in the manifest.\n")
		msg.WriteString("   Check the field name spelling and nesting level.\n")
	} else {
		msg.WriteString("   Check your --set flag syntax and ensure the field path is correct.\n")
		msg.WriteString("   Field paths must use dot notation (e.g., spec.container.app.image.repo).\n")
	}

	msg.WriteString("\n")
	msg.WriteString(bold("📋 Helpful Commands:\n\n"))
	msg.WriteString("   • View manifest structure:  " + cyan("planton load-manifest --kustomize-dir _kustomize --overlay prod") + "\n")
	msg.WriteString("   • See available fields:     " + cyan("planton load-manifest --help") + "\n")
	msg.WriteString("\n")

	msg.WriteString(bold("📚 Documentation: ") + cyan("https://github.com/plantonhq/planton/tree/main/apis\n"))
	msg.WriteString("\n")

	return errors.New(msg.String())
}
