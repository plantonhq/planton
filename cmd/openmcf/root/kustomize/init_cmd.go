//go:build !codegen
// +build !codegen

package kustomize

import (
	"fmt"
	"os"

	"github.com/plantonhq/openmcf/internal/cli/cliprint"
	"github.com/plantonhq/openmcf/pkg/kustomize/initializer"
	"github.com/plantonhq/openmcf/pkg/kustomize/schema"
	"github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "Initialize kustomize schema integration for OpenMCF resources",
	Long: `Write the universal OpenMCF kustomize OpenAPI schema into _kustomize directories
and add the "openapi:" reference to overlay kustomization.yaml files.

Use --dir to initialize a single _kustomize directory, or --scan to walk a
directory tree and initialize every _kustomize directory found.

The command is idempotent: re-running regenerates the schema file (picking up
new kinds or field changes) and skips overlays that already have the openapi
reference.`,
	Run: initHandler,
}

func init() {
	Init.Flags().String("dir", "", "path to a single _kustomize directory to initialize")
	Init.Flags().String("scan", "", "root directory to scan for _kustomize directories")
}

func initHandler(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	scan, _ := cmd.Flags().GetString("scan")

	if dir == "" && scan == "" {
		cliprint.PrintError("Provide either --dir or --scan")
		cmd.Usage()
		os.Exit(1)
	}

	if dir != "" && scan != "" {
		cliprint.PrintError("Provide only one of --dir or --scan, not both")
		os.Exit(1)
	}

	if dir != "" {
		initSingleDir(dir)
		return
	}

	scanAndInitAll(scan)
}

func initSingleDir(dir string) {
	cliprint.PrintStep(fmt.Sprintf("Initializing kustomize schema in %s", dir))

	result, err := initializer.InitDir(dir)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed: %v", err))
		os.Exit(1)
	}

	printResult(result)
}

func scanAndInitAll(rootDir string) {
	cliprint.PrintStep(fmt.Sprintf("Scanning %s for _kustomize directories...", rootDir))

	results, err := initializer.ScanAndInit(rootDir)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed: %v", err))
		os.Exit(1)
	}

	if len(results) == 0 {
		cliprint.PrintWarning("No _kustomize directories found")
		return
	}

	for _, r := range results {
		printResult(r)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Initialized %d _kustomize directories", len(results)))
}

func printResult(r *initializer.InitResult) {
	if r.SchemaWritten {
		cliprint.PrintSuccess(fmt.Sprintf("Wrote %s to %s", schema.SchemaFileName, r.Dir))
	}
	for _, overlay := range r.OverlaysUpdated {
		cliprint.PrintSuccess(fmt.Sprintf("  Updated overlays/%s/kustomization.yaml", overlay))
	}
	if len(r.OverlaysUpdated) == 0 && r.SchemaWritten {
		cliprint.PrintInfo("  All overlays already configured")
	}
}
