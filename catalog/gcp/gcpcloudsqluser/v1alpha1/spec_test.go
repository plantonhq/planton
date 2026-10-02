package gcpcloudsqluserv1alpha1

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
	ginkgo.RunSpecs(t, "GcpCloudSqlUserSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func fromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func ptr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}

var _ = ginkgo.Describe("GcpCloudSqlUserSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalBuiltIn := func() *GcpCloudSqlUser {
		return &GcpCloudSqlUser{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudSqlUser",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-user",
			},
			Spec: &GcpCloudSqlUserSpec{
				Instance: litRef("orders-db"),
				UserName: "orders-app",
				Password: "AppPassword123!",
			},
		}
	}

	expectValid := func(r *GcpCloudSqlUser) {
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	}

	expectInvalid := func(r *GcpCloudSqlUser, substr string) {
		err := validator.Validate(r)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), substr)).To(
			gomega.BeTrue(), "expected error to contain %q, got: %s", substr, err)
	}

	ginkgo.It("accepts a minimal built-in user", func() {
		expectValid(minimalBuiltIn())
	})

	ginkgo.It("accepts a passwordless built-in user (password set post-deploy)", func() {
		r := minimalBuiltIn()
		r.Spec.Password = ""
		expectValid(r)
	})

	ginkgo.It("rejects a missing instance", func() {
		r := minimalBuiltIn()
		r.Spec.Instance = nil
		expectInvalid(r, "instance")
	})

	ginkgo.It("accepts an instance reference", func() {
		r := minimalBuiltIn()
		r.Spec.Instance = fromRef("orders-db")
		expectValid(r)
	})

	ginkgo.It("rejects a missing user_name", func() {
		r := minimalBuiltIn()
		r.Spec.UserName = ""
		expectInvalid(r, "user_name")
	})

	ginkgo.It("rejects an unknown type", func() {
		r := minimalBuiltIn()
		r.Spec.Type = ptr("SUPERUSER")
		expectInvalid(r, "type")
	})

	ginkgo.It("rejects a password on an IAM user", func() {
		r := minimalBuiltIn()
		r.Spec.Type = ptr("CLOUD_IAM_USER")
		r.Spec.UserName = "dev@example.com"
		expectInvalid(r, "must not set a password")
	})

	ginkgo.It("accepts an IAM service account user without a password", func() {
		r := minimalBuiltIn()
		r.Spec.Type = ptr("CLOUD_IAM_SERVICE_ACCOUNT")
		r.Spec.UserName = "ci-runner@my-project.iam.gserviceaccount.com"
		r.Spec.Password = ""
		expectValid(r)
	})

	ginkgo.It("accepts an IAM group without a password", func() {
		r := minimalBuiltIn()
		r.Spec.Type = ptr("CLOUD_IAM_GROUP")
		r.Spec.UserName = "db-readers@example.com"
		r.Spec.Password = ""
		expectValid(r)
	})

	ginkgo.It("rejects a password policy on an IAM user", func() {
		r := minimalBuiltIn()
		r.Spec.Type = ptr("CLOUD_IAM_USER")
		r.Spec.UserName = "dev@example.com"
		r.Spec.Password = ""
		r.Spec.PasswordPolicy = &GcpCloudSqlUserPasswordPolicy{EnableFailedAttemptsCheck: true}
		expectInvalid(r, "password_policy applies to BUILT_IN")
	})

	ginkgo.It("accepts a MySQL host-scoped user", func() {
		r := minimalBuiltIn()
		r.Spec.Host = "%"
		expectValid(r)
	})

	ginkgo.It("accepts a full password policy", func() {
		r := minimalBuiltIn()
		r.Spec.PasswordPolicy = &GcpCloudSqlUserPasswordPolicy{
			AllowedFailedAttempts:      intPtr(5),
			PasswordExpirationDuration: "2592000s",
			EnableFailedAttemptsCheck:  true,
			EnablePasswordVerification: true,
		}
		expectValid(r)
	})

	ginkgo.It("rejects a malformed expiration duration", func() {
		r := minimalBuiltIn()
		r.Spec.PasswordPolicy = &GcpCloudSqlUserPasswordPolicy{
			PasswordExpirationDuration: "30d",
		}
		expectInvalid(r, "seconds duration")
	})

	ginkgo.It("rejects zero allowed_failed_attempts", func() {
		r := minimalBuiltIn()
		r.Spec.PasswordPolicy = &GcpCloudSqlUserPasswordPolicy{
			AllowedFailedAttempts: intPtr(0),
		}
		expectInvalid(r, "allowed_failed_attempts")
	})

	ginkgo.It("rejects a wrong kind constant", func() {
		r := minimalBuiltIn()
		r.Kind = "GcpCloudSql"
		expectInvalid(r, "kind")
	})

	ginkgo.It("accepts database roles", func() {
		r := minimalBuiltIn()
		r.Spec.DatabaseRoles = []string{"cloudsqlsuperuser", "reporting_reader"}
		expectValid(r)
	})

	ginkgo.It("rejects duplicate database roles", func() {
		r := minimalBuiltIn()
		r.Spec.DatabaseRoles = []string{"cloudsqlsuperuser", "cloudsqlsuperuser"}
		expectInvalid(r, "database_roles")
	})

	ginkgo.It("accepts every valid deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			r := minimalBuiltIn()
			r.Spec.DeletionPolicy = policy
			expectValid(r)
		}
	})

	ginkgo.It("rejects an unknown deletion_policy", func() {
		r := minimalBuiltIn()
		r.Spec.DeletionPolicy = "RETAIN"
		expectInvalid(r, "deletion_policy")
	})

	ginkgo.Context("service account users", func() {
		ginkgo.It("accepts a service account by reference with the IAM service-account type", func() {
			r := minimalBuiltIn()
			r.Spec.UserName = ""
			r.Spec.Password = ""
			r.Spec.Type = ptr("CLOUD_IAM_SERVICE_ACCOUNT")
			r.Spec.ServiceAccount = fromRef("orders-runner")
			expectValid(r)
		})

		ginkgo.It("accepts a literal service account email", func() {
			r := minimalBuiltIn()
			r.Spec.UserName = ""
			r.Spec.Password = ""
			r.Spec.Type = ptr("CLOUD_IAM_SERVICE_ACCOUNT")
			r.Spec.ServiceAccount = litRef("orders-runner@my-project.iam.gserviceaccount.com")
			expectValid(r)
		})

		ginkgo.It("rejects a service account without the IAM service-account type", func() {
			r := minimalBuiltIn()
			r.Spec.UserName = ""
			r.Spec.Password = ""
			r.Spec.ServiceAccount = litRef("orders-runner@my-project.iam.gserviceaccount.com")
			expectInvalid(r, "CLOUD_IAM_SERVICE_ACCOUNT")
		})

		ginkgo.It("rejects both user_name and service_account", func() {
			r := minimalBuiltIn()
			r.Spec.Password = ""
			r.Spec.Type = ptr("CLOUD_IAM_SERVICE_ACCOUNT")
			r.Spec.ServiceAccount = litRef("orders-runner@my-project.iam.gserviceaccount.com")
			expectInvalid(r, "exactly one")
		})

		ginkgo.It("rejects neither user_name nor service_account", func() {
			r := minimalBuiltIn()
			r.Spec.UserName = ""
			expectInvalid(r, "exactly one")
		})
	})
})
