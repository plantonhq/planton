package module

import (
	"strings"

	alicloudnasfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnasfilesystem/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudNasFileSystem *alicloudnasfilesystemv1.AlicloudNasFileSystem
	Tags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudnasfilesystemv1.AlicloudNasFileSystemStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudNasFileSystem = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudNasFileSystem.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}

func fileSystemType(spec *alicloudnasfilesystemv1.AlicloudNasFileSystemSpec) string {
	if spec.FileSystemType != nil && *spec.FileSystemType != "" {
		return *spec.FileSystemType
	}
	return "standard"
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
