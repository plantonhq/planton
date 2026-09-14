package gcpserviceaccountv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpServiceAccountSpec Suite")
}

func boolPtr(b bool) *bool {
	return &b
}

var _ = ginkgo.Describe("GcpServiceAccountSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpServiceAccount.
	minimal := func() *GcpServiceAccount {
		return &GcpServiceAccount{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpServiceAccount",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-service-account",
			},
			Spec: &GcpServiceAccountSpec{
				ServiceAccountId: "test-sa-123",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec with only service_account_id", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully populated spec", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
		}
		msg.Spec.DisplayName = "CI/CD Deployer"
		msg.Spec.Description = "Deploys application releases from the pipeline"
		msg.Spec.Disabled = boolPtr(false)
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			Algorithm:      "KEY_ALG_RSA_2048",
			PrivateKeyType: "TYPE_GOOGLE_CREDENTIALS_FILE",
			Keepers:        map[string]string{"rotation": "2026-08"},
			DeletionPolicy: "DELETE",
		}
		msg.Spec.ProjectIamRoles = []string{"roles/logging.logWriter", "roles/storage.objectViewer"}
		msg.Spec.OrgId = "123456789012"
		msg.Spec.OrgIamRoles = []string{"roles/resourcemanager.organizationViewer"}
		msg.Spec.CreateIgnoreAlreadyExists = true
		msg.Spec.DeletionPolicy = "PREVENT"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a user_managed_key declared and switched off", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{Enabled: proto.Bool(false), Algorithm: "KEY_ALG_RSA_2048"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an empty user_managed_key (GCP defaults: 2048-bit RSA JSON key)", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept the upload flow: public_key_data without key types", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			PublicKeyData: "TUlJQ1dnSUJBQUtCZ1FD...",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 6-character service_account_id (lower bound)", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "abc123"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 30-character service_account_id (upper bound)", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "a" + strings.Repeat("b", 29)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a description at exactly 256 characters", func() {
		msg := minimal()
		msg.Spec.Description = strings.Repeat("d", 256)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept org_iam_roles when org_id is set", func() {
		msg := minimal()
		msg.Spec.OrgId = "123456789012"
		msg.Spec.OrgIamRoles = []string{"roles/resourcemanager.organizationViewer"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing service_account_id", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a service_account_id shorter than 6 characters", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "abc12"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a service_account_id longer than 30 characters", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "a" + strings.Repeat("b", 30)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a service_account_id starting with a digit", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "1invalid-sa"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a service_account_id with uppercase letters", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "Invalid-SA-Id"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a service_account_id ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.ServiceAccountId = "test-sa-123-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Description = strings.Repeat("d", 257)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject org_iam_roles without org_id", func() {
		msg := minimal()
		msg.Spec.OrgIamRoles = []string{"roles/resourcemanager.organizationViewer"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "ABANDON"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject public_key_data combined with private_key_type (upload vs generate)", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			PublicKeyData:  "TUlJQ1dnSUJBQUtCZ1FD...",
			PrivateKeyType: "TYPE_GOOGLE_CREDENTIALS_FILE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject public_key_data combined with public_key_type (upload vs generate)", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			PublicKeyData: "TUlJQ1dnSUJBQUtCZ1FD...",
			PublicKeyType: "TYPE_X509_PEM_FILE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid key algorithm", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			Algorithm: "KEY_ALG_ED25519",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid key deletion_policy (ABANDON is not a key policy)", func() {
		msg := minimal()
		msg.Spec.UserManagedKey = &GcpServiceAccountUserManagedKey{
			DeletionPolicy: "ABANDON",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
