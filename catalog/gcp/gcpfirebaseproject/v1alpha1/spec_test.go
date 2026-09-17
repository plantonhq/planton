package gcpfirebaseprojectv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFirebaseProjectSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFirebaseProjectSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFirebaseProject {
		return &GcpFirebaseProject{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirebaseProject",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-firebase-project",
			},
			Spec: &GcpFirebaseProjectSpec{},
		}
	}

	expectError := func(target *GcpFirebaseProject, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept an entirely empty spec (enable Firebase on the default project)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project reference and every default_storage_location shape", func() {
		for _, loc := range []string{"US", "EU", "NAM4", "us-central1", "europe-west1"} {
			target := minimal()
			target.Spec.ProjectId = litRef("my-gcp-project-123")
			target.Spec.DefaultStorageLocation = loc
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), loc)
		}
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("should accept App Check service configs for every supported service and every mode", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseProjectAppCheck{
			ServiceConfigs: []*GcpFirebaseProjectAppCheckServiceConfig{
				{ServiceId: "firestore.googleapis.com", EnforcementMode: "ENFORCED"},
				{ServiceId: "firebasestorage.googleapis.com", EnforcementMode: "UNENFORCED"},
				{ServiceId: "firebasedatabase.googleapis.com"},
				{ServiceId: "identitytoolkit.googleapis.com", EnforcementMode: "ENFORCED"},
			},
			ResourcePolicies: []*GcpFirebaseProjectAppCheckResourcePolicy{
				{
					ServiceId:       "oauth2.googleapis.com",
					TargetResource:  "//oauth2.googleapis.com/projects/123456789/oauthClients/123456789-abc.apps.googleusercontent.com",
					EnforcementMode: "UNENFORCED",
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a malformed default_storage_location", func() {
		target := minimal()
		target.Spec.DefaultStorageLocation = "us central"
		expectError(target, "default_storage_location")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "DETACH"
		expectError(target, "deletion_policy must be one of")
	})

	ginkgo.It("should reject an App Check service Google does not support", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseProjectAppCheck{
			ServiceConfigs: []*GcpFirebaseProjectAppCheckServiceConfig{
				{ServiceId: "fcm.googleapis.com", EnforcementMode: "ENFORCED"},
			},
		}
		expectError(target, "service_id must be one of")
	})

	ginkgo.It("should reject a duplicated App Check service", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseProjectAppCheck{
			ServiceConfigs: []*GcpFirebaseProjectAppCheckServiceConfig{
				{ServiceId: "firestore.googleapis.com", EnforcementMode: "ENFORCED"},
				{ServiceId: "firestore.googleapis.com", EnforcementMode: "UNENFORCED"},
			},
		}
		expectError(target, "at most once")
	})

	ginkgo.It("should reject an unknown enforcement_mode", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseProjectAppCheck{
			ServiceConfigs: []*GcpFirebaseProjectAppCheckServiceConfig{
				{ServiceId: "firestore.googleapis.com", EnforcementMode: "OFF"},
			},
		}
		expectError(target, "enforcement_mode must be one of")
	})

	ginkgo.It("should reject a resource policy on the wrong service or with a malformed target", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseProjectAppCheck{
			ResourcePolicies: []*GcpFirebaseProjectAppCheckResourcePolicy{
				{ServiceId: "firestore.googleapis.com", TargetResource: "//oauth2.googleapis.com/projects/1/oauthClients/x"},
			},
		}
		expectError(target, "oauth2.googleapis.com")

		target.Spec.AppCheck.ResourcePolicies[0] = &GcpFirebaseProjectAppCheckResourcePolicy{
			ServiceId: "oauth2.googleapis.com", TargetResource: "projects/1/oauthClients/x",
		}
		expectError(target, "target_resource")
	})

	ginkgo.It("should reject a wrong kind or apiVersion", func() {
		target := minimal()
		target.Kind = "GcpFirebase"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
