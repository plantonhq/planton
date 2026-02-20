package module

import (
	"github.com/pkg/errors"
	ocikmskeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocikmskey/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var protectionModeMap = map[ocikmskeyv1.OciKmsKeySpec_ProtectionMode]string{
	ocikmskeyv1.OciKmsKeySpec_hsm:      "HSM",
	ocikmskeyv1.OciKmsKeySpec_software: "SOFTWARE",
	ocikmskeyv1.OciKmsKeySpec_external: "EXTERNAL",
}

var algorithmMap = map[ocikmskeyv1.OciKmsKeySpec_KeyShape_Algorithm]string{
	ocikmskeyv1.OciKmsKeySpec_KeyShape_aes:   "AES",
	ocikmskeyv1.OciKmsKeySpec_KeyShape_rsa:   "RSA",
	ocikmskeyv1.OciKmsKeySpec_KeyShape_ecdsa: "ECDSA",
}

var curveIdMap = map[ocikmskeyv1.OciKmsKeySpec_KeyShape_CurveId]string{
	ocikmskeyv1.OciKmsKeySpec_KeyShape_nist_p256: "NIST_P256",
	ocikmskeyv1.OciKmsKeySpec_KeyShape_nist_p384: "NIST_P384",
	ocikmskeyv1.OciKmsKeySpec_KeyShape_nist_p521: "NIST_P521",
}

func Resources(ctx *pulumi.Context, stackInput *ocikmskeyv1.OciKmsKeyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	if err := key(ctx, locals, ociProvider); err != nil {
		return errors.Wrap(err, "failed to create kms key")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}

