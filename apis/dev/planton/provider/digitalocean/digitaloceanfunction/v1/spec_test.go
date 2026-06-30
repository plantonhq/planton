package digitaloceanfunctionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
)

func TestDigitalOceanFunctionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanFunctionSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanFunctionSpec validations", func() {

	// Helper function to create a minimal valid HTTP function spec
	makeValidHttpFunctionSpec := func() *DigitalOceanFunctionSpec {
		return &DigitalOceanFunctionSpec{
			FunctionName: "api-handler",
			Region:       digitalocean.DigitalOceanRegion_nyc3,
			Runtime:      DigitalOceanFunctionRuntime_nodejs_18,
			GithubSource: &DigitalOceanFunctionGithubSource{
				Repo:         "myorg/my-functions",
				Branch:       "main",
				DeployOnPush: true,
			},
			SourceDirectory: "/functions/api-handler",
			MemoryMb:        256,
			TimeoutMs:       3000,
			IsWeb:           true,
		}
	}

	// Helper function to create a scheduled function spec
	makeValidScheduledFunctionSpec := func() *DigitalOceanFunctionSpec {
		return &DigitalOceanFunctionSpec{
			FunctionName: "nightly-cleanup",
			Region:       digitalocean.DigitalOceanRegion_nyc3,
			Runtime:      DigitalOceanFunctionRuntime_python_311,
			GithubSource: &DigitalOceanFunctionGithubSource{
				Repo:   "myorg/my-functions",
				Branch: "main",
			},
			SourceDirectory: "/functions/cleanup",
			MemoryMb:        512,
			TimeoutMs:       60000,
			CronSchedule:    "0 0 * * *",
			IsWeb:           false,
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid HTTP function spec", func() {
			spec := makeValidHttpFunctionSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a minimal valid scheduled function spec", func() {
			spec := makeValidScheduledFunctionSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with empty function_name", func() {
			spec := makeValidHttpFunctionSpec()
			spec.FunctionName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing region", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_digital_ocean_function_runtime_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("function_name validation", func() {
		ginkgo.It("accepts function_name with 64 characters (max)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.FunctionName = "a123456789b123456789c123456789d123456789e123456789f123456789abcd" // 64 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects function_name exceeding 64 characters", func() {
			spec := makeValidHttpFunctionSpec()
			spec.FunctionName = "a123456789b123456789c123456789d123456789e123456789f123456789abcde" // 65 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts function_name with hyphens and numbers", func() {
			spec := makeValidHttpFunctionSpec()
			spec.FunctionName = "api-handler-v2"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("GitHub source validation", func() {
		ginkgo.It("accepts valid GitHub repo format (owner/repo)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Repo = "planton/functions"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts repo with hyphens and underscores", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Repo = "my-org_123/my-repo_456"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects invalid repo format (missing slash)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Repo = "invalidrepo"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects repo with special characters", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Repo = "my@org/my$repo"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts valid branch names", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Branch = "feature/new-api"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects empty branch name", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.Branch = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts deploy_on_push = true", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.DeployOnPush = true
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts deploy_on_push = false", func() {
			spec := makeValidHttpFunctionSpec()
			spec.GithubSource.DeployOnPush = false
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("source_directory validation", func() {
		ginkgo.It("accepts source_directory with leading slash", func() {
			spec := makeValidHttpFunctionSpec()
			spec.SourceDirectory = "/src/functions/api"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts source_directory without leading slash", func() {
			spec := makeValidHttpFunctionSpec()
			spec.SourceDirectory = "functions/api"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects empty source_directory", func() {
			spec := makeValidHttpFunctionSpec()
			spec.SourceDirectory = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Runtime validation", func() {
		ginkgo.It("accepts Node.js 18 runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_nodejs_18
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Node.js 20 runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_nodejs_20
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Python 3.11 runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_python_311
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Go 1.21 runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_go_121
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts PHP 8.2 runtime", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_php_82
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Memory allocation validation", func() {
		ginkgo.It("accepts 128 MB memory", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 128
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 256 MB memory (default)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 256
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 512 MB memory", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 512
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 1024 MB memory", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 1024
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 2048 MB memory", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 2048
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects invalid memory value (not in allowed list)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 384 // Not in allowed values
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects memory value of 0", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Timeout validation", func() {
		ginkgo.It("accepts 3000ms timeout (default)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.TimeoutMs = 3000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 60000ms timeout (1 minute)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.TimeoutMs = 60000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 300000ms timeout (5 minutes, max)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.TimeoutMs = 300000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects timeout exceeding 300000ms", func() {
			spec := makeValidHttpFunctionSpec()
			spec.TimeoutMs = 300001
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts minimal timeout of 1ms", func() {
			spec := makeValidHttpFunctionSpec()
			spec.TimeoutMs = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Environment variables validation", func() {
		ginkgo.It("accepts function with environment_variables", func() {
			spec := makeValidHttpFunctionSpec()
			spec.EnvironmentVariables = map[string]string{
				"LOG_LEVEL": "info",
				"NODE_ENV":  "production",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function with secret_environment_variables", func() {
			spec := makeValidHttpFunctionSpec()
			spec.SecretEnvironmentVariables = map[string]string{
				"DB_URL":  "postgresql://user:pass@host/db",
				"API_KEY": "secret-key-123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function with both environment and secret variables", func() {
			spec := makeValidHttpFunctionSpec()
			spec.EnvironmentVariables = map[string]string{
				"LOG_LEVEL": "debug",
			}
			spec.SecretEnvironmentVariables = map[string]string{
				"DB_PASSWORD": "secret-password",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function with empty environment_variables", func() {
			spec := makeValidHttpFunctionSpec()
			spec.EnvironmentVariables = map[string]string{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("HTTP vs Scheduled function patterns", func() {
		ginkgo.It("accepts HTTP function with is_web = true", func() {
			spec := makeValidHttpFunctionSpec()
			spec.IsWeb = true
			spec.CronSchedule = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts scheduled function with is_web = false and cron_schedule", func() {
			spec := makeValidScheduledFunctionSpec()
			spec.IsWeb = false
			spec.CronSchedule = "0 * * * *"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts valid cron schedule (hourly)", func() {
			spec := makeValidScheduledFunctionSpec()
			spec.CronSchedule = "0 * * * *"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts valid cron schedule (daily at midnight)", func() {
			spec := makeValidScheduledFunctionSpec()
			spec.CronSchedule = "0 0 * * *"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts valid cron schedule (every 15 minutes)", func() {
			spec := makeValidScheduledFunctionSpec()
			spec.CronSchedule = "*/15 * * * *"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Entrypoint validation", func() {
		ginkgo.It("accepts function with entrypoint for Go", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Runtime = DigitalOceanFunctionRuntime_go_120
			spec.Entrypoint = "main"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function with entrypoint for Node.js", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Entrypoint = "handler"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function without entrypoint (optional)", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Entrypoint = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Production patterns", func() {
		ginkgo.It("accepts production API function with secrets and monitoring", func() {
			spec := &DigitalOceanFunctionSpec{
				FunctionName: "prod-api",
				Region:       digitalocean.DigitalOceanRegion_nyc3,
				Runtime:      DigitalOceanFunctionRuntime_nodejs_20,
				GithubSource: &DigitalOceanFunctionGithubSource{
					Repo:         "mycompany/production-apis",
					Branch:       "main",
					DeployOnPush: true,
				},
				SourceDirectory: "/functions/user-api",
				MemoryMb:        1024,
				TimeoutMs:       15000,
				EnvironmentVariables: map[string]string{
					"NODE_ENV":  "production",
					"LOG_LEVEL": "info",
					"REGION":    "us-east",
				},
				SecretEnvironmentVariables: map[string]string{
					"DATABASE_URL":   "postgresql://prod-db:5432/users",
					"API_SECRET_KEY": "super-secret-key",
					"STRIPE_API_KEY": "sk_live_xxx",
				},
				IsWeb: true,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts background job with high memory and long timeout", func() {
			spec := &DigitalOceanFunctionSpec{
				FunctionName: "data-processor",
				Region:       digitalocean.DigitalOceanRegion_sfo3,
				Runtime:      DigitalOceanFunctionRuntime_python_311,
				GithubSource: &DigitalOceanFunctionGithubSource{
					Repo:         "mycompany/data-jobs",
					Branch:       "production",
					DeployOnPush: false,
				},
				SourceDirectory: "/jobs/processor",
				MemoryMb:        2048,
				TimeoutMs:       300000, // 5 minutes max
				SecretEnvironmentVariables: map[string]string{
					"S3_ACCESS_KEY": "aws-key",
					"S3_SECRET_KEY": "aws-secret",
				},
				CronSchedule: "0 2 * * *", // 2 AM daily
				IsWeb:        false,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Region validation", func() {
		ginkgo.It("accepts NYC3 region", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Region = digitalocean.DigitalOceanRegion_nyc3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts SFO3 region", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Region = digitalocean.DigitalOceanRegion_sfo3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts AMS3 region", func() {
			spec := makeValidHttpFunctionSpec()
			spec.Region = digitalocean.DigitalOceanRegion_ams3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Edge cases", func() {
		ginkgo.It("accepts function with minimal memory and timeout", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 128
			spec.TimeoutMs = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts function with maximum memory and timeout", func() {
			spec := makeValidHttpFunctionSpec()
			spec.MemoryMb = 2048
			spec.TimeoutMs = 300000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts source_directory with complex path", func() {
			spec := makeValidHttpFunctionSpec()
			spec.SourceDirectory = "/src/backend/services/api/functions/v2/users"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
