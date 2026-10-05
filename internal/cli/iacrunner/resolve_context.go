package iacrunner

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/cli/prompt"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/iacinput/providerdetect"
	"github.com/plantonhq/planton/pkg/iac/localmodule"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/spf13/cobra"
)

// ResolveContext reads command flags, resolves manifest from various sources,
// validates the manifest, detects provisioner, and returns a ready-to-execute Context.
func ResolveContext(cmd *cobra.Command) (*Context, error) {
	ctx := &Context{}

	// Get module directory. Empty is a legal value: it means the user made no
	// explicit choice, and the module resolvers quietly probe the current
	// directory before falling through to the download/staging paths.
	moduleDir, err := cmd.Flags().GetString(string(flag.ModuleDir))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get module-dir flag")
	}
	ctx.ModuleDir = moduleDir

	// Get value overrides
	valueOverrides, err := cmd.Flags().GetStringToString(string(flag.Set))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get set flag")
	}
	ctx.ValueOverrides = valueOverrides

	// Check which manifest source is being used for informative messages
	kustomizeDir, _ := cmd.Flags().GetString(string(flag.KustomizeDir))
	overlay, _ := cmd.Flags().GetString(string(flag.Overlay))

	if kustomizeDir != "" && overlay != "" {
		cliprint.PrintStep(fmt.Sprintf("Building manifest from kustomize overlay: %s", overlay))
	} else {
		cliprint.PrintStep("Loading manifest...")
	}

	// Resolve manifest path with priority: --iac-input > --manifest > --input-dir > --kustomize-dir + --overlay
	targetManifestPath, isTemp, err := climanifest.ResolveManifestPath(cmd)
	if err != nil {
		// Check for clipboard-specific errors and display beautifully
		if climanifest.HandleClipboardError(err) {
			// Return error to signal failure, but error display already handled
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to resolve manifest")
	}
	if isTemp {
		ctx.AddCleanupFunc(func() { os.Remove(targetManifestPath) })
	}
	ctx.ManifestPath = targetManifestPath

	cliprint.PrintSuccess("Manifest loaded")

	// Apply value overrides if any (creates new temp file if overrides exist)
	if len(valueOverrides) > 0 {
		cliprint.PrintStep(fmt.Sprintf("Applying %d field override(s)...", len(valueOverrides)))
	}

	finalManifestPath, isTempOverrides, err := manifest.ApplyOverridesToFile(targetManifestPath, valueOverrides)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply overrides to manifest")
	}
	if isTempOverrides {
		ctx.AddCleanupFunc(func() { os.Remove(finalManifestPath) })
		ctx.ManifestPath = finalManifestPath
		cliprint.PrintSuccess("Overrides applied")
	}

	// Validate manifest before proceeding (after overrides are applied)
	cliprint.PrintStep("Validating manifest...")
	if err := manifest.Validate(ctx.ManifestPath); err != nil {
		// Check for manifest load errors (proto unmarshaling) and display beautifully
		if manifest.HandleManifestLoadError(err) {
			return nil, err
		}
		return nil, errors.Wrap(err, "manifest validation failed")
	}
	cliprint.PrintSuccess("Manifest validated")

	// Load manifest to extract provisioner
	cliprint.PrintStep("Detecting provisioner...")
	manifestObject, err := manifest.LoadWithOverrides(ctx.ManifestPath, valueOverrides)
	if err != nil {
		// Check for manifest load errors (proto unmarshaling) and display beautifully
		if manifest.HandleManifestLoadError(err) {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to load manifest")
	}
	ctx.ManifestObject = manifestObject

	// The manifest's provisioner, else the kind's sole declared engine, else ask among the
	// engines the kind runs on.
	provType, err := provisioner.ForManifest(manifestObject)
	if err != nil {
		return nil, errors.Wrap(err, "invalid provisioner in manifest")
	}
	if provType == provisioner.ProvisionerTypeUnspecified {
		cliprint.PrintInfo("Provisioner not specified in manifest")
		allowed, err := provisioner.AllowedForManifest(manifestObject)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read the engines this kind runs on")
		}
		provType, err = prompt.PromptForProvisioner(allowed...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get provisioner")
		}
	} else if note := provisioner.SoleEngineNote(manifestObject, provType); note != "" {
		cliprint.PrintInfo(note)
	}
	ctx.ProvisionerType = provType

	cliprint.PrintSuccess(fmt.Sprintf("Using provisioner: %s", provType.String()))

	// Resolve kube context: flag takes priority over manifest annotation
	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}
	ctx.KubeContext = kubeCtx

	// Handle --local-module flag: derive module directory from local planton repo
	localModule, _ := cmd.Flags().GetBool(string(flag.LocalModule))
	if localModule {
		iacProv := provType.ModuleFamily()
		derivedModuleDir, err := localmodule.GetModuleDir(ctx.ManifestPath, cmd, iacProv)
		if err != nil {
			if lmErr, ok := err.(*localmodule.Error); ok {
				lmErr.PrintError()
			}
			return nil, errors.Wrap(err, "failed to derive module directory from local repo")
		}
		ctx.ModuleDir = derivedModuleDir
	}

	// Get IaC input file path if provided
	iacInputFilePath, _ := cmd.Flags().GetString(string(flag.IacInput))
	if iacInputFilePath != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using IaC input file: %s", iacInputFilePath))
	}
	ctx.IacInputFilePath = iacInputFilePath

	// Get other execution flags
	ctx.ModuleVersion, _ = cmd.Flags().GetString(string(flag.ModuleVersion))
	ctx.NoCleanup, _ = cmd.Flags().GetBool(string(flag.NoCleanup))
	ctx.ShowDiff, _ = cmd.Flags().GetBool(string(flag.Diff))

	// Detect required provider from manifest
	cliprint.PrintStep("Detecting required provider...")
	manifestBytes, err := os.ReadFile(ctx.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read manifest for provider detection")
	}

	detectionResult, err := providerdetect.DetectFromManifest(manifestBytes)
	if err != nil {
		// Show guidance for kind detection failure
		cliprint.PrintKindDetectionError(providerdetect.KindDetectionErrorGuidance())
		return nil, err
	}

	ctx.DetectionResult = detectionResult

	// Show detected resource information
	cliprint.PrintSuccess(fmt.Sprintf("Detected resource: %s (%s provider)",
		detectionResult.KindName, providerdetect.ProviderDisplayName(detectionResult.Provider)))

	// Get provider config from flags
	cliprint.PrintStep("Preparing execution...")
	providerConfig, err := iacinputproviderconfig.GetFromFlags(cmd.Flags(), detectionResult)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get provider config from flags")
	}

	// Validate provider config if explicitly provided via -p flag
	if providerConfig.Path != "" {
		if err := providerdetect.ValidateProviderConfig(providerConfig.Path, detectionResult.Provider); err != nil {
			guidance := providerdetect.InvalidProviderConfigGuidance(detectionResult, err)
			cliprint.PrintInvalidProviderConfig("Invalid provider config", guidance)
			return nil, errors.Wrap(err, "provider config validation failed")
		}
		cliprint.PrintProviderConfigLoaded(providerdetect.ProviderDisplayName(detectionResult.Provider))
	}

	ctx.ProviderConfig = providerConfig
	cliprint.PrintSuccess("Execution prepared")

	return ctx, nil
}
