package digitaloceanappv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanAppSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanAppSpec Validation Tests")
}

func validImageApp() *DigitalOceanApp {
	return &DigitalOceanApp{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanApp",
		Metadata:   &shared.CloudResourceMetadata{Name: "demo-app"},
		Spec: &DigitalOceanAppSpec{
			AppName: "demo-app",
			Region:  digitalocean.DigitalOceanAppRegion_nyc,
			Services: []*DigitalOceanAppService{
				{
					Name: "web",
					Image: &digitalocean.DigitalOceanAppImageSource{
						RegistryType: digitalocean.DigitalOceanAppRegistryType_docker_hub,
						Registry:     "library",
						Repository:   "nginx",
						Tag:          "latest",
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanAppSpec", func() {
	ginkgo.It("accepts a minimal app with a docker hub service", func() {
		gomega.Expect(protovalidate.Validate(validImageApp())).To(gomega.BeNil())
	})

	ginkgo.It("rejects an app with no components", func() {
		app := validImageApp()
		app.Spec.Services = nil
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an app name longer than 32 characters", func() {
		app := validImageApp()
		app.Spec.AppName = "this-name-is-way-too-long-for-app-platform"
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a service with two sources", func() {
		app := validImageApp()
		app.Spec.Services[0].Git = &digitalocean.DigitalOceanAppGitSource{
			RepoCloneUrl: "https://github.com/example/app.git",
			Branch:       "main",
		}
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a service that sets instance_count while autoscaling", func() {
		app := validImageApp()
		app.Spec.Services[0].InstanceCount = 2
		app.Spec.Services[0].Autoscaling = &digitalocean.DigitalOceanAppAutoscaling{
			MinInstanceCount: 1,
			MaxInstanceCount: 4,
			CpuPercent:       80,
		}
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a service with autoscaling and no instance_count", func() {
		app := validImageApp()
		app.Spec.Services[0].Autoscaling = &digitalocean.DigitalOceanAppAutoscaling{
			MinInstanceCount: 1,
			MaxInstanceCount: 4,
			CpuPercent:       80,
		}
		gomega.Expect(protovalidate.Validate(app)).To(gomega.BeNil())
	})

	ginkgo.It("rejects a worker that sets drain_seconds", func() {
		app := validImageApp()
		drain := uint32(15)
		app.Spec.Workers = []*DigitalOceanAppWorker{
			{
				Name: "worker",
				Git: &digitalocean.DigitalOceanAppGitSource{
					RepoCloneUrl: "https://github.com/example/app.git",
					Branch:       "main",
				},
				Termination: &digitalocean.DigitalOceanAppTermination{DrainSeconds: &drain},
			},
		}
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects docr images that set a registry hostname", func() {
		app := validImageApp()
		app.Spec.Services[0].Image = &digitalocean.DigitalOceanAppImageSource{
			RegistryType: digitalocean.DigitalOceanAppRegistryType_docr,
			Registry:     "registry.digitalocean.com",
			Repository:   "myapp",
			Tag:          "v1",
		}
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an image with both tag and digest", func() {
		app := validImageApp()
		app.Spec.Services[0].Image.Digest = "sha256:abc"
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an env var with neither plaintext nor secret", func() {
		app := validImageApp()
		app.Spec.Envs = []*digitalocean.DigitalOceanAppEnvVar{
			{Key: "EMPTY"},
		}
		gomega.Expect(protovalidate.Validate(app)).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a secret env var", func() {
		app := validImageApp()
		app.Spec.Envs = []*digitalocean.DigitalOceanAppEnvVar{
			{Key: "API_KEY", Value: &digitalocean.DigitalOceanAppEnvVar_Secret{Secret: "s3cret"}},
		}
		gomega.Expect(protovalidate.Validate(app)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a domain pointing at a DNS zone", func() {
		app := validImageApp()
		app.Spec.Domains = []*digitalocean.DigitalOceanAppDomain{
			{
				Name: "www.example.com",
				Type: "PRIMARY",
				Zone: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example.com"},
				},
			},
		}
		gomega.Expect(protovalidate.Validate(app)).To(gomega.BeNil())
	})
})
