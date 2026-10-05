package awsredshiftserverlessnamespacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsRedshiftServerlessNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRedshiftServerlessNamespaceSpec Custom Validation Tests")
}

// validNamespace returns a minimal valid namespace that individual tests
// mutate into specific scenarios.
func validNamespace() *AwsRedshiftServerlessNamespace {
	return &AwsRedshiftServerlessNamespace{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsRedshiftServerlessNamespace",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-namespace",
		},
		Spec: &AwsRedshiftServerlessNamespaceSpec{
			Region:              "us-west-2",
			ManageAdminPassword: true,
		},
	}
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

var _ = ginkgo.Describe("AwsRedshiftServerlessNamespaceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal managed-password namespace", func() {
			err := protovalidate.Validate(validNamespace())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a namespace with no admin credentials at all", func() {
			input := validNamespace()
			input.Spec.ManageAdminPassword = false
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a plaintext-password namespace", func() {
			input := validNamespace()
			input.Spec.ManageAdminPassword = false
			input.Spec.AdminUsername = "warehouseadmin"
			input.Spec.AdminUserPassword = "Correct1HorseBattery"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a fully composed production namespace", func() {
			input := validNamespace()
			input.Spec.DbName = "analytics"
			input.Spec.AdminUsername = "warehouseadmin"
			input.Spec.AdminPasswordSecretKmsKeyId = literal("arn:aws:kms:us-west-2:123456789012:key/11111111-2222-3333-4444-555555555555")
			input.Spec.KmsKeyId = literal("arn:aws:kms:us-west-2:123456789012:key/66666666-7777-8888-9999-000000000000")
			input.Spec.IamRoles = []*foreignkeyv1.StringValueOrRef{
				literal("arn:aws:iam::123456789012:role/redshift-spectrum"),
				literal("arn:aws:iam::123456789012:role/redshift-copy"),
			}
			input.Spec.DefaultIamRoleArn = literal("arn:aws:iam::123456789012:role/redshift-copy")
			input.Spec.LogExports = []string{"connectionlog", "useractivitylog", "userlog"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing region", func() {
			input := validNamespace()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a plaintext password alongside the managed strategy (password_xor_managed)", func() {
			input := validNamespace()
			input.Spec.AdminUserPassword = "Correct1HorseBattery"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("pick one password strategy"))
		})

		ginkgo.It("rejects a secret KMS key without the managed strategy (secret_kms_requires_managed)", func() {
			input := validNamespace()
			input.Spec.ManageAdminPassword = false
			input.Spec.AdminPasswordSecretKmsKeyId = literal("arn:aws:kms:us-west-2:123456789012:key/11111111-2222-3333-4444-555555555555")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("manage_admin_password"))
		})

		ginkgo.It("rejects an unknown log export type", func() {
			input := validNamespace()
			input.Spec.LogExports = []string{"connectionlog", "querylog"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate log export types", func() {
			input := validNamespace()
			input.Spec.LogExports = []string{"userlog", "userlog"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
