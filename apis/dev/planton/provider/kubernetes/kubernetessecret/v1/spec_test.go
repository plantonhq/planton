package kubernetessecretv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func stringPtr(s string) *string {
	return &s
}

func TestKubernetesSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSecretSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesSecretSpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal valid Opaque secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "my-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"api-key": "supersecret",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Opaque secret with multiple keys", func() {
			spec := &KubernetesSecretSpec{
				Name:      "multi-key-secret",
				Namespace: stringPtr("production"),
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{
							"username": "admin",
							"password": "s3cret",
							"api-key":  "abc123",
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid TLS secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "tls-cert",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
						TlsKey: "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid DockerConfigJson secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "registry-creds",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "https://index.docker.io/v1/",
						Username:       "myuser",
						Password:       "mypassword",
						Email:          "user@example.com",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a DockerConfigJson secret without email", func() {
			spec := &KubernetesSecretSpec{
				Name: "registry-creds-no-email",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "_json_key",
						Password:       "{\"type\":\"service_account\"}",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid BasicAuth secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "basic-auth-creds",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "admin",
						Password: "password123",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid SSHAuth secret", func() {
			spec := &KubernetesSecretSpec{
				Name: "ssh-key",
				SecretData: &KubernetesSecretSpec_SshAuth{
					SshAuth: &KubernetesSecretSshAuthData{
						SshPrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nb3Blb...\n-----END OPENSSH PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a secret with labels and annotations", func() {
			spec := &KubernetesSecretSpec{
				Name:      "labeled-secret",
				Namespace: stringPtr("kube-system"),
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
				},
				Annotations: map[string]string{
					"description": "Platform API credentials",
				},
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a secret with immutable flag", func() {
			spec := &KubernetesSecretSpec{
				Name:      "immutable-secret",
				Immutable: true,
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"config": "frozen-value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a name with dots (DNS subdomain)", func() {
			spec := &KubernetesSecretSpec{
				Name: "my.dotted.secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects empty secret name", func() {
			spec := &KubernetesSecretSpec{
				Name: "",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name with uppercase letters", func() {
			spec := &KubernetesSecretSpec{
				Name: "MySecret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name starting with a dot", func() {
			spec := &KubernetesSecretSpec{
				Name: ".hidden-secret",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name ending with a hyphen", func() {
			spec := &KubernetesSecretSpec{
				Name: "my-secret-",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects secret name longer than 253 characters", func() {
			longName := "a"
			for i := 0; i < 253; i++ {
				longName += "a"
			}
			spec := &KubernetesSecretSpec{
				Name: longName,
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace with uppercase letters", func() {
			spec := &KubernetesSecretSpec{
				Name:      "my-secret",
				Namespace: stringPtr("MyNamespace"),
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace longer than 63 characters", func() {
			spec := &KubernetesSecretSpec{
				Name:      "my-secret",
				Namespace: stringPtr("this-is-a-very-long-namespace-name-that-exceeds-the-maximum-length-allowed"),
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{"key": "value"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects spec without any secret data", func() {
			spec := &KubernetesSecretSpec{
				Name: "empty-secret",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Opaque secret with empty data map", func() {
			spec := &KubernetesSecretSpec{
				Name: "empty-opaque",
				SecretData: &KubernetesSecretSpec_Opaque{
					Opaque: &KubernetesSecretOpaqueData{
						Data: map[string]string{},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects TLS secret with empty certificate", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-tls",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "",
						TlsKey: "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects TLS secret with empty key", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-tls-key",
				SecretData: &KubernetesSecretSpec_Tls{
					Tls: &KubernetesSecretTlsData{
						TlsCrt: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
						TlsKey: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty registry server", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "",
						Username:       "user",
						Password:       "pass",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty username", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker-user",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "",
						Password:       "pass",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects DockerConfigJson with empty password", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-docker-pass",
				SecretData: &KubernetesSecretSpec_DockerConfigJson{
					DockerConfigJson: &KubernetesSecretDockerConfigJsonData{
						RegistryServer: "gcr.io",
						Username:       "user",
						Password:       "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects BasicAuth with empty username", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-basic-auth",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "",
						Password: "password",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects BasicAuth with empty password", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-basic-auth-pass",
				SecretData: &KubernetesSecretSpec_BasicAuth{
					BasicAuth: &KubernetesSecretBasicAuthData{
						Username: "admin",
						Password: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects SSHAuth with empty private key", func() {
			spec := &KubernetesSecretSpec{
				Name: "bad-ssh",
				SecretData: &KubernetesSecretSpec_SshAuth{
					SshAuth: &KubernetesSecretSshAuthData{
						SshPrivateKey: "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
