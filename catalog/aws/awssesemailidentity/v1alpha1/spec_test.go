package awssesemailidentityv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsSesEmailIdentitySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSesEmailIdentitySpec Validation Suite")
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

func minimalIdentity() *AwsSesEmailIdentity {
	return &AwsSesEmailIdentity{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsSesEmailIdentity",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "prod-sender",
		},
		Spec: &AwsSesEmailIdentitySpec{
			Region:        "us-west-2",
			EmailIdentity: "example.com",
		},
	}
}

var _ = ginkgo.Describe("AwsSesEmailIdentitySpec validations", func() {

	ginkgo.It("accepts a minimal domain identity", func() {
		err := protovalidate.Validate(minimalIdentity())
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Easy DKIM on a domain identity", func() {
		input := minimalIdentity()
		input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
			NextSigningKeyLength: "RSA_2048_BIT",
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a configuration set reference", func() {
		input := minimalIdentity()
		input.Spec.ConfigurationSet = &fkv1.StringValueOrRef{
			LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
				ValueFrom: &fkv1.ValueFromRef{
					Kind:      catalogkind.CatalogKind_AwsSesConfigurationSet,
					Name:      "txn-set",
					FieldPath: "status.outputs.configuration_set_name",
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.Context("CEL: dkim_requires_domain_identity", func() {
		ginkgo.It("fails when dkim_signing is set on an email-address identity", func() {
			input := minimalIdentity()
			input.Spec.EmailIdentity = "sender@example.com"
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
				NextSigningKeyLength: "RSA_2048_BIT",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("CEL: policy_names_unique", func() {
		ginkgo.It("fails when policy names are duplicated", func() {
			policyDoc, err := structpb.NewStruct(map[string]any{
				"Version": "2012-10-17",
				"Statement": []any{
					map[string]any{
						"Effect":   "Allow",
						"Action":   []any{"ses:SendEmail"},
						"Resource": "*",
					},
				},
			})
			gomega.Expect(err).To(gomega.BeNil())

			policy := func() *AwsSesEmailIdentityPolicy {
				return &AwsSesEmailIdentityPolicy{Name: "dup", Policy: policyDoc}
			}
			input := minimalIdentity()
			input.Spec.Policies = []*AwsSesEmailIdentityPolicy{policy(), policy()}
			err = protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("CEL: dkim_arm_required", func() {
		ginkgo.It("fails when dkim_signing is declared empty (configures nothing)", func() {
			input := minimalIdentity()
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts the BYODKIM arm without a key length", func() {
			input := minimalIdentity()
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
				DomainSigningPrivateKey: "dGVzdA==",
				DomainSigningSelector:   "selector1",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("email_forwarding_enabled tri-state", func() {
		ginkgo.It("accepts an explicit false (pins the position)", func() {
			input := minimalIdentity()
			input.Spec.EmailForwardingEnabled = boolPtr(false)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("CEL: byodkim_pair", func() {
		ginkgo.It("fails when only domain_signing_private_key is set", func() {
			input := minimalIdentity()
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
				DomainSigningPrivateKey: "dGVzdA==",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when only domain_signing_selector is set", func() {
			input := minimalIdentity()
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
				DomainSigningSelector: "selector1",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("CEL: easy_dkim_xor_byodkim", func() {
		ginkgo.It("fails when Easy DKIM and BYODKIM are combined", func() {
			input := minimalIdentity()
			input.Spec.DkimSigning = &AwsSesEmailIdentityDkimSigning{
				NextSigningKeyLength:    "RSA_2048_BIT",
				DomainSigningPrivateKey: "dGVzdA==",
				DomainSigningSelector:   "selector1",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("mail_from required fields", func() {
		ginkgo.It("accepts mail_from with mail_from_domain", func() {
			input := minimalIdentity()
			input.Spec.MailFrom = &AwsSesEmailIdentityMailFrom{
				MailFromDomain: "mail.example.com",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when mail_from_domain is missing", func() {
			input := minimalIdentity()
			input.Spec.MailFrom = &AwsSesEmailIdentityMailFrom{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails on an invalid behavior_on_mx_failure", func() {
			input := minimalIdentity()
			input.Spec.MailFrom = &AwsSesEmailIdentityMailFrom{
				MailFromDomain:      "mail.example.com",
				BehaviorOnMxFailure: stringPtr("IGNORE"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
