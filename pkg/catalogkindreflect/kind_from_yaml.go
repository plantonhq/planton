package catalogkindreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/apis/gvk"
	"github.com/plantonhq/planton/shared/catalogkind"
	log "github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

func ExtractKindFromYaml(yamlManifestBytes []byte) (catalogkind.CatalogKind, error) {
	gvk := new(gvk.GVK)
	if err := goyaml.Unmarshal(yamlManifestBytes, gvk); err != nil {
		return 0, errors.Wrap(err, "failed to yaml unmarshal into gvk object")
	}
	log.Debugf("detected apiVersion: %s and kind: %s", gvk.ApiVersion, gvk.Kind)
	catalogKind, err := KindByKindName(gvk.Kind)
	if err != nil {
		return catalogkind.CatalogKind_unspecified,
			errors.Wrapf(err, "failed to detect catalog-kind by kind %s", gvk.Kind)
	}
	return catalogKind, nil
}
