package gcpartifactregistryrepov1alpha1

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
	ginkgo.RunSpecs(t, "GcpArtifactRegistryRepoSpec Suite")
}

var _ = ginkgo.Describe("GcpArtifactRegistryRepoSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	literal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpArtifactRegistryRepo.
	minimal := func() *GcpArtifactRegistryRepo {
		return &GcpArtifactRegistryRepo{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpArtifactRegistryRepo",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-artifact-repo",
			},
			Spec: &GcpArtifactRegistryRepoSpec{
				Location: "us-central1",
				Format:   "DOCKER",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (location + format only)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a literal project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = literal("my-gcp-project-123")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a project_id reference", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "main-project"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit repository_id", func() {
		msg := minimal()
		msg.Spec.RepositoryId = "app-images"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a multi-region location", func() {
		msg := minimal()
		msg.Spec.Location = "us"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every well-known format", func() {
		for _, format := range []string{"DOCKER", "MAVEN", "NPM", "PYTHON", "GO", "APT", "YUM", "GENERIC", "KFP"} {
			msg := minimal()
			msg.Spec.Format = format
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "format %s should be accepted", format)
		}
	})

	ginkgo.It("should accept a future format value the API adds (free-string contract)", func() {
		msg := minimal()
		msg.Spec.Format = "RUBY"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept explicit STANDARD_REPOSITORY mode", func() {
		msg := minimal()
		msg.Spec.Mode = "STANDARD_REPOSITORY"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept description and labels", func() {
		msg := minimal()
		msg.Spec.Description = "container images for the api service"
		msg.Spec.Labels = map[string]string{"team": "platform", "env": "prod"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a CMEK key reference", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = literal("projects/p/locations/us-central1/keyRings/ring/cryptoKeys/key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept docker_config with immutable tags", func() {
		msg := minimal()
		msg.Spec.DockerConfig = &GcpArtifactRegistryRepoDockerConfig{ImmutableTags: true}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept maven_config with RELEASE version policy", func() {
		msg := minimal()
		msg.Spec.Format = "MAVEN"
		msg.Spec.MavenConfig = &GcpArtifactRegistryRepoMavenConfig{VersionPolicy: "RELEASE"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept maven_config with SNAPSHOT policy and overwrites", func() {
		msg := minimal()
		msg.Spec.Format = "MAVEN"
		msg.Spec.MavenConfig = &GcpArtifactRegistryRepoMavenConfig{
			VersionPolicy:           "SNAPSHOT",
			AllowSnapshotOverwrites: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a DELETE cleanup policy with an older_than condition", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "delete-old-untagged",
			Action: "DELETE",
			Condition: &GcpArtifactRegistryRepoCleanupPolicyCondition{
				OlderThan: "2592000s",
				TagState:  "UNTAGGED",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a KEEP cleanup policy with most_recent_versions", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "keep-last-10",
			Action: "KEEP",
			MostRecentVersions: &GcpArtifactRegistryRepoCleanupPolicyMostRecentVersions{
				KeepCount: 10,
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a KEEP cleanup policy with a newer_than condition", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "keep-recent",
			Action: "KEEP",
			Condition: &GcpArtifactRegistryRepoCleanupPolicyCondition{
				NewerThan: "604800s",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cleanup_policy_dry_run", func() {
		msg := minimal()
		msg.Spec.CleanupPolicyDryRun = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a REMOTE_REPOSITORY caching Docker Hub", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a remote repository with a common_repository upstream", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			CommonRepository: &GcpArtifactRegistryRepoRemoteCommonRepository{
				Uri: literal("https://registry.company.com"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a remote apt repository upstream", func() {
		msg := minimal()
		msg.Spec.Format = "APT"
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			AptRepository: &GcpArtifactRegistryRepoRemoteAptRepository{
				RepositoryBase: "DEBIAN",
				RepositoryPath: "debian/dists/bookworm",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a remote yum repository upstream", func() {
		msg := minimal()
		msg.Spec.Format = "YUM"
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			YumRepository: &GcpArtifactRegistryRepoRemoteYumRepository{
				RepositoryBase: "ROCKY",
				RepositoryPath: "pub/rocky/9/BaseOS/x86_64/os",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept remote upstream credentials with a secret version reference", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
			UpstreamCredentials: &GcpArtifactRegistryRepoRemoteUpstreamCredentials{
				Username:              "ci-bot",
				PasswordSecretVersion: "projects/my-project/secrets/dockerhub-token/versions/latest",
			},
			DisableUpstreamValidation: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a VIRTUAL_REPOSITORY with prioritized upstreams", func() {
		msg := minimal()
		msg.Spec.Mode = "VIRTUAL_REPOSITORY"
		msg.Spec.VirtualRepositoryConfig = &GcpArtifactRegistryRepoVirtualConfig{
			UpstreamPolicies: []*GcpArtifactRegistryRepoVirtualUpstreamPolicy{
				{
					Id:         "team-repo",
					Repository: literal("projects/p/locations/us-central1/repositories/team-images"),
					Priority:   100,
				},
				{
					Id:         "dockerhub-cache",
					Repository: literal("projects/p/locations/us-central1/repositories/dockerhub-remote"),
					Priority:   50,
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept vulnerability scanning INHERITED and DISABLED", func() {
		for _, v := range []string{"INHERITED", "DISABLED"} {
			msg := minimal()
			msg.Spec.VulnerabilityScanningEnablement = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "value %s should be accepted", v)
		}
	})

	ginkgo.It("should accept additive iam_members incl. a public reader", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpArtifactRegistryRepoIamMember{
			{Role: "roles/artifactregistry.writer", Member: literal("serviceAccount:ci@my-project.iam.gserviceaccount.com")},
			{Role: "roles/artifactregistry.reader", Member: literal("allUsers")},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec without location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.ToLower(err.Error())).To(gomega.ContainSubstring("location"))
	})

	ginkgo.It("should reject a spec without format", func() {
		msg := minimal()
		msg.Spec.Format = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.ToLower(err.Error())).To(gomega.ContainSubstring("format"))
	})

	ginkgo.It("should reject an invalid repository_id (uppercase)", func() {
		msg := minimal()
		msg.Spec.RepositoryId = "App-Images"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("repository_id"))
	})

	ginkgo.It("should reject an invalid repository_id (leading digit)", func() {
		msg := minimal()
		msg.Spec.RepositoryId = "1images"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown mode", func() {
		msg := minimal()
		msg.Spec.Mode = "PULL_THROUGH"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mode"))
	})

	ginkgo.It("should reject remote_repository_config without REMOTE_REPOSITORY mode", func() {
		msg := minimal()
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("REMOTE_REPOSITORY"))
	})

	ginkgo.It("should reject REMOTE_REPOSITORY mode without remote_repository_config", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("remote_repository_config"))
	})

	ginkgo.It("should reject virtual_repository_config without VIRTUAL_REPOSITORY mode", func() {
		msg := minimal()
		msg.Spec.VirtualRepositoryConfig = &GcpArtifactRegistryRepoVirtualConfig{
			UpstreamPolicies: []*GcpArtifactRegistryRepoVirtualUpstreamPolicy{
				{Id: "a", Repository: literal("projects/p/locations/l/repositories/r")},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("VIRTUAL_REPOSITORY"))
	})

	ginkgo.It("should reject VIRTUAL_REPOSITORY mode without virtual_repository_config", func() {
		msg := minimal()
		msg.Spec.Mode = "VIRTUAL_REPOSITORY"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("virtual_repository_config"))
	})

	ginkgo.It("should reject a virtual config with zero upstream policies", func() {
		msg := minimal()
		msg.Spec.Mode = "VIRTUAL_REPOSITORY"
		msg.Spec.VirtualRepositoryConfig = &GcpArtifactRegistryRepoVirtualConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("upstream_policies"))
	})

	ginkgo.It("should reject a remote config with no upstream source", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one upstream"))
	})

	ginkgo.It("should reject a remote config with two upstream sources", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
			NpmPublicRepository:    "NPMJS",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one upstream"))
	})

	ginkgo.It("should reject an unknown docker public upstream", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "QUAY_IO",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("DOCKER_HUB"))
	})

	ginkgo.It("should reject an unknown apt repository base", func() {
		msg := minimal()
		msg.Spec.Format = "APT"
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			AptRepository: &GcpArtifactRegistryRepoRemoteAptRepository{
				RepositoryBase: "FEDORA",
				RepositoryPath: "some/path",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject upstream credentials with a malformed secret version", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
			UpstreamCredentials: &GcpArtifactRegistryRepoRemoteUpstreamCredentials{
				Username:              "ci-bot",
				PasswordSecretVersion: "dockerhub-token",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("password_secret_version"))
	})

	ginkgo.It("should reject upstream credentials without a username", func() {
		msg := minimal()
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			DockerPublicRepository: "DOCKER_HUB",
			UpstreamCredentials: &GcpArtifactRegistryRepoRemoteUpstreamCredentials{
				PasswordSecretVersion: "projects/p/secrets/s/versions/latest",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown maven version policy", func() {
		msg := minimal()
		msg.Spec.Format = "MAVEN"
		msg.Spec.MavenConfig = &GcpArtifactRegistryRepoMavenConfig{VersionPolicy: "MILESTONE"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a cleanup policy without an id", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Action:    "DELETE",
			Condition: &GcpArtifactRegistryRepoCleanupPolicyCondition{OlderThan: "2592000s"},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a cleanup policy with an unknown action", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "p",
			Action: "ARCHIVE",
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject most_recent_versions on a DELETE policy", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "bad",
			Action: "DELETE",
			MostRecentVersions: &GcpArtifactRegistryRepoCleanupPolicyMostRecentVersions{
				KeepCount: 5,
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("KEEP"))
	})

	ginkgo.It("should reject a KEEP policy with no criteria", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "bad-keep",
			Action: "KEEP",
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed cleanup duration", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "bad-duration",
			Action: "DELETE",
			Condition: &GcpArtifactRegistryRepoCleanupPolicyCondition{
				OlderThan: "30d",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown tag_state", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "bad-tag-state",
			Action: "DELETE",
			Condition: &GcpArtifactRegistryRepoCleanupPolicyCondition{
				TagState: "ORPHANED",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject most_recent_versions with a non-positive keep_count", func() {
		msg := minimal()
		msg.Spec.CleanupPolicies = []*GcpArtifactRegistryRepoCleanupPolicy{{
			Id:     "zero-keep",
			Action: "KEEP",
			MostRecentVersions: &GcpArtifactRegistryRepoCleanupPolicyMostRecentVersions{
				KeepCount: 0,
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown vulnerability scanning value", func() {
		msg := minimal()
		msg.Spec.VulnerabilityScanningEnablement = "ENABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an iam_member without a role", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpArtifactRegistryRepoIamMember{
			{Member: literal("allUsers")},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an iam_member without a member", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpArtifactRegistryRepoIamMember{
			{Role: "roles/artifactregistry.reader"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a virtual upstream policy without a repository", func() {
		msg := minimal()
		msg.Spec.Mode = "VIRTUAL_REPOSITORY"
		msg.Spec.VirtualRepositoryConfig = &GcpArtifactRegistryRepoVirtualConfig{
			UpstreamPolicies: []*GcpArtifactRegistryRepoVirtualUpstreamPolicy{
				{Id: "missing-repo"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		msg := minimal()
		msg.Kind = "GcpArtifactRegistry"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a remote apt repository from the Debian snapshot archive", func() {
		msg := minimal()
		msg.Spec.Format = "APT"
		msg.Spec.Mode = "REMOTE_REPOSITORY"
		msg.Spec.RemoteRepositoryConfig = &GcpArtifactRegistryRepoRemoteConfig{
			AptRepository: &GcpArtifactRegistryRepoRemoteAptRepository{
				RepositoryBase: "DEBIAN_SNAPSHOT",
				RepositoryPath: "debian/20240101T000000Z/dists/bookworm",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept the explicit VERSION_POLICY_UNSPECIFIED maven sentinel", func() {
		msg := minimal()
		msg.Spec.Format = "MAVEN"
		msg.Spec.MavenConfig = &GcpArtifactRegistryRepoMavenConfig{VersionPolicy: "VERSION_POLICY_UNSPECIFIED"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
