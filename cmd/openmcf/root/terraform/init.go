package terraform

import (
	"fmt"
	"os"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	tfpb "github.com/plantonhq/openmcf/apis/org/openmcf/shared/iac/terraform"
	"github.com/plantonhq/openmcf/internal/cli/cliprint"
	"github.com/plantonhq/openmcf/internal/cli/flag"
	"github.com/plantonhq/openmcf/internal/cli/ui"
	"github.com/plantonhq/openmcf/internal/cli/workspace"
	"github.com/plantonhq/openmcf/internal/manifest"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/iac/localmodule"
	"github.com/plantonhq/openmcf/pkg/iac/provisioner"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kubecontext"
	"github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "run terraform init",
	Run:   initHandler,
}

func init() {
	Init.PersistentFlags().StringArray(string(flag.BackendConfig), []string{},
		"Backend configuration key=value pairs")

	Init.PersistentFlags().String(string(flag.BackendType), tfpb.TerraformBackendType_local.String(),
		"Backend type (local, s3, gcs, azurerm, etc.)")

	Init.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.")
}

func initHandler(cmd *cobra.Command, args []string) {
	if err := provisioner.HclBinaryTerraform.CheckAvailable(); err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}

	inputDir, err := cmd.Flags().GetString(string(flag.InputDir))
	flag.HandleFlagErr(err, flag.InputDir)

	moduleDir, err := cmd.Flags().GetString(string(flag.ModuleDir))
	flag.HandleFlagErrAndValue(err, flag.ModuleDir, moduleDir)

	valueOverrides, err := cmd.Flags().GetStringToString(string(flag.Set))
	flag.HandleFlagErr(err, flag.Set)

	backendTypeString, err := cmd.Flags().GetString(string(flag.BackendType))
	flag.HandleFlagErrAndValue(err, flag.BackendType, backendTypeString)

	backendConfigList, err := cmd.Flags().GetStringArray(string(flag.BackendConfig))
	flag.HandleFlagErr(err, flag.BackendConfig)

	backendType := tfbackend.BackendTypeFromString(backendTypeString)

	targetManifestPath := inputDir + "/target.yaml"

	if inputDir == "" {
		targetManifestPath, err = cmd.Flags().GetString(string(flag.Manifest))
		flag.HandleFlagErrAndValue(err, flag.Manifest, targetManifestPath)
	}

	providerConfig, err := stackinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to get provider config: %v", err))
		os.Exit(1)
	}

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		cliprint.PrintError("failed to load manifest file")
		os.Exit(1)
	}

	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to extract kind name from manifest: %v", err))
		os.Exit(1)
	}

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

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

	pathResult, err := tofumodule.GetModulePath(moduleDir, kindName, moduleVersion, noCleanup)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to get terraform module directory: %v", err))
		os.Exit(1)
	}

	if pathResult.ShouldCleanup {
		defer func() {
			if cleanupErr := pathResult.CleanupFunc(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup workspace copy: %v\n", cleanupErr)
			}
		}()
	}

	modulePath := pathResult.ModulePath

	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to build stack input yaml: %v", err))
		os.Exit(1)
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		cliprint.PrintError("failed to get workspace directory")
		os.Exit(1)
	}

	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}

	providerConfigEnvVars, err := tofumodule.GetProviderConfigEnvVars(stackInputYaml, workspaceDir, kubeCtx)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("failed to get credential env vars: %v", err))
		os.Exit(1)
	}

	cliprint.PrintHandoff("Terraform")

	err = tofumodule.Init(
		cmd.Context(),
		"terraform",
		modulePath,
		manifestObject,
		backendType,
		backendConfigList,
		providerConfigEnvVars,
		false, // isReconfigure - not supported in legacy commands
		false,
		nil,
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
