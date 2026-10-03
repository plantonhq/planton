package gcpsecretmanagersecretv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSecretManagerSecretSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSecretManagerSecretSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSecretManagerSecret {
		return &GcpSecretManagerSecret{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSecretManagerSecret",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-secret",
			},
			Spec: &GcpSecretManagerSecretSpec{},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal global secret (automatic replication)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal and explicit secret_id", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.SecretId = "db_password-prod"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional secret with regional CMEK", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.CustomerManagedEncryption = &GcpSecretManagerSecretCmek{
			KmsKey: litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept explicit auto replication with CMEK", func() {
		target := minimal()
		target.Spec.Replication = &GcpSecretManagerSecretReplication{
			Auto: &GcpSecretManagerSecretReplicationAuto{
				CustomerManagedEncryption: &GcpSecretManagerSecretCmek{
					KmsKey: litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k"),
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept user-managed replication with per-replica CMEK", func() {
		target := minimal()
		target.Spec.Replication = &GcpSecretManagerSecretReplication{
			UserManaged: &GcpSecretManagerSecretReplicationUserManaged{
				Replicas: []*GcpSecretManagerSecretReplica{
					{
						Location: "us-east1",
						CustomerManagedEncryption: &GcpSecretManagerSecretCmek{
							KmsKey: litRef("projects/p/locations/us-east1/keyRings/r/cryptoKeys/k"),
						},
					},
					{
						Location: "us-west1",
						CustomerManagedEncryption: &GcpSecretManagerSecretCmek{
							KmsKey: litRef("projects/p/locations/us-west1/keyRings/r/cryptoKeys/k"),
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an initial version with base64 payload and DISABLE policy", func() {
		target := minimal()
		target.Spec.InitialVersion = &GcpSecretManagerSecretInitialVersion{
			Data:           litRef("aGVsbG8="),
			IsBase64:       true,
			DeletionPolicy: "DISABLE",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept iam members with and without conditions", func() {
		target := minimal()
		target.Spec.IamMembers = []*GcpSecretManagerSecretIamMember{
			{
				Role:   "roles/secretmanager.secretAccessor",
				Member: litRef("serviceAccount:app@my-project.iam.gserviceaccount.com"),
			},
			{
				Role:   "roles/secretmanager.viewer",
				Member: litRef("group:platform@example.com"),
				Condition: &GcpSecretManagerSecretIamCondition{
					Title:      "expires-2027",
					Expression: `request.time < timestamp("2027-01-01T00:00:00Z")`,
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept rotation with topics", func() {
		target := minimal()
		target.Spec.Rotation = &GcpSecretManagerSecretRotation{
			RotationPeriod:   "2592000s",
			NextRotationTime: "2026-09-01T00:00:00Z",
		}
		target.Spec.Topics = []*foreignkeyv1.StringValueOrRef{
			litRef("projects/my-project/topics/secret-rotation"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept expiry via ttl OR expire_time", func() {
		viaTtl := minimal()
		viaTtl.Spec.Ttl = "7776000s"
		gomega.Expect(validator.Validate(viaTtl)).To(gomega.Succeed())

		viaTime := minimal()
		viaTime.Spec.ExpireTime = "2027-01-01T00:00:00Z"
		gomega.Expect(validator.Validate(viaTime)).To(gomega.Succeed())
	})

	ginkgo.It("should accept aliases, annotations, tags, destroy TTL, and guards", func() {
		target := minimal()
		target.Spec.VersionAliases = map[string]string{"prod": "1"}
		target.Spec.Annotations = map[string]string{"owner": "platform-team"}
		target.Spec.Tags = map[string]string{"tagKeys/123": "tagValues/456"}
		target.Spec.VersionDestroyTtl = "86400s"
		target.Spec.DeletionProtection = true
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject replication on a regional secret", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.Replication = &GcpSecretManagerSecretReplication{
			Auto: &GcpSecretManagerSecretReplicationAuto{},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "global")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject regional CMEK on a global secret", func() {
		target := minimal()
		target.Spec.CustomerManagedEncryption = &GcpSecretManagerSecretCmek{
			KmsKey: litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k"),
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a replication message with both or neither arm", func() {
		both := minimal()
		both.Spec.Replication = &GcpSecretManagerSecretReplication{
			Auto:        &GcpSecretManagerSecretReplicationAuto{},
			UserManaged: &GcpSecretManagerSecretReplicationUserManaged{Replicas: []*GcpSecretManagerSecretReplica{{Location: "us-east1"}}},
		}
		gomega.Expect(validator.Validate(both)).ToNot(gomega.Succeed())

		neither := minimal()
		neither.Spec.Replication = &GcpSecretManagerSecretReplication{}
		gomega.Expect(validator.Validate(neither)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject user-managed replication without replicas", func() {
		target := minimal()
		target.Spec.Replication = &GcpSecretManagerSecretReplication{
			UserManaged: &GcpSecretManagerSecretReplicationUserManaged{},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both expire_time and ttl", func() {
		target := minimal()
		target.Spec.Ttl = "7776000s"
		target.Spec.ExpireTime = "2027-01-01T00:00:00Z"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "one of")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject rotation without topics", func() {
		target := minimal()
		target.Spec.Rotation = &GcpSecretManagerSecretRotation{
			RotationPeriod:   "2592000s",
			NextRotationTime: "2026-09-01T00:00:00Z",
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "topics")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a rotation_period without next_rotation_time", func() {
		target := minimal()
		target.Spec.Rotation = &GcpSecretManagerSecretRotation{RotationPeriod: "2592000s"}
		target.Spec.Topics = []*foreignkeyv1.StringValueOrRef{litRef("projects/p/topics/t")}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a sub-hour rotation_period", func() {
		target := minimal()
		target.Spec.Rotation = &GcpSecretManagerSecretRotation{
			RotationPeriod:   "1800s",
			NextRotationTime: "2026-09-01T00:00:00Z",
		}
		target.Spec.Topics = []*foreignkeyv1.StringValueOrRef{litRef("projects/p/topics/t")}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a sub-day version_destroy_ttl", func() {
		target := minimal()
		target.Spec.VersionDestroyTtl = "3600s"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject more than ten topics", func() {
		target := minimal()
		for i := 0; i < 11; i++ {
			target.Spec.Topics = append(target.Spec.Topics, litRef("projects/p/topics/t"))
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid secret_id", func() {
		target := minimal()
		target.Spec.SecretId = "has spaces"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an initial version without data", func() {
		target := minimal()
		target.Spec.InitialVersion = &GcpSecretManagerSecretInitialVersion{}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid version deletion_policy", func() {
		target := minimal()
		target.Spec.InitialVersion = &GcpSecretManagerSecretInitialVersion{
			Data:           litRef("hello"),
			DeletionPolicy: "PREVENT", // version-level accepts DELETE/DISABLE/ABANDON only
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "DISABLE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an iam member without role or member", func() {
		noRole := minimal()
		noRole.Spec.IamMembers = []*GcpSecretManagerSecretIamMember{{
			Member: litRef("serviceAccount:app@p.iam.gserviceaccount.com"),
		}}
		gomega.Expect(validator.Validate(noRole)).ToNot(gomega.Succeed())

		noMember := minimal()
		noMember.Spec.IamMembers = []*GcpSecretManagerSecretIamMember{{
			Role: "roles/secretmanager.secretAccessor",
		}}
		gomega.Expect(validator.Validate(noMember)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an iam condition without title or expression", func() {
		target := minimal()
		target.Spec.IamMembers = []*GcpSecretManagerSecretIamMember{{
			Role:      "roles/secretmanager.secretAccessor",
			Member:    litRef("serviceAccount:app@p.iam.gserviceaccount.com"),
			Condition: &GcpSecretManagerSecretIamCondition{Title: "only-title"},
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
