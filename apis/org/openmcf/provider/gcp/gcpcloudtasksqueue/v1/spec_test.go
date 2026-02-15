package gcpcloudtasksqueuev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudTasksQueueSpec Suite")
}

var _ = ginkgo.Describe("GcpCloudTasksQueueSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpCloudTasksQueue.
	minimal := func() *GcpCloudTasksQueue {
		return &GcpCloudTasksQueue{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudTasksQueue",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-queue",
			},
			Spec: &GcpCloudTasksQueueSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				QueueName: "my-task-queue",
				Location:  "us-central1",
			},
		}
	}

	// Helper to build a StringValueOrRef with a literal value.
	strRef := func(val string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: val,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept queue_name with letters only", func() {
		msg := minimal()
		msg.Spec.QueueName = "mytaskqueue"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept queue_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.QueueName = "task-queue-123"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept single character queue_name", func() {
		msg := minimal()
		msg.Spec.QueueName = "q"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept 63-character queue_name", func() {
		msg := minimal()
		msg.Spec.QueueName = "a" + strings.Repeat("b", 62)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept desired_state RUNNING", func() {
		msg := minimal()
		msg.Spec.DesiredState = "RUNNING"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept desired_state PAUSED", func() {
		msg := minimal()
		msg.Spec.DesiredState = "PAUSED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty desired_state (defaults to RUNNING)", func() {
		msg := minimal()
		msg.Spec.DesiredState = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rate_limits with dispatches and concurrency", func() {
		msg := minimal()
		msg.Spec.RateLimits = &GcpCloudTasksQueueRateLimits{
			MaxDispatchesPerSecond:  500.0,
			MaxConcurrentDispatches: 100,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept retry_config with all fields", func() {
		msg := minimal()
		msg.Spec.RetryConfig = &GcpCloudTasksQueueRetryConfig{
			MaxAttempts:      5,
			MaxRetryDuration: "3600s",
			MinBackoff:       "0.100s",
			MaxBackoff:       "3600s",
			MaxDoublings:     16,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept retry_config with unlimited attempts (-1)", func() {
		msg := minimal()
		msg.Spec.RetryConfig = &GcpCloudTasksQueueRetryConfig{
			MaxAttempts: -1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept stackdriver_logging_config with 0.0 ratio", func() {
		msg := minimal()
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: 0.0,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept stackdriver_logging_config with 1.0 ratio", func() {
		msg := minimal()
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: 1.0,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept stackdriver_logging_config with 0.5 ratio", func() {
		msg := minimal()
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: 0.5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept http_target with POST method", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HttpMethod: "POST",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept http_target with all HTTP methods", func() {
		methods := []string{"POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}
		for _, method := range methods {
			msg := minimal()
			msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
				HttpMethod: method,
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "method %s should be valid", method)
		}
	})

	ginkgo.It("should accept http_target with oidc_token", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HttpMethod: "POST",
			OidcToken: &GcpCloudTasksQueueOidcToken{
				ServiceAccountEmail: strRef("task-invoker@my-project.iam.gserviceaccount.com"),
				Audience:            "https://my-service-abc123-uc.a.run.app",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept http_target with oauth_token", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			OauthToken: &GcpCloudTasksQueueOAuthToken{
				ServiceAccountEmail: strRef("task-invoker@my-project.iam.gserviceaccount.com"),
				Scope:               "https://www.googleapis.com/auth/cloud-platform",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept http_target with header_overrides", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HeaderOverrides: []*GcpCloudTasksQueueHttpHeaderOverride{
				{Key: "Content-Type", Value: "application/json"},
				{Key: "X-Custom-Header", Value: "custom-value"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept http_target with uri_override", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Scheme:      "HTTPS",
				Host:        "my-service-abc123-uc.a.run.app",
				EnforceMode: "ALWAYS",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept uri_override with IF_NOT_EXISTS enforce_mode", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Host:        "example.com",
				EnforceMode: "IF_NOT_EXISTS",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept uri_override with HTTP scheme", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Scheme: "HTTP",
				Host:   "internal-service.local",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept uri_override with path and query_params", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Host:        "api.example.com",
				Port:        "8080",
				Path:        "/v1/tasks/process",
				QueryParams: "priority=high&retry=true",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full-featured spec", func() {
		msg := minimal()
		msg.Spec.DesiredState = "RUNNING"
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HttpMethod: "POST",
			HeaderOverrides: []*GcpCloudTasksQueueHttpHeaderOverride{
				{Key: "Content-Type", Value: "application/json"},
			},
			OidcToken: &GcpCloudTasksQueueOidcToken{
				ServiceAccountEmail: strRef("task-invoker@my-project.iam.gserviceaccount.com"),
				Audience:            "https://my-service.run.app",
			},
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Scheme:      "HTTPS",
				Host:        "my-service-abc123-uc.a.run.app",
				Path:        "/v1/process",
				EnforceMode: "ALWAYS",
			},
		}
		msg.Spec.RateLimits = &GcpCloudTasksQueueRateLimits{
			MaxDispatchesPerSecond:  500.0,
			MaxConcurrentDispatches: 100,
		}
		msg.Spec.RetryConfig = &GcpCloudTasksQueueRetryConfig{
			MaxAttempts:      5,
			MaxRetryDuration: "3600s",
			MinBackoff:       "1s",
			MaxBackoff:       "3600s",
			MaxDoublings:     16,
		}
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: 0.1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing queue_name", func() {
		msg := minimal()
		msg.Spec.QueueName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject queue_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.QueueName = "1queue"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject queue_name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.QueueName = "-queue"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject queue_name with underscores", func() {
		msg := minimal()
		msg.Spec.QueueName = "my_queue"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject queue_name with dots", func() {
		msg := minimal()
		msg.Spec.QueueName = "my.queue"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject queue_name exceeding 63 characters", func() {
		msg := minimal()
		msg.Spec.QueueName = "a" + strings.Repeat("b", 63)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid desired_state", func() {
		msg := minimal()
		msg.Spec.DesiredState = "STOPPED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid http_method", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HttpMethod: "CONNECT",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid uri_override scheme", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				Scheme: "FTP",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid uri_override enforce_mode", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			UriOverride: &GcpCloudTasksQueueUriOverride{
				EnforceMode: "NEVER",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both oauth_token and oidc_token set", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			OauthToken: &GcpCloudTasksQueueOAuthToken{
				ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
			},
			OidcToken: &GcpCloudTasksQueueOidcToken{
				ServiceAccountEmail: strRef("sa@project.iam.gserviceaccount.com"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject oauth_token without service_account_email", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			OauthToken: &GcpCloudTasksQueueOAuthToken{
				Scope: "https://www.googleapis.com/auth/cloud-platform",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject oidc_token without service_account_email", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			OidcToken: &GcpCloudTasksQueueOidcToken{
				Audience: "https://my-service.run.app",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject sampling_ratio above 1.0", func() {
		msg := minimal()
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: 1.5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject sampling_ratio below 0.0", func() {
		msg := minimal()
		msg.Spec.StackdriverLoggingConfig = &GcpCloudTasksQueueLoggingConfig{
			SamplingRatio: -0.1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject header_override with missing key", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HeaderOverrides: []*GcpCloudTasksQueueHttpHeaderOverride{
				{Key: "", Value: "some-value"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject header_override with missing value", func() {
		msg := minimal()
		msg.Spec.HttpTarget = &GcpCloudTasksQueueHttpTarget{
			HeaderOverrides: []*GcpCloudTasksQueueHttpHeaderOverride{
				{Key: "Content-Type", Value: ""},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// Verify proto is importable and not nil (compile-time sanity).
	ginkgo.It("should produce a non-nil proto message", func() {
		msg := minimal()
		gomega.Expect(proto.Clone(msg)).ToNot(gomega.BeNil())
	})
})
