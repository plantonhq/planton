package tofu

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/localmodule"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Apply = &cobra.Command{
	Use:   "apply",
	Short: "run tofu apply",
	Run:   applyHandler,
}

func init() {
	Apply.PersistentFlags().Bool(string(flag.AutoApprove), false, "Skip interactive approval of plan before applying")
	Apply.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")
	Apply.PersistentFlags().Bool(string(flag.NoCleanup), false, "Do not cleanup the workspace copy after execution (keeps cloned modules)")
}

func applyHandler(cmd *cobra.Command, args []string) {
	isAutoApprove, err := cmd.Flags().GetBool(string(flag.AutoApprove))
	flag.HandleFlagErr(err, flag.AutoApprove)

	moduleDir, err := cmd.Flags().GetString(string(flag.ModuleDir))
	flag.HandleFlagErrAndValue(err, flag.ModuleDir, moduleDir)

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
		log.Fatalf("failed to resolve manifest: %v", err)
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
	providerConfig, err := stackinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		log.Fatalf("failed to get provider config: %v", err)
	}
	cliprint.PrintSuccess("Execution prepared")

	// Load manifest to extract kube context
	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		log.Fatalf("Failed to load manifest: %v", err)
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
		terraform.TerraformOperationType_apply,
		valueOverrides,
		isAutoApprove,
		false,
		false, // isReconfigure - not supported in legacy commands
		moduleVersion,
		noCleanup,
		kubeCtx,
		providerConfig,
		nil, // backendConfig - uses manifest labels for direct commands
	)
	if err != nil {
		ui.ErrorWithoutExit("OpenTofu Execution Failed", err.Error(),
			"Check the module configuration for syntax errors",
			"Ensure all required provider credentials are configured")
		cliprint.PrintTofuFailure()
		os.Exit(1)
	}
	cliprint.PrintTofuSuccess()
}
