// Package manifestcr is the Pulumi side of a Kubernetes manifest projection
// kind: it applies the custom resource pkg/kubernetes/manifestprojection
// renders from the manifest, so a projection kind's Pulumi module and its
// generated Terraform module build the same object from the same projection.
//
// A projection kind's Pulumi module is therefore a few lines: create the
// provider, call Apply, export the outputs. It never maps spec fields by hand,
// because every field the spec gains reaches the object with no module change,
// exactly as it does through Terraform.
package manifestcr

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Apply renders the manifest and creates its custom resource. The Pulumi
// resource is named after the object (metadata.name). No await: the custom
// resource is configuration its controller consumes, and applying it is the
// whole contract -- the same posture as the Terraform module's
// kubectl_manifest without a wait.
func Apply(ctx *pulumi.Context, manifest manifestprojection.Manifest, opts ...pulumi.ResourceOption) (*manifestprojection.Object, error) {
	obj, err := manifestprojection.Render(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "render custom resource")
	}

	meta := &metav1.ObjectMetaArgs{
		Name:   pulumi.String(obj.Name),
		Labels: pulumi.ToStringMap(obj.Labels),
	}
	if obj.Namespace != "" {
		meta.Namespace = pulumi.String(obj.Namespace)
	}
	if len(obj.Annotations) > 0 {
		meta.Annotations = pulumi.ToStringMap(obj.Annotations)
	}

	if _, err := apiextensions.NewCustomResource(ctx, obj.Name, &apiextensions.CustomResourceArgs{
		ApiVersion:  pulumi.String(obj.APIVersion),
		Kind:        pulumi.String(obj.Kind),
		Metadata:    meta,
		OtherFields: kubernetes.UntypedArgs{"spec": obj.Spec},
	}, opts...); err != nil {
		return nil, errors.Wrapf(err, "create %s %s", obj.Kind, obj.Name)
	}
	return obj, nil
}
