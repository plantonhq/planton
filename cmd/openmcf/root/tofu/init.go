package tofu

import (
	"fmt"
	"os"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/iac/terraform"
	"github.com/plantonhq/openmcf/internal/cli/cliprint"
	"github.com/plantonhq/openmcf/internal/cli/flag"
	"github.com/plantonhq/openmcf/internal/cli/ui"
	"github.com/plantonhq/openmcf/internal/cli/workspace"
	"github.com/plantonhq/openmcf/internal/manifest"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/iac/localmodule"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kubecontext"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "run tofu init",
	Run:   initHandler,
}

func init() {
	Init.PersistentFlags().StringArray(string(flag.BackendConfig), []string{},
		"Configuration to be merged with what is in the\n                          "+
			"configuration file's 'backend' block. "+
			"This can be\n                          either a path to an HCL file with key/value\n                          "+
			"assignments (same format as terraform.tfvars) or a\n                          'key=value' format, and can be "+
			"specified multiple\n                          times. The backend type must be in the "+
			"configuration\n                          itself.")

	Init.PersistentFlags().String(string(flag.BackendType), terraform.TerraformBackendType_local.String(),
		"Specifies the backend type that Terraform will use to store and manage the state.\n"+
			"This must match one of the supported Terraform backends, such as 'local', 's3', 'gcs',\n"+
			"'azurerm', 'remote', 'consul', 'http', 'etcdv3', 'manta', 'swift', 'artifactory', or\n"+
			"'oss'. By default, it uses 'local', which stores the Terraform state on the local\n"+
			"filesystem.\n\n"+
			"If you choose a different backend (e.g., 's3'), you can then supply additional\n"+
			"configuration parameters using the '--backend-config' flag. For example, when using\n"+
			"'s3', you might provide a bucket name, key, region, and a DynamoDB table for locking,\n"+
			"either via a path to an HCL file or via key-value pairs.\n\n"+
			"This option can be used multiple times if you need to override certain backend\n"+
			"attributes. The backend type itself, however, must be declared in your Terraform\n"+
			"configuration using a 'terraform { backend \"<type>\" {} }' block. The '--backend-type'\n"+
			"flag will then instruct Terraform which backend configuration block to use.\n\n"+
			"Example:\n"+
			"  --backend-type=s3 --backend-config=bucket=my-terraform-bucket --backend-config=key=state.tfstate")

	Init.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")

}

func initHandler(cmd *cobra.Command, args []string) {
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
		log.Fatalf("failed to get provider config: %v", err)
	}

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		log.Fatalf("failed to override values in target manifest file")
	}

	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		log.Fatalf("failed to extract kind name from manifest proto %v", err)
	}

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

	// Handle --local-module flag: derive module directory from local openmcf repo
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
		log.Fatalf("failed to get tofu module directory %v", err)
	}

	// Setup cleanup to run after execution
	if pathResult.ShouldCleanup {
		defer func() {
			if cleanupErr := pathResult.CleanupFunc(); cleanupErr != nil {
				log.Warnf("failed to cleanup workspace copy: %v", cleanupErr)
			}
		}()
	}

	tofuModulePath := pathResult.ModulePath

	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		log.Fatalf("failed to build stack input yaml %v", err)
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		log.Fatalf("failed to get workspace directory")
	}

	// Resolve kube context: flag takes priority over manifest label
	kubeCtx, _ := cmd.Flags().GetString(string(flag.KubeContext))
	if kubeCtx == "" {
		kubeCtx = kubecontext.ExtractFromManifest(manifestObject)
	}
	if kubeCtx != "" {
		cliprint.PrintInfo(fmt.Sprintf("Using kubectl context: %s", kubeCtx))
	}

	providerConfigEnvVars, err := tofumodule.GetProviderConfigEnvVars(stackInputYaml, workspaceDir, kubeCtx)
	if err != nil {
		log.Fatalf("failed to get credential env vars %v", err)
	}

	cliprint.PrintHandoff("OpenTofu")

	err = tofumodule.Init(
		cmd.Context(),
		"tofu",
		tofuModulePath,
		manifestObject,
		backendType,
		backendConfigList,
		providerConfigEnvVars,
		false, // isReconfigure - not supported in legacy commands
		false,
		nil,
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
