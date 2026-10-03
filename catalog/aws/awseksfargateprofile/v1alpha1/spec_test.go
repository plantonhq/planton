package awseksfargateprofilev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEksFargateProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksFargateProfileSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidProfile is the common case: one namespace selector over a
// private subnet pair.
func minimalValidProfile() *AwsEksFargateProfile {
	return &AwsEksFargateProfile{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEksFargateProfile",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "serverless",
		},
		Spec: &AwsEksFargateProfileSpec{
			Region:              "us-west-2",
			ClusterName:         literalRef("platform"),
			PodExecutionRoleArn: literalRef("arn:aws:iam::123456789012:role/EksFargatePodExecutionRole"),
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-0123456789abcdef0"),
				literalRef("subnet-0123456789abcdef1"),
			},
			Selectors: []*AwsEksFargateProfileSelector{
				{Namespace: "serverless"},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsEksFargateProfileSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_fargate_profile", func() {

			ginkgo.It("should not return a validation error for a minimal profile", func() {
				err := protovalidate.Validate(minimalValidProfile())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept label-scoped selectors", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{
						Namespace: "batch",
						Labels:    map[string]string{"compute": "fargate"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept wildcard namespaces", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{Namespace: "team-*"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the AWS maximum of five selectors", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{Namespace: "a"}, {Namespace: "b"}, {Namespace: "c"},
					{Namespace: "d"}, {Namespace: "e"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-subnet profile", func() {
				input := minimalValidProfile()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					literalRef("subnet-0123456789abcdef0"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_eks_fargate_profile", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidProfile()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when cluster_name is missing", func() {
				input := minimalValidProfile()
				input.Spec.ClusterName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when pod_execution_role_arn is missing", func() {
				input := minimalValidProfile()
				input.Spec.PodExecutionRoleArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when subnet_ids is empty", func() {
				input := minimalValidProfile()
				input.Spec.SubnetIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when selectors is empty", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than five selectors (the AWS limit)", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{Namespace: "a"}, {Namespace: "b"}, {Namespace: "c"},
					{Namespace: "d"}, {Namespace: "e"}, {Namespace: "f"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a selector without a namespace", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{Labels: map[string]string{"compute": "fargate"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than five label pairs on a selector (the AWS limit)", func() {
				input := minimalValidProfile()
				input.Spec.Selectors = []*AwsEksFargateProfileSelector{
					{
						Namespace: "batch",
						Labels: map[string]string{
							"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
