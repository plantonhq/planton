package gcpworkflowv1alpha1

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
	ginkgo.RunSpecs(t, "GcpWorkflowSpec Suite")
}

const minimalSource = `main:
  steps:
    - hello:
        return: "hello world"
`

var _ = ginkgo.Describe("GcpWorkflowSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpWorkflow {
		return &GcpWorkflow{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpWorkflow",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-workflow",
			},
			Spec: &GcpWorkflowSpec{
				SourceContents: minimalSource,
			},
		}
	}

	ginkgo.It("accepts the minimal manifest", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.Context("source_contents", func() {
		ginkgo.It("is required (the API truth the provider defers to 8.0.0)", func() {
			m := minimal()
			m.Spec.SourceContents = ""
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})

		ginkgo.It("enforces the API's 128KB size cap", func() {
			m := minimal()
			m.Spec.SourceContents = strings.Repeat("a", 131072)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.SourceContents = strings.Repeat("a", 131073)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("description", func() {
		ginkgo.It("enforces the provider's 1000-character cap", func() {
			m := minimal()
			m.Spec.Description = strings.Repeat("d", 1000)
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())

			m.Spec.Description = strings.Repeat("d", 1001)
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("call_log_level", func() {
		ginkgo.It("accepts the provider's ValidateEnum values and rejects others", func() {
			for _, v := range []string{"", "CALL_LOG_LEVEL_UNSPECIFIED", "LOG_ALL_CALLS", "LOG_ERRORS_ONLY", "LOG_NONE"} {
				m := minimal()
				m.Spec.CallLogLevel = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.CallLogLevel = "LOG_EVERYTHING"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("execution_history_level", func() {
		ginkgo.It("accepts the provider's ValidateEnum values and rejects others", func() {
			for _, v := range []string{"", "EXECUTION_HISTORY_LEVEL_UNSPECIFIED", "EXECUTION_HISTORY_BASIC", "EXECUTION_HISTORY_DETAILED"} {
				m := minimal()
				m.Spec.ExecutionHistoryLevel = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.ExecutionHistoryLevel = "EXECUTION_HISTORY_FULL"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})

	ginkgo.Context("user_env_vars", func() {
		ginkgo.It("enforces the API's 20-entry cap", func() {
			m := minimal()
			m.Spec.UserEnvVars = map[string]string{}
			for i := 0; i < 20; i++ {
				m.Spec.UserEnvVars["KEY_"+strings.Repeat("A", i+1)] = "v"
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "20 entries")

			m.Spec.UserEnvVars["ONE_MORE"] = "v"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed(), "21 entries")
		})
	})

	ginkgo.Context("secret_env_vars", func() {
		withSecrets := func() *GcpWorkflow {
			m := minimal()
			m.Spec.Region = "us-central1"
			m.Spec.UserEnvVars = map[string]string{"MODE": "prod"}
			m.Spec.SecretEnvVars = map[string]string{"API_TOKEN": "tok"}
			return m
		}

		ginkgo.It("accepts secret values beside literals", func() {
			gomega.Expect(validator.Validate(withSecrets())).To(gomega.Succeed())
		})

		ginkgo.It("counts both maps toward the API's 20-variable cap", func() {
			m := withSecrets()
			for i := 0; i < 18; i++ {
				m.Spec.SecretEnvVars["SECRET_"+strings.Repeat("A", i+1)] = "s"
			}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "1 literal + 19 secrets")
			m.Spec.SecretEnvVars["ONE_MORE"] = "s"
			err := validator.Validate(m)
			gomega.Expect(err).To(gomega.HaveOccurred(), "1 literal + 20 secrets")
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("together accept at most 20"))
		})

		ginkgo.It("refuses a variable in both maps", func() {
			m := withSecrets()
			m.Spec.UserEnvVars["API_TOKEN"] = "literal"
			err := validator.Validate(m)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("never both"))
		})

		ginkgo.It("needs a region for the secrets to replicate in", func() {
			m := withSecrets()
			m.Spec.Region = ""
			err := validator.Validate(m)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("secret_env_vars needs region"))
			m.Spec.SecretEnvVars = nil
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "a region is optional without secrets")
		})

		ginkgo.It("needs a literal service account named by email", func() {
			m := withSecrets()
			m.Spec.ServiceAccount = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "wf@acme.iam.gserviceaccount.com"}}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
			m.Spec.ServiceAccount.LiteralOrRef = &foreignkeyv1.StringValueOrRef_Value{Value: "projects/acme/serviceAccounts/wf@acme.iam.gserviceaccount.com"}
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed())
			m.Spec.ServiceAccount.LiteralOrRef = &foreignkeyv1.StringValueOrRef_Value{Value: "projects/acme/serviceAccounts/104510491012345678901"}
			err := validator.Validate(m)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must name the account by email"))
			m.Spec.SecretEnvVars = nil
			gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "a unique id is fine without secrets")
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				m := minimal()
				m.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})
})
