//go:build !codegen
// +build !codegen

package kustomize

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/kustomize/schema"
	"github.com/spf13/cobra"
)

var Schema = &cobra.Command{
	Use:   "schema",
	Short: "Generate the universal kustomize OpenAPI schema for all Planton resource kinds",
	Long: `Generate a single kustomize-compatible OpenAPI schema JSON that covers all
Planton cloud resource kinds. The schema declares strategic merge patch
directives (x-kubernetes-patch-merge-key, x-kubernetes-patch-strategy) for
list fields that should merge by name instead of being replaced.

Only kinds with merge-worthy fields are included. The output is suitable
for the "openapi:" directive in kustomization.yaml.`,
	Run: schemaHandler,
}

func init() {
	Schema.Flags().StringP("output", "o", "", "write schema to file instead of stdout")
}

func schemaHandler(cmd *cobra.Command, args []string) {
	cliprint.PrintStep("Generating universal kustomize OpenAPI schema...")

	data, err := schema.Generate()
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to generate schema: %v", err))
		os.Exit(1)
	}

	output, _ := cmd.Flags().GetString("output")
	if output != "" {
		if err := os.WriteFile(output, append(data, '\n'), 0644); err != nil {
			cliprint.PrintError(fmt.Sprintf("Failed to write file: %v", err))
			os.Exit(1)
		}
		cliprint.PrintSuccess(fmt.Sprintf("Schema written to %s", output))
		return
	}

	fmt.Println(string(data))
}
