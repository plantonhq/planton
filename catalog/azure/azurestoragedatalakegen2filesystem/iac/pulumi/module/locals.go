package module

import (
	azurestoragedatalakegen2filesystemv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragedatalakegen2filesystem/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageDataLakeGen2Filesystem *azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2Filesystem
	StorageAccountId                   string
}

// aceScopeStrings maps the spec's ace scope enum to the data plane's
// lowercase wire values. Unspecified is deliberately absent: the module
// leaves the input unset so Azure's own default (access) applies --
// matching the Terraform module's null.
var aceScopeStrings = map[azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceScope]string{
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceScope_ACCESS:  "access",
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceScope_DEFAULT: "default",
}

// aceTypeStrings maps the spec's ace type enum to the data plane's
// lowercase wire values.
var aceTypeStrings = map[azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceType]string{
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceType_USER:  "user",
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceType_GROUP: "group",
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceType_MASK:  "mask",
	azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemAceType_OTHER: "other",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemIacInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageDataLakeGen2Filesystem = iacInput.Target
	locals.StorageAccountId = iacInput.Target.Spec.StorageAccountId.GetValue()

	// No Azure tags: ARM does not support tags on this resource (the
	// properties map is the filesystem-level metadata surface), so the
	// platform's identity tags live on the parent account.

	return locals
}
