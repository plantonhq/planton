package manifest

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/kustomize/builder"
	"github.com/spf13/cobra"
)

// ResolveManifestPath determines the target manifest path based on flag priority.
// Priority order:
//  1. --clipboard flag (if provided, read manifest from system clipboard)
//  2. --stack-input flag (if provided, extract manifest from target field)
//  3. --manifest flag (if provided, use it directly)
//  4. --input-dir flag (if provided, use inputDir/target.yaml)
//  5. --kustomize-dir + --overlay flags (if both provided, build kustomize manifest)
//  6. Error if none of the above are provided
//
// Returns:
//   - manifestPath: The resolved path to the manifest file
//   - isTemp: Whether the manifest file is temporary and should be cleaned up
//   - error: Any error encountered during resolution
func ResolveManifestPath(cmd *cobra.Command) (string, bool, error) {
	// Priority 1: Check for --clipboard flag (reads manifest from system clipboard)
	clipboardManifest, isTemp, err := resolveFromClipboard(cmd)
	if err != nil {
		return "", false, err
	}
	if clipboardManifest != "" {
		return clipboardManifest, isTemp, nil
	}

	// Priority 2: Check for --stack-input flag (extracts manifest from target field)
	stackInputManifest, isTemp, err := resolveFromStackInput(cmd)
	if err != nil {
		return "", false, err
	}
	if stackInputManifest != "" {
		return stackInputManifest, isTemp, nil
	}

	// Priority 3: Check for --manifest flag
	manifestPath, err := cmd.Flags().GetString(string(flag.Manifest))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get manifest flag")
	}
	if manifestPath != "" {
		return manifestPath, false, nil
	}

	// Priority 4: Check for --input-dir flag
	inputDir, err := cmd.Flags().GetString(string(flag.InputDir))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get input-dir flag")
	}
	if inputDir != "" {
		return inputDir + "/target.yaml", false, nil
	}

	// Priority 5: Check for --kustomize-dir and --overlay flags
	kustomizeDir, err := cmd.Flags().GetString(string(flag.KustomizeDir))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get kustomize-dir flag")
	}

	overlay, err := cmd.Flags().GetString(string(flag.Overlay))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get overlay flag")
	}

	// Both kustomize-dir and overlay must be provided together
	if kustomizeDir != "" && overlay != "" {
		tempManifestPath, err := builder.BuildManifest(kustomizeDir, overlay)
		if err != nil {
			return "", false, errors.Wrap(err, "failed to build kustomize manifest")
		}
		return tempManifestPath, true, nil
	}

	// If only one of kustomize-dir or overlay is provided, that's an error
	if kustomizeDir != "" || overlay != "" {
		return "", false, errors.New("both --kustomize-dir and --overlay flags must be provided together")
	}

	// No valid manifest source provided
	return "", false, errors.New("must provide one of: --clipboard, --stack-input, --manifest, --input-dir, or (--kustomize-dir + --overlay)")
}
