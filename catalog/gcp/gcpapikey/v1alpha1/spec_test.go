package gcpapikeyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpApiKeySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpApiKeySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpApiKey {
		return &GcpApiKey{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpApiKey",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-api-key",
			},
			Spec: &GcpApiKeySpec{
				KeyId: "firebase-android-key",
			},
		}
	}

	expectError := func(target *GcpApiKey, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal unrestricted key", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project reference, a display name, and a service-account binding", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.DisplayName = "Firebase Android key"
		target.Spec.ServiceAccountEmail = litRef("sender@my-gcp-project-123.iam.gserviceaccount.com")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("should accept an Android restriction with colon and colon-free fingerprints", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			AndroidKeyRestrictions: &GcpApiKeyAndroidKeyRestrictions{
				AllowedApplications: []*GcpApiKeyAndroidApplication{
					{PackageName: "ai.planton.mobile", Sha1Fingerprint: "DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09"},
					{PackageName: "ai.planton.mobile", Sha1Fingerprint: "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept iOS, browser, and server arms each on their own, with API targets", func() {
		arms := []*GcpApiKeyRestrictions{
			{IosKeyRestrictions: &GcpApiKeyIosKeyRestrictions{AllowedBundleIds: []string{"ai.planton.mobile"}}},
			{BrowserKeyRestrictions: &GcpApiKeyBrowserKeyRestrictions{AllowedReferrers: []string{"https://app.example.com/*"}}},
			{ServerKeyRestrictions: &GcpApiKeyServerKeyRestrictions{AllowedIps: []string{"203.0.113.7", "2001:db8::/32"}}},
		}
		for _, arm := range arms {
			arm.ApiTargets = []*GcpApiKeyApiTarget{
				{Service: "fcmregistrations.googleapis.com"},
				{Service: "translate.googleapis.com", Methods: []string{"google.cloud.translate.v2.*"}},
			}
			target := minimal()
			target.Spec.Restrictions = arm
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept API targets alone (no client arm)", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			ApiTargets: []*GcpApiKeyApiTarget{{Service: "maps-backend.googleapis.com"}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing key_id", func() {
		target := minimal()
		target.Spec.KeyId = ""
		expectError(target, "key_id")
	})

	ginkgo.It("should reject a key_id that is not an RFC 1034 label", func() {
		for _, bad := range []string{"Firebase-Key", "1key", "key_id", "trailing-", "a-very-long-key-id-that-goes-well-past-the-sixty-three-character-limit-google-allows"} {
			target := minimal()
			target.Spec.KeyId = bad
			expectError(target, "key_id")
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "FORCE"
		expectError(target, "deletion_policy must be one of")
	})

	ginkgo.It("should reject two client arms on one key", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			IosKeyRestrictions:     &GcpApiKeyIosKeyRestrictions{AllowedBundleIds: []string{"ai.planton.mobile"}},
			BrowserKeyRestrictions: &GcpApiKeyBrowserKeyRestrictions{AllowedReferrers: []string{"https://app.example.com/*"}},
		}
		expectError(target, "at most one of")
	})

	ginkgo.It("should reject a hollow client arm", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			IosKeyRestrictions: &GcpApiKeyIosKeyRestrictions{},
		}
		expectError(target, "allowed_bundle_ids")

		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			AndroidKeyRestrictions: &GcpApiKeyAndroidKeyRestrictions{},
		}
		expectError(target, "allowed_applications")
	})

	ginkgo.It("should reject a malformed Android package name or fingerprint", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			AndroidKeyRestrictions: &GcpApiKeyAndroidKeyRestrictions{
				AllowedApplications: []*GcpApiKeyAndroidApplication{
					{PackageName: "mobile", Sha1Fingerprint: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709"},
				},
			},
		}
		expectError(target, "package_name")

		target.Spec.Restrictions.AndroidKeyRestrictions.AllowedApplications[0] = &GcpApiKeyAndroidApplication{
			PackageName: "ai.planton.mobile", Sha1Fingerprint: "not-a-fingerprint",
		}
		expectError(target, "sha1_fingerprint")
	})

	ginkgo.It("should reject an API target that is not a googleapis.com service", func() {
		target := minimal()
		target.Spec.Restrictions = &GcpApiKeyRestrictions{
			ApiTargets: []*GcpApiKeyApiTarget{{Service: "example.com"}},
		}
		expectError(target, "googleapis.com")
	})

	ginkgo.It("should reject a wrong kind or apiVersion", func() {
		target := minimal()
		target.Kind = "GcpApiKeys"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
