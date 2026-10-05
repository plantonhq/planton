package module

import (
	gcpbackendbucketv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbackendbucket/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpBackendBucket *gcpbackendbucketv1alpha1.GcpBackendBucket

	// The cloud-side name defaults to metadata.name when the spec leaves
	// backend_bucket_name empty — the same naming basis every kind uses.
	BackendBucketName string
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpbackendbucketv1alpha1.GcpBackendBucketIacInput) *Locals {
	target := iacInput.Target

	backendBucketName := target.Spec.BackendBucketName
	if backendBucketName == "" {
		backendBucketName = target.Metadata.Name
	}

	return &Locals{
		GcpBackendBucket:  target,
		BackendBucketName: backendBucketName,
	}
}
