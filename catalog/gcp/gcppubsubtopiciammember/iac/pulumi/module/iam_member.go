package module

import (
	"github.com/pkg/errors"
	gcppubsubtopiciammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubtopiciammember/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// iamMember applies the single ADDITIVE grant ON the topic: one role, to one
// member, on one topic. Additive means the provider merges this (role,
// member) pair into the topic's IAM policy without touching any other
// member's bindings on the same role, and destroy subtracts only this pair.
//
// Every argument is immutable (ForceNew): an IAM grant has no update -- any
// change replaces it, which is also how the API behaves.
func iamMember(ctx *pulumi.Context, target *gcppubsubtopiciammemberv1alpha1.GcpPubSubTopicIamMember, gcpProvider *gcp.Provider) error {
	spec := target.Spec

	// The topic arrives as its full name (projects/<project>/topics/<topic>);
	// the provider reads the project from it, so no project argument is set.
	args := &pubsub.TopicIAMMemberArgs{
		Topic:  pulumi.String(spec.Topic.GetValue()),
		Role:   pulumi.String(spec.Role.GetValue()),
		Member: pulumi.String(spec.Member.GetValue()),
	}

	// No condition: Pub/Sub topics do not accept conditional role bindings
	// (the provider never requests the version-3 policy conditions need).

	createdMember, err := pubsub.NewTopicIAMMember(ctx, "iam-member", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub topic IAM member")
	}

	ctx.Export(OpTopic, createdMember.Topic)
	ctx.Export(OpRole, createdMember.Role)
	ctx.Export(OpMember, createdMember.Member)
	ctx.Export(OpEtag, createdMember.Etag)

	return nil
}
