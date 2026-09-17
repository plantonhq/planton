module github.com/plantonhq/planton

go 1.26.6

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20251209175733-2a1774d88802.1
	buf.build/go/protovalidate v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2 v2.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates v1.4.0
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/Masterminds/sprig v2.15.0+incompatible
	github.com/aws/aws-sdk-go-v2 v1.43.6
	github.com/aws/aws-sdk-go-v2/config v1.32.25
	github.com/aws/aws-sdk-go-v2/credentials v1.19.24
	github.com/aws/aws-sdk-go-v2/service/acm v1.30.6
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.49.2
	github.com/aws/aws-sdk-go-v2/service/amp v1.48.3
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.42.5
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.36.0
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.41.0
	github.com/aws/aws-sdk-go-v2/service/appsync v1.56.6
	github.com/aws/aws-sdk-go-v2/service/athena v1.59.0
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.68.1
	github.com/aws/aws-sdk-go-v2/service/backup v1.60.1
	github.com/aws/aws-sdk-go-v2/service/batch v1.67.0
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.66.2
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.58.5
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.53.0
	github.com/aws/aws-sdk-go-v2/service/budgets v1.46.6
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.31.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.65.4
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.58.5
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.62.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.79.0
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.71.0
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.48.0
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.64.0
	github.com/aws/aws-sdk-go-v2/service/configservice v1.68.5
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.67.6
	github.com/aws/aws-sdk-go-v2/service/dlm v1.40.1
	github.com/aws/aws-sdk-go-v2/service/docdb v1.49.7
	github.com/aws/aws-sdk-go-v2/service/dsql v1.16.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.59.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.307.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.59.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.86.2
	github.com/aws/aws-sdk-go-v2/service/efs v1.43.0
	github.com/aws/aws-sdk-go-v2/service/eks v1.88.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.54.5
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.55.6
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.46.8
	github.com/aws/aws-sdk-go-v2/service/firehose v1.44.2
	github.com/aws/aws-sdk-go-v2/service/fsx v1.67.0
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.37.0
	github.com/aws/aws-sdk-go-v2/service/glue v1.148.0
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.85.5
	github.com/aws/aws-sdk-go-v2/service/iam v1.54.7
	github.com/aws/aws-sdk-go-v2/service/kafka v1.54.2
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.44.4
	github.com/aws/aws-sdk-go-v2/service/kms v1.53.6
	github.com/aws/aws-sdk-go-v2/service/lambda v1.94.1
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.35.0
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.41.7
	github.com/aws/aws-sdk-go-v2/service/neptune v1.46.2
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.73.0
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.34.5
	github.com/aws/aws-sdk-go-v2/service/organizations v1.53.8
	github.com/aws/aws-sdk-go-v2/service/pipes v1.26.2
	github.com/aws/aws-sdk-go-v2/service/rds v1.119.5
	github.com/aws/aws-sdk-go-v2/service/redshift v1.63.5
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.35.10
	github.com/aws/aws-sdk-go-v2/service/route53 v1.46.2
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.48.6
	github.com/aws/aws-sdk-go-v2/service/s3 v1.104.0
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.18.6
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.10.5
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.257.0
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.20.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.6
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.43.6
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.63.0
	github.com/aws/aws-sdk-go-v2/service/sfn v1.44.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.40.3
	github.com/aws/aws-sdk-go-v2/service/sqs v1.37.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.73.5
	github.com/aws/aws-sdk-go-v2/service/sts v1.43.3
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.47.6
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.75.0
	github.com/aws/smithy-go v1.27.8
	github.com/blang/semver v3.5.1+incompatible
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/digitalocean/godo v1.202.0
	github.com/fatih/color v1.18.0
	github.com/google/cel-go v0.26.1
	github.com/google/go-containerregistry v0.21.8
	github.com/google/uuid v1.6.0
	github.com/gruntwork-io/terratest v0.56.0
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/iancoleman/strcase v0.3.0
	github.com/nats-io/nats.go v1.52.0
	github.com/nikolalohinski/gonja/v2 v2.8.0
	github.com/oklog/ulid/v2 v2.1.0
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/onsi/gomega v1.38.2
	github.com/pkg/errors v0.9.1
	github.com/pseudomuto/protoc-gen-doc v1.5.1
	github.com/pulumi/pulumi-auth0/sdk/v3 v3.35.0
	github.com/pulumi/pulumi-aws-native/sdk v1.14.0
	github.com/pulumi/pulumi-aws/sdk/v7 v7.41.0
	github.com/pulumi/pulumi-azure-native-sdk/machinelearningservices/v3 v3.12.1
	github.com/pulumi/pulumi-azure-native-sdk/v3 v3.12.1
	github.com/pulumi/pulumi-azure/sdk/v6 v6.38.0
	github.com/pulumi/pulumi-cloudflare/sdk/v6 v6.19.0
	github.com/pulumi/pulumi-digitalocean/sdk/v4 v4.53.0
	github.com/pulumi/pulumi-gcp/sdk/v9 v9.37.0
	github.com/pulumi/pulumi-kubernetes/sdk/v4 v4.33.0
	github.com/pulumi/pulumi-random/sdk/v4 v4.16.7
	github.com/pulumi/pulumi-tls/sdk/v4 v4.11.1
	github.com/pulumi/pulumi/sdk/v3 v3.261.0
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stigmer/stigmer/sdk/go/v3 v3.12.4
	github.com/stretchr/testify v1.11.1
	github.com/zclconf/go-cty v1.16.3
	github.com/zyedidia/clipboard v1.0.4
	go.temporal.io/api v1.63.0
	go.temporal.io/sdk v1.46.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/term v0.45.0
	golang.org/x/text v0.40.0
	google.golang.org/api v0.206.0
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.20.2
	k8s.io/apimachinery v0.35.1
	k8s.io/client-go v0.35.1
	sigs.k8s.io/kustomize/api v0.20.1
	sigs.k8s.io/kustomize/kyaml v0.20.1
	sigs.k8s.io/yaml v1.6.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go/auth v0.10.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.5 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.31 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.12.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.31.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.36.6 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/charmbracelet/bubbles v1.0.0 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/x/ansi v0.11.7 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/containerd/containerd v1.7.30 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.5.0 // indirect
	github.com/djherbis/times v1.5.0 // indirect
	github.com/docker/cli v29.6.2+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.9.3 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg/v2 v2.0.2 // indirect
	github.com/go-git/go-billy/v6 v6.0.0-alpha.1 // indirect
	github.com/go-git/go-git/v6 v6.0.0-alpha.4 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-test/deep v1.1.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.0 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter/v2 v2.2.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/hashicorp/terraform-json v0.23.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.6.0 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lucasb-eyer/go-colorful v1.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/mattn/go-zglob v0.0.2-0.20190814121620-e3c945676326 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nexus-rpc/nexus-proto-annotations v0.1.0 // indirect
	github.com/nexus-rpc/sdk-go v0.6.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opentracing/basictracer-go v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pgavlin/fx v0.1.6 // indirect
	github.com/pgavlin/fx/v2 v2.0.12 // indirect
	github.com/pjbgf/sha1cd v0.6.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pseudomuto/protokit v0.2.0 // indirect
	github.com/pulumi/appdash v0.0.0-20231130102222-75f619a67231 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rubenv/sql-migrate v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/tmccombs/hcl2json v0.6.4 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/zalando/go-keyring v0.2.8 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/featuregate v1.59.0 // indirect
	go.opentelemetry.io/collector/pdata v1.59.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelslog v0.18.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/log v0.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260622175928-b703f567277d // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.35.1 // indirect
	k8s.io/apiextensions-apiserver v0.35.1 // indirect
	k8s.io/apiserver v0.35.1 // indirect
	k8s.io/cli-runtime v0.35.1 // indirect
	k8s.io/component-base v0.35.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250910181357-589584f1c912 // indirect
	k8s.io/kubectl v0.35.1 // indirect
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4 // indirect
	lukechampine.com/frand v1.5.1 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)
