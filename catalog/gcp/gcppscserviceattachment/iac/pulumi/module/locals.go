package module

import (
	gcppscserviceattachmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppscserviceattachment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share.
type Locals struct {
	GcpPscServiceAttachment *gcppscserviceattachmentv1alpha1.GcpPscServiceAttachment

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	// AttachmentName is the spec's attachment_name, or metadata.name when
	// the spec leaves it empty — the same naming basis every kind uses.
	AttachmentName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcppscserviceattachmentv1alpha1.GcpPscServiceAttachmentStackInput) *Locals {
	target := stackInput.Target

	attachmentName := target.Spec.AttachmentName
	if attachmentName == "" {
		attachmentName = target.Metadata.Name
	}

	return &Locals{
		GcpPscServiceAttachment: target,
		ProjectId:               target.Spec.ProjectId.GetValue(),
		AttachmentName:          attachmentName,
	}
}
