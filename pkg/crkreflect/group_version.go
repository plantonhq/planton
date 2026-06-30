package crkreflect

import (
	"fmt"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	log "github.com/sirupsen/logrus"
)

func GroupVersion(kind cloudresourcekind.CloudResourceKind) string {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		log.Errorf("failed to get kindMeta from Kind: %v", err)
		return ""
	}
	providerMeta, err := ProviderMeta(kind)
	if err != nil {
		log.Errorf("failed to extract group meta by kind %s with error %s", kind.String(), err.Error())
		return ""
	}
	return fmt.Sprintf("%s/%s", providerMeta.Group, kindMeta.Version.String())
}
