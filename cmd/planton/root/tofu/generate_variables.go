package tofu

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var GenerateVariables = &cobra.Command{
	Use:   "generate-variables <deployment-component>",
	Short: "Generate Terraform variables for a specified deployment component",
	Long: `The "generate-variables" command takes a specified planton 
deployment component type (e.g., "S3Bucket", "RedisKubernetes") and generates 
Terraform variable definitions (variables.tf) and a corresponding 
terraform.tfvars file.

This command instantiates an empty object of the specified component kind 
under the hood, and then converts that empty object into a Terraform-compatible 
variables file. These variables can then be passed into Terraform modules, 
streamlining infrastructure provisioning and ensuring a consistent, 
declarative workflow.`,
	Example: `
  # Generate variables for an S3Bucket deployment component
  planton tofu generate-variables S3Bucket

  # Generate variables for a RedisKubernetes deployment component
  planton tofu generate-variables RedisKubernetes
`,
	Args: cobra.ExactArgs(1), // "s3-bucket", "redis-kubernetes", etc.
	Run:  generateVariablesHandler,
}

func init() {
	GenerateVariables.Flags().String(string(flag.OutputFile), "", "output file for Terraform variables")
}

func generateVariablesHandler(cmd *cobra.Command, args []string) {
	kindName := args[0]

	outputFile, err := cmd.Flags().GetString(string(flag.OutputFile))
	flag.HandleFlagErr(err, flag.OutputFile)

	cloudResourceKind := crkreflect.KindFromString(kindName)

	manifestObject := crkreflect.ToMessageMap[cloudResourceKind]

	if manifestObject == nil {
		log.Fatalf("proto message not found for %s cloudResourceKind", cloudResourceKind.String())
	}

	variablesTfContent, err := generators.ProtoToVariablesTF(manifestObject)
	if err != nil {
		log.Fatal("failed to generate Terraform variables: ", err)
	}
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(variablesTfContent), 0644); err != nil {
			log.Fatalf("failed to write Terraform variables to file %s: %v", outputFile, err)
		}
		log.Infof("Terraform variables written to file %s", outputFile)
	} else {
		println(variablesTfContent)
	}
}
