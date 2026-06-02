package flag

import (
	log "github.com/sirupsen/logrus"
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
	OpenMCFGitRepo  Flag = "openmcf-git-repo"
	ProviderConfig  Flag = "provider-config"
	Reconfigure     Flag = "reconfigure"
	Set             Flag = "set"
	Stack           Flag = "stack"
	StackInput      Flag = "stack-input"
	Yes             Flag = "yes"
)

func HandleFlagErrAndValue(err error, flag Flag, flagVal string) {
	if err != nil {
		log.Fatalf("error parsing %s flag. err %v", flag, err)
	}
	if flagVal == "" {
		log.Fatalf("please provide %s", flag)
	}
}

func HandleFlagErr(err error, flag Flag) {
	if err != nil {
		log.Fatalf("error parsing %s flag. err %v", flag, err)
	}
}
