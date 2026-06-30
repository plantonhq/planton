package module

import (
	"github.com/pkg/errors"
	kubernetesnatsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesnats/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream NATS Helm chart and tailors it to the spec.
// Returns the chart resource and auth password (if basic_auth is enabled).
// The chart resource is used as a dependency for NACK deployment.
// The auth password is needed for NACK controller to connect to NATS.
func helmChart(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (pulumi.Resource, pulumi.StringOutput, error) {
	sc := locals.KubernetesNats.Spec.ServerContainer

	// Track auth password for NACK controller (only set for basic_auth)
	var authPassword pulumi.StringOutput

	//----------------------------------------------------------------------
	// Helm values
	//----------------------------------------------------------------------
	values := pulumi.Map{}
	config := pulumi.Map{}

	// --------------- container resources ---------------
	values["container"] = pulumi.Map{
		"merge": pulumi.Map{
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"cpu":    pulumi.String(sc.Resources.Limits.Cpu),
					"memory": pulumi.String(sc.Resources.Limits.Memory),
				},
				"requests": pulumi.Map{
					"cpu":    pulumi.String(sc.Resources.Requests.Cpu),
					"memory": pulumi.String(sc.Resources.Requests.Memory),
				},
			},
		},
	}

	// ---------------------- clustering -----------------
	clusterEnabled := sc.Replicas > 1
	config["cluster"] = pulumi.Map{
		"enabled":  pulumi.Bool(clusterEnabled),
		"replicas": pulumi.Int(sc.Replicas),
	}

	// ---------------------- JetStream ------------------
	if locals.KubernetesNats.Spec.DisableJetStream {
		config["jetstream"] = pulumi.Map{"enabled": pulumi.Bool(false)}
	} else {
		config["jetstream"] = pulumi.Map{
			"enabled": pulumi.Bool(true),
			"fileStore": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"pvc": pulumi.Map{
					"size": pulumi.String(sc.DiskSize),
				},
			},
		}
	}

	// --------------------- authentication --------------
	if auth := locals.KubernetesNats.Spec.Auth; auth != nil && auth.Enabled {
		switch auth.Scheme {

		// bearer-token
		case kubernetesnatsv1.KubernetesNatsAuthScheme_bearer_token:
			var token *random.RandomPassword
			token, err := random.NewRandomPassword(ctx, "auth-token",
				&random.RandomPasswordArgs{Length: pulumi.Int(32), Special: pulumi.Bool(false)},
				pulumi.Provider(kubernetesProvider))
			if err != nil {
				return nil, pulumi.StringOutput{}, errors.Wrap(err, "generate auth token")
			}

			bearerOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
			_, err = kubernetescorev1.NewSecret(ctx, locals.AuthSecretName,
				&kubernetescorev1.SecretArgs{
					Metadata: &kubernetesmeta.ObjectMetaArgs{
						Name:      pulumi.String(locals.AuthSecretName),
						Namespace: pulumi.String(locals.Namespace),
						Labels:    pulumi.ToStringMap(locals.Labels),
					},
					StringData: pulumi.StringMap{vars.AdminAuthSecretKey: token.Result},
					Type:       pulumi.String("Opaque"),
				}, bearerOpts...)
			if err != nil {
				return nil, pulumi.StringOutput{}, errors.Wrap(err, "create bearer-token secret")
			}

			values["auth"] = pulumi.Map{
				"enabled": pulumi.Bool(true),
				"token": pulumi.Map{
					"users": pulumi.Array{
						pulumi.Map{
							"existingSecret": pulumi.Map{
								"name": pulumi.String(locals.AuthSecretName),
								"key":  pulumi.String(vars.AdminAuthSecretKey),
							},
						},
					},
				},
			}

		// basic-auth
		// ---------------- basic auth ------------------
		case kubernetesnatsv1.KubernetesNatsAuthScheme_basic_auth:
			var adminPass *random.RandomPassword
			adminPass, err := random.NewRandomPassword(ctx, "auth-password",
				&random.RandomPasswordArgs{Length: pulumi.Int(32), Special: pulumi.Bool(false)},
				pulumi.Provider(kubernetesProvider))
			if err != nil {
				return nil, pulumi.StringOutput{}, errors.Wrap(err, "generate admin password")
			}

			// Store password for NACK controller
			authPassword = adminPass.Result

			basicOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
			_, err = kubernetescorev1.NewSecret(ctx, locals.AuthSecretName,
				&kubernetescorev1.SecretArgs{
					Metadata: &kubernetesmeta.ObjectMetaArgs{
						Name:      pulumi.String(locals.AuthSecretName),
						Namespace: pulumi.String(locals.Namespace),
						Labels:    pulumi.ToStringMap(locals.Labels),
					},
					StringData: pulumi.StringMap{
						vars.NatsUserSecretKeyUsername: pulumi.String(vars.AdminUsername),
						vars.NatsUserSecretKeyPassword: adminPass.Result,
					},
					Type: pulumi.String("Opaque"),
				}, basicOpts...)
			if err != nil {
				return nil, pulumi.StringOutput{}, errors.Wrap(err, "create admin secret")
			}

			// optional no-auth user
			if na := auth.NoAuthUser; na != nil && na.Enabled {
				noAuthOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
				_, err = kubernetescorev1.NewSecret(ctx, locals.NoAuthUserSecretName,
					&kubernetescorev1.SecretArgs{
						Metadata: &kubernetesmeta.ObjectMetaArgs{
							Name:      pulumi.String(locals.NoAuthUserSecretName),
							Namespace: pulumi.String(locals.Namespace),
							Labels:    pulumi.ToStringMap(locals.Labels),
						},
						StringData: pulumi.StringMap{
							vars.NatsUserSecretKeyUsername: pulumi.String(vars.NoAuthUsername),
							vars.NatsUserSecretKeyPassword: pulumi.String(vars.NoAuthPassword),
						},
						Type: pulumi.String("Opaque"),
					}, noAuthOpts...)
				if err != nil {
					return nil, pulumi.StringOutput{}, errors.Wrap(err, "create no-auth secret")
				}
			}

			// surface password-key so ops can fetch it quickly
			ctx.Export(OpAuthSecretKey, pulumi.String(vars.NatsUserSecretKeyPassword))

			// 1. enable basic-auth but leave the chart’s users list empty
			values["auth"] = pulumi.Map{
				"enabled": pulumi.Bool(true),
				"basic":   pulumi.Map{}, // users will be injected via config.patch
			}

			// ------------------------------------------------------------------
			// 1) add env-vars via container.patch  ➜  they are appended, nothing
			//    pre-existing (POD_NAME / SERVER_NAME) is removed
			// ------------------------------------------------------------------
			envPatch := pulumi.Array{
				pulumi.Map{
					"op":   pulumi.String("add"),
					"path": pulumi.String("/env/-"),
					"value": pulumi.Map{
						"name": pulumi.String(vars.AdminUserPasswordEnvVarName),
						"valueFrom": pulumi.Map{
							"secretKeyRef": pulumi.Map{
								"name": pulumi.String(locals.AuthSecretName),
								"key":  pulumi.String(vars.NatsUserSecretKeyPassword),
							},
						},
					},
				},
			}

			// attach the env patch
			values["container"].(pulumi.Map)["patch"] = envPatch

			//------------------------------------------------------------------
			// 3. build users array that references the env-vars (no secrets here)
			//------------------------------------------------------------------
			users := pulumi.Array{
				pulumi.Map{
					"username": pulumi.String(vars.AdminUsername),
					// WARNING: Helm renders "$VAR" in quotes, so NATS never substitutes it.
					// As a workaround use literal password
					//"password": pulumi.Sprintf("$%s", vars.AdminUserPasswordEnvVarName),
					"password": adminPass.Result,
				},
			}

			// optional no-auth user
			if na := auth.NoAuthUser; na != nil && na.Enabled {
				users = append(users, pulumi.Map{
					"username": pulumi.String(vars.NoAuthUsername),
					"password": pulumi.String(vars.NoAuthPassword),
					"permissions": pulumi.Map{
						"publish":   pulumi.ToStringArray(na.PublishSubjects),
						"subscribe": pulumi.Array{},
					},
				})
			}

			//------------------------------------------------------------------
			// 4. patch the config: add users (+ no_auth_user if set)
			//------------------------------------------------------------------
			patches := pulumi.Array{
				pulumi.Map{
					"op":    pulumi.String("add"),
					"path":  pulumi.String("/authorization"),
					"value": pulumi.Map{},
				},
				pulumi.Map{
					"op":    pulumi.String("add"),
					"path":  pulumi.String("/authorization/users"),
					"value": users,
				},
			}
			if auth.NoAuthUser != nil && auth.NoAuthUser.Enabled {
				patches = append(patches, pulumi.Map{
					"op":    pulumi.String("add"),
					"path":  pulumi.String("/no_auth_user"),
					"value": pulumi.String(vars.NoAuthUsername),
				})
			}

			// merge with any existing patches
			if p, ok := config["patch"]; ok {
				config["patch"] = append(p.(pulumi.Array), patches...)
			} else {
				config["patch"] = patches
			}
		}
	}

	// ------------------------- TLS ---------------------
	if locals.KubernetesNats.Spec.TlsEnabled {
		values["tls"] = pulumi.Map{
			"enabled": pulumi.Bool(true),
			"secret":  pulumi.Map{"name": pulumi.String(locals.TlsSecretName)},
		}
	}

	// ---------------------- nats-box -------------------
	if locals.KubernetesNats.Spec.DisableNatsBox {
		values["natsbox"] = pulumi.Map{"enabled": pulumi.Bool(false)}
	}

	// attach chart-level config and deploy
	values["config"] = config

	chartOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	natsChart, err := helmv3.NewChart(ctx,
		locals.KubernetesNats.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(locals.NatsHelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values:    values,
			FetchArgs: helmv3.FetchArgs{Repo: pulumi.String(vars.HelmChartRepoUrl)},
		}, chartOpts...)
	if err != nil {
		return nil, pulumi.StringOutput{}, errors.Wrap(err, "deploying NATS chart failed")
	}

	ctx.Export(OpMetricsEndpoint, pulumi.Sprintf(
		"http://nats-prom.%s.svc.cluster.local:7777/metrics", locals.Namespace))
	return natsChart, authPassword, nil
}
