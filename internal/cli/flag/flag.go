package flag

import (
	"fmt"

	"github.com/plantonhq/planton/internal/cli/ui"
)

type Flag string

const (
	AutoApprove     Flag = "auto-approve"
	BackendBucket   Flag = "backend-bucket"
	BackendConfig   Flag = "backend-config"
	BackendEndpoint Flag = "backend-endpoint"
	BackendKey      Flag = "backend-key"
	BackendRegion   Flag = "backend-region"
	BackendType     Flag = "backend-type"
	BackendUrl      Flag = "backend-url"
	Clipboard       Flag = "clipboard"
	Destroy         Flag = "destroy"
	Diff            Flag = "diff"
	Force           Flag = "force"
	InputDir        Flag = "input-dir"
	KubeContext     Flag = "kube-context"
	KustomizeDir    Flag = "kustomize-dir"
	LocalModule     Flag = "local-module"
	Manifest        Flag = "manifest"
	ModuleDir       Flag = "module-dir"
	ModuleVersion   Flag = "module-version"
	NoCleanup       Flag = "no-cleanup"
	OutputDir       Flag = "output-dir"
	OutputFile      Flag = "output-file"
	Overlay         Flag = "overlay"
	PlantonGitRepo  Flag = "planton-git-repo"
	ProviderConfig  Flag = "provider-config"
	Reconfigure     Flag = "reconfigure"
	Set             Flag = "set"
	Stack           Flag = "stack"
	IacInput        Flag = "iac-input"
	Yes             Flag = "yes"
)

// HandleFlagErr stops the command when cobra could not read a flag. That only
// happens when a handler asks for a flag its command never registered (a
// programming error), so the message names the flag and the fix.
func HandleFlagErr(err error, flag Flag) {
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the --%s flag could not be read: %v", flag, err),
			"the command asked for a flag it does not register; this is a defect in the CLI, not in your invocation",
			"report it at https://github.com/plantonhq/planton/issues with the exact command you ran",
		)
	}
}

// Require stops the command when a flag that has NO default and no other
// source (a manifest annotation, an environment variable, a probe of the
// current directory) was left empty. Most engine flags have one of those and
// must not use Require: --module-dir, for example, is resolved by the module
// runtime (current directory, then the published module for this release),
// so an empty value is a legal choice, not a mistake.
//
// example is the flag as the user would type it, filled in
// (`--manifest path/to/manifest.yaml`), so the next step is copyable.
func Require(err error, flag Flag, value string, example string) {
	HandleFlagErr(err, flag)
	if value == "" {
		ui.Failure(
			fmt.Sprintf("--%s is empty", flag),
			fmt.Sprintf("this command has no other source for %s and cannot proceed without it", flag),
			fmt.Sprintf("pass %s", example),
		)
	}
}
