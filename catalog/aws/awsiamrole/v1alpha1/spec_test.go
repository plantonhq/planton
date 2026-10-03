package awsiamrolev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsIamRoleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsIamRoleSpec Validation Tests")
}

func literalArn(arn string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: arn},
	}
}

// lambdaTrustPolicy is the common case: allow Lambda to assume the role.
func lambdaTrustPolicy() *structpb.Struct {
	doc, err := structpb.NewStruct(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []interface{}{
			map[string]interface{}{
				"Effect":    "Allow",
				"Principal": map[string]interface{}{"Service": "lambda.amazonaws.com"},
				"Action":    "sts:AssumeRole",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return doc
}

// minimalValidRole is the common case: a region and a trust policy.
func minimalValidRole() *AwsIamRole {
	return &AwsIamRole{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsIamRole",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-role",
		},
		Spec: &AwsIamRoleSpec{
			Region: "us-west-2",
			Trust:  &AwsIamRoleSpec_TrustPolicy{TrustPolicy: lambdaTrustPolicy()},
		},
	}
}

// minimalOidcTrust is the EKS IRSA shape: a provider pair plus one exact
// service-account subject.
func minimalOidcTrust() *AwsIamRoleOidcTrust {
	return &AwsIamRoleOidcTrust{
		ProviderArn: literalArn("arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED"),
		ProviderUrl: literalArn("oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED"),
		Subjects:    []string{"system:serviceaccount:kube-system:external-dns"},
	}
}

var _ = ginkgo.Describe("AwsIamRoleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_iam_role", func() {

			ginkgo.It("should not return a validation error for a minimal role", func() {
				err := protovalidate.Validate(minimalValidRole())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a fully-specified role", func() {
				input := minimalValidRole()
				input.Spec.Description = "Execution role for the analytics lambda"
				input.Spec.Path = "/service-roles/"
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					literalArn("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				}
				input.Spec.MaxSessionDuration = 7200
				input.Spec.PermissionsBoundary = literalArn("arn:aws:iam::123456789012:policy/boundaries/workload")
				input.Spec.ForceDetachPolicies = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for oidc_trust with an exact subject", func() {
				input := minimalValidRole()
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: minimalOidcTrust()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for oidc_trust wired by reference", func() {
				// The one-run composition shape: the OIDC provider deploys in
				// the same run and both halves flow in by reference.
				input := minimalValidRole()
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: &AwsIamRoleOidcTrust{
					ProviderArn: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Kind: catalogkind.CatalogKind_AwsIamOidcProvider,
								Name: "eks-oidc",
							},
						},
					},
					ProviderUrl: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Kind:      catalogkind.CatalogKind_AwsIamOidcProvider,
								Name:      "eks-oidc",
								FieldPath: "status.outputs.provider_url",
							},
						},
					},
					Subjects: []string{"system:serviceaccount:karpenter:karpenter"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for oidc_trust with only wildcard subjects", func() {
				// The CI-federation shape (e.g. GitHub Actions).
				input := minimalValidRole()
				oidcTrust := minimalOidcTrust()
				oidcTrust.Subjects = nil
				oidcTrust.WildcardSubjects = []string{"repo:my-org/my-repo:*"}
				oidcTrust.Audiences = []string{"sts.amazonaws.com"}
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: oidcTrust}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when managed policies are references", func() {
				input := minimalValidRole()
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{Name: "s3-read-only"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with inline policies", func() {
				input := minimalValidRole()
				inline, err := structpb.NewStruct(map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []interface{}{
						map[string]interface{}{
							"Effect":   "Allow",
							"Action":   "s3:ListBucket",
							"Resource": "arn:aws:s3:::demo-bucket",
						},
					},
				})
				gomega.Expect(err).To(gomega.BeNil())
				input.Spec.InlinePolicies = map[string]*structpb.Struct{"bucketAccess": inline}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error at the session-duration bounds", func() {
				input := minimalValidRole()
				input.Spec.MaxSessionDuration = 3600
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.MaxSessionDuration = 43200
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata set", func() {
				input := minimalValidRole()
				input.Metadata = &shared.CatalogObjectMetadata{
					Name:   "full-role",
					Org:    "acme-corp",
					Env:    "production",
					Labels: map[string]string{"team": "platform"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_iam_role", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidRole()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRole()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidRole()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AwsIamRole{
					ApiVersion: "aws.planton.dev/v1alpha1",
					Kind:       "AwsIamRole",
					Metadata:   &shared.CatalogObjectMetadata{Name: "test-role"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidRole()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither trust arm is set", func() {
				input := minimalValidRole()
				input.Spec.Trust = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when oidc_trust has no subjects at all", func() {
				input := minimalValidRole()
				oidcTrust := minimalOidcTrust()
				oidcTrust.Subjects = nil
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: oidcTrust}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when oidc_trust omits provider_arn", func() {
				input := minimalValidRole()
				oidcTrust := minimalOidcTrust()
				oidcTrust.ProviderArn = nil
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: oidcTrust}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when oidc_trust omits provider_url", func() {
				input := minimalValidRole()
				oidcTrust := minimalOidcTrust()
				oidcTrust.ProviderUrl = nil
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: oidcTrust}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an oidc_trust subject entry is empty", func() {
				input := minimalValidRole()
				oidcTrust := minimalOidcTrust()
				oidcTrust.Subjects = []string{""}
				input.Spec.Trust = &AwsIamRoleSpec_OidcTrust{OidcTrust: oidcTrust}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description exceeds 1000 characters", func() {
				input := minimalValidRole()
				long := make([]byte, 1001)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Description = string(long)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when description contains typographic quotes", func() {
				input := minimalValidRole()
				input.Spec.Description = "Role for the “analytics” pipeline"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path does not begin and end with '/'", func() {
				input := minimalValidRole()
				input.Spec.Path = "service-roles"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when path contains an empty segment", func() {
				input := minimalValidRole()
				input.Spec.Path = "/service//roles/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_session_duration is below 3600", func() {
				input := minimalValidRole()
				input.Spec.MaxSessionDuration = 1800
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_session_duration exceeds 43200", func() {
				input := minimalValidRole()
				input.Spec.MaxSessionDuration = 86400
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a managed policy entry is empty", func() {
				input := minimalValidRole()
				input.Spec.ManagedPolicyArns = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an inline policy name exceeds 128 characters", func() {
				input := minimalValidRole()
				long := make([]byte, 129)
				for i := range long {
					long[i] = 'a'
				}
				doc, err := structpb.NewStruct(map[string]interface{}{"Version": "2012-10-17"})
				gomega.Expect(err).To(gomega.BeNil())
				input.Spec.InlinePolicies = map[string]*structpb.Struct{string(long): doc}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
