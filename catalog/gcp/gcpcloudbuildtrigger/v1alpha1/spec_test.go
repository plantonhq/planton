package gcpcloudbuildtriggerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudBuildTriggerSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind cloudresourcekind.CloudResourceKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

const repositoryName = "projects/acme-ci/locations/us-central1/connections/acme-github/repositories/acme-orders"

var _ = ginkgo.Describe("GcpCloudBuildTriggerSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	step := func() *GcpCloudBuildTriggerBuildStep {
		return &GcpCloudBuildTriggerBuildStep{Name: "ubuntu", Args: []string{"echo", "hello"}}
	}

	// The canonical shape: a Pub/Sub trigger with an inline build.
	pubsub := func() *GcpCloudBuildTrigger {
		return &GcpCloudBuildTrigger{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudBuildTrigger",
			Metadata:   &shared.CloudResourceMetadata{Name: "on-release"},
			Spec: &GcpCloudBuildTriggerSpec{
				ServiceAccount: reference(cloudresourcekind.CloudResourceKind_GcpServiceAccount, "builder"),
				PubsubConfig:   &GcpCloudBuildTriggerPubsubConfig{Topic: reference(cloudresourcekind.CloudResourceKind_GcpPubSubTopic, "releases")},
				Build: &GcpCloudBuildTriggerBuild{
					Steps:   []*GcpCloudBuildTriggerBuildStep{step()},
					Options: &GcpCloudBuildTriggerBuildOptions{Logging: "CLOUD_LOGGING_ONLY"},
				},
			},
		}
	}

	// A push trigger on a linked repository reading cloudbuild.yaml.
	repository := func() *GcpCloudBuildTrigger {
		msg := pubsub()
		msg.Spec.PubsubConfig = nil
		msg.Spec.Build = nil
		msg.Spec.Location = "us-central1"
		msg.Spec.Filename = "cloudbuild.yaml"
		msg.Spec.RepositoryEventConfig = &GcpCloudBuildTriggerRepositoryEventConfig{
			Repository: reference(cloudresourcekind.CloudResourceKind_GcpCloudBuildRepository, "orders"),
			Push:       &GcpCloudBuildTriggerPushFilter{Branch: "^main$"},
		}
		return msg
	}

	ginkgo.It("should accept the canonical shapes and a fully declared build", func() {
		gomega.Expect(validator.Validate(pubsub())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(repository())).To(gomega.Succeed())

		full := pubsub()
		full.Spec.ProjectId = literal("acme-ci")
		full.Spec.TriggerName = "On-release-2"
		full.Spec.Description = "Builds every release"
		full.Spec.Tags = []string{"release"}
		full.Spec.Substitutions = map[string]string{"_DEPLOY_ENV": "prod"}
		full.Spec.ServiceAccount = literal("projects/acme-ci/serviceAccounts/builder@acme-ci.iam.gserviceaccount.com")
		full.Spec.ApprovalConfig = &GcpCloudBuildTriggerApprovalConfig{ApprovalRequired: true}
		full.Spec.Filter = "_ACTION == 'INSERT'"
		full.Spec.SourceToBuild = &GcpCloudBuildTriggerSourceToBuild{Repository: literal(repositoryName), Ref: "refs/heads/main", RepoType: "GITHUB"}
		full.Spec.Build = &GcpCloudBuildTriggerBuild{
			Steps: []*GcpCloudBuildTriggerBuildStep{
				{Name: "gcr.io/cloud-builders/docker", Id: "build", Args: []string{"build", "-t", "img", "."}, Env: []string{"A=1"}, SecretEnv: []string{"TOKEN"}, Timeout: "300s", AllowExitCodes: []int32{1}, Volumes: []*GcpCloudBuildTriggerBuildStepVolume{{Name: "cache", Path: "/cache"}}},
				{Name: "ubuntu", Script: "echo done", WaitFor: []string{"build"}, Volumes: []*GcpCloudBuildTriggerBuildStepVolume{{Name: "cache", Path: "/cache"}}},
			},
			Timeout:    "1200s",
			Images:     []string{"img"},
			LogsBucket: reference(cloudresourcekind.CloudResourceKind_GcpGcsBucket, "build-logs"),
			Artifacts: &GcpCloudBuildTriggerBuildArtifacts{
				Objects:        &GcpCloudBuildTriggerBuildArtifactsObjects{Location: "gs://acme-artifacts/", Paths: []string{"dist/*"}},
				MavenArtifacts: []*GcpCloudBuildTriggerBuildArtifactsMavenArtifact{{Repository: "https://us-central1-maven.pkg.dev/acme-ci/maven", Path: "app.jar"}},
			},
			Options: &GcpCloudBuildTriggerBuildOptions{
				MachineType:          "E2_HIGHCPU_8",
				DiskSizeGb:           200,
				WorkerPool:           reference(cloudresourcekind.CloudResourceKind_GcpCloudBuildWorkerPool, "private"),
				Logging:              "GCS_ONLY",
				SourceProvenanceHash: []string{"SHA256"},
			},
			Source:           &GcpCloudBuildTriggerBuildSource{StorageSource: &GcpCloudBuildTriggerBuildSourceStorageSource{Bucket: "acme-src", Object: "src.tar.gz"}},
			AvailableSecrets: &GcpCloudBuildTriggerBuildAvailableSecrets{SecretManager: []*GcpCloudBuildTriggerBuildSecretManagerSecret{{Env: "TOKEN", VersionName: reference(cloudresourcekind.CloudResourceKind_GcpSecretManagerSecret, "token")}}},
			Secrets:          []*GcpCloudBuildTriggerBuildSecret{{KmsKeyName: literal("projects/acme-ci/locations/global/keyRings/ci/cryptoKeys/builds"), SecretEnv: map[string]string{"OLD": "Y2lwaGVy"}}},
		}
		gomega.Expect(validator.Validate(full)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one build configuration", func() {
		msg := pubsub()
		msg.Spec.Filename = "cloudbuild.yaml"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = pubsub()
		msg.Spec.Build = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = pubsub()
		msg.Spec.Build = nil
		msg.Spec.GitFileSource = &GcpCloudBuildTriggerGitFileSource{Path: "cloudbuild.yaml", RepoType: "GITHUB", Repository: literal(repositoryName), Revision: "refs/heads/main"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Filename = "cloudbuild.yaml"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse github together with trigger_template", func() {
		msg := repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.Github = &GcpCloudBuildTriggerGithub{Owner: "acme", Name: "orders", Push: &GcpCloudBuildTriggerPushFilter{Tag: "^v"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.TriggerTemplate = &GcpCloudBuildTriggerTriggerTemplate{BranchName: "^main$"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Github = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of pull_request or push on every repository source", func() {
		pr := &GcpCloudBuildTriggerPullRequestFilter{Branch: "^main$"}
		push := &GcpCloudBuildTriggerPushFilter{Branch: "^main$"}

		msg := repository()
		msg.Spec.RepositoryEventConfig.PullRequest = pr
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.RepositoryEventConfig.Push = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.RepositoryEventConfig.PullRequest = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.Github = &GcpCloudBuildTriggerGithub{Owner: "acme", Name: "orders", PullRequest: pr, Push: push}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Github = &GcpCloudBuildTriggerGithub{Owner: "acme", Name: "orders"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.BitbucketServerTriggerConfig = &GcpCloudBuildTriggerBitbucketServerTriggerConfig{BitbucketServerConfigResource: "projects/p/locations/global/bitbucketServerConfigs/b", ProjectKey: "TEST", RepoSlug: "repo", Push: push}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.BitbucketServerTriggerConfig.PullRequest = pr
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.DeveloperConnectEventConfig = &GcpCloudBuildTriggerDeveloperConnectEventConfig{GitRepositoryLink: "projects/p/locations/us-central1/connections/c/gitRepositoryLinks/l", PullRequest: &GcpCloudBuildTriggerPullRequestFilter{}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.DeveloperConnectEventConfig.PullRequest = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require the pull request branch where Google does", func() {
		msg := repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.Github = &GcpCloudBuildTriggerGithub{Owner: "acme", Name: "orders", PullRequest: &GcpCloudBuildTriggerPullRequestFilter{CommentControl: "COMMENTS_ENABLED"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Github.PullRequest.Branch = "^main$"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Github.PullRequest.CommentControl = "ALWAYS"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.BitbucketServerTriggerConfig = &GcpCloudBuildTriggerBitbucketServerTriggerConfig{BitbucketServerConfigResource: "projects/p/locations/global/bitbucketServerConfigs/b", ProjectKey: "TEST", RepoSlug: "repo", PullRequest: &GcpCloudBuildTriggerPullRequestFilter{}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of branch or tag on a push filter", func() {
		msg := repository()
		msg.Spec.RepositoryEventConfig.Push = &GcpCloudBuildTriggerPushFilter{Branch: "^main$", Tag: "^v"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.RepositoryEventConfig.Push = &GcpCloudBuildTriggerPushFilter{InvertRegex: true}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one revision on trigger_template and repo_source", func() {
		msg := repository()
		msg.Spec.RepositoryEventConfig = nil
		msg.Spec.TriggerTemplate = &GcpCloudBuildTriggerTriggerTemplate{RepoName: "orders"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.TriggerTemplate = &GcpCloudBuildTriggerTriggerTemplate{BranchName: "^main$", TagName: "^v"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = pubsub()
		msg.Spec.Build.Source = &GcpCloudBuildTriggerBuildSource{RepoSource: &GcpCloudBuildTriggerBuildSourceRepoSource{RepoName: "orders", CommitSha: "abc123"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Build.Source.RepoSource.BranchName = "^main$"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Source.RepoSource = &GcpCloudBuildTriggerBuildSourceRepoSource{CommitSha: "abc123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one build source arm", func() {
		msg := pubsub()
		msg.Spec.Build.Source = &GcpCloudBuildTriggerBuildSource{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Source = &GcpCloudBuildTriggerBuildSource{
			RepoSource:    &GcpCloudBuildTriggerBuildSourceRepoSource{RepoName: "orders", TagName: "^v"},
			StorageSource: &GcpCloudBuildTriggerBuildSourceStorageSource{Bucket: "b", Object: "o.tar.gz"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold source_to_build and git_file_source to one repository", func() {
		msg := pubsub()
		msg.Spec.SourceToBuild = &GcpCloudBuildTriggerSourceToBuild{Uri: "https://github.com/acme/orders", Ref: "refs/heads/main", RepoType: "GITHUB"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.SourceToBuild.Repository = literal(repositoryName)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.SourceToBuild = &GcpCloudBuildTriggerSourceToBuild{Ref: "refs/heads/main", RepoType: "GITHUB"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.SourceToBuild = &GcpCloudBuildTriggerSourceToBuild{Uri: "https://github.com/acme/orders", Ref: "main", RepoType: "GITHUB"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.SourceToBuild = &GcpCloudBuildTriggerSourceToBuild{Uri: "https://gitlab.com/acme/orders", Ref: "refs/heads/main", RepoType: "GITLAB"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = pubsub()
		msg.Spec.Build = nil
		msg.Spec.GitFileSource = &GcpCloudBuildTriggerGitFileSource{Path: "cloudbuild.yaml", RepoType: "GITHUB"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.GitFileSource.Uri = "https://github.com/acme/orders"
		msg.Spec.GitFileSource.Repository = reference(cloudresourcekind.CloudResourceKind_GcpCloudBuildRepository, "orders")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.GitFileSource = &GcpCloudBuildTriggerGitFileSource{RepoType: "GITHUB"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse a step that mixes script with entrypoint or args", func() {
		msg := pubsub()
		msg.Spec.Build.Steps = []*GcpCloudBuildTriggerBuildStep{{Name: "ubuntu", Script: "echo hi", Args: []string{"x"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Steps = []*GcpCloudBuildTriggerBuildStep{{Name: "ubuntu", Script: "echo hi", Entrypoint: "bash"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Steps = []*GcpCloudBuildTriggerBuildStep{{Name: "ubuntu", Script: "echo hi"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require at least one step with an image, and complete volumes", func() {
		msg := pubsub()
		msg.Spec.Build.Steps = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Steps = []*GcpCloudBuildTriggerBuildStep{{Args: []string{"echo"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Steps = []*GcpCloudBuildTriggerBuildStep{{Name: "ubuntu", Volumes: []*GcpCloudBuildTriggerBuildStepVolume{{Name: "cache"}}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold secrets to their shapes", func() {
		msg := pubsub()
		msg.Spec.Build.AvailableSecrets = &GcpCloudBuildTriggerBuildAvailableSecrets{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.AvailableSecrets = &GcpCloudBuildTriggerBuildAvailableSecrets{SecretManager: []*GcpCloudBuildTriggerBuildSecretManagerSecret{{Env: "TOKEN", VersionName: literal("projects/p/secrets/token")}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.AvailableSecrets.SecretManager[0].VersionName = literal("projects/p/secrets/token/versions/latest")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Build.AvailableSecrets.SecretManager[0].Env = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = pubsub()
		msg.Spec.Build.Secrets = []*GcpCloudBuildTriggerBuildSecret{{KmsKeyName: literal("builds")}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Build.Secrets = []*GcpCloudBuildTriggerBuildSecret{{}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold the webhook secret, topic, and repository to their shapes", func() {
		msg := pubsub()
		msg.Spec.PubsubConfig = nil
		msg.Spec.WebhookConfig = &GcpCloudBuildTriggerWebhookConfig{Secret: reference(cloudresourcekind.CloudResourceKind_GcpSecretManagerSecret, "hook")}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.WebhookConfig.Secret = literal("projects/p/secrets/hook")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.WebhookConfig.Secret = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = pubsub()
		msg.Spec.PubsubConfig.Topic = literal("releases")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.PubsubConfig.Topic = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = repository()
		msg.Spec.RepositoryEventConfig.Repository = literal("acme-orders")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse malformed names, substitutions, enums, and service accounts", func() {
		for _, mutate := range []func(*GcpCloudBuildTrigger){
			func(m *GcpCloudBuildTrigger) { m.Spec.TriggerName = "-release" },
			func(m *GcpCloudBuildTrigger) { m.Spec.TriggerName = "on_release" },
			func(m *GcpCloudBuildTrigger) { m.Spec.Substitutions = map[string]string{"DEPLOY_ENV": "prod"} },
			func(m *GcpCloudBuildTrigger) { m.Spec.Substitutions = map[string]string{"_deploy": "prod"} },
			func(m *GcpCloudBuildTrigger) { m.Spec.IncludeBuildLogs = "ALWAYS" },
			func(m *GcpCloudBuildTrigger) {
				m.Spec.ServiceAccount = literal("builder@acme-ci.iam.gserviceaccount.com")
			},
			func(m *GcpCloudBuildTrigger) { m.Spec.DeletionPolicy = "KEEP" },
			func(m *GcpCloudBuildTrigger) { m.Spec.Build.Options.Logging = "STDOUT" },
			func(m *GcpCloudBuildTrigger) { m.Spec.Build.Options.LogStreamingOption = "STREAM" },
			func(m *GcpCloudBuildTrigger) { m.Spec.Build.Options.RequestedVerifyOption = "MAYBE" },
			func(m *GcpCloudBuildTrigger) { m.Spec.Build.Options.SourceProvenanceHash = []string{"SHA1"} },
			func(m *GcpCloudBuildTrigger) { m.Spec.Build.Options.DiskSizeGb = -1 },
		} {
			msg := pubsub()
			mutate(msg)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})
})
