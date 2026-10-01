package module

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                 = "name"
	OpServiceAccount       = "service_account"
	OpServiceAccountMember = "service_account_member"
)

// exportOutputs exports the config's name and its publisher, both as the
// bare email and in IAM member form ("serviceAccount:" + email, composed
// exactly as the Terraform module's outputs.tf does) -- the value a
// GcpPubSubTopicIamMember's member references to grant roles/pubsub.publisher.
func exportOutputs(ctx *pulumi.Context, name pulumi.StringOutput, serviceAccount pulumi.StringOutput) {
	ctx.Export(OpName, name)
	ctx.Export(OpServiceAccount, serviceAccount)
	ctx.Export(OpServiceAccountMember, pulumi.Sprintf("serviceAccount:%s", serviceAccount))
}
