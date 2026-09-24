package digitaloceanfunctionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
)

func TestDigitalOceanFunctionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanFunctionSpec Validation Suite")
}

// validGitFunction mirrors the hello-world scenario: project.yml sits at the
// repository root, so source_directory is deliberately unset.
func validGitFunction() *DigitalOceanFunctionSpec {
	return &DigitalOceanFunctionSpec{
		AppName:      "hello-fn",
		FunctionName: "hello",
		Region:       digitalocean.DigitalOceanAppRegion_nyc,
		Git: &digitalocean.DigitalOceanAppGitSource{
			RepoCloneUrl: "https://github.com/digitalocean/sample-functions-nodejs-helloworld.git",
			Branch:       "master",
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanFunctionSpec", func() {
	ginkgo.It("accepts a public git source with project.yml at the repo root", func() {
		gomega.Expect(protovalidate.Validate(validGitFunction())).To(gomega.BeNil())
	})

	ginkgo.It("accepts a linked github source with a source_directory", func() {
		spec := &DigitalOceanFunctionSpec{
			AppName:      "my-functions",
			FunctionName: "api",
			Region:       digitalocean.DigitalOceanAppRegion_nyc,
			Github: &digitalocean.DigitalOceanAppGithubSource{
				Repo:   "myorg/my-functions",
				Branch: "main",
			},
			SourceDirectory: "functions/api",
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.Describe("app_name (the App Platform app name: 2-32, letter-first, account-unique)", func() {
		ginkgo.It("rejects a missing app_name", func() {
			spec := validGitFunction()
			spec.AppName = ""
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app_name longer than 32 characters (the API's limit)", func() {
			spec := validGitFunction()
			spec.AppName = "planton-oss-e2e-digitaloceanfunction-hello"
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app_name that starts with a digit", func() {
			spec := validGitFunction()
			spec.AppName = "9abc"
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app_name with uppercase or underscores", func() {
			spec := validGitFunction()
			spec.AppName = "Hello_World"
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts the shortest legal app_name", func() {
			spec := validGitFunction()
			spec.AppName = "ab"
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("function_name (the component name: the same API rule)", func() {
		ginkgo.It("rejects an empty function_name", func() {
			spec := validGitFunction()
			spec.FunctionName = ""
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a function_name with a dot", func() {
			spec := validGitFunction()
			spec.FunctionName = "web.1"
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a function_name ending with a hyphen", func() {
			spec := validGitFunction()
			spec.FunctionName = "hello-"
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.It("rejects a missing region", func() {
		spec := validGitFunction()
		spec.Region = digitalocean.DigitalOceanAppRegion_digital_ocean_app_region_unspecified
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts an unset source_directory (project.yml at the repository root)", func() {
		spec := validGitFunction()
		spec.SourceDirectory = ""
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("rejects two sources", func() {
		spec := validGitFunction()
		spec.Github = &digitalocean.DigitalOceanAppGithubSource{
			Repo:   "myorg/my-functions",
			Branch: "main",
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a spec with no source", func() {
		spec := validGitFunction()
		spec.Git = nil
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an env var with no value", func() {
		spec := validGitFunction()
		spec.Envs = []*digitalocean.DigitalOceanAppEnvVar{{Key: "EMPTY"}}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a plaintext env var", func() {
		spec := validGitFunction()
		spec.Envs = []*digitalocean.DigitalOceanAppEnvVar{
			{Key: "NODE_ENV", Value: &digitalocean.DigitalOceanAppEnvVar_Plaintext{Plaintext: "production"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})
})
