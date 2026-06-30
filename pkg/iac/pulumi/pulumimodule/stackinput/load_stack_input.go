package stackinput

import (
	"fmt"
	"os"

	"buf.build/go/protovalidate"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput/fieldsextractor"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

const (
	PulumiConfigKey   = "planton:stack-input"
	FilePathEnvVar    = "STACK_INPUT_YAML_FILE"
	YamlContentEnvVar = "STACK_INPUT_YAML"
)

func LoadStackInput(ctx *pulumi.Context, stackInput proto.Message) error {
	stackInputString, ok := ctx.GetConfig(PulumiConfigKey)
	var jsonBytes, stackInputYamlBytes []byte
	var err error

	if !ok {
		yamlContent := os.Getenv(YamlContentEnvVar)
		if yamlContent != "" {
			stackInputYamlBytes = []byte(yamlContent)
		} else {
			stackInputFilePath := os.Getenv(FilePathEnvVar)
			if stackInputFilePath == "" {
				return errors.Errorf("stack-input not found in pulumi config %s or in %s environment variable",
					PulumiConfigKey, FilePathEnvVar)
			}
			stackInputYamlBytes, err = os.ReadFile(stackInputFilePath)
			if err != nil {
				return errors.Wrap(err, "failed to read input file")
			}
		}

		jsonBytes, err = yaml.YAMLToJSON(stackInputYamlBytes)
		if err != nil {
			return errors.Wrap(err, "failed to load yaml to json")
		}
	} else {
		jsonBytes, err = yaml.YAMLToJSON([]byte(stackInputString))
		if err != nil {
			return errors.Wrap(err, "failed to load yaml to json")
		}
	}

	if err := protojson.Unmarshal(jsonBytes, stackInput); err != nil {
		return errors.Wrap(err, "failed to load json into proto message")
	}

	targetSpec, err := fieldsextractor.ExtractApiResourceSpecField(stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to extract api resource spec field")
	}

	v, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages((*targetSpec).Interface()),
	)
	if err != nil {
		fmt.Println("failed to initialize validator:", err)
	}

	if err = v.Validate((*targetSpec).Interface()); err != nil {
		return errors.Errorf("%s", err)
	}
	return nil
}
