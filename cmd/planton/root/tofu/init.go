package tofu

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/localmodule"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/kubernetes/kubecontext"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/iac/terraform"
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
	flag.HandleFlagErr(err, flag.ModuleDir)

	valueOverrides, err := cmd.Flags().GetStringToString(string(flag.Set))
	flag.HandleFlagErr(err, flag.Set)

	backendTypeString, err := cmd.Flags().GetString(string(flag.BackendType))
	flag.HandleFlagErr(err, flag.BackendType)

	backendConfigList, err := cmd.Flags().GetStringArray(string(flag.BackendConfig))
	flag.HandleFlagErr(err, flag.BackendConfig)

	backendType := tfbackend.BackendTypeFromString(backendTypeString)

	targetManifestPath := inputDir + "/target.yaml"

	if inputDir == "" {
		targetManifestPath, err = cmd.Flags().GetString(string(flag.Manifest))
		flag.Require(err, flag.Manifest, targetManifestPath, "--manifest path/to/manifest.yaml (or --input-dir <dir> holding target.yaml)")
	}

	providerConfig, err := iacinputproviderconfig.GetFromFlagsSimple(cmd.Flags())
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the provider configuration could not be read: %v", err),
			"--provider-config names a file or value the CLI cannot parse",
			"fix the path or the file, or drop --provider-config to use this machine's own credentials",
		)
	}

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the manifest at %s could not be loaded with the --set overrides applied: %v", targetManifestPath, err),
			"either the file does not load into the kind it declares, or a --set path names a field the kind does not have",
			fmt.Sprintf("run `planton validate-manifest -f %s`, and check each --set key against `planton explain <kind>`", targetManifestPath),
		)
	}

	kindName, err := catalogkindreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the manifest's kind could not be determined: %v", err),
			"the manifest loaded, but its apiVersion and kind do not name a catalog kind this CLI knows",
			"check `kind:` and `apiVersion:` against `planton catalog search <name>`",
		)
	}

	noCleanup, _ := cmd.Flags().GetBool(string(flag.NoCleanup))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))

	// Handle --local-module flag: derive module directory from local planton repo
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
		ui.Failure(
			fmt.Sprintf("no OpenTofu module could be resolved for %s: %v", kindName, err),
			"the current directory holds no module, and the published module for this release could not be downloaded or staged",
			"pass --module-dir <dir> to run a module you have on disk, or check network access to downloads.planton.dev",
		)
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

	iacInputYaml, err := iacinput.BuildIacInputYaml(manifestObject, providerConfig)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the IaC input could not be assembled from the manifest: %v", err),
			"the manifest and provider configuration loaded, but could not be combined into the input the module reads",
			"report it at https://github.com/plantonhq/planton/issues with the manifest (secrets removed)",
		)
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the CLI workspace directory could not be prepared: %v", err),
			"the CLI keeps module copies and generated files under ~/.planton and could not create or read that directory",
			"check that your home directory is writable, or set HOME to a writable location",
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

	providerConfigEnvVars, err := tofumodule.GetProviderConfigEnvVars(iacInputYaml, workspaceDir, kubeCtx)
	if err != nil {
		ui.EngineFailure("Provider credentials could not be prepared", err,
			"check the provider configuration's fields against `planton explain <provider connection kind>`")
		os.Exit(1)
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
		if !ui.EngineFailure("OpenTofu Execution Failed", err,
			"Check the module configuration for syntax errors",
			"Ensure all required provider credentials are configured") {
			cliprint.PrintTofuFailure()
		}
		os.Exit(1)
	}
	cliprint.PrintTofuSuccess()
}
