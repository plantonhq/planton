package awseksaddonv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEksAddonSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksAddonSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidAddon is the common case: an AWS-built add-on on the AWS
// default version, adopting any bootstrap self-managed copy.
func minimalValidAddon() *AwsEksAddon {
	return &AwsEksAddon{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEksAddon",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "platform-vpc-cni",
		},
		Spec: &AwsEksAddonSpec{
			Region:                   "us-west-2",
			ClusterName:              literalRef("platform"),
			AddonName:                "vpc-cni",
			ResolveConflictsOnCreate: "OVERWRITE",
		},
	}
}

var _ = ginkgo.Describe("AwsEksAddonSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_addon", func() {

			ginkgo.It("should not return a validation error for a minimal add-on", func() {
				err := protovalidate.Validate(minimalValidAddon())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned version with drift handling", func() {
				input := minimalValidAddon()
				input.Spec.AddonVersion = "v1.18.1-eksbuild.3"
				input.Spec.ResolveConflictsOnUpdate = "PRESERVE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a plain semver version without a build suffix", func() {
				input := minimalValidAddon()
				input.Spec.AddonVersion = "v1.18.1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IRSA service-account role", func() {
				input := minimalValidAddon()
				input.Spec.AddonName = "aws-ebs-csi-driver"
				input.Spec.ServiceAccountRoleArn = literalRef("arn:aws:iam::123456789012:role/EbsCsiDriverRole")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept pod-identity associations", func() {
				input := minimalValidAddon()
				input.Spec.AddonName = "aws-ebs-csi-driver"
				input.Spec.PodIdentityAssociations = []*AwsEksAddonPodIdentityAssociation{
					{
						RoleArn:        literalRef("arn:aws:iam::123456789012:role/EbsCsiDriverRole"),
						ServiceAccount: "ebs-csi-controller-sa",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept configuration values and preserve-on-delete", func() {
				input := minimalValidAddon()
				input.Spec.AddonName = "coredns"
				input.Spec.ConfigurationValues = `{"replicaCount":3}`
				input.Spec.Preserve = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom namespace config", func() {
				input := minimalValidAddon()
				input.Spec.NamespaceConfig = &AwsEksAddonNamespaceConfig{
					Namespace: "platform-addons",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_eks_addon", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidAddon()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when cluster_name is missing", func() {
				input := minimalValidAddon()
				input.Spec.ClusterName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when addon_name is empty", func() {
				input := minimalValidAddon()
				input.Spec.AddonName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a version missing the leading v", func() {
				input := minimalValidAddon()
				input.Spec.AddonVersion = "1.18.1-eksbuild.3"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a partial version", func() {
				input := minimalValidAddon()
				input.Spec.AddonVersion = "v1.18"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject PRESERVE at create time (AWS's create/update asymmetry)", func() {
				input := minimalValidAddon()
				input.Spec.ResolveConflictsOnCreate = "PRESERVE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown resolve_conflicts_on_update value", func() {
				input := minimalValidAddon()
				input.Spec.ResolveConflictsOnUpdate = "KEEP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a pod-identity association without a role", func() {
				input := minimalValidAddon()
				input.Spec.PodIdentityAssociations = []*AwsEksAddonPodIdentityAssociation{
					{ServiceAccount: "ebs-csi-controller-sa"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a pod-identity association without a service account", func() {
				input := minimalValidAddon()
				input.Spec.PodIdentityAssociations = []*AwsEksAddonPodIdentityAssociation{
					{RoleArn: literalRef("arn:aws:iam::123456789012:role/EbsCsiDriverRole")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a namespace that is not an RFC 1123 label", func() {
				input := minimalValidAddon()
				input.Spec.NamespaceConfig = &AwsEksAddonNamespaceConfig{
					Namespace: "Platform_Addons",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
