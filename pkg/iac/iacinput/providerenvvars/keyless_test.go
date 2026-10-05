package providerenvvars

import (
	"context"
	"sort"
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsprovider "github.com/plantonhq/planton/catalog/aws"
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// The variable each provider's environment-reading engine takes its keyless credential from. A
// provider config that gains a web_identity field is found by the gate below without anyone
// listing it, and fails until its keyless arm and its row here exist.
var keylessCredentialVariable = map[catalogkind.CatalogProvider]string{
	catalogkind.CatalogProvider_aws:   "AWS_ACCESS_KEY_ID",
	catalogkind.CatalogProvider_gcp:   "GOOGLE_OAUTH_ACCESS_TOKEN",
	catalogkind.CatalogProvider_azure: "ARM_OIDC_TOKEN",
}

// Every variable any provider reads a credential from. On an engine that builds its providers,
// a keyless configuration emits none of them; on every engine it never emits a stored secret.
var credentialVariables = []string{
	"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
	"GOOGLE_OAUTH_ACCESS_TOKEN", "GOOGLE_CREDENTIALS",
	"ARM_CLIENT_SECRET", "ARM_OIDC_TOKEN",
}

var storedSecretVariables = []string{"GOOGLE_CREDENTIALS", "ARM_CLIENT_SECRET"}

// keylessProviderConfigs finds every registered provider config carrying a web_identity field
// (package dev.planton.<provider>), and fills every string field of it and of its web identity with
// a placeholder, so a keyless arm that also reads a stored credential would be caught emitting it.
func keylessProviderConfigs(t *testing.T) map[catalogkind.CatalogProvider][]byte {
	t.Helper()
	configs := map[catalogkind.CatalogProvider][]byte{}
	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		desc := mt.Descriptor()
		name := string(desc.FullName())
		if !strings.HasPrefix(name, "dev.planton.") || !strings.HasSuffix(string(desc.Name()), "ProviderConfig") {
			return true
		}
		webIdentity := desc.Fields().ByName("web_identity")
		if webIdentity == nil || webIdentity.Kind() != protoreflect.MessageKind {
			return true
		}
		providerName := strings.TrimPrefix(string(desc.ParentFile().Package()), "dev.planton.")
		value, known := catalogkind.CatalogProvider_value[providerName]
		require.Truef(t, known, "%s carries web_identity, but %q names no CatalogProvider", name, providerName)

		msg := mt.New()
		fillStrings(msg)
		wi := msg.Mutable(webIdentity).Message()
		fillStrings(wi)
		yaml, err := protojson.Marshal(msg.Interface())
		require.NoError(t, err)
		configs[catalogkind.CatalogProvider(value)] = yaml
		return true
	})
	require.NotEmpty(t, configs, "no provider config with web_identity is registered; the gate would prove nothing")
	return configs
}

func fillStrings(msg protoreflect.Message) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if f.Kind() == protoreflect.StringKind && f.Cardinality() != protoreflect.Repeated {
			msg.Set(f, protoreflect.ValueOfString("placeholder-"+string(f.Name())))
		}
	}
}

func keylessEnvVars(provider catalogkind.CatalogProvider, yaml []byte, opts Options, r keylessResolvers) (map[string]string, error) {
	if provider == catalogkind.CatalogProvider_aws {
		return loadAwsEnvVars(yaml, true, "us-east-1", opts, r.aws)
	}
	return loadProviderEnvVars(yaml, provider, opts, r)
}

// A keyless configuration never leaves the provider on the machine's ambient identity: on an
// engine that reads its environment it becomes that engine's keyless credential, on an engine that
// builds its providers it adds none (the builder owns the exchange), and a caller naming no engine
// is refused. Never, on any engine, does a keyless configuration emit a stored secret.
func TestEveryKeylessProvider_NeverFallsBackToAmbient(t *testing.T) {
	configs := keylessProviderConfigs(t)
	providers := make([]catalogkind.CatalogProvider, 0, len(configs))
	for p := range configs {
		providers = append(providers, p)
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i] < providers[j] })

	exchanging := keylessResolvers{
		aws: func(context.Context, string, *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
			return awssdk.Credentials{AccessKeyID: "exchanged", SecretAccessKey: "exchanged", SessionToken: "exchanged"}, nil
		},
		gcp: func(context.Context, *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
			return "exchanged", nil
		},
	}
	refusing := keylessResolvers{
		aws: func(context.Context, string, *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
			t.Fatal("an engine that builds its providers must not be handed an exchange")
			return awssdk.Credentials{}, nil
		},
		gcp: func(context.Context, *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
			t.Fatal("an engine that builds its providers must not be handed an exchange")
			return "", nil
		},
	}

	for _, provider := range providers {
		yaml := configs[provider]
		t.Run(provider.String(), func(t *testing.T) {
			variable, known := keylessCredentialVariable[provider]
			require.Truef(t, known, "%s has a keyless configuration and no keyless credential variable in this gate: "+
				"teach the loader its keyless arm for each engine, then add its row here", provider)

			t.Run("reads its environment", func(t *testing.T) {
				env, err := keylessEnvVars(provider, yaml, Options{Engine: EngineReadsEnvironment}, exchanging)
				require.NoError(t, err)
				assert.NotEmptyf(t, env[variable], "keyless %s on OpenTofu must emit %s, or the provider signs in as whatever the machine holds", provider, variable)
				assert.NotContains(t, env[variable], "placeholder-access", "the keyless credential must come from the web identity, not a stored credential")
				for _, stored := range storedSecretVariables {
					assert.NotContainsf(t, env, stored, "keyless %s must never emit the stored secret %s", provider, stored)
				}
			})
			t.Run("builds its providers", func(t *testing.T) {
				env, err := keylessEnvVars(provider, yaml, Options{Engine: EngineBuildsProviders}, refusing)
				require.NoError(t, err)
				for _, credential := range credentialVariables {
					assert.NotContainsf(t, env, credential, "keyless %s on Pulumi must leave %s to the builder", provider, credential)
				}
			})
			t.Run("names no engine", func(t *testing.T) {
				_, err := keylessEnvVars(provider, yaml, Options{}, refusing)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Options.Engine")
			})
		})
	}
}
