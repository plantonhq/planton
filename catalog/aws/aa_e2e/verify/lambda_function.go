package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	pkgerrors "github.com/pkg/errors"
)

// lambdaFunctionVerifier verifies an AwsLambda function via GetFunction,
// keyed on the function_name output. When the outputs report satellite
// resources -- a published version, aliases, a function URL -- existence is
// asserted for each of them from the resource's own state (GetFunction with
// a qualifier, GetAlias, ListFunctionUrlConfigs): the satellites are where
// versioning defects hide, and an apply that "succeeded" says nothing about
// whether the alias or URL actually materialized. Absence needs only the
// function probe -- versions, aliases, and URLs are children of the function
// and cannot outlive it.
type lambdaFunctionVerifier struct{}

func (*lambdaFunctionVerifier) IDOutputKey() string { return "function_name" }

func (*lambdaFunctionVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := lambdaFunctionExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslambda verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awslambda function %q not found after deploy", id)
	}
	return nil
}

func (*lambdaFunctionVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := lambdaFunctionExists(ctx, cfg, id, region)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslambda verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awslambda function %q still exists after destroy", id)
	}
	return nil
}

func (v *lambdaFunctionVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	name := stringOutputMap(outputs, "function_name")
	if name == "" {
		return pkgerrors.New("awslambda verify-exists: no function_name in outputs")
	}
	if err := v.VerifyExists(ctx, cfg, name, region); err != nil {
		return err
	}
	client := lambdaClient(cfg, region)

	// The published version output is non-empty exactly when the spec asked
	// for publishing; the qualified GetFunction proves the version object
	// itself exists, not merely that the create call returned a number.
	if version := stringOutputMap(outputs, "version"); version != "" {
		if _, err := client.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: aws.String(name),
			Qualifier:    aws.String(version),
		}); err != nil {
			return pkgerrors.Wrapf(err, "awslambda %q: published version %q not retrievable", name, version)
		}
	}

	aliasArns, err := lambdaAliasArns(outputs)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslambda %q verify-exists", name)
	}
	for alias, wantArn := range aliasArns {
		got, err := client.GetAlias(ctx, &lambda.GetAliasInput{
			FunctionName: aws.String(name),
			Name:         aws.String(alias),
		})
		if err != nil {
			return pkgerrors.Wrapf(err, "awslambda %q: alias %q not retrievable", name, alias)
		}
		if wantArn != "" && aws.ToString(got.AliasArn) != wantArn {
			return pkgerrors.Errorf("awslambda %q: alias %q ARN is %q, outputs claim %q",
				name, alias, aws.ToString(got.AliasArn), wantArn)
		}
	}

	// A function URL may be qualifier-scoped, and the qualifier is not an
	// output -- listing the function's URL configs covers both the $LATEST
	// and alias-qualified shapes without knowing which one was declared.
	if wantURL := stringOutputMap(outputs, "function_url"); wantURL != "" {
		listed, err := client.ListFunctionUrlConfigs(ctx, &lambda.ListFunctionUrlConfigsInput{
			FunctionName: aws.String(name),
		})
		if err != nil {
			return pkgerrors.Wrapf(err, "awslambda %q: listing function URL configs failed", name)
		}
		found := false
		for _, urlConfig := range listed.FunctionUrlConfigs {
			if aws.ToString(urlConfig.FunctionUrl) == wantURL {
				found = true
				break
			}
		}
		if !found {
			return pkgerrors.Errorf("awslambda %q: function URL %q not among the function's URL configs", name, wantURL)
		}
	}
	return nil
}

func (v *lambdaFunctionVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	name := stringOutputMap(outputs, "function_name")
	if name == "" {
		return pkgerrors.New("awslambda verify-absent: no function_name in outputs")
	}
	return v.VerifyAbsent(ctx, cfg, name, region)
}

// lambdaAliasArns reads the alias_arns output map (alias name -> alias ARN).
// A missing or empty map means the spec declared no aliases.
func lambdaAliasArns(outputs map[string]interface{}) (map[string]string, error) {
	raw, ok := outputs["alias_arns"]
	if !ok || raw == nil {
		return nil, nil
	}
	arns := map[string]string{}
	switch m := raw.(type) {
	case map[string]interface{}:
		for alias, arn := range m {
			if s, isStr := arn.(string); isStr {
				arns[alias] = s
			} else {
				arns[alias] = ""
			}
		}
	case map[string]string:
		for alias, arn := range m {
			arns[alias] = arn
		}
	default:
		return nil, pkgerrors.Errorf("alias_arns has unexpected type %T", raw)
	}
	return arns, nil
}

func lambdaClient(cfg aws.Config, region string) *lambda.Client {
	return lambda.NewFromConfig(cfg, func(o *lambda.Options) {
		if region != "" {
			o.Region = region
		}
	})
}

func lambdaFunctionExists(ctx context.Context, cfg aws.Config, id, region string) (bool, error) {
	out, err := lambdaClient(cfg, region).GetFunction(ctx, &lambda.GetFunctionInput{FunctionName: &id})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	if out.Configuration == nil {
		return false, nil
	}
	return true, nil
}
