package awseksaccessentryv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsEksAccessEntrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEksAccessEntrySpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidEntry is the common case: a STANDARD entry for a team role
// with one cluster-scoped view policy.
func minimalValidEntry() *AwsEksAccessEntry {
	return &AwsEksAccessEntry{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsEksAccessEntry",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "platform-viewers",
		},
		Spec: &AwsEksAccessEntrySpec{
			Region:       "us-west-2",
			ClusterName:  literalRef("platform"),
			PrincipalArn: literalRef("arn:aws:iam::123456789012:role/TeamViewerRole"),
			PolicyAssociations: []*AwsEksAccessEntryPolicyAssociation{
				{
					PolicyArn: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy",
					AccessScope: &AwsEksAccessEntryAccessScope{
						Type: "cluster",
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsEksAccessEntrySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_eks_access_entry", func() {

			ginkgo.It("should not return a validation error for a minimal entry", func() {
				err := protovalidate.Validate(minimalValidEntry())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace-scoped admin association", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations = []*AwsEksAccessEntryPolicyAssociation{
					{
						PolicyArn: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy",
						AccessScope: &AwsEksAccessEntryAccessScope{
							Type:       "namespace",
							Namespaces: []string{"team-a", "team-b"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an RBAC group mapping without policy associations", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations = nil
				input.Spec.KubernetesGroups = []string{"platform-operators"}
				input.Spec.UserName = "ops:{{SessionName}}"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit STANDARD type", func() {
				input := minimalValidEntry()
				input.Spec.Type = "STANDARD"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a bare node-type entry", func() {
				input := minimalValidEntry()
				input.Spec.Type = "HYBRID_LINUX"
				input.Spec.PolicyAssociations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_eks_access_entry", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidEntry()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when cluster_name is missing", func() {
				input := minimalValidEntry()
				input.Spec.ClusterName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when principal_arn is missing", func() {
				input := minimalValidEntry()
				input.Spec.PrincipalArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown entry type", func() {
				input := minimalValidEntry()
				input.Spec.Type = "SSM_LINUX"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject policy associations on a node-type entry", func() {
				input := minimalValidEntry()
				input.Spec.Type = "EC2_LINUX"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject kubernetes_groups on a node-type entry", func() {
				input := minimalValidEntry()
				input.Spec.Type = "FARGATE_LINUX"
				input.Spec.PolicyAssociations = nil
				input.Spec.KubernetesGroups = []string{"platform-operators"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a reserved system: group", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations = nil
				input.Spec.KubernetesGroups = []string{"system:masters"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a policy association without a policy ARN", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations = []*AwsEksAccessEntryPolicyAssociation{
					{
						AccessScope: &AwsEksAccessEntryAccessScope{Type: "cluster"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a policy association without an access scope", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations = []*AwsEksAccessEntryPolicyAssociation{
					{PolicyArn: "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown access-scope type", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations[0].AccessScope.Type = "workspace"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a namespace scope without namespaces", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations[0].AccessScope.Type = "namespace"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a cluster scope that lists namespaces", func() {
				input := minimalValidEntry()
				input.Spec.PolicyAssociations[0].AccessScope.Namespaces = []string{"team-a"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
