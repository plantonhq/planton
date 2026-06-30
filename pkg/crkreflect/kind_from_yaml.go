package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/apis/gvk"
	log "github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

func ExtractKindFromYaml(yamlManifestBytes []byte) (cloudresourcekind.CloudResourceKind, error) {
	gvk := new(gvk.GVK)
	if err := goyaml.Unmarshal(yamlManifestBytes, gvk); err != nil {
		return 0, errors.Wrap(err, "failed to yaml unmarshal into gvk object")
	}
	log.Debugf("detected apiVersion: %s and kind: %s", gvk.ApiVersion, gvk.Kind)
	cloudResourceKind, err := KindByKindName(gvk.Kind)
	if err != nil {
		return cloudresourcekind.CloudResourceKind_unspecified,
			errors.Wrapf(err, "failed to detect cloud-resource-kind by kind %s", gvk.Kind)
	}
	return cloudResourceKind, nil
}
