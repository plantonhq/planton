package root

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/plantonhq/planton/internal/cli/iacrunner"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/cli/prompt"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/localmodule"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumistack"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "initialize backend/stack using the provisioner specified in manifest",
	Long: `Initialize infrastructure backend or stack by automatically routing to the appropriate provisioner
(Pulumi, Tofu, or Terraform) based on the manifest label 'planton.dev/provisioner'.

If the provisioner label is not present, you will be prompted to select one interactively.`,
	Example: `
	# Initialize from clipboard (manifest content already copied)
	planton init --clipboard
	planton init -c

	# Initialize with manifest file
	planton init -f manifest.yaml
	planton init --manifest manifest.yaml

	# Initialize with stack input file (extracts manifest from target field)
	planton init -i stack-input.yaml

	# Initialize with kustomize
	planton init --kustomize-dir _kustomize --overlay prod

	# Initialize with tofu-specific backend config
	planton init -f manifest.yaml --backend-type s3 --backend-config bucket=my-bucket
	`,
	Run: initHandler,
}

func init() {
	iacflags.AddManifestSourceFlags(Init)
	iacflags.AddProviderConfigFlags(Init)
	iacflags.AddExecutionFlags(Init)
	iacflags.AddPulumiFlags(Init)
	iacflags.AddTofuInitFlags(Init)
}

func initHandler(cmd *cobra.Command, args []string) {
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
		cliprint.PrintError(fmt.Sprintf("Failed to resolve manifest: %v", err))
		os.Exit(1)
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

	// Load manifest to extract provisioner
	cliprint.PrintStep("Detecting provisioner...")
	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to load manifest: %v", err))
		os.Exit(1)
	}

	// Extract provisioner from manifest
	provType, err := provisioner.ExtractFromManifest(manifestObject)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Invalid provisioner in manifest: %v", err))
		os.Exit(1)
	}

	// If provisioner not specified in manifest, prompt user
	if provType == provisioner.ProvisionerTypeUnspecified {
		cliprint.PrintInfo("Provisioner not specified in manifest")
		provType, err = prompt.PromptForProvisioner()
		if err != nil {
			cliprint.PrintError(fmt.Sprintf("Failed to get provisioner: %v", err))
			os.Exit(1)
		}
	}

	cliprint.PrintSuccess(fmt.Sprintf("Using provisioner: %s", provType.String()))

	// Resolve kube context: flag takes priority over manifest label
	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}

	// Handle --local-module flag: derive module directory from local planton repo
	localModule, _ := cmd.Flags().GetBool(string(flag.LocalModule))
	if localModule {
		var iacProv shared.IacProvisioner
		switch provType {
		case provisioner.ProvisionerTypePulumi:
			iacProv = shared.IacProvisioner_pulumi
		case provisioner.ProvisionerTypeTofu, provisioner.ProvisionerTypeTerraform:
			iacProv = shared.IacProvisioner_terraform
		}
		moduleDir, err = localmodule.GetModuleDir(targetManifestPath, cmd, iacProv)
		if err != nil {
			if lmErr, ok := err.(*localmodule.Error); ok {
				lmErr.PrintError()
			} else {
				cliprint.PrintError(err.Error())
			}
			os.Exit(1)
		}
	}

	// Prepare provider config
	cliprint.PrintStep("Preparing execution...")
	providerConfig, err := stackinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get provider config: %v", err))
		os.Exit(1)
	}
	cliprint.PrintSuccess("Execution prepared")

	// Route to appropriate provisioner
	switch provType {
	case provisioner.ProvisionerTypePulumi:
		initWithPulumi(cmd, moduleDir, targetManifestPath, valueOverrides)
	case provisioner.ProvisionerTypeTofu:
		initWithTofu(cmd, moduleDir, targetManifestPath, valueOverrides, kubeCtx, manifestObject, providerConfig)
	case provisioner.ProvisionerTypeTerraform:
		initWithTerraform(cmd, moduleDir, targetManifestPath, valueOverrides, kubeCtx, manifestObject, providerConfig)
	default:
		cliprint.PrintError("Unknown provisioner type")
		os.Exit(1)
	}
}

func initWithPulumi(cmd *cobra.Command, moduleDir, targetManifestPath string, valueOverrides map[string]string) {
	// Stack can be provided via flag or extracted from manifest
	stackFqdn, err := cmd.Flags().GetString(string(flag.Stack))
	flag.HandleFlagErr(err, flag.Stack)

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

	err = pulumistack.Init(moduleDir, stackFqdn, targetManifestPath, valueOverrides, moduleVersion, noCleanup)
	if err != nil {
		cliprint.PrintPulumiFailure()
		os.Exit(1)
	}
	cliprint.PrintPulumiSuccess()
}

func initWithTofu(cmd *cobra.Command, moduleDir, targetManifestPath string, valueOverrides map[string]string,
	kubeContext string, manifestObject proto.Message, providerConfig *stackinputproviderconfig.ProviderConfig) {

	backendTypeString, err := cmd.Flags().GetString(string(flag.BackendType))
	flag.HandleFlagErrAndValue(err, flag.BackendType, backendTypeString)

	backendConfigList, err := cmd.Flags().GetStringArray(string(flag.BackendConfig))
	flag.HandleFlagErr(err, flag.BackendConfig)

	isReconfigure, _ := cmd.Flags().GetBool(string(flag.Reconfigure))

	backendType := tfbackend.BackendTypeFromString(backendTypeString)

	// Extract kind name for module path resolution
	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to extract kind name from manifest proto: %v", err))
		os.Exit(1)
	}

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

	pathResult, err := tofumodule.GetModulePath(moduleDir, kindName, moduleVersion, noCleanup)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get tofu module directory: %v", err))
		os.Exit(1)
	}

	// Setup cleanup to run after execution
	if pathResult.ShouldCleanup {
		defer func() {
			if cleanupErr := pathResult.CleanupFunc(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup workspace copy: %v\n", cleanupErr)
			}
		}()
	}

	tofuModulePath := pathResult.ModulePath

	// Build stack input YAML
	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to build stack input yaml: %v", err))
		os.Exit(1)
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		cliprint.PrintError("Failed to get workspace directory")
		os.Exit(1)
	}

	providerConfigEnvVars, err := tofumodule.GetProviderConfigEnvVars(stackInputYaml, workspaceDir, kubeContext)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get credential env vars: %v", err))
		os.Exit(1)
	}

	// Display backend configuration if available (before handoff)
	if backendInfo := iacrunner.ExtractBackendConfigForDisplay(manifestObject, "tofu"); backendInfo != nil {
		cliprint.PrintBackendConfig(backendInfo.BackendType, backendInfo.Bucket, backendInfo.Key)
	}

	// Display module path
	cliprint.PrintModulePath(tofuModulePath)

	cliprint.PrintHandoff("OpenTofu")

	err = tofumodule.Init(
		cmd.Context(),
		"tofu",
		tofuModulePath,
		manifestObject,
		backendType,
		backendConfigList,
		providerConfigEnvVars,
		isReconfigure,
		false,
		nil,
	)
	if err != nil {
		cliprint.PrintTofuFailure()
		os.Exit(1)
	}
	cliprint.PrintTofuSuccess()
}

func initWithTerraform(cmd *cobra.Command, moduleDir, targetManifestPath string, valueOverrides map[string]string,
	kubeContext string, manifestObject proto.Message, providerConfig *stackinputproviderconfig.ProviderConfig) {

	backendTypeString, err := cmd.Flags().GetString(string(flag.BackendType))
	flag.HandleFlagErrAndValue(err, flag.BackendType, backendTypeString)

	backendConfigList, err := cmd.Flags().GetStringArray(string(flag.BackendConfig))
	flag.HandleFlagErr(err, flag.BackendConfig)

	isReconfigure, _ := cmd.Flags().GetBool(string(flag.Reconfigure))

	backendType := tfbackend.BackendTypeFromString(backendTypeString)

	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to extract kind name from manifest proto: %v", err))
		os.Exit(1)
	}

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

	pathResult, err := tofumodule.GetModulePath(moduleDir, kindName, moduleVersion, noCleanup)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get terraform module directory: %v", err))
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
		cliprint.PrintError(fmt.Sprintf("Failed to build stack input yaml: %v", err))
		os.Exit(1)
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		cliprint.PrintError("Failed to get workspace directory")
		os.Exit(1)
	}

	providerConfigEnvVars, err := tofumodule.GetProviderConfigEnvVars(stackInputYaml, workspaceDir, kubeContext)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get credential env vars: %v", err))
		os.Exit(1)
	}

	// Display backend configuration if available (before handoff)
	if backendInfo := iacrunner.ExtractBackendConfigForDisplay(manifestObject, "terraform"); backendInfo != nil {
		cliprint.PrintBackendConfig(backendInfo.BackendType, backendInfo.Bucket, backendInfo.Key)
	}

	// Display module path
	cliprint.PrintModulePath(modulePath)

	cliprint.PrintHandoff("Terraform")

	err = tofumodule.Init(
		cmd.Context(),
		"terraform",
		modulePath,
		manifestObject,
		backendType,
		backendConfigList,
		providerConfigEnvVars,
		isReconfigure,
		false,
		nil,
	)
	if err != nil {
		cliprint.PrintTerraformFailure()
		os.Exit(1)
	}
	cliprint.PrintTerraformSuccess()
}
