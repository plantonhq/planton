package tofu

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/localmodule"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/iac/terraform"
	"github.com/spf13/cobra"
)

var Refresh = &cobra.Command{
	Use:   "refresh",
	Short: "run tofu refresh",
	Run:   refreshHandler,
}

func init() {
	Refresh.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")
	Refresh.PersistentFlags().Bool(string(flag.NoCleanup), false, "Do not cleanup the workspace copy after execution (keeps cloned modules)")
}

func refreshHandler(cmd *cobra.Command, args []string) {
	moduleDir, err := cmd.Flags().GetString(string(flag.ModuleDir))
	flag.HandleFlagErr(err, flag.ModuleDir)

	valueOverrides, err := cmd.Flags().GetStringToString(string(flag.Set))
	flag.HandleFlagErr(err, flag.Set)

	// Check which manifest source is being used for informative messages
	kustomizeDir, _ := cmd.Flags().GetString(string(flag.KustomizeDir))
	overlay, _ := cmd.Flags().GetString(string(flag.Overlay))

	if kustomizeDir != "" && overlay != "" {
		cliprint.PrintStep(fmt.Sprintf("Building manifest from kustomize overlay: %s", overlay))
	} else {
		cliprint.PrintStep("Loading manifest...")
	}

	// Resolve manifest path with priority: --manifest > --input-dir > --kustomize-dir + --overlay
	targetManifestPath, isTemp, err := climanifest.ResolveManifestPath(cmd)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("no manifest could be resolved: %v", err),
			"the command needs a resource manifest and none of --manifest, --input-dir, or --kustomize-dir with --overlay supplied one it could read",
			"pass --manifest path/to/manifest.yaml",
		)
	}
	if isTemp {
		defer os.Remove(targetManifestPath)
	}

	cliprint.PrintSuccess("Manifest loaded")

	// Apply value overrides if any (creates new temp file if overrides exist)
	if len(valueOverrides) > 0 {
		cliprint.PrintStep(fmt.Sprintf("Applying %d field override(s)...", len(valueOverrides)))
	}

	finalManifestPath, isTempOverrides, err := manifest.ApplyOverridesToFile(targetManifestPath, valueOverrides)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if isTempOverrides {
		defer os.Remove(finalManifestPath)
		targetManifestPath = finalManifestPath
		cliprint.PrintSuccess("Overrides applied")
	}

	// Validate manifest before proceeding (after overrides are applied)
	cliprint.PrintStep("Validating manifest...")
	if err := manifest.Validate(targetManifestPath); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cliprint.PrintSuccess("Manifest validated")

	// Handle --local-module flag: derive module directory from local planton repo
	localModule, _ := cmd.Flags().GetBool(string(flag.LocalModule))
	if localModule {
		var err error
		moduleDir, err = localmodule.GetModuleDir(targetManifestPath, cmd, shared.IacProvisioner_terraform)
		if err != nil {
			if lmErr, ok := err.(*localmodule.Error); ok {
				lmErr.PrintError()
			} else {
				cliprint.PrintError(err.Error())
			}
			os.Exit(1)
		}
	}

	cliprint.PrintStep("Preparing OpenTofu execution...")
	providerConfig, err := iacinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the provider configuration could not be read: %v", err),
			"--provider-config names a file or value the CLI cannot parse",
			"fix the path or the file, or drop --provider-config to use this machine's own credentials",
		)
	}
	cliprint.PrintSuccess("Execution prepared")

	// Load manifest to extract kube context
	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the manifest at %s could not be loaded: %v", targetManifestPath, err),
			"the file is readable but its contents do not load into the kind it declares",
			fmt.Sprintf("run `planton validate-manifest -f %s` for the field-level report", targetManifestPath),
		)
	}

	// Resolve kube context: flag takes priority over manifest label
	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}

	cliprint.PrintHandoff("OpenTofu")

	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))
	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))

	err = tofumodule.RunCommand(
		"tofu",
		moduleDir,
		targetManifestPath,
		terraform.TerraformOperationType_refresh,
		valueOverrides,
		true,
		false,
		false, // isReconfigure - not supported in legacy commands
		moduleVersion,
		noCleanup,
		kubeCtx,
		providerConfig,
		nil, // backendConfig - uses manifest annotations for direct commands
	)
	if err != nil {
		if !ui.EngineFailure("OpenTofu Execution Failed", err,
			"Check the module configuration for syntax errors",
			"Ensure all required provider credentials are configured") {
			cliprint.PrintTofuFailure()
		}
		os.Exit(1)
	}
	cliprint.PrintTofuSuccess()
}
