package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// workloadIdentityBinding grants “roles/iam.workloadIdentityUser” on the
// Google Service Account (GSA) to the Kubernetes Service Account (KSA)
// specified in the component spec.
func workloadIdentityBinding(ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider) error {

	member := fmt.Sprintf(
		"serviceAccount:%s.svc.id.goog[%s/%s]",
		locals.GcpGkeWorkloadIdentityBinding.Spec.ProjectId.GetValue(),
		locals.GcpGkeWorkloadIdentityBinding.Spec.KsaNamespace,
		locals.GcpGkeWorkloadIdentityBinding.Spec.KsaName,
	)

	createdIamMember, err := serviceaccount.NewIAMMember(
		ctx,
		"workload-identity-binding",
		&serviceaccount.IAMMemberArgs{
			ServiceAccountId: pulumi.String(
				locals.GcpGkeWorkloadIdentityBinding.Spec.ServiceAccountEmail.GetValue()),
			Role:   pulumi.String("roles/iam.workloadIdentityUser"),
			Member: pulumi.String(member),
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create IAM member for workload identity binding")
	}

	// Export outputs expected by the proto’s StackOutputs message.
	ctx.Export(OpMember, createdIamMember.Member)
	ctx.Export(OpServiceAccountEmail,
		pulumi.String(locals.GcpGkeWorkloadIdentityBinding.Spec.ServiceAccountEmail.GetValue()))

	return nil
}
