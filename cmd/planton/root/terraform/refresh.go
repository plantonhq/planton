package terraform

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
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/spf13/cobra"
)

var Refresh = &cobra.Command{
	Use:   "refresh",
	Short: "run terraform refresh",
	Run:   refreshHandler,
}

func init() {
	Refresh.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.")
	Refresh.PersistentFlags().Bool(string(flag.NoCleanup), false, "Do not cleanup the workspace copy after execution")
}

func refreshHandler(cmd *cobra.Command, args []string) {
	if err := provisioner.HclBinaryTerraform.CheckAvailable(); err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}

	moduleDir, err := cmd.Flags().GetString(string(flag.ModuleDir))
	flag.HandleFlagErrAndValue(err, flag.ModuleDir, moduleDir)

	valueOverrides, err := cmd.Flags().GetStringToString(string(flag.Set))
	flag.HandleFlagErr(err, flag.Set)

	kustomizeDir, _ := cmd.Flags().GetString(string(flag.KustomizeDir))
	overlay, _ := cmd.Flags().GetString(string(flag.Overlay))

	if kustomizeDir != "" && overlay != "" {
		cliprint.PrintStep(fmt.Sprintf("Building manifest from kustomize overlay: %s", overlay))
	} else {
		cliprint.PrintStep("Loading manifest...")
	}

	targetManifestPath, isTemp, err := climanifest.ResolveManifestPath(cmd)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to resolve manifest: %v", err))
		os.Exit(1)
	}
	if isTemp {
		defer os.Remove(targetManifestPath)
	}

	cliprint.PrintSuccess("Manifest loaded")

	if len(valueOverrides) > 0 {
		cliprint.PrintStep(fmt.Sprintf("Applying %d field override(s)...", len(valueOverrides)))
	}

	finalManifestPath, isTempOverrides, err := manifest.ApplyOverridesToFile(targetManifestPath, valueOverrides)
	if err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}
	if isTempOverrides {
		defer os.Remove(finalManifestPath)
		targetManifestPath = finalManifestPath
		cliprint.PrintSuccess("Overrides applied")
	}

	cliprint.PrintStep("Validating manifest...")
	if err := manifest.Validate(targetManifestPath); err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}
	cliprint.PrintSuccess("Manifest validated")

	localModule, _ := cmd.Flags().GetBool(string(flag.LocalModule))
	if localModule {
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

	cliprint.PrintStep("Preparing Terraform execution...")
	providerConfig, err := stackinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to get provider config: %v", err))
		os.Exit(1)
	}
	cliprint.PrintSuccess("Execution prepared")

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to load manifest: %v", err))
		os.Exit(1)
	}

	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}

	cliprint.PrintHandoff("Terraform")

	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))
	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))

	err = tofumodule.RunCommand(
		"terraform",
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
		nil, // backendConfig - uses manifest labels for direct commands
	)
	if err != nil {
		ui.ErrorWithoutExit("Terraform Execution Failed", err.Error(),
			"Check the module configuration for syntax errors",
			"Ensure all required provider credentials are configured")
		cliprint.PrintTerraformFailure()
		os.Exit(1)
	}
	cliprint.PrintTerraformSuccess()
}
