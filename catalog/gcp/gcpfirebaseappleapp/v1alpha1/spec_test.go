package gcpfirebaseappleappv1alpha1

import (
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
	ginkgo.RunSpecs(t, "GcpFirebaseAppleAppSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const devicecheckKey = "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg\n-----END PRIVATE KEY-----\n"

var _ = ginkgo.Describe("GcpFirebaseAppleAppSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFirebaseAppleApp {
		return &GcpFirebaseAppleApp{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirebaseAppleApp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-firebase-apple-app",
			},
			Spec: &GcpFirebaseAppleAppSpec{
				DisplayName: "My iOS App",
				BundleId:    "com.example.app",
			},
		}
	}

	expectError := func(target *GcpFirebaseAppleApp, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the minimal registration (display name and bundle id)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept the full identity surface and every deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.ProjectId = litRef("my-gcp-project-123")
			target.Spec.ApiKeyId = litRef("9f3a2c1e-4b5d-4e6f-8a7b-0c1d2e3f4a5b")
			target.Spec.AppStoreId = "1234567890"
			target.Spec.TeamId = "ABCDE12345"
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("should accept every bundle id shape Apple allows", func() {
		for _, id := range []string{"com.example.app", "ai.planton.mobile", "com.example.my-app", "com.example.App2", "MyApp"} {
			target := minimal()
			target.Spec.BundleId = id
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should accept App Attest and DeviceCheck together with a team id, and debug tokens", func() {
		target := minimal()
		target.Spec.TeamId = "ABCDE12345"
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			AppAttest: &GcpFirebaseAppleAppAppCheckAppAttest{TokenTtl: "3600s"},
			DeviceCheck: &GcpFirebaseAppleAppAppCheckDeviceCheck{
				KeyId:      "ABC123DEFG",
				PrivateKey: devicecheckKey,
				TokenTtl:   "1800s",
			},
			DebugTokens: []*GcpFirebaseAppleAppAppCheckDebugToken{
				{DisplayName: "simulator", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		// The rule-001 guard for a defaulted switch: an explicit false is a
		// legal, meaningful shape ("declared, off"), never a validation error.
		target.Spec.AppCheck.AppAttest.Enabled = proto.Bool(false)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an App Attest block declared off without a team id", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			AppAttest: &GcpFirebaseAppleAppAppCheckAppAttest{Enabled: proto.Bool(false)},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept debug tokens alone without a team id", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			DebugTokens: []*GcpFirebaseAppleAppAppCheckDebugToken{
				{DisplayName: "simulator", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing display name or bundle id", func() {
		target := minimal()
		target.Spec.DisplayName = ""
		expectError(target, "display_name")

		target = minimal()
		target.Spec.BundleId = ""
		expectError(target, "bundle_id")
	})

	ginkgo.It("should reject a bundle id with characters Apple does not allow", func() {
		for _, id := range []string{"com.example.my_app", "com.example app", "com/example/app"} {
			target := minimal()
			target.Spec.BundleId = id
			expectError(target, "bundle_id must use only")
		}
	})

	ginkgo.It("should reject a malformed app_store_id or team_id", func() {
		target := minimal()
		target.Spec.AppStoreId = "id1234567890"
		expectError(target, "app_store_id")

		target = minimal()
		target.Spec.TeamId = "abcde12345"
		expectError(target, "team_id must be the 10-character")

		target = minimal()
		target.Spec.TeamId = "ABCDE1234"
		expectError(target, "team_id must be the 10-character")
	})

	ginkgo.It("should reject App Attest or DeviceCheck without a team id", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			AppAttest: &GcpFirebaseAppleAppAppCheckAppAttest{},
		}
		expectError(target, "team_id is required")

		target = minimal()
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			DeviceCheck: &GcpFirebaseAppleAppAppCheckDeviceCheck{KeyId: "ABC123DEFG", PrivateKey: devicecheckKey},
		}
		expectError(target, "team_id is required")
	})

	ginkgo.It("should reject a DeviceCheck block missing its key id or private key, or with a malformed key id", func() {
		target := minimal()
		target.Spec.TeamId = "ABCDE12345"
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			DeviceCheck: &GcpFirebaseAppleAppAppCheckDeviceCheck{PrivateKey: devicecheckKey},
		}
		expectError(target, "key_id")

		target.Spec.AppCheck.DeviceCheck = &GcpFirebaseAppleAppAppCheckDeviceCheck{KeyId: "ABC123DEFG"}
		expectError(target, "private_key")

		target.Spec.AppCheck.DeviceCheck = &GcpFirebaseAppleAppAppCheckDeviceCheck{KeyId: "abc", PrivateKey: devicecheckKey}
		expectError(target, "device_check.key_id must be the 10-character")
	})

	ginkgo.It("should reject a malformed token_ttl", func() {
		target := minimal()
		target.Spec.TeamId = "ABCDE12345"
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			AppAttest: &GcpFirebaseAppleAppAppCheckAppAttest{TokenTtl: "60m"},
		}
		expectError(target, "token_ttl must be a duration")
	})

	ginkgo.It("should reject duplicated debug token names", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAppleAppAppCheck{
			DebugTokens: []*GcpFirebaseAppleAppAppCheckDebugToken{
				{DisplayName: "sim", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
				{DisplayName: "sim", Token: "7d2e9f1b-4a6c-4e8d-b0f3-1c5a7e9b2d4f"},
			},
		}
		expectError(target, "at most once")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "DETACH"
		expectError(target, "deletion_policy must be one of")
	})

	ginkgo.It("should reject a wrong kind or apiVersion", func() {
		target := minimal()
		target.Kind = "GcpFirebaseIosApp"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
