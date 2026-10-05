package gcpcloudbuildconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudBuildConnectionSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

func secret(name string) *foreignkeyv1.StringValueOrRef {
	return reference(catalogkind.CatalogKind_GcpSecretManagerSecret, name)
}

func token(name string) *GcpCloudBuildConnectionUserTokenCredential {
	return &GcpCloudBuildConnectionUserTokenCredential{UserTokenSecretVersion: secret(name)}
}

var _ = ginkgo.Describe("GcpCloudBuildConnectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudBuildConnection {
		return &GcpCloudBuildConnection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudBuildConnection",
			Metadata:   &shared.CatalogObjectMetadata{Name: "github"},
			Spec:       &GcpCloudBuildConnectionSpec{Location: "us-central1"},
		}
	}

	ginkgo.It("should accept every code host", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())

		github := minimal()
		github.Spec.ProjectId = reference(catalogkind.CatalogKind_GcpProject, "ci")
		github.Spec.ConnectionId = "acme-github"
		github.Spec.Annotations = map[string]string{"owner": "platform"}
		github.Spec.GithubConfig = &GcpCloudBuildConnectionGithubConfig{
			AppInstallationId:    12345678,
			AuthorizerCredential: &GcpCloudBuildConnectionOauthCredential{OauthTokenSecretVersion: literal("projects/ci/secrets/github-token/versions/latest")},
		}
		gomega.Expect(validator.Validate(github)).To(gomega.Succeed())

		ghe := minimal()
		ghe.Spec.GithubEnterpriseConfig = &GcpCloudBuildConnectionGithubEnterpriseConfig{
			HostUri:                    "https://github.example.com",
			AppId:                      42,
			AppInstallationId:          7,
			AppSlug:                    "cloud-build",
			PrivateKeySecretVersion:    secret("ghe-key"),
			WebhookSecretSecretVersion: secret("ghe-webhook"),
			SslCa:                      "-----BEGIN CERTIFICATE-----",
			ServiceDirectoryConfig:     &GcpCloudBuildConnectionServiceDirectoryConfig{Service: "projects/ci/locations/us-central1/namespaces/scm/services/ghe"},
		}
		gomega.Expect(validator.Validate(ghe)).To(gomega.Succeed())

		gitlab := minimal()
		gitlab.Spec.GitlabConfig = &GcpCloudBuildConnectionGitlabConfig{
			AuthorizerCredential:       token("gitlab-api"),
			ReadAuthorizerCredential:   token("gitlab-read"),
			WebhookSecretSecretVersion: secret("gitlab-webhook"),
		}
		gomega.Expect(validator.Validate(gitlab)).To(gomega.Succeed())

		cloud := minimal()
		cloud.Spec.BitbucketCloudConfig = &GcpCloudBuildConnectionBitbucketCloudConfig{
			Workspace:                  "acme",
			AuthorizerCredential:       token("bb-admin"),
			ReadAuthorizerCredential:   token("bb-read"),
			WebhookSecretSecretVersion: secret("bb-webhook"),
		}
		gomega.Expect(validator.Validate(cloud)).To(gomega.Succeed())

		dc := minimal()
		dc.Spec.BitbucketDataCenterConfig = &GcpCloudBuildConnectionBitbucketDataCenterConfig{
			HostUri:                    "https://bitbucket.example.com",
			AuthorizerCredential:       token("bbdc-admin"),
			ReadAuthorizerCredential:   token("bbdc-read"),
			WebhookSecretSecretVersion: secret("bbdc-webhook"),
		}
		gomega.Expect(validator.Validate(dc)).To(gomega.Succeed())
	})

	ginkgo.It("should refuse two code hosts", func() {
		msg := minimal()
		msg.Spec.GithubConfig = &GcpCloudBuildConnectionGithubConfig{AppInstallationId: 1}
		msg.Spec.GithubEnterpriseConfig = &GcpCloudBuildConnectionGithubEnterpriseConfig{HostUri: "https://github.example.com"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require each host's mandatory fields", func() {
		msg := minimal()
		msg.Spec.GithubEnterpriseConfig = &GcpCloudBuildConnectionGithubEnterpriseConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.GitlabConfig = &GcpCloudBuildConnectionGitlabConfig{AuthorizerCredential: token("a"), ReadAuthorizerCredential: token("r")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "gitlab without a webhook secret")

		msg = minimal()
		msg.Spec.BitbucketCloudConfig = &GcpCloudBuildConnectionBitbucketCloudConfig{
			AuthorizerCredential: token("a"), ReadAuthorizerCredential: token("r"), WebhookSecretSecretVersion: secret("w"),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "bitbucket cloud without a workspace")

		msg = minimal()
		msg.Spec.BitbucketDataCenterConfig = &GcpCloudBuildConnectionBitbucketDataCenterConfig{
			HostUri: "https://bitbucket.example.com", AuthorizerCredential: token("a"), WebhookSecretSecretVersion: secret("w"),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "bitbucket data center without a read credential")

		msg = minimal()
		msg.Spec.GitlabConfig = &GcpCloudBuildConnectionGitlabConfig{
			AuthorizerCredential: &GcpCloudBuildConnectionUserTokenCredential{}, ReadAuthorizerCredential: token("r"), WebhookSecretSecretVersion: secret("w"),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a credential without its token")
	})

	ginkgo.It("should require full secret version names on literals", func() {
		msg := minimal()
		msg.Spec.GithubConfig = &GcpCloudBuildConnectionGithubConfig{
			AuthorizerCredential: &GcpCloudBuildConnectionOauthCredential{OauthTokenSecretVersion: literal("github-token")},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.GitlabConfig = &GcpCloudBuildConnectionGitlabConfig{
			AuthorizerCredential:       token("a"),
			ReadAuthorizerCredential:   token("r"),
			WebhookSecretSecretVersion: literal("projects/ci/secrets/webhook"),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a full Service Directory service name", func() {
		msg := minimal()
		msg.Spec.GithubEnterpriseConfig = &GcpCloudBuildConnectionGithubEnterpriseConfig{
			HostUri:                "https://github.example.com",
			ServiceDirectoryConfig: &GcpCloudBuildConnectionServiceDirectoryConfig{Service: "ghe"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a location and reject bad IDs and deletion policies", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		for _, id := range []string{"acme github", "acme/github", "acme#1"} {
			msg = minimal()
			msg.Spec.ConnectionId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		for _, id := range []string{"acme-github", "acme_github.v2", "a~b@c"} {
			msg = minimal()
			msg.Spec.ConnectionId = id
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), id)
		}

		msg = minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
