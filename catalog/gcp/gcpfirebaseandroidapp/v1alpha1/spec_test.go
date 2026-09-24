package gcpfirebaseandroidappv1alpha1

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
	ginkgo.RunSpecs(t, "GcpFirebaseAndroidAppSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFirebaseAndroidAppSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFirebaseAndroidApp {
		return &GcpFirebaseAndroidApp{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirebaseAndroidApp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-firebase-android-app",
			},
			Spec: &GcpFirebaseAndroidAppSpec{
				DisplayName: "My Android App",
				PackageName: "com.example.app",
			},
		}
	}

	expectError := func(target *GcpFirebaseAndroidApp, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the minimal registration (display name and package name)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Firebase project reference, an API key reference, and every deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.ProjectId = litRef("my-gcp-project-123")
			target.Spec.ApiKeyId = litRef("9f3a2c1e-4b5d-4e6f-8a7b-0c1d2e3f4a5b")
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("should accept every valid package name shape", func() {
		for _, pkg := range []string{"com.example.app", "ai.planton.mobile", "org.example.my_app.v2", "a.b"} {
			target := minimal()
			target.Spec.PackageName = pkg
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), pkg)
		}
	})

	ginkgo.It("should accept certificate fingerprints with and without colon separators", func() {
		target := minimal()
		target.Spec.Sha1Hashes = []string{
			"da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"DA:39:A3:EE:5E:6B:4B:0D:32:55:BF:EF:95:60:18:90:AF:D8:07:09",
		}
		target.Spec.Sha256Hashes = []string{
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			"E3:B0:C4:42:98:FC:1C:14:9A:FB:F4:C8:99:6F:B9:24:27:AE:41:E4:64:9B:93:4C:A4:95:99:1B:78:52:B8:55",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept App Check with Play Integrity (present, explicitly on, explicitly off) and debug tokens", func() {
		target := minimal()
		target.Spec.Sha256Hashes = []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}
		target.Spec.AppCheck = &GcpFirebaseAndroidAppAppCheck{
			PlayIntegrity: &GcpFirebaseAndroidAppAppCheckPlayIntegrity{TokenTtl: "3600s"},
			DebugTokens: []*GcpFirebaseAndroidAppAppCheckDebugToken{
				{DisplayName: "ci-emulator", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
				{DisplayName: "dev-laptop", Token: "7d2e9f1b-4a6c-4e8d-b0f3-1c5a7e9b2d4f"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		// The rule-001 guard for a defaulted switch: an explicit false is a
		// legal, meaningful shape ("declared, off"), never a validation error.
		target.Spec.AppCheck.PlayIntegrity.Enabled = proto.Bool(false)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.AppCheck.PlayIntegrity.Enabled = proto.Bool(true)
		target.Spec.AppCheck.PlayIntegrity.TokenTtl = "604800s"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing display name or package name", func() {
		target := minimal()
		target.Spec.DisplayName = ""
		expectError(target, "display_name")

		target = minimal()
		target.Spec.PackageName = ""
		expectError(target, "package_name")
	})

	ginkgo.It("should reject a package name that is not a Java package", func() {
		for _, pkg := range []string{"app", "com.example.1app", "com.example-app", "com..app", ".com.example"} {
			target := minimal()
			target.Spec.PackageName = pkg
			expectError(target, "package_name must be a valid Java package name")
		}
	})

	ginkgo.It("should reject malformed or duplicated certificate fingerprints", func() {
		target := minimal()
		target.Spec.Sha1Hashes = []string{"da39a3ee5e6b4b0d3255bfef95601890afd8070"}
		expectError(target, "SHA-1")

		target = minimal()
		target.Spec.Sha256Hashes = []string{"not-a-fingerprint"}
		expectError(target, "SHA-256")

		target = minimal()
		target.Spec.Sha1Hashes = []string{
			"da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"da39a3ee5e6b4b0d3255bfef95601890afd80709",
		}
		expectError(target, "sha1_hashes")
	})

	ginkgo.It("should reject a malformed token_ttl", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAndroidAppAppCheck{
			PlayIntegrity: &GcpFirebaseAndroidAppAppCheckPlayIntegrity{TokenTtl: "1h"},
		}
		expectError(target, "token_ttl must be a duration")
	})

	ginkgo.It("should reject a debug token without a name or a value, and duplicated names", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseAndroidAppAppCheck{
			DebugTokens: []*GcpFirebaseAndroidAppAppCheckDebugToken{{Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"}},
		}
		expectError(target, "display_name")

		target.Spec.AppCheck.DebugTokens = []*GcpFirebaseAndroidAppAppCheckDebugToken{{DisplayName: "ci"}}
		expectError(target, "token")

		target.Spec.AppCheck.DebugTokens = []*GcpFirebaseAndroidAppAppCheckDebugToken{
			{DisplayName: "ci", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
			{DisplayName: "ci", Token: "7d2e9f1b-4a6c-4e8d-b0f3-1c5a7e9b2d4f"},
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
		target.Kind = "GcpFirebaseApp"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
